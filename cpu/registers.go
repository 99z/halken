package cpu

import (
	"encoding/binary"
)

// 8 bit registers
type Registers struct {
	a byte // Accumulator
	f byte // Flags

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

// incrementSP converts current SP to an integer,
// increments it by 1, then stores it back as 2 bytes
func (regs *Registers) incrementSP(amt uint16) {
	spInt := binary.LittleEndian.Uint16(regs.sp)
	spInt += amt
	binary.LittleEndian.PutUint16(regs.sp, spInt)
}

func (regs *Registers) incrementPC(amt uint16) {
	pcInt := binary.LittleEndian.Uint16(regs.PC)
	pcInt += amt
	binary.LittleEndian.PutUint16(regs.PC, pcInt)
}
