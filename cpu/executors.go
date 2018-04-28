// Package cpu executors contains implementations of all non-CB prefixed
// Sharp LR35902 instructions
// Non-CB prefixed means these are only run if the instruction we received
// is not `CB .. ..`
// Executor function field on the Instruction struct executes these
package cpu

import (
	"encoding/binary"

	"../mmu"
)

// GbMMU variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

// LDrr -> e.g. LD A,B
// Loads value in one register to another
func (gbcpu *GBCPU) LDrr(to, from *byte) {
	*to = *from
}

// LDrrnn -> e.g. LD BC,i16
// Loads 2 8-bit immediate operands into register pair
func (gbcpu *GBCPU) LDrrnn(reg1, reg2 *byte) {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.writePair(reg1, reg2, operands)
}

// LDSPHL -> e.g. LD SP,HL
// Loads bytes from register pair HL into SP
func (gbcpu *GBCPU) LDSPHL() {
	gbcpu.Regs.sp = []byte{gbcpu.Regs.l, gbcpu.Regs.h}
}

// LDHLSPs -> e.g. LD BC,SP+s8
// Loads value of SP + signed 8-bit value into register pair
// HC and C are a little weird for this instruction
// https://stackoverflow.com/questions/5159603/gbz80-how-does-ld-hl-spe-affect-h-and-c-flags
func (gbcpu *GBCPU) LDHLSPs() {
	operand := gbcpu.getOperands(1)[0]
	sp := int(binary.LittleEndian.Uint16(gbcpu.Regs.sp))
	result := 0

	if operand > 127 {
		result = sp - (256 - int(operand))
	} else {
		result = sp + int(operand)
	}

	check := sp ^ int(operand) ^ ((sp + int(operand)) & 0xFFFF)

	if (check & 0x100) != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	if (check & 0x10) != 0 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(uint16(result))

	gbcpu.Regs.clearZero()
	gbcpu.Regs.clearSubtract()
}

// LDrn -> e.g. LD B,i8
// Loads 1 8-bit immediate operand into a register
func (gbcpu *GBCPU) LDrn(reg *byte) {
	operand := gbcpu.getOperands(1)
	*reg = operand[0]

}

// INCrr -> e.g. INC BC
// Increments register pair by 1
// Flags: none
func (gbcpu *GBCPU) INCrr(reg1, reg2 *byte) {
	rr := gbcpu.Regs.JoinRegs(reg1, reg2)
	rr++
	*reg1, *reg2 = gbcpu.Regs.SplitWord(rr)
}

// DECrr -> e.g. DEC BC
// Decrements register pair by 1
// Flags: none
func (gbcpu *GBCPU) DECrr(reg1, reg2 *byte) {
	// gbcpu.Regs.decrementPair(reg1, reg2, 1)
	rr := gbcpu.Regs.JoinRegs(reg1, reg2)
	rr--
	*reg1, *reg2 = gbcpu.Regs.SplitWord(rr)
}

// INCr -> e.g. INC B
// Adds 1 to register, sets new value
// Flags: Z0H-
func (gbcpu *GBCPU) INCr(reg *byte) {
	hc := (((*reg & 0xf) + (1 & 0xf)) & 0x10) == 0x10
	*reg++

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// INCHL -> e.g. INC (HL)
// Increment value at addr (HL)
// Flags: Z0H-
func (gbcpu *GBCPU) INCHL() {
	addr := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	result := GbMMU.Memory[addr] + 1

	if (result^0x01^GbMMU.Memory[addr])&0x10 == 0x10 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	GbMMU.WriteByte(addr, result)

	if GbMMU.Memory[addr] == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearSubtract()
}

// DECHL -> e.g. DEC (HL)
// Decrement value at addr (HL)
// Flags: Z1H-
func (gbcpu *GBCPU) DECHL() {
	addr := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	result := GbMMU.Memory[addr] - 1

	if (result^0x01^GbMMU.Memory[addr])&0x10 == 0x10 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	GbMMU.WriteByte(addr, result)

	if GbMMU.Memory[addr] == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.setSubtract()
}

// INCSP -> e.g. INC SP
// Increment value of stack pointer by 1
// Flags: none
func (gbcpu *GBCPU) INCSP() {
	newSP := gbcpu.sliceToInt(gbcpu.Regs.sp)
	newSP++
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, newSP)
}

// DECSP -> e.g. DEC SP
// Decrement value of stack pointer by 1
// Flags: none
func (gbcpu *GBCPU) DECSP() {
	newSP := gbcpu.sliceToInt(gbcpu.Regs.sp)
	newSP--
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, newSP)
}

