// Package lcd models all functionality related to the Game Boy's video output
package lcd

import (
	"image"
	"image/color"

	"../cpu"
	"../io"
	"../mmu"
	"../timer"
	"github.com/hajimehoshi/ebiten"
)

// GBLCD represents the state of the LCD
// mode: which mode the GB is in - VBlank, HBlank, OAM read, VRAM read
// modeClock: clock cycle counter, changes which mode we're in
// currentLine: what "scanline" is being drawn
// Window: image of current window
type GBLCD struct {
	mode        uint8
	modeClock   int16
	currentLine uint16
	Window      image.Image
}

// Pixel holds color & x, y point
type Pixel struct {
	Point image.Point
	Color color.RGBA
}

// Sprite holds Tile data as a Pixel array & x, y point
// x, y point is the upper-left corner of where to render the sprite
// Tile is an array of pixels that make up actual data to draw to screen
type Sprite struct {
	Point image.Point
	Tile  []*Pixel
}

// Constants for registers related to LCD
const (
	lcdc = 0xFF40
	stat = 0xFF41
	scy  = 0xFF42
	scx  = 0xFF43
	ly   = 0xFF44
	lyc  = 0xFF45

	// # of bytes in a tile
	tileBytes = 16
)

// InitLCD sets LCD initial values
// Only current one I'm aware of that we need nonzero is the mode
func (gblcd *GBLCD) InitLCD() {
	gblcd.mode = 2
}

// Injected variables from main.go
var (
	GbMMU   *mmu.GBMMU
	GbCPU   *cpu.GBCPU
	GbTimer *timer.GBTimer
	GbIO    *io.GBIO
)

// lcdEnabled returns 1 if LCD is enabled, 0 if not
// 7th bit of LCDC register indicates if LCD is on or off
func lcdEnabled() byte {
	if GbMMU.Memory[lcdc]&(1<<7) != 0 {
		return 1
	}

	return 0
}

// UpdateLCD updates the status of the LCD
// First, checks if LCD is enabled. If not, set modeClock, currentLine, and LY
// register value to 0. Also set LCD STAT register to default vlaue of 0x80.
// Setting these values ensures that they don't get incremented when LCD is off
// If it is enabled, then add to modeClock and set LCD status
func (gblcd *GBLCD) UpdateLCD(cycles int, screen *ebiten.Image) {
	if lcdEnabled() == 0 {
		gblcd.modeClock = 0
		gblcd.currentLine = 0
		GbMMU.Memory[ly] = 0

		// Clear LCD status
		GbMMU.Memory[stat] = 0x80
	} else {
		gblcd.modeClock += int16(cycles)
		gblcd.setLCDStatus(screen)
	}
}

func (gblcd *GBLCD) DebugDrawBG() {
	// If the third bit of LCDC is 1, this indicated we should use the second
	// background map, located elsewhere in memory
	useAltbgmap := GbMMU.Memory[lcdc]&(1<<3) != 0

	// By default, use the background map located at 0x9800 - 0x9C00
	bgmap := GbMMU.Memory[0x9800:0x9C00]

	if useAltbgmap {
		// Change background map location if bit above was set
		bgmap = GbMMU.Memory[0x9C00:0x9FFF]
	}

	populated := gblcd.populateBackgroundTiles(bgmap)
	bgImage := gblcd.generateBackgroundImage(populated)
	gblcd.placeWindow(bgImage)
}

func (gblcd *GBLCD) populateBackgroundTiles(bgmap []byte) [][]*Pixel {
	var filledBackground [][]*Pixel

	for _, tileID := range bgmap {
		tile := gblcd.renderTile(int(tileID), false)
		filledBackground = append(filledBackground, tile)
	}

	return filledBackground
}

