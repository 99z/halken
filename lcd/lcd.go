package lcd

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"os"
	"time"

	"../cpu"
	"../mmu"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type GBLCD struct {
	// 160*144 screen size, where each xy can be an RGBA value
	screen      [23040]color.RGBA
	mode        uint8
	tileset     [2]byte
	modeClock   int16
	currentLine uint16
}

const (
	LCDC = 0xFF40
	STAT = 0xFF41
	LY   = 0xFF44
)

func (gblcd *GBLCD) reset() {
	// Initialize screen to white
	for i := range gblcd.screen {
		gblcd.screen[i] = color.RGBA{255, 255, 255, 1}
	}

	gblcd.modeClock = 456
}

const maxCycles = 69905

var (
	GbMMU *mmu.GBMMU
	GbCPU *cpu.GBCPU

	frames = 0
	second = time.Tick(time.Second)
)

func (gblcd *GBLCD) Run(screen *ebiten.Image) error {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%v", ebiten.CurrentFPS()))
	// Logical update
	gblcd.Update(screen)

	// Graphics update

	return nil
}

func (gblcd *GBLCD) Update(screen *ebiten.Image) {
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

		//fmt.Printf("%02X:%02X\t%02X\t%v\n", opcode[1], opcode[0], operation, GbCPU.Instrs[operation])
		delay := GbCPU.Instrs[operation].Executor()
		//fmt.Printf("LCD STAT: %02X\n", GbMMU.Memory[0xFF41])
		//fmt.Printf("LY: %02X\n", GbMMU.Memory[0xFF44])

		if opcode[0] == 68 && opcode[1] == 203 {
			os.Exit(0)
		}

		// Update cycles
		updateCycles += int(GbCPU.Instrs[operation].TCycles) + delay

		// Update graphics
		gblcd.updateGraphics(int(GbCPU.Instrs[operation].TCycles)+delay, screen)

		if GbCPU.Jumped {
			continue
		} else {
			nextInstr := binary.LittleEndian.Uint16(GbCPU.Regs.PC) + GbCPU.Instrs[operation].NumOperands
			// Interesting problem if we don't make a new byte array here
			// TODO Explain exactly what it is... when I understand it
			nextInstrAdddr := make([]byte, 2)
			binary.LittleEndian.PutUint16(nextInstrAdddr, nextInstr)
			GbCPU.Regs.PC = nextInstrAdddr
		}
	}
}

func (gblcd *GBLCD) updateGraphics(cycles int, screen *ebiten.Image) {
	gblcd.modeClock += int16(cycles)
	gblcd.setLCDStatus(screen)
}

func lcdEnabled() byte {
	return GbMMU.Memory[STAT] & (1 << 7)
}

func (gblcd *GBLCD) setLCDStatus(screen *ebiten.Image) {
	switch gblcd.mode {
	// OAM read mode
	case 2:
		if gblcd.modeClock >= 80 {
			gblcd.modeClock = 0
			gblcd.mode = 3
		}
	// VRAM read mode
	case 3:
		if gblcd.modeClock >= 172 {
			gblcd.modeClock = 0
			gblcd.mode = 0

			// TODO Write scanline to framebuffer
		}
	// HBlank
	case 0:
		if gblcd.modeClock >= 204 {
			gblcd.modeClock = 0
			gblcd.currentLine++
			GbMMU.Memory[LY]++

			if gblcd.currentLine == 143 {
				gblcd.mode = 1
				// TODO draw image to screen
				// screen.Fill(color.RGBA{255, 255, 255, 255})
			} else {
				gblcd.mode = 2
			}
		}
	// VBlank
	case 1:
		if gblcd.modeClock >= 456 {
			gblcd.modeClock = 0
			gblcd.currentLine++
			GbMMU.Memory[LY]++

			if gblcd.currentLine > 153 {
				gblcd.mode = 2
				gblcd.currentLine = 0
				GbMMU.Memory[LY] = 0
				// screen.Fill(color.RGBA{0, 0, 0, 255})
			}
		}
	}
}

func (gblcd *GBLCD) drawScanline() {

}
