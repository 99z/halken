package lcd

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"../cpu"
	"../mmu"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const maxCycles = 69905

var (
	GbMMU *mmu.GBMMU
	GbCPU *cpu.GBCPU

	frames = 0
	second = time.Tick(time.Second)
)

func Run(screen *ebiten.Image) error {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%v", ebiten.CurrentFPS()))
	// Logical update
	Update()

	// Graphics update

	return nil
}

func Update() {
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
		delay := GbCPU.Instrs[operation].Executor()
		fmt.Printf("LCD STAT: %02X\n", GbMMU.Memory[0xFF41])
		fmt.Printf("LY: %02X\n", GbMMU.Memory[0xFF44])

		if opcode[0] == 68 && opcode[1] == 203 {
			fmt.Println(GbMMU.Memory[40960:40963])
			os.Exit(0)
		}

		// Update cycles
		updateCycles += int(GbCPU.Instrs[operation].TCycles) + delay

		// Update graphics
		updateGraphics(int(GbCPU.Instrs[operation].TCycles) + delay)

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

func updateGraphics(cycles int) {
	setLCDStatus()

	if lcdEnabled() != 0 {
		GbMMU.ScanlineCount -= int16(cycles)
	} else {
		return
	}

	if GbMMU.ScanlineCount <= 0 {
		GbMMU.Memory[0xFF44]++
		GbMMU.ScanlineCount += 456

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
		var mode2Bounds int16 = 376
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
