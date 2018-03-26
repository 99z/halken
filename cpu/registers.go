package cpu

import (
	"encoding/binary"
	"fmt"
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

// GB register initial values
// http://bgb.bircd.org/pandocs.htm#powerupsequence
func (regs *Registers) InitRegs() {
	regs.writePair(&regs.a, &regs.f, []byte{0xB0, 0x01})
	regs.writePair(&regs.b, &regs.c, []byte{0x13, 0x00})
	regs.writePair(&regs.d, &regs.e, []byte{0xD8, 0x00})
	regs.writePair(&regs.h, &regs.l, []byte{0x4D, 0x01})

	regs.sp = []byte{0xFF, 0xFE}
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

	if newL == 0 {
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
	spInt -= amt
	binary.LittleEndian.PutUint16(regs.sp, spInt)
}

func (regs *Registers) incrementPC(amt uint16) {
	pcInt := binary.LittleEndian.Uint16(regs.PC)
	pcInt += amt
	binary.LittleEndian.PutUint16(regs.PC, pcInt)
}

func (regs *Registers) setZero() {
	regs.f |= (1 << 7)
}

func (regs *Registers) clearZero() {
	mask := ^(1 << 7)
	regs.f &= byte(mask)
}

func (regs *Registers) getZero() byte {
	return regs.f & (1 << 7)
}

func (regs *Registers) setSubtract() {
	regs.f |= (1 << 6)
}

func (regs *Registers) clearSubtract() {
	mask := ^(1 << 6)
	regs.f &= byte(mask)
}

func (regs *Registers) setHalfCarry() {
	regs.f |= (1 << 5)
}

func (regs *Registers) clearHalfCarry() {
	mask := ^(1 << 5)
	regs.f &= byte(mask)
}

func (regs *Registers) setCarry() {
	regs.f |= (1 << 4)
}

func (regs *Registers) clearCarry() {
	mask := ^(1 << 4)
	regs.f &= byte(mask)
}

func (regs *Registers) getCarry() byte {
	return regs.f & (1 << 4)
}

func (regs *Registers) Dump() {
	fmt.Printf("AF: %02X %02X\n", regs.a, regs.f)
	fmt.Printf("BC: %02X %02X\n", regs.b, regs.c)
	fmt.Printf("DE: %02X %02X\n", regs.d, regs.e)
	fmt.Printf("HL: %02X %02X\n", regs.h, regs.l)
	fmt.Printf("SP: %04X\n", regs.sp)
	fmt.Printf("PC: %04X\n", regs.PC)
}
