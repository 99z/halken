package cpu

// Holds all relevant information for CPU instructions
// Don't think we need to store opcode in the struct since
// it will be equal to the byte key of the instruction in the map
type Instruction struct {
	Mnemonic    string
	Cycles      int
	NumOperands int
	Operands    [2]byte // [2]byte will hold 8-bit and 16-bit operands
}

// Non-CB prefixed instructions
// 'n' is an 8bit operation, 'nn' is 16-bit
var instructions map[byte]Instruction = map[byte]Instruction{
	0x00: Instruction{"NOP", 4, 0, [2]byte{}},
	0x01: Instruction{"LD BC,nn", 12, 2, [2]byte{}},
	0x02: Instruction{"LD (BC),A", 8, 0, [2]byte{}},
	0x03: Instruction{"INC BC", 8, 0, [2]byte{}},
	0x04: Instruction{"INC B", 4, 0, [2]byte{}},
	0x05: Instruction{"DEC B", 4, 0, [2]byte{}},
	0x06: Instruction{"LD B,n", 8, 1, [2]byte{}},
	0x07: Instruction{"RLCA", 4, 0, [2]byte{}},
	0x08: Instruction{"LD (nn),SP", 20, 2, [2]byte{}},
	0x09: Instruction{"ADD HL,BC", 8, 0, [2]byte{}},
	0x0A: Instruction{"LD A,(BC)", 8, 0, [2]byte{}},
	0x0B: Instruction{"DEC BC", 8, 0, [2]byte{}},
	0x0C: Instruction{"INC C", 4, 0, [2]byte{}},
	0x0D: Instruction{"DEC C", 4, 0, [2]byte{}},
	0x0E: Instruction{"LD C,n", 8, 1, [2]byte{}},
	0x0F: Instruction{"RRCA", 4, 0, [2]byte{}},
	0x10: Instruction{"STOP", 4, 1, [2]byte{}},
	0x11: Instruction{"LD DE,nn", 12, 2, [2]byte{}},
	0x12: Instruction{"LD (DE),A", 8, 0, [2]byte{}},
	0x13: Instruction{"INC DE", 4, 0, [2]byte{}},
	0x14: Instruction{"INC D", 4, 0, [2]byte{}},
	0x15: Instruction{"DEC D", 4, 0, [2]byte{}},
	0x16: Instruction{"LD D,n", 8, 1, [2]byte{}},
	0x17: Instruction{"RLA", 4, 0, [2]byte{}},
	0x18: Instruction{"JR n", 12, 1, [2]byte{}},
	0x19: Instruction{"ADD HL,DE", 8, 0, [2]byte{}},
	0x1A: Instruction{"LD A,(DE)", 8, 0, [2]byte{}},
	0x1B: Instruction{"DEC DE", 8, 0, [2]byte{}},
	0x1C: Instruction{"INC E", 4, 0, [2]byte{}},
	0x1D: Instruction{"DEC E", 4, 0, [2]byte{}},
	0x1E: Instruction{"LD E,n", 8, 1, [2]byte{}},
	0x1F: Instruction{"RRA", 4, 0, [2]byte{}},
	0x20: Instruction{"JR NZ, n", 8, 0, [2]byte{}},
	0x21: Instruction{"LD HL,nn", 12, 2, [2]byte{}},
	0x22: Instruction{"LD (HL+),A", 8, 0, [2]byte{}},
	0x23: Instruction{"INC HL", 8, 0, [2]byte{}},
	0x24: Instruction{"INC H", 4, 0, [2]byte{}},
	0x25: Instruction{"DEC H", 4, 0, [2]byte{}},
	0x26: Instruction{"LD H,n", 8, 1, [2]byte{}},
	0x27: Instruction{"DAA", 4, 0, [2]byte{}},
	0x28: Instruction{"JR Z,n", 8, 1, [2]byte{}},
	0x29: Instruction{"ADD HL,HL", 8, 0, [2]byte{}},
	0x2A: Instruction{"LD A,(HL+)", 8, 0, [2]byte{}},
	0x2B: Instruction{"DEC HL", 8, 0, [2]byte{}},
	0x2C: Instruction{"INC L", 4, 0, [2]byte{}},
	0x2D: Instruction{"DEC L", 4, 0, [2]byte{}},
	0x2E: Instruction{"LD L,n", 8, 1, [2]byte{}},
	0x2F: Instruction{"CPL", 4, 0, [2]byte{}},
	0x30: Instruction{"JR NC,n", 8, 1, [2]byte{}},
	0x31: Instruction{"LD SP,nn", 12, 2, [2]byte{}},
	0x32: Instruction{"LD (HL-),A", 8, 0, [2]byte{}},
	0x33: Instruction{"INC SP", 8, 0, [2]byte{}},
	0x34: Instruction{"INC (HL)", 12, 0, [2]byte{}},
	0x35: Instruction{"DEC (HL)", 12, 0, [2]byte{}},
	0x36: Instruction{"LD (HL),n", 12, 1, [2]byte{}},
	0x37: Instruction{"SCF", 4, 0, [2]byte{}},
	0x38: Instruction{"JR C,n", 8, 1, [2]byte{}},
	0x39: Instruction{"ADD HL,SP", 8, 0, [2]byte{}},
	0x3A: Instruction{"LD A,(HL-)", 8, 0, [2]byte{}},
	0x3B: Instruction{"DEC SP", 8, 0, [2]byte{}},
	0x3C: Instruction{"INC A", 4, 0, [2]byte{}},
	0x3D: Instruction{"DEC A", 4, 0, [2]byte{}},
	0x3E: Instruction{"LD A,n", 8, 1, [2]byte{}},
	0x3F: Instruction{"CCF", 4, 0, [2]byte{}},
}

// CB prefixed instructions
// CB is the prefix byte. Like the Z80, the Sharp LR35902 will
// look up a CB prefixed instruction in a different instruction bank
var instructionsCB map[byte]Instruction = map[byte]Instruction{
	0x00: Instruction{"RLC B", 8, 0, [2]byte{}},
}
