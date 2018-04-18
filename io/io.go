package io

import (
	"../mmu"
	"github.com/hajimehoshi/ebiten"
)

var GbMMU *mmu.GBMMU

type GBIO struct {
	rows [2]byte
	col  byte
}

const buttonCol byte = 0x10
const dpadCol byte = 0x20

func (gbio *GBIO) InitIO() {
	gbio.rows[0], gbio.rows[1] = 0x0F, 0x0F
	gbio.col = 0
	GbMMU.Memory[0xFF00] = 0xCF
}

func (gbio *GBIO) ReadInput() {
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		GbMMU.Memory[0xFF00] |= 0x2
		// fmt.Println(GbMMU.Memory[0xFF00])
	}
}
