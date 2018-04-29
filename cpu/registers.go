package cpu

import (
	"encoding/binary"
	"fmt"
)

// Registers represents Sharp LR35902 registers
// Each individual register is a byte, but AF, BC, DE, HL can be addressed
// as pairs (single 16-bit value)
// Stack pointer and program counter are 2 bytes always
// Could have also represented SP/PC as a single uint16. However, seems
// more semantically accurate to do them as byte slices w/ 2 values
// Notes:
// Carry flag - https://stackoverflow.com/questions/31409444/what-is-the-behavior-of-the-carry-flag-for-cp-on-a-game-boy
// Half carry flag - http://stackoverflow.com/questions/8868396/gbz80-what-constitutes-a-half-carry/8874607#8874607
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
	PC []byte // Program counter
}

// InitRegs sets post-bootrom register values
// GB register initial values:
// http://bgb.bircd.org/pandocs.htm#powerupsequence
func (regs *Registers) InitRegs() {
	regs.a, regs.f = 0x01, 0xB0
	regs.b, regs.c = 0x00, 0x13
	regs.d, regs.e = 0x00, 0xD8
	regs.h, regs.l = 0x01, 0x4D
	regs.sp = []byte{0xFF, 0xFE}
}

// SplitWord splits a 16 bit integer into 2 bytes
func (regs *Registers) SplitWord(rr uint16) (byte, byte) {
	return byte(rr >> 8), byte(rr)
}

// JoinRegs takes 2 bytes and returns a 16-bit integer
// Almost exclusively used to manipulate the value of a reg pair
func (regs *Registers) JoinRegs(reg1, reg2 *byte) uint16 {
	result := uint16(*reg2) | uint16(*reg1)<<8

	// Handle OOB
	// TODO Probably a potential issue elsewhere, change mem model to fix
	// if result >= 0xFFFF {
	// 	result--
	// }

	return result
}

// incrementSP converts current SP to an integer,
// increments it by amt, then stores it back as 2 bytes
func (regs *Registers) incrementSP(amt byte) {
	spInt := binary.LittleEndian.Uint16(regs.sp)
	spInt += uint16(amt)
	binary.LittleEndian.PutUint16(regs.sp, spInt)
}

// decrementSP converts current SP to an integer,
// decrements it by amt, then stores it back as 2 bytes
func (regs *Registers) decrementSP(amt byte) {
	spInt := binary.LittleEndian.Uint16(regs.sp)
	spInt -= uint16(amt)
	binary.LittleEndian.PutUint16(regs.sp, spInt)
}

// incrementPC converts current PC to an integer,
// increments it by amt, then stores it back as 2 bytes
func (regs *Registers) incrementPC(amt uint16) {
	pcInt := binary.LittleEndian.Uint16(regs.PC)
	pcInt += amt
	binary.LittleEndian.PutUint16(regs.PC, pcInt)
}

// setZero sets the 7th bit of register F
func (regs *Registers) setZero() {
	regs.f |= (1 << 7)
}

// clearZero clears the 7th bit of register F
func (regs *Registers) clearZero() {
	regs.f &^= (1 << 7)
}

// getZero returns the 7th bit of register F
// Returns 1 if set, 0 if not set
func (regs *Registers) getZero() byte {
	if regs.f&(1<<7) != 0 {
		return 1
	}

	return 0
}

// setSubtract sets the 6th bit of register F
func (regs *Registers) setSubtract() {
	regs.f |= (1 << 6)
}

// clearSubtract clears the 6th bit of register F
func (regs *Registers) clearSubtract() {
	mask := ^(1 << 6)
	regs.f &= byte(mask)
}

// getSubtract returns the 6th bit of register F
// Returns 1 if set, 0 if not set
func (regs *Registers) getSubtract() byte {
	if regs.f&(1<<6) != 0 {
		return 1
	}

	return 0
}

// setHalfCarry sets the 5th bit of register F
func (regs *Registers) setHalfCarry() {
	regs.f |= (1 << 5)
}

// clearHalfCarry clears the 5th bit of register F
func (regs *Registers) clearHalfCarry() {
	mask := ^(1 << 5)
	regs.f &= byte(mask)
}

// getHalfCarry returns the 5th bit of register F
// Returns 1 if set, 0 if not set
func (regs *Registers) getHalfCarry() byte {
	if regs.f&(1<<5) != 0 {
		return 1
	}

	return 0
}

// setCarry sets the 4th bit of register F
func (regs *Registers) setCarry() {
	regs.f |= (1 << 4)
}

// clearCarry clears the 4th bit of register F
func (regs *Registers) clearCarry() {
	mask := ^(1 << 4)
	regs.f &= byte(mask)
}

// getCarry returns the 4th bit of register F
// Returns 1 if set, 0 if not set
func (regs *Registers) getCarry() byte {
	if regs.f&(1<<4) != 0 {
		return 1
	}

	return 0
}

// Dump prints register values for debugging
func (regs *Registers) Dump() {
	fmt.Printf("AF: %02X %02X\n", regs.a, regs.f)
	fmt.Printf("BC: %02X %02X\n", regs.b, regs.c)
	fmt.Printf("DE: %02X %02X\n", regs.d, regs.e)
	fmt.Printf("HL: %02X %02X\n", regs.h, regs.l)
	fmt.Printf("SP: %04X\n", regs.sp)
	fmt.Printf("PC: %04X\n", regs.PC)
}
