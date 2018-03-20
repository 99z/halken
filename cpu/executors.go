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
func (gbcpu *GBCPU) LDSPHL(reg1, reg2 *byte) {
	gbcpu.Regs.sp = []byte{*reg1, *reg2}
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
}

func (gbcpu *GBCPU) LDrrr(reg1, reg2, op *byte) {
	gbcpu.Regs.writePair(reg1, reg2, []byte{*op, *op})
}

func (gbcpu *GBCPU) INCrr(reg1, reg2 *byte) {
	*reg1++
	*reg2++
}

// INCr -> e.g. INC B
// Adds 1 to register, sets new value
// Flags: Z0H-
func (gbcpu *GBCPU) INCr(reg *byte) {
	*reg++
	sum := (*reg & 0xf) + (1 & 0xf)

	// Check for zero
	if sum == 0x0 {
		gbcpu.Regs.f |= (1 << 7)
	} else {
		gbcpu.Regs.f |= (0 << 7)
	}

	// Check for half carry
	if sum&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.f |= (1 << 5)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.f |= (0 << 5)
	}

	// Set subtract flag to zero
	gbcpu.Regs.f |= (0 << 6)
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
	val := gbcpu.getValCartAddr(reg1, reg2, 2)
	intVal := gbcpu.sliceToInt(val)
	intVal--
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, intVal)
	*reg1 = buf[0]
	*reg2 = buf[1]
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
	*reg--
	sub := (*reg & 0xf) - (1 & 0xf)

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for half carry
	if sub&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag
	gbcpu.Regs.setSubtract(1)
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
	gbcpu.Regs.setCarry(carry)
	gbcpu.Regs.a = (carry | carry>>7)

	if carry != 0x0 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setZero(0)
}

// RLA rotates register A left through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: 000C
func (gbcpu *GBCPU) RLA() {
	carry := (gbcpu.Regs.a >> 7)
	oldCarry := gbcpu.Regs.getCarry()
	gbcpu.Regs.a = ((gbcpu.Regs.a << 1) | oldCarry)

	if carry != 0x0 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setZero(0)
}

// RRCA performs 8-bit rotation to the right
// Rotated bit is copied to carry
// Flags: 000C
// GB opcodes list show Z being set to zero, but other sources disagree
// Z80 does not modify Z, other emulator sources do
func (gbcpu *GBCPU) RRCA() {
	carry := gbcpu.Regs.a >> 1
	gbcpu.Regs.setCarry(carry)
	gbcpu.Regs.a = (carry | carry<<7)

	if carry != 0x0 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setZero(0)
}

// RRA rotates register A right through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: 000C
func (gbcpu *GBCPU) RRA() {
	carry := (gbcpu.Regs.a << 7)
	oldCarry := gbcpu.Regs.getCarry()
	gbcpu.Regs.a = ((gbcpu.Regs.a >> 1) | oldCarry)

	if carry != 0x0 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setZero(0)
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
	val := gbcpu.getValCartAddr(&operands[1], &operands[0], 2)
	gbcpu.Regs.sp = val
}

func (gbcpu *GBCPU) LDSPnn() {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.sp = operands
}

// ADDrr -> e.g. ADD A,B
// Values of reg1 and reg2 are added together
// Result is written into reg1
// Flags: Z0HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) ADDrr(reg1, reg2 *byte) {
	result := *reg1 + *reg2
	*reg1 = result

	sum := (*reg1 & 0xf) + (*reg2 & 0xf)

	// Check for zero
	if sum == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for carry
	if sum > 255 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	// Check for half carry
	if sum&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract(0)
}

func (gbcpu *GBCPU) ADDrn(reg *byte) {
	operands := gbcpu.getOperands(1)
	*reg = *reg + operands[0]
}

func (gbcpu *GBCPU) ADCrn(reg *byte) {
	operands := gbcpu.getOperands(1)
	result := *reg + operands[0] + ((gbcpu.Regs.f >> 4) & 1)
	*reg = result
}

// ADCrr -> e.g. ADC A,B
// Values of reg1, reg2 and carry flag are added together
// Result is written into reg1
// Flags: Z0HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) ADCrr(reg1, reg2 *byte) {
	result := *reg1 + *reg2 + ((gbcpu.Regs.f >> 4) & 1)
	*reg1 = result

	sum := (*reg1 & 0xf) + (*reg2 & 0xf)

	// Check for zero
	if sum == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for carry
	if sum > 255 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	// Check for half carry
	if sum&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract(0)
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
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) ADDHLrr(reg1, reg2 *byte) {
	hlInt := gbcpu.sliceToInt([]byte{gbcpu.Regs.h, gbcpu.Regs.l})
	rrInt := gbcpu.sliceToInt([]byte{*reg1, *reg2})
	sum := rrInt + hlInt

	newH := gbcpu.Regs.h + *reg1
	newL := gbcpu.Regs.l + *reg2
	gbcpu.Regs.writePair(&gbcpu.Regs.h, &gbcpu.Regs.l, []byte{newH, newL})

	// Check for half carry
	if sum&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Check for carry
	if sum > 65535 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}
}