// DECr -> e.g. DEC B
// Subtracts 1 from register, sets new value
// Flags: Z1H-
func (gbcpu *GBCPU) DECr(reg *byte) {
	hc := (((*reg & 0xf) - (1 & 0xf)) & 0x10) == 0x10
	*reg--

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag
	gbcpu.Regs.setSubtract()

}

// RLCA performs 8-bit rotation to the left
// Rotated bit is copied to carry
// Flags: 000C
// GB opcodes list show Z being set to zero, but other sources disagree
// Z80 does not modify Z, other emulator sources do
// Reference: https://hax.iimarckus.org/topic/1617/
func (gbcpu *GBCPU) RLCA() {
	carry := gbcpu.Regs.a >> 7
	gbcpu.Regs.a = (gbcpu.Regs.a<<1 | carry)

	if carry != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearZero()
}

// RLA rotates register A left through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: 000C
func (gbcpu *GBCPU) RLA() {
	lsBit := 0

	if gbcpu.Regs.a&0x80 == 0x80 {
		lsBit = 1
	}

	gbcpu.Regs.a = gbcpu.Regs.a << 1

	if gbcpu.Regs.getCarry() != 0 {
		gbcpu.Regs.a ^= 0x01
	}

	if lsBit == 1 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearZero()

}

// RRCA performs 8-bit rotation to the right
// Rotated bit is copied to carry
// Flags: 000C
// GB opcodes list show Z being set to zero, but other sources disagree
// Z80 does not modify Z, other emulator sources do
func (gbcpu *GBCPU) RRCA() {
	carry := gbcpu.Regs.a << 7
	gbcpu.Regs.a = (gbcpu.Regs.a>>1 | carry)

	if carry != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearZero()
}

// RRA rotates register A right through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: 000C
func (gbcpu *GBCPU) RRA() {
	msBit := 0

	if gbcpu.Regs.a&0x01 == 0x01 {
		msBit = 1
	}

	gbcpu.Regs.a = gbcpu.Regs.a >> 1

	if gbcpu.Regs.getCarry() != 0 {
		gbcpu.Regs.a ^= 0x80
	}

	if msBit == 1 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearZero()

}

// RST pushes current PC + 3 onto stack
// MSB of PC is set to 0x00, LSB is set to argument
// Handles 0000,0008,0010,0018,0020,0028,0030,0038 which come from ROM
// Flags: none
func (gbcpu *GBCPU) RST(imm byte) {
	nextInstr := gbcpu.sliceToInt(gbcpu.Regs.PC) + 1
	nextInstrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(nextInstrBytes, nextInstr)
	gbcpu.pushByteToStack(nextInstrBytes[1])
	gbcpu.pushByteToStack(nextInstrBytes[0])

	gbcpu.Regs.PC = []byte{imm, 0x00}
	gbcpu.Jumped = true
}

// RSTI runs interrupt handlers
// Does not have a corresponding opcode, but is slightly different than RST
// Flags: none
func (gbcpu *GBCPU) RSTI(addr byte) {
	// Disable further interrupts while vblank is executing
	gbcpu.IME = 0
	gbcpu.EIReceived = false

	// Push current PC to stack
	gbcpu.pushByteToStack(gbcpu.Regs.PC[1])
	gbcpu.pushByteToStack(gbcpu.Regs.PC[0])

	// Jump to vblank handler address
	gbcpu.Regs.PC = []byte{addr, 0x00}
	gbcpu.Jumped = true
}

// LDaaSP -> e.g. LD (a16),SP
// Loads value of SP into addr provided by operands
// Since SP is 2 bytes, we write to (a16) and (a16 + 1)
// Flags: none
func (gbcpu *GBCPU) LDaaSP() {
	operands := gbcpu.getOperands(2)
	addrInc := binary.LittleEndian.Uint16(operands) + 1
	GbMMU.WriteByte(gbcpu.sliceToInt(operands), gbcpu.Regs.sp[0])
	GbMMU.WriteByte(addrInc, gbcpu.Regs.sp[1])
}

// LDSPnn -> e.g. LD SP,i16
// Loads 16 bit value from next 2 bytes into SP
// Flags: none
func (gbcpu *GBCPU) LDSPnn() {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.sp = operands
}

