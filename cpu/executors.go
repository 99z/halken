package cpu

import (
	"encoding/binary"
	"fmt"

	"../mmu"
)

// GbMMU variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

// LDrr -> e.g. LD A,B
// Loads value in one register to another
func (gbcpu *GBCPU) LDrr(to, from *byte) {
	*to = *from
	fmt.Printf("A: %02X, B: %02X\n", *to, *from)
}

// LDrrnn -> e.g. LD BC,i16
// Loads 2 8-bit immediate operands into register pair
func (gbcpu *GBCPU) LDrrnn(reg1, reg2 *byte) {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.writePair(reg1, reg2, operands)
}

// LDSPHL -> e.g. LD SP,HL
// Loads bytes from register pair HL into SP
func (gbcpu *GBCPU) LDSPHL(reg1, reg2 *byte) {
	gbcpu.Regs.sp = []byte{*reg2, *reg1}
	gbcpu.Regs.Dump()
}

// LDrrSPs -> e.g. LD BC,SP+s8
// Loads value of SP + signed 8-bit value into register pair
func (gbcpu *GBCPU) LDrrSPs(reg1, reg2 *byte) {
	operand := gbcpu.getOperands(1)
	// Get SP as a 16-bit int, add operand to it
	val := gbcpu.sliceToInt(gbcpu.Regs.sp) + uint16(operand[0])
	gbcpu.Regs.writePairFromInt(reg1, reg2, val)

	// HC and C are a little weird for this instruction
	// https://stackoverflow.com/questions/5159603/gbz80-how-does-ld-hl-spe-affect-h-and-c-flags
	if val >= 0 {
		newHC := ((gbcpu.sliceToInt(gbcpu.Regs.sp) & 0xFF) + uint16(operand[0])) > 0xFF
		var newHCVal uint8
		if newHC {
			newHCVal = 1
		}
		newC := ((gbcpu.sliceToInt(gbcpu.Regs.sp) & 0xF) + (uint16(operand[0]) & 0xF)) > 0xF
		var newCVal uint8
		if newC {
			newCVal = 1
		}
		gbcpu.Regs.f |= (newHCVal << 4)
		gbcpu.Regs.f |= (newCVal << 5)
	} else {
		newHC := (val & 0xFF) <= (gbcpu.sliceToInt(gbcpu.Regs.sp) & 0xFF)
		var newHCVal uint8
		if newHC {
			newHCVal = 1
		}
		newC := (val & 0xF) <= (gbcpu.sliceToInt(gbcpu.Regs.sp) & 0xF)
		var newCVal uint8
		if newC {
			newCVal = 1
		}
		gbcpu.Regs.f |= (newHCVal << 4)
		gbcpu.Regs.f |= (newCVal << 5)
	}
}

// LDrn -> e.g. LD B,i8
// Loads 1 8-bit immediate operand into a register
func (gbcpu *GBCPU) LDrn(reg *byte) {
	operand := gbcpu.getOperands(1)
	*reg = operand[0]
	gbcpu.Regs.Dump()
}

func (gbcpu *GBCPU) LDrrr(reg1, reg2, op *byte) {
	gbcpu.Regs.writePair(reg1, reg2, []byte{*op, *op})
}

func (gbcpu *GBCPU) INCrr(reg1, reg2 *byte) {
	gbcpu.Regs.incrementPair(reg1, reg2, 1)
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

// Increment value at memory location reg1reg2
func (gbcpu *GBCPU) INCaa(reg1, reg2 *byte) {
	val := gbcpu.getValCartAddr(reg1, reg2, 2)
	intVal := gbcpu.sliceToInt(val)
	intVal++
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, intVal)
	*reg1 = buf[0]
	*reg2 = buf[1]
}

func (gbcpu *GBCPU) DECaa(reg1, reg2 *byte) {
	GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*reg2, *reg1})]--
}

func (gbcpu *GBCPU) INCSP() {
	newSP := gbcpu.sliceToInt(gbcpu.Regs.sp)
	newSP++
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, newSP)
}

