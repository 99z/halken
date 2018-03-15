package cpu

// Reference: http://www.zilog.com/docs/z80/um0080.pdf
// Page 80 discusses clocks
type GBCPU struct {
	// Total machine cycles
	mCycles	int
	// Total time cycles
	tCycles	int
	regs	*Registers
	Instrs	map[byte]Instruction
}

func (gbcpu *GBCPU) InitCPU() {
	gbcpu.regs = new(Registers)
	gbcpu.regs.pc = 0x0150
	gbcpu.loadInstructions()
}

func (gbcpu *GBCPU) readPC() {
	
}