// ADDrr -> e.g. ADD A,B
// Values of reg1 and reg2 are added together
// Result is written into reg1
// Flags: Z0HC
func (gbcpu *GBCPU) ADDrr(reg1, reg2 *byte) {
	oldVal := *reg1
	result := *reg1 + *reg2
	hc := (((*reg1 & 0xf) + (*reg2 & 0xf)) & 0x10) == 0x10
	*reg1 = result

	// Check for zero
	if *reg1 == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	// Occurred if byte overflows
	if *reg1 < oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// ADDAn -> e.g. ADD A,i8
// Values of reg and i8 are added together
// Result is written into reg
// Flags: Z0HC
func (gbcpu *GBCPU) ADDAn() {
	oldVal := gbcpu.Regs.a
	operand := gbcpu.getOperands(1)[0]
	hc := (((gbcpu.Regs.a & 0xf) + (operand & 0xf)) & 0x10) == 0x10
	gbcpu.Regs.a = gbcpu.Regs.a + operand

	// Check for zero
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if gbcpu.Regs.a < oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// ADCAn -> e.g. ADC A,i8
// Values of operand and carry flag are added to reg
// Result is written into reg
// Flags: Z0HC
func (gbcpu *GBCPU) ADCAn() {
	carry := int(gbcpu.Regs.getCarry())
	operand := gbcpu.getOperands(1)[0]

	// Check for carry
	if ((int(gbcpu.Regs.a) & 0xFF) + (int(operand) & 0xFF) + carry) > 0xFF {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if ((int(gbcpu.Regs.a) & 0xF) + (int(operand) & 0xF) + carry) > 0xF {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a += operand + byte(carry)

	// Check for zero
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// ADCrr -> e.g. ADC A,B
// Values of reg1, reg2 and carry flag are added together
// Result is written into reg1
// Flags: Z0HC
func (gbcpu *GBCPU) ADCrr(reg1, reg2 *byte) {
	carry := int(gbcpu.Regs.getCarry())

	if ((int(*reg1) & 0xF) + (int(*reg2) & 0xF) + carry) > 0xF {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	if ((int(*reg1) & 0xFF) + (int(*reg2) & 0xFF) + carry) > 0xFF {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	*reg1 += *reg2 + byte(carry)

	if *reg1 == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearSubtract()
}

// ADCAHL -> e.g. ADC A,(HL)
// Adds value at addr (HL) + carry bit to reg A
// Flags: Z0HC
func (gbcpu *GBCPU) ADCAHL() {
	carry := int(gbcpu.Regs.getCarry())
	operand := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]

	// Check for carry
	if ((int(gbcpu.Regs.a) & 0xFF) + (int(operand) & 0xFF) + carry) > 0xFF {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if ((int(gbcpu.Regs.a) & 0xF) + (int(operand) & 0xF) + carry) > 0xF {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a += operand + byte(carry)

	// Check for zero
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// ADDAHL -> e.g. ADD A,(HL)
// Adds value at addr (HL) to reg A
// Flags: Z0HC
func (gbcpu *GBCPU) ADDAHL() {
	operand := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]
	oldVal := gbcpu.Regs.a
	result := gbcpu.Regs.a + operand
	hc := (((gbcpu.Regs.a & 0xf) + (operand & 0xf)) & 0x10) == 0x10
	gbcpu.Regs.a = result

	// Check for zero
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	// Occurred if byte overflows
	if gbcpu.Regs.a < oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.clearSubtract()
}

// ADDHLrr -> e.g. ADD HL,BC
// Values of HL and reg1reg2 are added together
// Result is written into HL
// Flags: -0HC
func (gbcpu *GBCPU) ADDHLrr(reg1, reg2 *byte) {
	hlInt := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	rrInt := gbcpu.Regs.JoinRegs(reg1, reg2)
	sum := rrInt + hlInt

	// Check for carry
	if sum < hlInt {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if (sum^rrInt^hlInt)&0x1000 == 0x1000 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(sum)

	gbcpu.Regs.clearSubtract()
}

// ADDSPs -> e.g. ADD SP,s8
// Adds signed 8-bit value to SP
// Sets SP to new value
// Flags: 00HC
func (gbcpu *GBCPU) ADDSPs() {
	operand := gbcpu.getOperands(1)[0]
	sp := int(binary.LittleEndian.Uint16(gbcpu.Regs.sp))
	result := 0

	if operand > 127 {
		result = sp - (256 - int(operand))
	} else {
		result = sp + int(operand)
	}

	check := sp ^ int(operand) ^ ((sp + int(operand)) & 0xFFFF)

	if (check & 0x100) == 0x100 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	if (check & 0x10) == 0x10 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, uint16(result))

	gbcpu.Regs.clearZero()
	gbcpu.Regs.clearSubtract()
}

// ADDHLSP -> e.g. ADD HL,SP
// Adds HL and SP, then sets HL to result
// Flags: -0HC
func (gbcpu *GBCPU) ADDHLSP() {
	hl := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	sp := binary.LittleEndian.Uint16(gbcpu.Regs.sp)
	result := hl + sp

	if result < hl {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	if (result^sp^hl)&0x1000 == 0x1000 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(result)

	gbcpu.Regs.clearSubtract()
}

// ANDr -> e.g. AND B
// Bitwise AND of reg into A
// Flags: Z010
func (gbcpu *GBCPU) ANDr(reg *byte) {
	gbcpu.Regs.a &= *reg

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.setHalfCarry()
	gbcpu.Regs.clearCarry()
}

// ANDn -> e.g. AND i8
// Bitwise AND of i8 into A
// Flags: Z010
func (gbcpu *GBCPU) ANDn() {
	operands := gbcpu.getOperands(1)
	gbcpu.Regs.a &= operands[0]

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.setHalfCarry()
	gbcpu.Regs.clearCarry()
}

// ANDHL -> e.g. AND (HL)
// Bitwise AND of value at addr (HL) into A
// Flags: Z010
func (gbcpu *GBCPU) ANDHL() {
	val := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]
	gbcpu.Regs.a &= val

	if gbcpu.Regs.a == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.setHalfCarry()
	gbcpu.Regs.clearCarry()
}

// ORr -> e.g. OR B
// Bitwise OR of reg into A
// Flags: Z000
func (gbcpu *GBCPU) ORr(reg *byte) {
	gbcpu.Regs.a |= *reg

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// ORn -> e.g. OR i8
// Bitwise OR of i8 into A
// Flags: Z000
func (gbcpu *GBCPU) ORn() {
	operand := gbcpu.getOperands(1)
	gbcpu.Regs.a |= operand[0]

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// ORHL -> e.g. OR (HL)
// Bitwise OR of byte at addr
// Flags: Z000
func (gbcpu *GBCPU) ORHL() {
	val := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]
	gbcpu.Regs.a |= val

	if gbcpu.Regs.a == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// XORr -> e.g. XOR B
// Bitwise XOR of reg into A
// Flags: Z000
func (gbcpu *GBCPU) XORr(reg *byte) {
	gbcpu.Regs.a ^= *reg

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// XORn -> e.g. XOR i8
// Bitwise XOR of i8 into A
// Flags: Z000
func (gbcpu *GBCPU) XORn() {
	operand := gbcpu.getOperands(1)
	gbcpu.Regs.a ^= operand[0]

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// XORaa -> e.g. XOR (HL)
// Bitwise XOR of value at addr a1a2 into A
// Flags: Z000
func (gbcpu *GBCPU) XORaa(a1, a2 *byte) {
	val := GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*a2, *a1})]
	gbcpu.Regs.a ^= val

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set flags
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearCarry()
}

// SUBn -> e.g. SUB i8
// Value of i8 is subtracted from A
// Result is written into reg
// Flags: Z1HC
func (gbcpu *GBCPU) SUBn() {
	operand := gbcpu.getOperands(1)[0]

	if (int(gbcpu.Regs.a) & 0xFF) < (int(operand) & 0xFF) {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	if (int(gbcpu.Regs.a) & 0xF) < (int(operand) & 0xF) {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a -= operand

	if gbcpu.Regs.a == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set subtract flag
	gbcpu.Regs.setSubtract()
}

// SUBr -> e.g. SUB B
// Value of reg is subtracted from A
// Result is written into reg
// Flags: Z1HC
func (gbcpu *GBCPU) SUBr(reg *byte) {
	if (int(gbcpu.Regs.a) & 0xFF) < (int(*reg) & 0xFF) {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	if (int(gbcpu.Regs.a) & 0xF) < (int(*reg) & 0xF) {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a -= *reg

	if gbcpu.Regs.a == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Set subtract flag
	gbcpu.Regs.setSubtract()
}

// SUBHL -> e.g. SUB (HL)
// Subtract value at addr (HL) from A
// Write result to A
// Flags: Z1HC
func (gbcpu *GBCPU) SUBHL() {
	operand := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]
	oldVal := gbcpu.Regs.a
	hc := (((gbcpu.Regs.a & 0xf) - (operand & 0xf)) & 0x10) == 0x10
	gbcpu.Regs.a = gbcpu.Regs.a - operand

	// Check for zero
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if gbcpu.Regs.a > oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract()
}

// SBCrr -> e.g. SBC A,B
// Sum of reg2 and carry flag is subtracted from reg1
// Result is written into reg1
// Flags: Z1HC
func (gbcpu *GBCPU) SBCrr(reg1, reg2 *byte) {
	carry := gbcpu.Regs.getCarry()
	result := (int(*reg1) - int(*reg2)) - int(carry)

	if result < 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	result &= 0xFF

	if result == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if ((result ^ int(*reg1) ^ int(*reg2)) & 0x10) == 0x10 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	*reg1 = byte(result)

	gbcpu.Regs.setSubtract()
}

// SBCAn -> e.g. SBC A,i8
// Sum of i8 and carry flag is subtracted from reg
// Result is written into reg
// Flags: Z1HC
func (gbcpu *GBCPU) SBCAn() {
	carry := gbcpu.Regs.getCarry()
	operand := int(gbcpu.getOperands(1)[0])
	result := ((int(gbcpu.Regs.a)) - operand) - int(carry)

	// Check for carry
	if result < 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	result &= 0xFF

	// Check for zero
	if result == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for half carry
	if ((result ^ operand ^ int(gbcpu.Regs.a)) & 0x10) == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a = byte(result)

	// Set subtract flag
	gbcpu.Regs.setSubtract()
}

// SBCAHL -> e.g. SBC A,(HL)
// Subtract value at addr (HL) and carry flag from reg A
// Write result to reg A
// Flags: Z1HC
func (gbcpu *GBCPU) SBCAHL() {
	carry := gbcpu.Regs.getCarry()
	operand := GbMMU.Memory[gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)]
	result := (int(gbcpu.Regs.a) - int(operand)) - int(carry)

	if result < 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	result &= 0xFF

	if result == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if ((result ^ int(gbcpu.Regs.a) ^ int(operand)) & 0x10) == 0x10 {
		gbcpu.Regs.setHalfCarry()
	} else {
		gbcpu.Regs.clearHalfCarry()
	}

	gbcpu.Regs.a = byte(result)

	gbcpu.Regs.setSubtract()
}

// CPr -> e.g. CP B
// Subtraction from accumulator that doesn't update it
// Only updates flags
// Flags: Z1HC
func (gbcpu *GBCPU) CPr(reg *byte) {
	sub := gbcpu.Regs.a - *reg
	hc := (((gbcpu.Regs.a & 0xf) - (*reg & 0xf)) & 0x10) == 0x10

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if gbcpu.Regs.a < *reg {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract()
}

// CPn -> e.g. CP i8
// Subtraction from accumulator that doesn't update it
// Only updates flags
// Flags: Z1HC
func (gbcpu *GBCPU) CPn() {

	operand := gbcpu.getOperands(1)[0]
	oldVal := gbcpu.Regs.a
	hc := (((gbcpu.Regs.a & 0xf) - (operand & 0xf)) & 0x10) == 0x10
	sub := gbcpu.Regs.a - operand

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	// Carry is set if the sum overflows 0xFF
	// Thus, if the result of subtraction is greater than the
	// initial value, overflow must have occurred
	if sub > oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	// HC is set if a byte from the first nibble moves into the next
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract()
}

// CPaa -> e.g. CP (HL)
// Subtraction of value at addr from accumulator that doesn't update it
// Only updates flags
// Flags: Z1HC
func (gbcpu *GBCPU) CPaa(a1, a2 *byte) {
	operand := GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*a2, *a1})]
	oldVal := gbcpu.Regs.a
	hc := (((gbcpu.Regs.a & 0xf) - (operand & 0xf)) & 0x10) == 0x10
	sub := gbcpu.Regs.a - operand

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	// Carry is set if the sum overflows 0xFF
	// Thus, if the result of subtraction is greater than the
	// initial value, overflow must have occurred
	if sub > oldVal {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	// Check for half carry
	// HC is set if a byte from the first nibble moves into the next
	if hc {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry()
	} else {
		// Half-carry did not occur
		gbcpu.Regs.clearHalfCarry()
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract()
}

// PUSHrr -> e.g PUSH BC
// Copies reg1reg2 into addr (SP)
// Flags: none
func (gbcpu *GBCPU) PUSHrr(reg1, reg2 *byte) {
	gbcpu.pushByteToStack(*reg1)
	gbcpu.pushByteToStack(*reg2)
}

// POPrr copies 2 bytes at addr (SP) into the operand
// Flags: ZNHC
func (gbcpu *GBCPU) POPrr(reg1, reg2 *byte) {
	b1 := gbcpu.popByteFromStack()
	b2 := gbcpu.popByteFromStack()
	*reg1 = b2
	*reg2 = b1

	// Flags are masked out
	// https://forums.nesdev.com/viewtopic.php?f=20&t=12815
	gbcpu.Regs.f &= 0xF0
}

// LDraa -> e.g. LD A,(BC)
// Loads value at addr (a1a2) into reg
// Flags: none
func (gbcpu *GBCPU) LDraa(reg, a1, a2 *byte) {
	*reg = GbMMU.ReadByte(gbcpu.Regs.JoinRegs(a1, a2))
}

// LDaar -> e.g. LD (BC),A
// Loads reg into value at addr (a1a2)
// Flags: none
func (gbcpu *GBCPU) LDaar(a1, a2, reg *byte) {
	GbMMU.WriteByte(gbcpu.Regs.JoinRegs(a1, a2), *reg)
}

// LDaaA -> e.g. LD (a16),A
// Loads reg A into addr specified by next 2 bytes
// Flags: none
func (gbcpu *GBCPU) LDaaA(reg *byte) {
	operands := gbcpu.getOperands(2)
	operandsInt := gbcpu.sliceToInt(operands)
	GbMMU.WriteByte(operandsInt, *reg)
}

// LDAaa -> e.g. LD A,(a16)
// Loads value at addr specified by next 2 bytes into A
// Flags: none
func (gbcpu *GBCPU) LDAaa(reg *byte) {
	operands := gbcpu.getOperands(2)
	addr := gbcpu.sliceToInt(operands)
	*reg = GbMMU.ReadByte(addr)
}

// LDffCA -> e.g. LD ($FF00+C),A
// Sets value at addr (0xFF00+C) to A
func (gbcpu *GBCPU) LDffCA() {
	GbMMU.WriteByte(0xFF00+uint16(gbcpu.Regs.c), gbcpu.Regs.a)
}

// LDAffC -> e.g. LD A,($FF00+C)
// Sets A to value at addr (0xFF00+C)
func (gbcpu *GBCPU) LDAffC() {
	gbcpu.Regs.a = GbMMU.ReadByte(0xFF00 + uint16(gbcpu.Regs.c))
}

// LDffnA -> e.g. LD ($FF00+a8),A
// Loads A into value at addr ($FF00+a8)
func (gbcpu *GBCPU) LDffnA() {
	operand := gbcpu.getOperands(1)[0]
	GbMMU.WriteByte(0xFF00+uint16(operand), gbcpu.Regs.a)
}

// LDAffn -> e.g. LD A,($FF00+a8)
// Loads value at addr ($FF00+a8) into A
func (gbcpu *GBCPU) LDAffn() {
	operand := gbcpu.getOperands(1)[0]
	gbcpu.Regs.a = GbMMU.ReadByte(0xFF00 + uint16(operand))
}

// LDHLn -> e.g. LD (HL),i8
// Loads 8 bit immediate into addr (HL)
func (gbcpu *GBCPU) LDHLn() {
	operand := gbcpu.getOperands(1)[0]
	GbMMU.WriteByte(gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l), operand)
}

// LDDrHL -> e.g. LDD A,(HL)
// Load value at addr (HL) into reg
// Decrement HL
func (gbcpu *GBCPU) LDDrHL(reg *byte) {
	addr := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	*reg = GbMMU.Memory[addr]
	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(addr - 1)
}

// LDDHLr -> e.g. LDD (HL),A
// Load reg to value at addr (HL)
// Decrement HL
func (gbcpu *GBCPU) LDDHLr(reg *byte) {
	addr := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	GbMMU.WriteByte(addr, *reg)
	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(addr - 1)
}

// LDIHLA -> e.g. LDI (HL),A
// Set value at address a1a2 to value in reg
// Increment reg
func (gbcpu *GBCPU) LDIHLA() {
	GbMMU.WriteByte(gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l), gbcpu.Regs.a)
	hl := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	hl++
	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(hl)
}

// LDIAHL -> e.g. LDI A,(HL)
// Set value in reg to value at address a1a2
// Increment HL
func (gbcpu *GBCPU) LDIAHL() {
	gbcpu.Regs.a = GbMMU.ReadByte(gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l))
	hl := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	hl++
	gbcpu.Regs.h, gbcpu.Regs.l = gbcpu.Regs.SplitWord(hl)
}

// JPaa -> e.g. JP a16
// Jumps to addr a16
func (gbcpu *GBCPU) JPaa() {
	jmpAddr := gbcpu.getOperands(2)
	gbcpu.Regs.PC = jmpAddr
	gbcpu.Jumped = true
}

// JPHL -> e.g. JP (HL)
// Jumps to addr specified by addr (HL)
func (gbcpu *GBCPU) JPHL() {
	addr := gbcpu.Regs.JoinRegs(&gbcpu.Regs.h, &gbcpu.Regs.l)
	binary.LittleEndian.PutUint16(gbcpu.Regs.PC, addr)
	gbcpu.Jumped = true
}

// JPZaa -> e.g. JP Z,a16
// Jumps to a16 if Z is set
func (gbcpu *GBCPU) JPZaa() int {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z != 0 {
		gbcpu.JPaa()
		return 4
	}

	return 0
}

// JPNZaa -> e.g. JP NZ,a16
// Jumps to a16 if Z is not set
func (gbcpu *GBCPU) JPNZaa() int {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.JPaa()
		return 4
	}

	return 0
}

// JPCaa -> e.g. JP C,a16
// Jumps to a16 if C is set
func (gbcpu *GBCPU) JPCaa() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.JPaa()
		return 4
	}

	return 0
}

// JPNCaa -> e.g. JP NC,a16
// Jumps to a16 if C is not set
func (gbcpu *GBCPU) JPNCaa() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.JPaa()
		return 4
	}

	return 0
}

// CALLaa -> e.g. CALL $028B
// Pushes the addr at PC+3 to the stack
// Jumps to the address specified by next 2 bytes
func (gbcpu *GBCPU) CALLaa() {
	operands := gbcpu.getOperands(2)
	nextInstr := gbcpu.sliceToInt(gbcpu.Regs.PC) + 3
	nextInstrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(nextInstrBytes, nextInstr)
	gbcpu.pushByteToStack(nextInstrBytes[1])
	gbcpu.pushByteToStack(nextInstrBytes[0])
	gbcpu.Regs.PC = operands

	gbcpu.Jumped = true
}

// CALLZaa -> e.g. CALL Z,a16
// Performs CALL to a16 if Z is set
func (gbcpu *GBCPU) CALLZaa() int {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z != 0 {
		gbcpu.CALLaa()
		return 12
	}

	return 0
}

// CALLNZaa -> e.g. CALL NZ,a16
// Performs CALL to a16 if Z is not set
func (gbcpu *GBCPU) CALLNZaa() int {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.CALLaa()
		return 12
	}

	return 0
}

// CALLCaa -> e.g. CALL C,a16
// Performs CALL to a16 if C is set
func (gbcpu *GBCPU) CALLCaa() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.CALLaa()
		return 12
	}

	return 0
}

