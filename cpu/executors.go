// Contains functions to execute assembly instructions
package cpu

import (
	"../mmu"
	"encoding/binary"
)

// Variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

func (gbcpu *GBCPU) LDrrnn(reg1, reg2 *byte) {
	operands := gbcpu.getOperands(2)

	gbcpu.Regs.writePair(reg1, reg2, operands)
}

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

func (gbcpu *GBCPU) LDaaSP() {
	operands := gbcpu.getOperands(2)
	val := gbcpu.getValCartAddr(&operands[1], &operands[0], 2)
	gbcpu.Regs.sp = val
}

func (gbcpu *GBCPU) ADDrrrr(left1, left2, right1, right2 *byte) {
	*left1 = *right1
	*left2 = *right2
}

// a1, s2 are 8-bit components of a 16-bit address
// Loads value at location a1a2 into reg
func (gbcpu *GBCPU) LDraa(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = val[0]
}

func (gbcpu *GBCPU) LDaar(a1, a2, reg *byte) {
	val := gbcpu.getValCartAddr(a1, a2, 1)
	*reg = val[0]
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

func (gbcpu *GBCPU) DAA() {
	// TODO
	// Reference: http://forums.nesdev.com/viewtopic.php?t=9088
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