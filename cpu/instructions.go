// Contains instruction opcodes & definitions
// Reference: http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
package cpu

// Holds all relevant information for CPU instructions
// Don't think we need to store opcode in the struct since
// it will be equal to the byte key of the instruction in the map
type Instruction struct {
	Mnemonic    string
	// Number of T cycles instruction takes to execute
	// Divide by 4 to get number of M cycles
	tCycles     int
	NumOperands int
	Executor    func() // Executes appropriate function
}

// Non-CB prefixed instructions
// Parentheses indicate an address
// i8 is 8-bit immediate, i16 is 16-bit immediate
// a16 is a 16-bit address, a8 is an 8-bit address added to $FF00
// s8 is 8-bit signed data, added to PC to move it
//var Instructions map[byte]Instruction = 
//}

// CB prefixed instructions
// CB is the prefix byte. Like the Z80, the Sharp LR35902 will
// look up a CB prefixed instruction in a different instruction bank
// More info: http://www.z80.info/decoding.htm
var instructionsCB map[byte]Instruction = map[byte]Instruction{
	0x00: Instruction{"RLC B", 8, 0, func() {}},
}
