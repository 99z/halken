package main

// Main creates all Game Boy components and handles main loop

import (
	"encoding/binary"
	"fmt"
	"os"

	"./cpu"
	"./io"
	"./lcd"
	"./mmu"
	"./timer"
	"github.com/hajimehoshi/ebiten"
)

// GB can execute 4194304 (4.19MHz) cycles per second
// Since updates happen 60 times per second, we divide by 60 to get
// maximum number of cycles allowed to be executed per frame
const maxCycles = 69905

// Create Game Boy components
// The US patent provides good visual breakdown of components in figure 4
// Patent reference: https://patents.google.com/patent/US5184830A/en

// GbCPU represents GB's CPU - patent fig. 4, #24
var GbCPU = new(cpu.GBCPU)

// GbTimer represents GB's timer - patent fig. 4, #24d
var GbTimer = new(timer.GBTimer)

// GbMMU represents all available GB memory - patent fig. 4, #s 16, 28, 30, 36, 42
var GbMMU = new(mmu.GBMMU)

// GbLCD represents GB's LCD components and physical screen
// patent fig. 4, #s 14, 38, 40, 44, 48
var GbLCD = new(lcd.GBLCD)

// GbIO represents GB's "controller key matrix" and I/O port
// patent fig. 4, #s 18, 27
var GbIO = new(io.GBIO)

func main() {
	cartPath := os.Args[1]

	// Inject components into packages that need to use them
	cpu.GbMMU = GbMMU
	lcd.GbMMU = GbMMU
	timer.GbMMU = GbMMU

	mmu.GbIO = GbIO
	lcd.GbIO = GbIO

	lcd.GbCPU = GbCPU

	lcd.GbTimer = GbTimer

	// Call initialization functions for components
	// Necessary to set default values
	GbMMU.InitMMU()
	GbCPU.InitCPU()
	GbIO.InitIO()
	GbLCD.InitLCD()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	// Kick off main emulation loop & create graphics context
	ebiten.Run(run, 160, 144, 4, "Halken")
}

// run is the primary emulation loop, called 60 times per second by ebiten
func run(screen *ebiten.Image) error {
	// Read inputs prior to updating state
	GbIO.ReadInput()

	// Execute next instruction and update graphics state
	update(screen)

	// Update window, which is just an image
	GbLCD.RenderWindow()

	// Draw the window to the graphics context
	ebitenBG, _ := ebiten.NewImageFromImage(GbLCD.Window, ebiten.FilterDefault)
	opts := &ebiten.DrawImageOptions{}
	screen.DrawImage(ebitenBG, opts)

	return nil
}