func (gblcd *GBLCD) generateBackgroundImage(bgTiles [][]*Pixel) *image.RGBA {
	background := image.NewRGBA(image.Rect(0, 0, 256, 256))

	for i, tile := range bgTiles {
		for _, px := range tile {
			tileX := ((i % 32) * 8)
			tileY := ((i / 32) * 8)
			background.Set(px.Point.X+tileX, px.Point.Y+tileY, px.Color)
		}
	}

	return background
}

func (gblcd *GBLCD) placeWindow(bgImage *image.RGBA) {
	// SCX and SCY specify the upper-left location on the 256x256 background map
	// which is displayed on the upper-left corner of the LCD
	// Basically, it is the window's offset into the background map
	var yVal byte = GbMMU.Memory[scy]
	var xVal byte = GbMMU.Memory[scx]

	bounds := image.Rect(int(xVal), int(yVal), int(xVal)+160, int(yVal)+144)
	window := bgImage.SubImage(bounds)

	// Update Window value with new frame data
	gblcd.Window = window
}

// RenderWindow sets the window field on GBLCD to an image of the current frame
// This approach avoids needing to draw the entire background and then a window
// on top of it. Instead, we are rendering only what should be displayed in
// the window every frame.
func (gblcd *GBLCD) RenderWindow() {
	// If the third bit of LCDC is 1, this indicated we should use the second
	// background map, located elsewhere in memory
	useAltbgmap := GbMMU.Memory[lcdc]&(1<<3) != 0

	// By default, use the background map located at 0x9800 - 0x9C00
	bgmap := GbMMU.Memory[0x9800:0x9C00]

	if useAltbgmap {
		// Change background map location if bit above was set
		bgmap = GbMMU.Memory[0x9C00:0x9FFF]
	}

	// Create a blank 160x144 image representing the LCD
	window := image.NewRGBA(image.Rect(0, 0, 160, 144))

	// Create an empty 2D slice of Pixels
	// Each element is a 64-len slice, representing one tile's pixels (8x8)
	// Total is 18*20 tiles, 360 total for the window
	var tiles [][]*Pixel

	// SCX and SCY specify the upper-left location on the 256x256 background map
	// which is displayed on the upper-left corner of the LCD
	// Basically, it is the window's offset into the background map
	var yVal byte = GbMMU.Memory[scy]
	var xVal byte = GbMMU.Memory[scx]

	// We want to keep the initial value of SCX so we can reset our X offset
	// when we start drawing tiles on a new row
	var initialX byte = GbMMU.Memory[scx]

	// Here we calculate an offset into the background map using the initial
	// SCY and SCX values
	// Since the background map is a 1D array, and the window is a 1D array,
	// we can calculate what index to begin rendering from the background map
	// using these calculations
	yOff := int(yVal) * 256
	xOff := int(xVal) * 8
	offset := (yOff + xOff) / 64

	// Get tiles on background map beginning at calculated offset
	for height := 0; height < 18; height++ {
		for width := 0; width < 20; width++ {
			// Pass tile ID to renderTile
			// 64-len slice of pixels is returned, representing one tile
			tile := gblcd.renderTile(int(bgmap[offset]), false)
			tiles = append(tiles, tile)

			// Move to the next tile
			xVal += 8
			xOff = int(xVal) * 8

			// Calculate new offset
			offset = (yOff + xOff) / 64
		}

		// Move to next row
		yVal += 8

		// if yVal > 248 {
		// 	yVal = 0
		// }

		yOff = int(yVal) * 256

		// Set X to the X value of the top left corner
		xVal = initialX
		xOff = int(xVal) * 8
		offset = (yOff + xOff) / 64
	}

	// renderSprites returns a slice of Sprites, which we can render directly
	sprites := gblcd.renderSprites()

	// Render each tile
	for i, tile := range tiles {
		for _, px := range tile {
			tileX := ((i % 20) * 8)
			tileY := ((i / 20) * 8)
			window.Set(px.Point.X+tileX, px.Point.Y+tileY, px.Color)
		}
	}

	// Render each sprite
	for _, sprite := range sprites {
		for _, px := range sprite.Tile {
			window.Set(px.Point.X+sprite.Point.X, px.Point.Y+sprite.Point.Y, px.Color)
		}
	}

	// Update Window value with new frame data
	gblcd.Window = window
}

