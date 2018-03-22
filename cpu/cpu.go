// Package cpu emulates the Sharp LR35902
// executors contains implementations of assembly instructions
// registers contains register information and methods
// instructions contains all CPU instructions
package cpu

// GBCPU represents an instance of an LR35902
// Reference: http://www.zilog.com/docs/z80/um0080.pdf
// Page 80 discusses clocks
type GBCPU struct {
	// Total machine cycles
	mCycles int
	// Total time cycles
	tCycles int
	Regs    *Registers
	Instrs  map[byte]Instruction
	Jumped  bool
}

// InitCPU initializes a new CPU struct
// Sets Regs and Instrs fields
// Sets program counter to location
func (gbcpu *GBCPU) InitCPU() {
	gbcpu.Regs = new(Registers)
	// For now, start PC at usual jump destination after
	// cartridge header information
	gbcpu.Regs.PC = append(gbcpu.Regs.PC, 0x50, 0x01)
	gbcpu.loadInstructions()
}

func (gbcpu *GBCPU) readPC() {
	// TODO
	// Might need if decide not to export Regs
}

// func (gbcpu *GBCPU) pushByteToStack(data byte) {
// 	gbcpu.Regs.decrementSP(1)
// 	GbMMU.WriteByte(gbcpu.Regs.sp, data)
// }

// func (gbcpu *GBCPU) popByteFromStack() byte {
// 	result := gbcpu.Regs.sp[0]
// 	gbcpu.Regs.incrementSP(1)
// 	return result
// }
