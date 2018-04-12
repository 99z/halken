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
	window      *image.RGBA
	currentBG   *image.RGBA
}

type Pixel struct {
	Point image.Point
	Color color.RGBA
}

const (
	LCDC = 0xFF40
	STAT = 0xFF41
	SCY  = 0xFF42
	SCX  = 0xFF43
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
const tileBytes = 16

var (
	GbMMU *mmu.GBMMU
	GbCPU *cpu.GBCPU

	frames = 0
	second = time.Tick(time.Second)
)

// Set offsets to uint8s, just add and let it overflow
// Linear offset in bg map of first tile in window
// ((y offset * num pixels per row) + (x offset * 8)) / 8
func (gblcd *GBLCD) Run(screen *ebiten.Image) error {
	// Logical update
	gblcd.Update(screen)
	// tiles := loadTilesDebug(0x8000, 0x9000)
	// ebitenTiles, _ := ebiten.NewImageFromImage(tiles, ebiten.FilterDefault)
	// opts := &ebiten.DrawImageOptions{}
	// opts.GeoM.Translate(160, 0)
	// screen.DrawImage(ebitenTiles, opts)

	// testTile := renderTile(0x8390)
	// ebiTest, _ := ebiten.NewImageFromImage(testTile, ebiten.FilterDefault)
	// testOpts := &ebiten.DrawImageOptions{}
	// screen.DrawImage(ebiTest, testOpts)

	// scy := int(GbMMU.Memory[SCY])
	// scx := int(GbMMU.Memory[SCX])
	// gblcd.renderBackground(GbMMU.Memory[0x9800:0x9C00])

	gblcd.renderWindow(GbMMU.Memory[0x9800:0x9C00])

	ebitenBG, _ := ebiten.NewImageFromImage(gblcd.window, ebiten.FilterDefault)
	opts := &ebiten.DrawImageOptions{}
	screen.DrawImage(ebitenBG, opts)

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

		// fmt.Printf("%02X:%02X\t%02X\t%v\n", opcode[1], opcode[0], operation, GbCPU.Instrs[operation])
		// fmt.Printf("C67A: %02X\n", GbMMU.Memory[0xC67A])
		delay := GbCPU.Instrs[operation].Executor()

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

func (gblcd *GBLCD) renderWindow(bgmap []byte) {
	// ((y tile offset * num pixels per row) + (x tile offset * 8)) / 8
	window := image.NewRGBA(image.Rect(0, 0, 160, 144))
	var tiles [][]*Pixel

	var yVal byte = GbMMU.Memory[SCY]
	var xVal byte = GbMMU.Memory[SCX]
	var initialX byte = GbMMU.Memory[SCX]
	yOff := int(yVal) * 256
	xOff := int(xVal) * 8
	offset := (yOff + xOff) / 64

	for height := 0; height < 18; height++ {
		for width := 0; width < 20; width++ {
			tile := renderTile(int(bgmap[offset]) * 16)
			tiles = append(tiles, tile)

			// Move to the next tile
			xVal += 8
			xOff = int(xVal) * 8

			offset = (yOff + xOff) / 64
		}

		yVal += 8
		yOff = int(yVal) * 256
		xVal = initialX
		xOff = int(xVal) * 8
		offset = (yOff + xOff) / 64
	}

	for i, tile := range tiles {
		for _, px := range tile {
			tileX := ((i % 20) * 8)
			tileY := ((i / 20) * 8)
			window.Set(px.Point.X+tileX, px.Point.Y+tileY, px.Color)
		}
	}

	gblcd.window = window
}

// BG map is 32*32 bytes, each references a tile
func (gblcd *GBLCD) renderBackground(bgmap []byte) {
	bg := image.NewRGBA(image.Rect(0, 0, 256, 256))
	var tiles [][]*Pixel

	// Iterate over 1024 tile IDs
	for _, tileID := range bgmap {
		// Render tile referenced by value in bgmap
		tile := renderTile(int(tileID) * 16)
		tiles = append(tiles, tile)
	}

	for i, tile := range tiles {
		for _, px := range tile {
			tileX := ((i % 32) * 8)
			tileY := ((i / 32) * 8)
			bg.Set(px.Point.X+tileX, px.Point.Y+tileY, px.Color)
		}
	}

	gblcd.currentBG = bg
}

func renderTile(tileAddr int) []*Pixel {
	// tile := image.NewRGBA(image.Rect(0, 0, 8, 8))
	pixels := []*Pixel{}

	palette := [4]color.RGBA{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{170, 170, 170, 255},
		color.RGBA{85, 85, 85, 255},
		color.RGBA{0, 0, 0, 255},
	}

	tileAddr = tileAddr + 0x8000
	tileVals := GbMMU.Memory[tileAddr : tileAddr+16]
	// fmt.Printf("%04X\n", tileAddr)

	// Iterate over lines of tiles, represented by 2 bytes
	for line := 0; line < 8; line++ {
		hi := tileVals[line*2]
		lo := tileVals[line*2+1]

		// Iterate over individual pixels of tile lines
		for pix := 0; pix < 8; pix++ {
			hiBit := (hi >> (7 - uint8(pix))) & 1
			loBit := (lo >> (7 - uint8(pix))) & 1

			colorIndex := loBit + hiBit*2
			color := palette[colorIndex]
			pixX := pix
			pixY := line

			p := &Pixel{
				Point: image.Point{pixX, pixY},
				Color: color,
			}

			pixels = append(pixels, p)

			// tile.Set(pixX, pixY, color)
		}
	}

	// return tile
	return pixels
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