// renderSprites goes through the OAM and populates a slice of Sprites
// OAM always contains information about the Sprites currently on screen
// TODO Handle additional sprite properties like rotation
func (gblcd *GBLCD) renderSprites() []*Sprite {
	var sprites []*Sprite
	oam := GbMMU.Memory[0xFE00:0xFEA0]

	for i := 0; i < len(oam); i += 4 {
		// Get next sprite data
		spriteData := GbMMU.Memory[0xFE00+i : 0xFE00+i+4]

		yLoc := spriteData[0] - 16
		xLoc := spriteData[1] - 8
		tile := gblcd.renderTile(int(spriteData[2]), true)

		s := &Sprite{
			Point: image.Point{int(xLoc), int(yLoc)},
			Tile:  tile,
		}

		sprites = append(sprites, s)
	}

	return sprites
}

// renderTile returns a single tile as an array of Pixels
// Takes a tile ID and a bool indicator of if this is a sprite or not
// The tile ID is provided by the background map, and is used to calculate
// the location of the tile data in memory
func (gblcd *GBLCD) renderTile(tileID int, sprites bool) []*Pixel {
	pixels := []*Pixel{}

	// A pixel can be one of four "colors"
	// In actual GB it is really black, white, or one of two grays
	// The LCD itself causes the pale green we know and love
	// These can be set to anything, but I set them to greenish to mimic the LCD
	palette := [4]color.RGBA{
		color.RGBA{205, 255, 205, 255},
		color.RGBA{120, 170, 120, 255},
		color.RGBA{35, 85, 35, 255},
		color.RGBA{0, 0, 0, 255},
	}

	// The 4th bit of the LCDC register determines where this tile is located
	// in memory - one of two possible ranges
	loTiles := GbMMU.Memory[lcdc]&(1<<4) != 0

	// If the 4th bit was set, OR if we're rendering a sprite, our tile ID
	// is unsigned (0 - 255) and we can find its data by
	// adding the ID * 16 to the address 0x800
	if loTiles || sprites {
		tileID = 0x8000 + (tileID * 16)
	} else {
		// If we're in hi tiles set, tile locations are signed (-128 - +127)
		if tileID > 127 {
			// Calculate tile data location for a positive ID
			tileID = tileID - 128
			tileID = 0x8800 + (tileID * 16)
		} else {
			// Calculate the data location for a negative ID
			tileID = 0x8800 + ((tileID + 128) * 16)
		}
	}

	// A single 8x8 pixel tile is actually represented by 16 bytes
	// A tile ID is really its beginning location in memory
	// Its data is the location + next 16 bytes
	tileVals := GbMMU.Memory[tileID : tileID+16]

	// Here's where the calculation of a tile's pixel information happens
	// It is somewhat convoluted, but https://fms.komkon.org/GameBoy/Tech/Software.html
	// contains a good explanation in the "Video" section

	// Iterate over lines of tiles, represented by 2 bytes
	for line := 0; line < 8; line++ {
		// Get 2 bytes representing current line
		hi := tileVals[line*2]
		lo := tileVals[line*2+1]

		// Iterate over individual pixels of tile lines
		for pix := 0; pix < 8; pix++ {
			// TODO Maybe make color lookup more like hardware
			// http://www.codeslinger.co.uk/pages/projects/gameboy/graphics.html

			// Determine values of lo and hi bits for this pixel on the line
			hiBit := (lo >> (7 - uint8(pix))) & 1
			loBit := (hi >> (7 - uint8(pix))) & 1

			// The result of this calculation is the index of a color in
			// our palette
			colorIndex := loBit + hiBit*2
			color := palette[colorIndex]

			// If we're rendering a sprite and the calculated colorIndex is 0,
			// then this pixel of the sprite is simply not rendered
			// This is how we get a "transluscent" effect on sprites
			if sprites && colorIndex == 0 {
				continue
			}

			// X location of pixel is the value of our current iterator
			pixX := pix

			// Y location of pixel is value of outer loop iterator
			pixY := line

			// Create new pixel
			p := &Pixel{
				Point: image.Point{pixX, pixY},
				Color: color,
			}

			pixels = append(pixels, p)
		}
	}

	return pixels
}

