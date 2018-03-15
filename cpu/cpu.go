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
	gbcpu.loadInstructions()
}

func (gbcpu *GBCPU) readPC() {
	
}