func (gbcpu *GBCPU) ADDSPs() {
	operands := gbcpu.getOperands(1)
	val := uint16(operands[0]) + gbcpu.sliceToInt(gbcpu.Regs.sp)
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, val)
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
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Set flags
	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(1)
	gbcpu.Regs.setCarry(0)
}

func (gbcpu *GBCPU) ANDn() {
	operands := gbcpu.getOperands(1)
	gbcpu.Regs.a &= operands[0]
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
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
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Set flags
	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setCarry(0)
}

func (gbcpu *GBCPU) ORn() {
	operand := gbcpu.getOperands(1)
	gbcpu.Regs.a |= operand[0]
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
}

func (gbcpu *GBCPU) ORaa(a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	gbcpu.Regs.a |= val[0]
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
}

// XORr -> e.g. XOR B
// Bitwise XOR of reg into A
// Flags: Z000
func (gbcpu *GBCPU) XORr(reg *byte) {
	gbcpu.Regs.a ^= *reg

	// Check for zero
	if gbcpu.Regs.a == 0x00 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Set flags
	gbcpu.Regs.setSubtract(0)
	gbcpu.Regs.setHalfCarry(0)
	gbcpu.Regs.setCarry(0)
}

func (gbcpu *GBCPU) XORn() {
	operand := gbcpu.getOperands(1)
	gbcpu.Regs.a ^= operand[0]
	if gbcpu.Regs.a == 0 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
}

func (gbcpu *GBCPU) XORaa(a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	gbcpu.Regs.a ^= val[0]
	if gbcpu.Regs.a == 0 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reseto zero flag
	}
}

func (gbcpu *GBCPU) SUBn() {
	operand := gbcpu.getOperands(1)
	result := gbcpu.Regs.a - operand[0]
	gbcpu.Regs.a = result
}

// SUBr -> e.g. SUB B
// Value of reg is subtracted from A
// Result is written into reg
// Flags: Z1HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) SUBr(reg *byte) {
	*reg = gbcpu.Regs.a - *reg

	sub := (gbcpu.Regs.a & 0xf) - (*reg & 0xf)

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for carry
	if sub > 255 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	// Check for half carry
	if sub&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract(1)
}

func (gbcpu *GBCPU) SUBaa(a1, a2 *byte) {
	GbMMU.Cart.MBC[gbcpu.sliceToInt([]byte{*a1, *a2})]--
}

// SBCrr -> e.g. SBC A,B
// Sum of reg2 and carry flag is subtracted from reg1
// Result is written into reg1
// Flags: Z1HC
// TODO Double-check this carry calculation
func (gbcpu *GBCPU) SBCrr(reg1, reg2 *byte) {
	sum := *reg2 + ((gbcpu.Regs.f >> 4) & 1)
	result := *reg1 - sum
	*reg1 = result

	sub := (*reg1 & 0xf) - (sum & 0xf)

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for carry
	if sub > 255 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	// Check for half carry
	if sub&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract(1)
}

func (gbcpu *GBCPU) SBCrn(reg *byte) {
	operands := gbcpu.getOperands(1)
	result := *reg - operands[0] - ((gbcpu.Regs.f >> 4) & 1)
	*reg = result
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
	sub := (gbcpu.Regs.a & 0xf) - (*reg & 0xf)

	// Check for zero
	if sub == 0x0 {
		gbcpu.Regs.setZero(1)
	} else {
		gbcpu.Regs.setZero(0)
	}

	// Check for carry
	if sub > 255 {
		gbcpu.Regs.setCarry(1)
	} else {
		gbcpu.Regs.setCarry(0)
	}

	// Check for half carry
	if sub&0x10 == 0x10 {
		// Half-carry occurred
		gbcpu.Regs.setHalfCarry(1)
	} else {
		// Half-carry did not occur
		gbcpu.Regs.setHalfCarry(0)
	}

	// Set subtract flag to zero
	gbcpu.Regs.setSubtract(1)
}

func (gbcpu *GBCPU) CPn() {
	// TODO
}

func (gbcpu *GBCPU) CPaa(a1, a2 *byte) {
	// TODO
}

// PUSHRR copies reg1reg2 into SP
// Increments SP by 2
func (gbcpu *GBCPU) PUSHrr(reg1, reg2 *byte) {
	gbcpu.Regs.sp = []byte{*reg1, *reg2}
	gbcpu.Regs.incrementSP(2)
}

// a1, s2 are 8-bit components of a 16-bit address
// Loads value at location a1a2 into reg
func (gbcpu *GBCPU) LDraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = val[0]
}

func (gbcpu *GBCPU) LDaar(a1, a2, reg *byte) {
	addr := gbcpu.sliceToInt([]byte{*a1, *a2})

	GbMMU.Cart.MBC[addr] = *reg
}