// setLCDStatus changes the status of the LCD
// LCD can be in one of four modes
// Mode switching happens only when a certain condition is met
// The GB hardware is actually meant to simulate a CRT in terms of timings
// This is why we have HBlank/VBlank modes
func (gblcd *GBLCD) setLCDStatus(screen *ebiten.Image) {
	switch gblcd.mode {
	// Horizontal blanking mode
	// GB is in this mode when horizontal lines are being drawn
	case 0:
		// If clock cycles for this frame >= 204, increment the current line
		// and reset modeClock
		if gblcd.modeClock >= 204 {
			gblcd.modeClock = 0
			gblcd.currentLine++
			GbMMU.Memory[ly]++

			// If we've rendered the last line on the LCD, enter VBlank mode
			// and send a VBlank interrupt request
			if gblcd.currentLine == 143 {
				gblcd.mode = 1
				GbMMU.Memory[stat] |= (1 << 0)
				GbMMU.Memory[stat] &^= (1 << 1)

				// Request VBlank interrupt
				GbMMU.Memory[0xFF0F] |= (1 << 0)
			} else {
				// Enter OAM read mode
				GbMMU.Memory[stat] &^= (1 << 0)
				GbMMU.Memory[stat] |= (1 << 1)
				gblcd.mode = 2
			}
		}
	// Vertical blanking mode
	// GB is in this mode when all horizontal lines have been drawn and
	// we reset to the top left of the screen
	// This happens at the end of a frame and is longer than HBlank time
	case 1:
		if gblcd.modeClock >= 456 {
			gblcd.modeClock = 0
			gblcd.currentLine++
			GbMMU.Memory[ly]++

			if gblcd.currentLine > 153 {
				GbMMU.Memory[stat] &^= (1 << 0)
				GbMMU.Memory[stat] |= (1 << 1)
				gblcd.mode = 2
				GbMMU.Memory[ly] = 0
				gblcd.currentLine = 0
			}
		}
	// OAM read mode
	// This mode indicates the GB is accessing data stored in OAM
	case 2:
		if gblcd.modeClock >= 80 {
			gblcd.modeClock = 0
			GbMMU.Memory[stat] |= (1 << 0)
			GbMMU.Memory[stat] |= (1 << 1)
			gblcd.mode = 3
		}
	// VRAM read mode
	// This mode indicates the GB is accessing data stored in VRAM
	case 3:
		if gblcd.modeClock >= 172 {
			gblcd.modeClock = 0
			GbMMU.Memory[stat] &^= (1 << 0)
			GbMMU.Memory[stat] &^= (1 << 1)
			gblcd.mode = 0
		}
	}

	// Check if LY and LYC registers are the same value
	// If they are, we set the coincidence bit in the LCD STAT register
	// and if LCD STAT interrupts are enabled, we send one
	// Otherwise clear the coincidence bit in the LCD STAT register
	if GbMMU.Memory[ly] == GbMMU.Memory[lyc] {
		// Set coincidence bit
		GbMMU.Memory[stat] |= (1 << 2)

		// LCD STAT interrupt
		if GbMMU.Memory[0xFFFE]&(1<<1) != 0 {
			GbMMU.Memory[0xFF0F] |= (1 << 1)
		}
	} else {
		// Clear coincidence bit
		GbMMU.Memory[stat] &^= (1 << 2)
	}
}