// CALLNCaa -> e.g. CALL NC,a16
// Performs CALL to a16 if Z is not set
func (gbcpu *GBCPU) CALLNCaa() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.CALLaa()
		return 12
	}

	return 0
}

// JRn -> e.g. JR s8
// Add byte at PC + 1 to PC, and set PC to that value
func (gbcpu *GBCPU) JRn() {
	operand := gbcpu.getOperands(1)[0]
	gbcpu.Regs.incrementPC(2)

	if operand > 127 {
		location := 256 - uint16(operand)
		newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) - uint16(location)
		binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
	} else {
		newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) + uint16(operand)
		binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
	}
	gbcpu.Jumped = true
}

// JRZn -> e.g. JR Z,s8
// Performs JR to PC + s8 if Z is set
func (gbcpu *GBCPU) JRZn() int {
	operand := gbcpu.getOperands(1)[0]

	if gbcpu.Regs.getZero() != 0 {
		gbcpu.Regs.incrementPC(2)

		if operand > 127 {
			location := 256 - uint16(operand)
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) - uint16(location)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		} else {
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) + uint16(operand)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		}
		gbcpu.Jumped = true
		return 4
	}

	return 0
}

// JRNZn -> e.g. JR NZ,s8
// Performs JR to PC + s8 if Z is not set
func (gbcpu *GBCPU) JRNZn() int {
	operand := gbcpu.getOperands(1)[0]

	if gbcpu.Regs.getZero() == 0 {
		gbcpu.Regs.incrementPC(2)

		if operand > 127 {
			location := 256 - uint16(operand)
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) - uint16(location)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		} else {
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) + uint16(operand)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		}
		gbcpu.Jumped = true
		return 4
	}

	return 0
}

