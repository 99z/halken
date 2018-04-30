// Package io handles reading keyboard inputs and returning them
// when requested by the CPU
// Input detection is provided by ebiten and ReadInput is called every frame
package io

import (
	"github.com/hajimehoshi/ebiten"
)

// GBIO represents the controller key matrix
// Imran Nazar has a good description and diagram of how this works:
// http://imrannazar.com/GameBoy-Emulation-in-JavaScript:-Input
// When 0xFF00 is written to, one of two columns is selected as `col`
// One column has Down/Up/Left/Right d-pad buttons,
// the other has Start/Select/B/A.
// We can represent this using an array of 2 bytes
// The 2 elements represent the columns, and the value represents which
// buttons were pressed
type GBIO struct {
	buttons [2]byte
	col     byte
}

// InitIO initializes the GBIO struct
// Key values are set to 0x0F and column 0 is selected by default
func (gbio *GBIO) InitIO() {
	gbio.buttons[0], gbio.buttons[1] = 0x0F, 0x0F
	gbio.col = 0
}

// ReadInput is called from our main update loop
// Determines which buttons were pressed for this frame, sets bytes in
// buttons array accordingly
func (gbio *GBIO) ReadInput() {
	// Start button
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		gbio.buttons[0] &= 0x7
	} else {
		gbio.buttons[0] |= 0x8
	}

	// Select button
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		gbio.buttons[0] &= 0xB
	} else {
		gbio.buttons[0] |= 0x5
	}

	// B button
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		gbio.buttons[0] &= 0xD
	} else {
		gbio.buttons[0] |= 0x2
	}

	// A button
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		gbio.buttons[0] &= 0xE
	} else {
		gbio.buttons[0] |= 0x1
	}

	// D-pad up
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		gbio.buttons[1] &= 0xB
	} else {
		gbio.buttons[1] |= 0x4
	}

	// D-pad down
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		gbio.buttons[1] &= 0x7
	} else {
		gbio.buttons[1] |= 0x8
	}

	// D-pad left
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		gbio.buttons[1] &= 0xD
	} else {
		gbio.buttons[1] |= 0x2
	}

	// D-pad right
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		gbio.buttons[1] &= 0xE
	} else {
		gbio.buttons[1] |= 0x1
	}
}

// SetCol sets the column for inputs we should return to the CPU
// This is called when a write to 0xFF00 happens, handled by the MMU
// Will either be set to 0 or 1
func (gbio *GBIO) SetCol(data byte) {
	gbio.col = data & 0x30
}

// GetInput returns a byte representing which buttons were pressed for
// this frame
// The button returned is dependent on which column is selected
func (gbio *GBIO) GetInput() byte {
	switch gbio.col {
	case 0x10:
		return gbio.buttons[0]
	case 0x20:
		return gbio.buttons[1]
	default:
		return 0x00
	}
}
