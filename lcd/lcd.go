package lcd

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
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
	// Logical update
	gblcd.Update(screen)
	tiles := loadTilesDebug(0x8000, 0x9000)
	ebitenTiles, _ := ebiten.NewImageFromImage(tiles, ebiten.FilterDefault)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(160, 0)
	screen.DrawImage(ebitenTiles, opts)

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

		// if opcode[0] == 68 && opcode[1] == 203 {
		// 	os.Exit(0)
		// }

		ebitenutil.DebugPrint(screen, fmt.Sprintf("%02X:%02X", opcode[1], opcode[0]))

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

func loadTilesDebug(beg, end int) *image.RGBA {
	bg := image.NewRGBA(image.Rect(0, 0, 128, 128))
	// Tile set #1: 0x8000 - 0x87FF
	// Tile map #0: 0x9800 - 0x9BFF
	// Complete list: http://imrannazar.com/GameBoy-Emulation-in-JavaScript:-Graphics
	tiles := GbMMU.Memory[beg:end]
	const tileBytes = 16

	palette := [4]color.RGBA{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{170, 170, 170, 255},
		color.RGBA{85, 85, 85, 255},
		color.RGBA{0, 0, 0, 255},
	}

	// Iterate over 8x8 tiles
	numTiles := (end - beg) / 16
	for tile := 0; tile < numTiles; tile++ {
		tileX := (tile % 16) * 8
		tileY := (tile / 16) * 8
		// Iterate over lines of tiles, represented by 2 bytes
		for line := 0; line < 8; line++ {
			hi := tiles[(tile*tileBytes)+line*2]
			lo := tiles[(tile*tileBytes)+line*2+1]

			// Iterate over individual pixels of tile lines
			for pix := 0; pix < 8; pix++ {
				hiBit := (hi >> (7 - uint8(pix))) & 1
				loBit := (lo >> (7 - uint8(pix))) & 1

				colorIndex := loBit + hiBit*2
				color := palette[colorIndex]

				pixX := tileX + pix
				pixY := tileY + line

				bg.Set(pixX, pixY, color)
			}
		}
	}

	return bg
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
				// gblcd.drawDebugTiles(screen)
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

func (gblcd *GBLCD) drawDebugTiles(screen *ebiten.Image) {
	// square, _ := ebiten.NewImage(32, 32, ebiten.FilterNearest)
	// square.Fill(color.White)
	// opts := &ebiten.DrawImageOptions{}
	// opts.GeoM.Translate(200, float64(gblcd.modeClock))
	// screen.DrawImage(square, opts)

	// tiles := GbMMU.Memory[0x8300:0x8400]
}
