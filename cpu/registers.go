package cpu

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

func (regs *Registers) writePair(reg1, reg2 *byte, data []byte) {
	*reg1 = data[1]
	*reg2 = data[0]
}

func (regs *Registers) readPair(reg1, reg2 *byte) [2]byte {
	return [2]byte{*reg1, *reg2}
}
