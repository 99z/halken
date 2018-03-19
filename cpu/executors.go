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

func (gbcpu *GBCPU) INCr(reg *byte) {
	*reg++
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

func (gbcpu *GBCPU) DECr(reg *byte) {
	*reg--
}

func (gbcpu *GBCPU) DECrr(reg1, reg2 *byte) {
	*reg1--
	*reg2--
}

func (gbcpu *GBCPU) RLCA() {
	gbcpu.Regs.a = gbcpu.Regs.a << 8
}

func (gbcpu *GBCPU) RLA() {
	gbcpu.Regs.a = gbcpu.Regs.a << 9
}

func (gbcpu *GBCPU) RRCA() {
	gbcpu.Regs.a = gbcpu.Regs.a >> 8
}

func (gbcpu *GBCPU) RRA() {
	gbcpu.Regs.a = gbcpu.Regs.a >> 9
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

func (gbcpu *GBCPU) ADDrr(reg1, reg2 *byte) {
	result := *reg1 + *reg2
	*reg1 = result
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

// Same as ADD, but also adds the carry bit
func (gbcpu *GBCPU) ADCrr(reg1, reg2 *byte) {
	result := *reg1 + *reg2 + ((gbcpu.Regs.f >> 4) & 1)
	*reg1 = result
}

func (gbcpu *GBCPU) ADCraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = *reg + val[0] + ((gbcpu.Regs.f >> 4) & 1)
}

func (gbcpu *GBCPU) ADDraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = *reg + val[0]
}

func (gbcpu *GBCPU) ADDrrrr(left1, left2, right1, right2 *byte) {
	*left1 = *right1
	*left2 = *right2
}

func (gbcpu *GBCPU) ADDSPs() {
	operands := gbcpu.getOperands(1)
	val := uint16(operands[0]) + gbcpu.sliceToInt(gbcpu.Regs.sp)
	binary.LittleEndian.PutUint16(gbcpu.Regs.sp, val)
}

func (gbcpu *GBCPU) ADDrrSP(reg1, reg2 *byte) {
	gbcpu.Regs.writePair(reg1, reg2, []byte{gbcpu.Regs.sp[0], gbcpu.Regs.sp[1]})
}

func (gbcpu *GBCPU) ANDr(reg *byte) {
	gbcpu.Regs.a &= *reg
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
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

func (gbcpu *GBCPU) ORr(reg *byte) {
	gbcpu.Regs.a |= *reg
	if gbcpu.Regs.a == 0x00 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
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

func (gbcpu *GBCPU) XORr(reg *byte) {
	gbcpu.Regs.a ^= *reg
	if gbcpu.Regs.a == 0 {
		// TODO
		// Set zero flag
	} else {
		// TODO
		// Reset zero flag
	}
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

func (gbcpu *GBCPU) SUBr(reg *byte) {
	*reg = gbcpu.Regs.a - *reg
}

func (gbcpu *GBCPU) SUBaa(a1, a2 *byte) {
	GbMMU.Cart.MBC[gbcpu.sliceToInt([]byte{*a1, *a2})]--
}

func (gbcpu *GBCPU) SBCrr(reg1, reg2 *byte) {
	result := *reg1 - *reg2 - ((gbcpu.Regs.f >> 4) & 1)
	*reg1 = result
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

// Subtraction from accumulator that doesn't update it
// Only updates flags
func (gbcpu *GBCPU) CPr(reg *byte) {
	// TODO
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
