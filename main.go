package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"./cpu"
	"./mmu"
)

// Global variables for component structs
var GbCPU = new(cpu.GBCPU)
var GbMMU = new(mmu.GBMMU)

const maxCycles = 69905

func main() {
	cartPath := os.Args[1]

	cpu.GbMMU = GbMMU
	GbMMU.InitMMU()
	GbCPU.InitCPU()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	// fmt.Printf("Title: %s\nCGBFlag: %v\nType: %v\nROM: %v\nRAM: %v\n",
	// 	GbMMU.Cart.Title, GbMMU.Cart.CGBFlag, GbMMU.Cart.Type, GbMMU.Cart.ROMSize, GbMMU.Cart.RAMSize)

	// Update 60 times per second
	ticker := time.NewTicker(time.Second / 60)
	for range ticker.C {
		update()
	}
	time.Sleep(1)
}

func update() {
	// Main loop
	// 1. Execute next operation
	// 2. Update total cycles
	// 3. Update timers
	// 4. Update LCD
	// 5. Perform interrupts
	updateCycles := 0

	// 4194304 is max cycles that can be executed per second
	// Since running at 60 FPS, each cycle max must be 4194304/60 = 69905
	for updateCycles < maxCycles {
		GbCPU.Jumped = false
		opcode := GbCPU.Regs.PC[:]
		opcodeInt := binary.LittleEndian.Uint16(opcode)

		operation := GbMMU.Memory[opcodeInt]

		fmt.Printf("%02X:%02X\t%02X\t%v\n", opcode[1], opcode[0], operation, GbCPU.Instrs[operation])
		GbCPU.Instrs[operation].Executor()

		// Update cycles
		updateCycles += int(GbCPU.Instrs[operation].TCycles)

		// Update graphics
		updateGraphics(updateCycles)

		if GbCPU.Jumped {
			continue
		} else {
			nextInstr := binary.LittleEndian.Uint16(GbCPU.Regs.PC) + GbCPU.Instrs[operation].NumOperands
			binary.LittleEndian.PutUint16(GbCPU.Regs.PC, nextInstr)
		}
	}
}

func updateGraphics(cycles int) {
	setLCDStatus()

	if lcdEnabled() != 0 {
		GbMMU.ScanlineCount -= uint16(cycles)
	} else {
		return
	}

	if GbMMU.ScanlineCount <= 0 {
		GbMMU.Memory[0xFF44]++
		GbMMU.ScanlineCount = 456

		if GbMMU.Memory[0xFF44] > 153 {
			GbMMU.Memory[0xFF44] = 0
		}
	}
}

func lcdEnabled() byte {
	return GbMMU.Memory[0xFF40] & (1 << 7)
}

// Good info on setting LCD status
// http://www.codeslinger.co.uk/pages/projects/gameboy/lcd.html
func setLCDStatus() {
	// Get value of LCD status
	status := GbMMU.Memory[0xFF41]
	lcdStatus := lcdEnabled()

	if lcdStatus == 0 {
		GbMMU.ScanlineCount = 456
		GbMMU.Memory[0xFF44] = 0
		status &= 252
		status |= (1 << 0)
		GbMMU.Memory[0xFF41] = status
		return
	}

	currentLine := GbMMU.Memory[0xFF44]
	var currentMode byte = status & 0x3
	var mode byte
	var reqInt byte

	// If true, in VBlank, set mode to 1
	if currentLine >= 144 {
		mode = 1
		status |= (1 << 0)
		mask := ^(1 << 1)
		status &= byte(mask)
		reqInt = status & (1 << 4)
	} else {
		var mode2Bounds uint16 = 376
		mode3Bounds := mode2Bounds - 172

		// If true, in mode 2
		if GbMMU.ScanlineCount >= mode2Bounds {
			mode = 2
			status |= (1 << 1)
			mask := ^(1 << 0)
			status &= byte(mask)
			reqInt = status & (1 << 5)
		} else if GbMMU.ScanlineCount >= mode3Bounds {
			mode = 3
			status |= (1 << 1)
			status |= (1 << 0)
		} else {
			mode = 0
			mask := ^(1 << 1)
			status &= byte(mask)
			mask = ^(1 << 0)
			status &= byte(mask)
			reqInt = status & (1 << 4)
		}
	}

	if (reqInt != 0) && (mode != currentMode) {
		// TODO request interrupt
	}

	// Check the coincidence flag
	if GbMMU.Memory[0xFF44] == GbMMU.Memory[0xFF45] {
		status |= (1 << 2)

		if (status & (1 << 6)) != 0 {
			// TODO request interrupt
		}
	} else {
		mask := ^(1 << 2)
		status &= byte(mask)
	}

	GbMMU.Memory[0xFF41] = status
}