func (gbcpu *GBCPU) LDnnr(reg *byte) {
	operands := gbcpu.getOperands(2)
	addr := gbcpu.sliceToInt(operands)
	GbMMU.Cart.MBC[addr] = *reg
}

func (gbcpu *GBCPU) LDrnn(reg *byte) {
	operands := gbcpu.getOperands(2)
	addr := gbcpu.sliceToInt(operands)
	*reg = GbMMU.Cart.MBC[addr]
}

// LDffrr sets value at (0xFF00+reg1) to reg2
func (gbcpu *GBCPU) LDffrr(reg1, reg2 *byte) {
	GbMMU.Cart.MBC[0xFF00+uint16(*reg1)] = *reg2
}

func (gbcpu *GBCPU) LDrffr(reg1, reg2 *byte) {
	*reg1 = GbMMU.Cart.MBC[0xFF00+uint16(*reg2)]
}

func (gbcpu *GBCPU) LDrffn(reg *byte) {
	operand := gbcpu.getOperands(1)
	*reg = GbMMU.Cart.MBC[0xFF0+uint16(operand[0])]
}

func (gbcpu *GBCPU) LDaan(reg1, reg2 *byte) {
	operand := gbcpu.getOperands(1)
	gbcpu.Regs.writePair(reg1, reg2, []byte{operand[0], operand[0]})
}

func (gbcpu *GBCPU) LDDaaR(a1, a2, reg *byte) {
	GbMMU.Cart.MBC[binary.LittleEndian.Uint16([]byte{*a1, *a2})] = *reg
	*reg--
}

// Set value at address a1a2 to value in reg
// Increment reg
func (gbcpu *GBCPU) LDIaaR(a1, a2, reg *byte) {
	GbMMU.Cart.MBC[binary.LittleEndian.Uint16([]byte{*a1, *a2})] = *reg
	*reg++
}

// Set value in reg to  value at address a1a2
// Increment reg
func (gbcpu *GBCPU) LDIRaa(reg, a1, a2 *byte) {
	*reg = GbMMU.Cart.MBC[binary.LittleEndian.Uint16([]byte{*a1, *a2})]
	*reg++
}

func (gbcpu *GBCPU) JPaa() {
	jmpAddr := gbcpu.getOperands(2)
	gbcpu.Regs.PC = jmpAddr
}

func (gbcpu *GBCPU) JPrr(reg1, reg2 *byte) {
	gbcpu.Regs.PC = []byte{*reg1, *reg2}
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

func (gbcpu *GBCPU) CALLaa() {
	nextInstr := gbcpu.sliceToInt(gbcpu.Regs.PC) + 3
	begin := nextInstr + 1
	end := nextInstr + 3
	gbcpu.Regs.sp = GbMMU.Cart.MBC[begin:end]
	gbcpu.Regs.PC = gbcpu.getOperands(2)
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
	operand := gbcpu.getOperands(1)
	newPC := gbcpu.sliceToInt(gbcpu.Regs.PC) + uint16(operand[0])
	binary.LittleEndian.PutUint16(gbcpu.Regs.PC, newPC)
}

func (gbcpu *GBCPU) JRZn() {
	// TODO
}

func (gbcpu *GBCPU) JRNZn() {
	// TODO
}

func (gbcpu *GBCPU) JRCn() {
	// TODO
}

func (gbcpu *GBCPU) JRNCn() {
	// TODO
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
	gbcpu.Regs.PC = gbcpu.Regs.sp
	gbcpu.Regs.incrementSP(1)
}

func (gbcpu *GBCPU) RETI() {
	gbcpu.RET()
	// TODO Set flag for interrupts enabled
}

func (gbcpu *GBCPU) RETZ() {
	z := ((gbcpu.Regs.f >> 7) & 1)
	if z != 0 {
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

// POPrr pops 2 bytes from SP into operand
// Increments SP by 2
func (gbcpu *GBCPU) POPrr(reg1, reg2 *byte) {
	*reg1 = gbcpu.Regs.sp[0]
	*reg2 = gbcpu.Regs.sp[1]
	gbcpu.Regs.incrementSP(2)
}

func (gbcpu *GBCPU) EI() {
	// TODO
	// Enables interrupts
}

func (gbcpu *GBCPU) DI() {
	// TODO
	// Disables interrupts
}

func (gbcpu *GBCPU) sliceToInt(slice []byte) uint16 {
	return binary.LittleEndian.Uint16(slice)
}

func (gbcpu *GBCPU) getOperands(number uint16) []byte {
	begin := gbcpu.sliceToInt(gbcpu.Regs.PC) + 1
	end := gbcpu.sliceToInt(gbcpu.Regs.PC) + (1 + number)

	return GbMMU.Cart.MBC[begin:end]
}

func (gbcpu *GBCPU) getValCartAddr(a1, a2 *byte, number uint16) []byte {
	begin := binary.LittleEndian.Uint16([]byte{*a1, *a2})
	end := begin + (number - 1)
	return GbMMU.Cart.MBC[begin:end]
}