// JRCn -> e.g. JR C,s8
// Performs JR to PC + s8 if C is set
func (gbcpu *GBCPU) JRCn() int {
	operand := gbcpu.getOperands(1)[0]

	if gbcpu.Regs.getCarry() != 0 {
		gbcpu.Regs.incrementPC(2)

		if operand > 127 {
			location := 256 - uint16(operand)
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) - uint16(location)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		} else {
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) + uint16(operand)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		}
		gbcpu.Jumped = true
		return 4
	}

	return 0
}

// JRNCn -> e.g. JR NC,s8
// Performs JR to PC + s8 if C is not set
func (gbcpu *GBCPU) JRNCn() int {
	operand := gbcpu.getOperands(1)[0]

	if gbcpu.Regs.getCarry() == 0 {
		gbcpu.Regs.incrementPC(2)

		if operand > 127 {
			location := 256 - uint16(operand)
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) - uint16(location)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		} else {
			newPC := binary.LittleEndian.Uint16(gbcpu.Regs.PC) + uint16(operand)
			binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
		}
		gbcpu.Jumped = true
		return 4
	}

	return 0
}

// CCF clears the carry flag
// Flags: -00C
func (gbcpu *GBCPU) CCF() {
	if gbcpu.Regs.getCarry() != 0 {
		gbcpu.Regs.clearCarry()
	} else {
		gbcpu.Regs.setCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
}

// DAA Decimal adjust A
// TODO Maybe rewrite this to be less esoteric
// Notoriously a pain to implement
// Reference: http://forums.nesdev.com/viewtopic.php?t=9088
func (gbcpu *GBCPU) DAA() {
	a := uint16(gbcpu.Regs.a)

	if gbcpu.Regs.getSubtract() == 0 {
		if gbcpu.Regs.getHalfCarry() == 1 || (a&0xF) > 9 {
			a += 0x06
		}

		if gbcpu.Regs.getCarry() == 1 || a > 0x9F {
			a += 0x60
		}
	} else {
		if gbcpu.Regs.getHalfCarry() == 1 {
			a = (a - 6) & 0xFF
		}

		if gbcpu.Regs.getCarry() == 1 {
			a -= 0x60
		}
	}

	gbcpu.Regs.clearHalfCarry()

	if (a & 0x100) == 0x100 {
		gbcpu.Regs.setCarry()
	}

	a &= 0xFF

	if a == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.a = byte(a)
}

// SCF sets the carry flag
// Flags: -001
func (gbcpu *GBCPU) SCF() {
	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.setCarry()
}

// CPL performs bitwise complement of A into a
// Flags: -11-
func (gbcpu *GBCPU) CPL() {
	gbcpu.Regs.a = ^gbcpu.Regs.a
	gbcpu.Regs.setSubtract()
	gbcpu.Regs.setHalfCarry()
}

// RET pops the top of the stack into the program counter
func (gbcpu *GBCPU) RET() {
	b1 := gbcpu.popByteFromStack()
	b2 := gbcpu.popByteFromStack()
	gbcpu.Regs.PC = []byte{b1, b2}
	gbcpu.Jumped = true
}

// RETI is the same as RET but also enables interrupts
func (gbcpu *GBCPU) RETI() {
	gbcpu.IME = 1
	gbcpu.RET()
}

// RETZ performs RET if Z is set
func (gbcpu *GBCPU) RETZ() int {
	if gbcpu.Regs.getZero() != 0 {
		gbcpu.RET()
		return 12
	}

	return 0
}

// RETNZ performs RET if Z is not set
func (gbcpu *GBCPU) RETNZ() int {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.RET()
		return 12
	}

	return 0
}