func (gbcpu *GBCPU) DECSP() {
	newSP := gbcpu.sliceToInt(gbcpu.Regs.sp)
	newSP--
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, newSP)
}

func (gbcpu *GBCPU) INCrn(reg *byte) {
	operand := gbcpu.getOperands(1)
	*reg = operand[0]
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
	gbcpu.Regs.Dump()
}

func (gbcpu *GBCPU) DECrr(reg1, reg2 *byte) {
	*reg1--
	*reg2--
}

// RLCA performs 8-bit rotation to the left
// Rotated bit is copied to carry
// Flags: 000C
// GB opcodes list show Z being set to zero, but other sources disagree
// Z80 does not modify Z, other emulator sources do
// Reference: https://hax.iimarckus.org/topic/1617/
func (gbcpu *GBCPU) RLCA() {
	carry := gbcpu.Regs.a << 1
	gbcpu.Regs.a = (carry | carry>>7)

	if carry != 0x0 {
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
	gbcpu.Regs.Dump()
}

// RRCA performs 8-bit rotation to the right
// Rotated bit is copied to carry
// Flags: 000C
// GB opcodes list show Z being set to zero, but other sources disagree
// Z80 does not modify Z, other emulator sources do
func (gbcpu *GBCPU) RRCA() {
	carry := gbcpu.Regs.a >> 1
	gbcpu.Regs.a = (carry | carry<<7)

	if carry != 0x0 {
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
	gbcpu.Regs.Dump()
}

// RST pushes current PC + 3 onto stack
// MSB of PC is set to 0x00, LSB is set to argument
func (gbcpu *GBCPU) RST(imm byte) {
	val := gbcpu.sliceToInt(gbcpu.Regs.PC)
	val += 3
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, val)
	gbcpu.Regs.PC = []byte{imm, 0x00}
}

func (gbcpu *GBCPU) LDaaSP() {
	operands := gbcpu.getOperands(2)
	addr := gbcpu.sliceToInt(operands)
	GbMMU.Memory[addr] = gbcpu.Regs.sp[0]
	GbMMU.Memory[addr+1] = gbcpu.Regs.sp[1]
}

func (gbcpu *GBCPU) LDSPnn() {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.sp = operands
	gbcpu.Regs.Dump()
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

// ADDrn -> e.g. ADD A,i8
// Values of reg and i8 are added together
// Result is written into reg
// Flags: Z0HC
func (gbcpu *GBCPU) ADDrn(reg *byte) {
	oldVal := *reg
	operand := gbcpu.getOperands(1)[0]
	hc := (((*reg & 0xf) + (operand & 0xf)) & 0x10) == 0x10
	*reg = *reg + operand

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if *reg < oldVal {
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

// ADCrn -> e.g. ADC A,i8
// Values of operand and carry flag are added to reg
// Result is written into reg
// Flags: Z0HC
func (gbcpu *GBCPU) ADCrn(reg *byte) {
	gbcpu.Regs.Dump()
	oldVal := *reg
	operand := gbcpu.getOperands(1)[0]
	result := (operand + gbcpu.Regs.getCarry()) + gbcpu.Regs.a
	fmt.Printf("OPERAND: %v, CARRY: %v", operand, gbcpu.Regs.getCarry())
	hc := (((operand & 0xf) + (gbcpu.Regs.getCarry() & 0xf) + (gbcpu.Regs.a & 0xf)) & 0x10) == 0x10
	*reg = result

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if *reg < oldVal {
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
	gbcpu.Regs.Dump()
}

// ADCrr -> e.g. ADC A,B
// Values of reg1, reg2 and carry flag are added together
// Result is written into reg1
// Flags: Z0HC
func (gbcpu *GBCPU) ADCrr(reg1, reg2 *byte) {
	oldVal := *reg1
	result := *reg1 + *reg2 + gbcpu.Regs.getCarry()
	hc := (((*reg1 & 0xf) + (*reg2 & 0xf) + (gbcpu.Regs.getCarry() & 0xf)) & 0x10) == 0x10
	*reg1 = result

	// Don't think this is needed
	// Can just check for carry/hc with actual result
	// sum := (*reg1 & 0xF) + (*reg2 & 0xF)

	// Check for zero
	if *reg1 == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
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

func (gbcpu *GBCPU) ADCraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = *reg + val[0] + ((gbcpu.Regs.f >> 4) & 1)
}

func (gbcpu *GBCPU) ADDraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = *reg + val[0]
}

// ADDHLrr -> e.g. ADD HL,BC
// Values of HL and reg1reg2 are added together
// Result is written into HL
// Flags: -0HC
func (gbcpu *GBCPU) ADDHLrr(reg1, reg2 *byte) {
	hlInt := gbcpu.sliceToInt([]byte{gbcpu.Regs.h, gbcpu.Regs.l})
	rrInt := gbcpu.sliceToInt([]byte{*reg1, *reg2})
	sum := rrInt + hlInt
	hc := (((rrInt & 0xf) + (hlInt & 0xf)) & 0x10) == 0x10

	newH := gbcpu.Regs.h + *reg1
	newL := gbcpu.Regs.l + *reg2
	gbcpu.Regs.writePair(&gbcpu.Regs.h, &gbcpu.Regs.l, []byte{newH, newL})

	// Check for carry
	if sum < hlInt {
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
}

// ADDSPs -> e.g. ADD SP,s8
// Adds 8-bit value to SP
// Sets SP to new value
// Flags: 00HC
func (gbcpu *GBCPU) ADDSPs() {
	oldVal := gbcpu.sliceToInt(gbcpu.Regs.sp)
	operand := gbcpu.getOperands(1)[0]
	val := uint16(operand) + gbcpu.sliceToInt(gbcpu.Regs.sp)
	hc := (((uint16(operand) & 0xf) + (gbcpu.sliceToInt(gbcpu.Regs.sp) & 0xf)) & 0x10) == 0x10
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, val)

	// Check for carry
	if val < oldVal {
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

	gbcpu.Regs.clearZero()
	gbcpu.Regs.clearSubtract()
}

func (gbcpu *GBCPU) ADDrrSP(reg1, reg2 *byte) {
	gbcpu.Regs.writePair(reg1, reg2, []byte{gbcpu.Regs.sp[0], gbcpu.Regs.sp[1]})
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

func (gbcpu *GBCPU) ANDaa(a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	gbcpu.Regs.a &= val[0]
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
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

// ORaa -> e.g. OR (HL)
// Bitwise OR of byte at addr
// Flags: Z000
func (gbcpu *GBCPU) ORaa(a1, a2 *byte) {
	val := GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*a2, *a1})]
	gbcpu.Regs.a |= val
	if gbcpu.Regs.a == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}
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
	oldVal := gbcpu.Regs.a
	operand := gbcpu.getOperands(1)[0]
	result := gbcpu.Regs.a - operand
	hc := (((gbcpu.Regs.a & 0xf) - (operand & 0xf)) & 0x10) == 0x10
	gbcpu.Regs.a = result

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

	// Set subtract flag
	gbcpu.Regs.setSubtract()
	gbcpu.Regs.Dump()
}

// SUBr -> e.g. SUB B
// Value of reg is subtracted from A
// Result is written into reg
// Flags: Z1HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) SUBr(reg *byte) {
	oldVal := *reg
	hc := (((gbcpu.Regs.a & 0xf) - (*reg & 0xf)) & 0x10) == 0x10
	*reg = gbcpu.Regs.a - *reg

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if *reg > oldVal {
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

func (gbcpu *GBCPU) SUBaa(a1, a2 *byte) {
	GbMMU.Memory[gbcpu.sliceToInt([]byte{*a2, *a1})]--
}

// SBCrr -> e.g. SBC A,B
// Sum of reg2 and carry flag is subtracted from reg1
// Result is written into reg1
// Flags: Z1HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) SBCrr(reg1, reg2 *byte) {
	oldVal := *reg1
	sum := *reg2 + gbcpu.Regs.getCarry()
	result := *reg1 - sum
	hc := (((*reg1 & 0xf) - (sum & 0xf)) & 0x10) == 0x10
	*reg1 = result

	// Check for zero
	if *reg1 == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if *reg1 > oldVal {
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

// SBCrn -> e.g. SBC A,i8
// Sum of i8 and carry flag is subtracted from reg
// Result is written into reg
// Flags: Z1HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) SBCrn(reg *byte) {
	oldVal := *reg
	operand := gbcpu.getOperands(1)[0]
	sum := operand + gbcpu.Regs.getCarry()
	hc := (((*reg & 0xf) - (sum & 0xf)) & 0x10) == 0x10
	result := *reg - sum
	*reg = result

	// Check for zero
	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if *reg > oldVal {
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

func (gbcpu *GBCPU) SBCraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = *reg - val[0]
}

// CPr -> e.g. CP B
// Subtraction from accumulator that doesn't update it
// Only updates flags
// Flags: Z1HC
func (gbcpu *GBCPU) CPr(reg *byte) {
	oldVal := *reg
	sub := gbcpu.Regs.a - *reg
	hc := (((gbcpu.Regs.a & 0xf) - (*reg & 0xf)) & 0x10) == 0x10

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	// Check for carry
	if sub > oldVal {
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
	gbcpu.Regs.Dump()
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

func (gbcpu *GBCPU) CPaa(a1, a2 *byte) {
	// TODO
}

// PUSHrr
// Copies reg1reg2 into addr (SP)
func (gbcpu *GBCPU) PUSHrr(reg1, reg2 *byte) {
	//gbcpu.Regs.decrementSP(2)
	//addr := gbcpu.sliceToInt(gbcpu.Regs.sp)
	// GbMMU.Memory[addr] = *reg1
	// GbMMU.Memory[addr+1] = *reg2
	gbcpu.pushByteToStack(*reg1)
	gbcpu.pushByteToStack(*reg2)
	gbcpu.Regs.Dump()
	// fmt.Println(GbMMU.Memory[addr])
	// fmt.Println(GbMMU.Memory[addr+1])
}

// a1, s2 are 8-bit components of a 16-bit address
// Loads value at location a1a2 into reg
func (gbcpu *GBCPU) LDraa(reg, a1, a2 *byte) {
	// val := gbcpu.getValCartAddr(a1, a2, 1)
	// *reg = val[0]
	*reg = GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*a2, *a1})]
	fmt.Println(GbMMU.Memory[binary.LittleEndian.Uint16([]byte{*a2, *a1})])
	gbcpu.Regs.Dump()
}

func (gbcpu *GBCPU) LDaar(a1, a2, reg *byte) {
	addr := gbcpu.sliceToInt([]byte{*a2, *a1})

	GbMMU.Memory[addr] = *reg
}

func (gbcpu *GBCPU) LDnnr(reg *byte) {
	operands := gbcpu.getOperands(2)
	addr := gbcpu.sliceToInt(operands)
	GbMMU.Memory[addr] = *reg
}

func (gbcpu *GBCPU) LDrnn(reg *byte) {
	operands := gbcpu.getOperands(2)
	fmt.Println(operands)
	addr := gbcpu.sliceToInt(operands)
	*reg = GbMMU.Memory[addr]
	gbcpu.Regs.Dump()
}

// LDffrr sets value at (0xFF00+reg1) to reg2
func (gbcpu *GBCPU) LDffrr(reg1, reg2 *byte) {
	addr := make([]byte, 2)
	binary.LittleEndian.PutUint16(addr, 0xFF00+uint16(*reg1))
	GbMMU.WriteByte(addr, *reg2)
}

func (gbcpu *GBCPU) LDrffr(reg1, reg2 *byte) {
	*reg1 = GbMMU.Memory[0xFF00+uint16(*reg2)]
}

func (gbcpu *GBCPU) LDffnr(reg *byte) {
	operand := gbcpu.getOperands(1)
	addr := make([]byte, 2)
	binary.LittleEndian.PutUint16(addr, 0xFF00+uint16(operand[0]))
	GbMMU.WriteByte(addr, *reg)
}

func (gbcpu *GBCPU) LDrffn(reg *byte) {
	operand := gbcpu.getOperands(1)
	addr := make([]byte, 2)
	binary.LittleEndian.PutUint16(addr, 0xFF00+uint16(operand[0]))
	*reg = GbMMU.ReadByte(addr)
	fmt.Println(GbMMU.ReadByte(addr))
}

func (gbcpu *GBCPU) LDaan(reg1, reg2 *byte) {
	operand := gbcpu.getOperands(1)
	GbMMU.Memory[gbcpu.sliceToInt([]byte{*reg1, *reg2})] = operand[0]
}

func (gbcpu *GBCPU) LDDaaR(a1, a2, reg *byte) {
	GbMMU.Memory[gbcpu.sliceToInt([]byte{*a2, *a1})] = *reg
	*reg--
}

// Set value at address a1a2 to value in reg
// Increment reg
func (gbcpu *GBCPU) LDIaaR(a1, a2, reg *byte) {
	GbMMU.WriteByte([]byte{*a2, *a1}, *reg)
	gbcpu.Regs.incrementHL(1)
}

// Set value in reg to value at address a1a2
// Increment HL
func (gbcpu *GBCPU) LDIRaa(reg, a1, a2 *byte) {
	*reg = GbMMU.ReadByte([]byte{*a2, *a1})
	fmt.Printf("A: %02X\n", *reg)
	fmt.Printf("HL: %02X%02X\n", gbcpu.Regs.h, gbcpu.Regs.l)
	fmt.Printf("DE: %02X%02X\n", gbcpu.Regs.d, gbcpu.Regs.e)
	gbcpu.Regs.incrementHL(1)
}

func (gbcpu *GBCPU) JPaa() {
	jmpAddr := gbcpu.getOperands(2)
	gbcpu.Regs.PC = jmpAddr
	gbcpu.Jumped = true
}

func (gbcpu *GBCPU) JPHL(reg1, reg2 *byte) {
	gbcpu.Regs.PC = []byte{*reg2, *reg1}
	gbcpu.Jumped = true
}

func (gbcpu *GBCPU) JPZaa() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z != 0 {
		gbcpu.JPaa()
	}
}

func (gbcpu *GBCPU) JPNZaa() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.JPaa()
	}
}

func (gbcpu *GBCPU) JPCaa() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.JPaa()
	}
}

