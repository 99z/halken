package cpu

import (
	"encoding/binary"
)

// Registers represents LR35902 register
// Notes:
// Carry flag - https://stackoverflow.com/questions/31409444/what-is-the-behavior-of-the-carry-flag-for-cp-on-a-game-boy
// Half carry flag - http://stackoverflow.com/questions/8868396/gbz80-what-constitutes-a-half-carry/8874607#8874607
// Zero flag - set if result was zero
type Registers struct {
	a byte // Accumulator
	// Flags
	// ZNHC 0000
	// Z = zero, N = subtract, H = half carry, C = carry
	f byte

	b byte
	c byte

	d byte
	e byte

	h byte
	l byte

	sp []byte // Stack pointer
	// Program counter
	// I THINK this will just refer to an index of the ROM as a byte array
	PC []byte
}

// writePair writes 2 bytes to a register pair
// Order is swapped because the Game Boy is Little Endian
func (regs *Registers) writePair(reg1, reg2 *byte, data []byte) {
	*reg1 = data[1]
	*reg2 = data[0]
}

// writePairFromInt writes a 16-bit integer to a register pair
func (regs *Registers) writePairFromInt(reg1, reg2 *byte, data uint16) {
	splitInt := make([]byte, 2)
	binary.LittleEndian.PutUint16(splitInt, data)
	*reg1 = splitInt[1]
	*reg2 = splitInt[0]
}

func (regs *Registers) readPair(reg1, reg2 *byte) [2]byte {
	return [2]byte{*reg1, *reg2}
}

func (regs *Registers) incrementHL(amt uint8) {
	newL := regs.l + amt

	if newL > 255 {
		regs.l = 0
		regs.h++
	} else {
		regs.l = newL
	}
}

// incrementSP converts current SP to an integer,
// increments it by 1, then stores it back as 2 bytes
func (regs *Registers) incrementSP(amt uint16) {
	spInt := binary.LittleEndian.Uint16(regs.sp)
	spInt += amt
	binary.LittleEndian.PutUint16(regs.sp, spInt)
}

func (regs *Registers) decrementSP(amt uint16) {
	spInt := binary.LittleEndian.Uint16(regs.sp)
	spInt--
	newSP := make([]byte, 2)
	binary.LittleEndian.PutUint16(newSP, spInt)
}

func (regs *Registers) incrementPC(amt uint16) {
	pcInt := binary.LittleEndian.Uint16(regs.PC)
	pcInt += amt
	binary.LittleEndian.PutUint16(regs.PC, pcInt)
}

func (regs *Registers) setZero(val byte) {
	regs.f |= (val << 7)
}

func (regs *Registers) getZero() byte {
	return (regs.f >> 7) & 1
}

func (regs *Registers) setSubtract(val byte) {
	regs.f |= (val << 6)
}

func (regs *Registers) setHalfCarry(val byte) {
	regs.f |= (val << 5)
}

func (regs *Registers) setCarry(val byte) {
	regs.f |= (val << 4)
}

func (regs *Registers) getCarry() byte {
	return (regs.f >> 4) & 1
}