// RETC performs RET if C is set
func (gbcpu *GBCPU) RETC() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.RET()
		return 12
	}

	return 0
}

// RETNC performs RET if C is not set
func (gbcpu *GBCPU) RETNC() int {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.RET()
		return 12
	}

	return 0
}

// EI enables interrupts by setting IME to 1
// Enables AFTER instruction immediately after
func (gbcpu *GBCPU) EI() {
	gbcpu.EIReceived = true
}

// DI disables interrupts by setting IME to 0
func (gbcpu *GBCPU) DI() {
	gbcpu.IME = 0
	gbcpu.EIReceived = false
}

// HALT stops the CPU until an interrupt occurs
func (gbcpu *GBCPU) HALT() {
	// Save interrupt flag
	gbcpu.IFPreHalt = GbMMU.ReadByte(0xFF0F)

	// Halt CPU
	gbcpu.Halted = true
}

// CB executes a CB-prefixed instruction
func (gbcpu *GBCPU) CB() int {
	operand := gbcpu.getOperands(1)[0]
	gbcpu.InstrsCB[operand].Executor()
	return int(gbcpu.InstrsCB[operand].TCycles)
}

func (gbcpu *GBCPU) sliceToInt(slice []byte) uint16 {
	return binary.LittleEndian.Uint16(slice)
}

func (gbcpu *GBCPU) getOperands(number uint16) []byte {
	begin := gbcpu.sliceToInt(gbcpu.Regs.PC) + 1
	end := gbcpu.sliceToInt(gbcpu.Regs.PC) + (1 + number)

	operands := make([]byte, 2)
	copy(operands, GbMMU.Memory[begin:end])

	return operands
}