func (gbcpu *GBCPU) JPNCaa() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.JPaa()
	}
}

// CALLaa -> e.g. CALL $028B
// Pushes the addr at PC+3 to the stack
// Jumps to the address specified by next 2 bytes
func (gbcpu *GBCPU) CALLaa() {
	gbcpu.Regs.Dump()
	nextInstr := gbcpu.sliceToInt(gbcpu.Regs.PC) + 3
	nextInstrBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(nextInstrBytes, nextInstr)
	gbcpu.pushByteToStack(nextInstrBytes[1])
	gbcpu.pushByteToStack(nextInstrBytes[0])
	fmt.Printf("%v\n", gbcpu.getOperands(2))
	gbcpu.Regs.PC = gbcpu.getOperands(2)
	// binary.LittleEndian.PutUint16(gbcpu.Regs.sp, gbcpu.sliceToInt(gbcpu.Regs.PC)+1)

	gbcpu.Regs.Dump()
	gbcpu.Jumped = true
}

func (gbcpu *GBCPU) CALLZaa() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z != 0 {
		gbcpu.CALLaa()
	}
}

func (gbcpu *GBCPU) CALLNZaa() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.CALLaa()
	}
}

func (gbcpu *GBCPU) CALLCaa() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.CALLaa()
	}
}