// update:
// 1. Executes next operation
// 2. Updates total cycles
// 3. Updates timers
// 4. Updates window's state
// 5. Performs interrupts
// TODO This might be too much of a god function, maybe break down
func update(screen *ebiten.Image) {
	// Counter for total number of cycles executed for this frame
	updateCycles := 0

	for updateCycles < maxCycles {
		// First, we need to check if the CPU is halted
		// This is indicated by a boolean which gets set if HALT is called
		if !GbCPU.Halted {
			if GbCPU.EIReceived {
				// If we aren't halted, but EI (enable interrupts) was called
				// last cycle, we set the Interrupt Master Enable to 1
				// IME globally enables/disables interrupts
				GbCPU.IME = 1
			}

			// Set Jumped to false in case the last instruction was a jump
			GbCPU.Jumped = false

			// Get value of PC as an integer, find next opcode to execute
			// by getting byte at that index in memory
			opcode := GbCPU.Regs.PC[:]
			opcodeInt := binary.LittleEndian.Uint16(opcode)
			operation := GbMMU.Memory[opcodeInt]

			// Execute the next instruction
			// Delay is set to the return value of the instruction's executor
			// This is important because certain instructions take a different
			// number of cycles depending on if they "completed" or not
			// For example, RET Z takes 8 cycles by default, but takes 20 cycles
			// if the zero flag was set and the RET happened
			// If an instruction does not have conditional # of cycles, the
			// executor just returns 0
			delay := GbCPU.Instrs[operation].Executor()

			// Update total cycles executed for this frame
			// If the instruction had conditional number of cycles, delay will
			// be nonzero and added to the base # of cycles for the instruction
			updateCycles += int(GbCPU.Instrs[operation].TCycles) + delay

			// Update graphics
			// Note we do NOT pass updateCycles here since that represents the
			// total number of cycles in this frame
			// We only want to pass the number of cycles taken by the previous
			// instruction
			GbLCD.UpdateLCD(int(GbCPU.Instrs[operation].TCycles)+delay, screen)

			// Increment the timer
			// See timer/timer.go for details
			GbTimer.Increment(updateCycles)

			// If the last instruction changed the value of the program counter
			// then a jump occurred
			// In this case we don't want to increment the PC
			// Otherwise, increment the PC by the number of operands (bytes) of
			// the last instruction
			if GbCPU.Jumped {
				continue
			} else {
				nextInstr := binary.LittleEndian.Uint16(GbCPU.Regs.PC) + GbCPU.Instrs[operation].NumOperands
				// TODO Maybe don't need to do this anymore?
				nextInstrAdddr := make([]byte, 2)
				binary.LittleEndian.PutUint16(nextInstrAdddr, nextInstr)
				GbCPU.Regs.PC = nextInstrAdddr
			}

			// If Interrupt Master Flag is set, and Interrupt Enable register is
			// nonzero, and Interrupt Flag register is nonzero,
			// then we execute an interrupt
			if GbCPU.IME != 0 && GbMMU.Memory[0xFFFE] != 0 && GbMMU.Memory[0xFF0F] != 0 {
				// Get the bit of the interrupt to execute
				interrupt := GbMMU.Memory[0xFFFE] & GbMMU.Memory[0xFF0F]

				if interrupt&1 != 0 {
					// Run VBlank interrupt handler
					GbCPU.RSTI(0x40)

					// Clear VBlank interrupt request bit
					GbMMU.Memory[0xFF0F] &^= (1 << 0)
					updateCycles += 16
				} else if interrupt&4 != 0 {
					// Run timer interrupt handler
					GbCPU.RSTI(0x50)

					// Clear timer interrupt request bit
					GbMMU.Memory[0xFF0F] &^= (1 << 2)
					updateCycles += 16
				}
			}

			GbTimer.Increment(updateCycles)
		} else {
			// CPU is halted
			instrTotal := 0

			// Get the current value of the Interrupt Flag register
			currentIF := GbMMU.ReadData(0xFF0F)

			// If IF's current value is different than the value it was prior
			// to HALT being called, unhalt
			if currentIF != GbCPU.IFPreHalt {
				GbCPU.Halted = false
			}

			// Check for interrupts, as above
			if GbCPU.IME != 0 && GbMMU.Memory[0xFFFE] != 0 && GbMMU.Memory[0xFF0F] != 0 {
				interrupt := GbMMU.Memory[0xFFFE] & GbMMU.Memory[0xFF0F]

				if interrupt&1 != 0 {
					// Run interrupt handler
					GbCPU.RSTI(0x40)

					// Clear VBlank interrupt request bit
					GbMMU.Memory[0xFF0F] &^= (1 << 0)
					updateCycles += 16
					instrTotal += 16
				} else if interrupt&4 != 0 {
					GbCPU.RSTI(0x50)

					// Clear timer interrupt request bit
					GbMMU.Memory[0xFF0F] &^= (1 << 2)
					instrTotal += 16
					updateCycles += 16
				}
			}

			// Halted CPU still takes 1 cycle by default
			updateCycles++
			GbTimer.Increment(updateCycles)
			GbLCD.UpdateLCD(1+instrTotal, screen)
		}
	}
}
