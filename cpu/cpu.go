// Package cpu emulates the Sharp LR35902
// executors contains implementations of assembly instructions
// registers contains register information and methods
// instructions contains all CPU instructions
package cpu

// GBCPU represents an instance of an LR35902
// Reference: http://www.zilog.com/docs/z80/um0080.pdf
// Page 80 discusses clocks
type GBCPU struct {
	// Total time cycles
	TCycles  uint16
	Regs     *Registers
	Instrs   map[byte]Instruction
	InstrsCB map[byte]Instruction
	Jumped   bool
	// Interrupt master enabled flag
	// Not accessed by a mem address and not technically(?) a register
	// Accessed directly by the CPU
	IME        byte
	EIReceived bool
	Halted     bool
	// Interrupt flag prior to halting
	IFPreHalt byte
}

// InitCPU initializes a new CPU struct
// Sets Regs and Instrs fields
// Sets program counter to location
func (gbcpu *GBCPU) InitCPU() {
	gbcpu.IME = 0
	gbcpu.Halted = false
	gbcpu.EIReceived = false
	gbcpu.Regs = new(Registers)
	gbcpu.Regs.InitRegs()
	// For now, start PC at usual jump destination after
	// cartridge header information
	gbcpu.Regs.PC = append(gbcpu.Regs.PC, 0x00, 0x01)
	gbcpu.loadInstructions()
}

// pushByteToStack decrements the SP by 1, then writes a byte at the addr
// pointed to by the SP
func (gbcpu *GBCPU) pushByteToStack(data byte) {
	gbcpu.Regs.decrementSP(1)
	GbMMU.WriteData(gbcpu.sliceToInt(gbcpu.Regs.sp), data)
}

// popByteFromStack gets the byte at the addr pointed to by the SP
// then increments the SP by 1
func (gbcpu *GBCPU) popByteFromStack() byte {
	result := GbMMU.ReadData(gbcpu.sliceToInt(gbcpu.Regs.sp))
	gbcpu.Regs.incrementSP(1)
	return result
}