func (gbcpu *GBCPU) CALLNCaa() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.CALLaa()
	}
}

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

func (gbcpu *GBCPU) JRZn() {
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
	}
}

// Jumps if zero flag = 0
func (gbcpu *GBCPU) JRNZn() {
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
	}
}

func (gbcpu *GBCPU) JRCn() {
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
	}
}

// JRNCn -> e.g. JR NC,FC
// Adds operand to PC
// Jumps to new addr if carry is not set
func (gbcpu *GBCPU) JRNCn() {
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
	}
}

func (gbcpu *GBCPU) CCF() {
	// TODO
	// Inverts the carry flag
}

func (gbcpu *GBCPU) DAA() {
	// TODO
	// Reference: http://forums.nesdev.com/viewtopic.php?t=9088
}

func (gbcpu *GBCPU) SCF() {
	// TODO
	// Set carry flag
}

// Bitwise complement of A
func (gbcpu *GBCPU) CPL() {
	gbcpu.Regs.a = ^gbcpu.Regs.a
}

// RET pops the top of the stack into the program counter
func (gbcpu *GBCPU) RET() {
	b1 := gbcpu.popByteFromStack()
	b2 := gbcpu.popByteFromStack()
	gbcpu.Regs.PC = []byte{b1, b2}
	gbcpu.Jumped = true
}

