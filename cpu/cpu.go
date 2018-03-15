package cpu

// Reference: http://www.zilog.com/docs/z80/um0080.pdf
// Page 80 discusses clocks
type GBCPU struct {
	// Total machine cycles
	mCycles	int
	// Total time cycles
	tCycles	int
	Regs	*Registers
	Instrs	map[byte]Instruction
}

func (gbcpu *GBCPU) InitCPU() {
	gbcpu.Regs = new(Registers)
	// For now, start PC at usual jump destination after
	// cartridge header information
	gbcpu.Regs.PC = 0x0150
	gbcpu.loadInstructions()
}

func (gbcpu *GBCPU) readPC() {
	
}