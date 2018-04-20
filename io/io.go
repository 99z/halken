package io

import (
	"github.com/hajimehoshi/ebiten"
)

type GBIO struct {
	rows [2]byte
	col  byte
}

const buttonCol byte = 0x10
const dpadCol byte = 0x20

func (gbio *GBIO) InitIO() {
	gbio.rows[0], gbio.rows[1] = 0x0F, 0x0F
	gbio.col = 0
}

func (gbio *GBIO) ReadInput() {
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		gbio.rows[0] &= 0x7
	} else {
		gbio.rows[0] |= 0x8
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		gbio.rows[0] &= 0xB
	} else {
		gbio.rows[0] |= 0x5
	}

	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		gbio.rows[0] &= 0xD
	} else {
		gbio.rows[0] |= 0x2
	}

	if ebiten.IsKeyPressed(ebiten.KeyX) {
		gbio.rows[0] &= 0xE
	} else {
		gbio.rows[0] |= 0x1
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		gbio.rows[1] &= 0xB
	} else {
		gbio.rows[1] |= 0x4
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		gbio.rows[1] &= 0x7
	} else {
		gbio.rows[1] |= 0x8
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		gbio.rows[1] &= 0xD
	} else {
		gbio.rows[1] |= 0x2
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		gbio.rows[1] &= 0xE
	} else {
		gbio.rows[1] |= 0x1
	}
}

func (gbio *GBIO) SetCol(data byte) {
	gbio.col = data & 0x30
}

func (gbio *GBIO) GetInput() byte {
	switch gbio.col {
	case 0x10:
		return gbio.rows[0]
	case 0x20:
		return gbio.rows[1]
	default:
		return 0x00
	}
}