func (gbcpu *GBCPU) RETI() {
	gbcpu.RET()
	// TODO Set flag for interrupts enabled
}

func (gbcpu *GBCPU) RETZ() {
	gbcpu.Regs.Dump()
	if gbcpu.Regs.getZero() != 0 {
		gbcpu.RET()
	}
}

func (gbcpu *GBCPU) RETC() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c != 0 {
		gbcpu.RET()
	}
}

func (gbcpu *GBCPU) RETNC() {
	c := ((gbcpu.Regs.f >> 4) & 1)
	if c == 0 {
		gbcpu.RET()
	}
}

// RETNZ pops the top of the stack into the program counter
// if Z is not set
func (gbcpu *GBCPU) RETNZ() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z == 0 {
		gbcpu.RET()
	}
}

// POPrr copies 2 bytes at addr (SP) into the operand
// Flags: ZNHC
func (gbcpu *GBCPU) POPrr(reg1, reg2 *byte) {
	b1 := gbcpu.popByteFromStack()
	b2 := gbcpu.popByteFromStack()
	*reg1 = b2
	*reg2 = b1
	gbcpu.Regs.Dump()
}

// func (gbcpu *GBCPU) POPHL(reg1, reg2 *byte) {
// 	b1 := gbcpu.popByteFromStack()
// 	b2 := gbcpu.popByteFromStack()
// 	*reg1 = b1
// 	*reg2 = b2
// 	gbcpu.Regs.Dump()
// }

func (gbcpu *GBCPU) EI() {
	// TODO
	// Enables interrupts
}

func (gbcpu *GBCPU) DI() {
	gbcpu.Regs.Dump()
	// TODO
	// Disables interrupts
}

func (gbcpu *GBCPU) CB() {
	operand := gbcpu.getOperands(1)[0]
	fmt.Printf("Executing: %s\n", gbcpu.InstrsCB[operand].Mnemonic)
	gbcpu.InstrsCB[operand].Executor()
}

func (gbcpu *GBCPU) sliceToInt(slice []byte) uint16 {
	return binary.LittleEndian.Uint16(slice)
}

func (gbcpu *GBCPU) getOperands(number uint16) []byte {
	begin := gbcpu.sliceToInt(gbcpu.Regs.PC) + 1
	end := gbcpu.sliceToInt(gbcpu.Regs.PC) + (1 + number)

	return GbMMU.Memory[begin:end]
	// return []byte{args[1], args[0]}
}

func (gbcpu *GBCPU) getValCartAddr(a1, a2 *byte, number uint16) []byte {
	begin := binary.LittleEndian.Uint16([]byte{*a2, *a1})
	end := begin + (number - 1)
	return GbMMU.Memory[begin:end]
}
