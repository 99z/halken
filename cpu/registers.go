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

	sp [2]byte // Stack pointer
	// Program counter
	// I THINK this will just refer to an index of the ROM as a byte array
	pc [2]byte
}

func (regs *Registers) writeAF(data [2]byte) {
	regs.a = data[0]
	regs.f = data[1]
}

func (regs *Registers) readAF() [2]byte {
	return [2]byte{regs.a, regs.f}
}

func (regs *Registers) writeBC(data [2]byte) {
	regs.b = data[0]
	regs.c = data[1]
}

func (regs *Registers) readBC() [2]byte {
	return [2]byte{regs.b, regs.c}
}

func (regs *Registers) writeDE(data [2]byte) {
	regs.d = data[0]
	regs.e = data[1]
}

func (regs *Registers) readDE() [2]byte {
	return [2]byte{regs.d, regs.e}
}

func (regs *Registers) writeHL(data [2]byte) {
	regs.h = data[0]
	regs.l = data[1]
}

func (regs *Registers) readHL() [2]byte {
	return [2]byte{regs.h, regs.l}
}
