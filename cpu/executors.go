// Contains functions to execute assembly instructions
package cpu

import (
	"../mmu"
	"encoding/binary"
)

// Variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

func (gbcpu *GBCPU) LDrr_nn(reg1, reg2 *byte) {
	operands := gbcpu.getOperands(2)

	// Little Endian, so reversed
	gbcpu.Regs.writePair(reg1, reg2, operands)
}

func (gbcpu *GBCPU) LDr_n(reg *byte) {
	operand := gbcpu.getOperands(1)
	*reg = operand[0]
}

func (gbcpu *GBCPU) LDrr_r(reg1, reg2, op *byte) {
	gbcpu.Regs.writePair(reg1, reg2, []byte{*op, *op})
}

func (gbcpu *GBCPU) INCrr(reg1, reg2 *byte) {
	*reg1++
	*reg2++
}

func (gbcpu *GBCPU) INCr(reg *byte) {
	*reg++
}

func (gbcpu *GBCPU) DECr(reg *byte) {
	*reg--
}

func (gbcpu *GBCPU) DECrr(reg1, reg2 *byte) {
	*reg1--
	*reg2--
}

func (gbcpu *GBCPU) RLCA(reg *byte) {
	*reg = *reg << 8
}

func (gbcpu *GBCPU) LDnn_SP() {
	operands := gbcpu.getOperands(2)
	gbcpu.Regs.sp = operands
}

func (gbcpu *GBCPU) ADDrr_rr(left1, left2, right1, right2 *byte) {
	*left1 = *right1
	*left2 = *right2
}

// a1, s2 are 8-bit components of a 16-bit address
// Loads value at location a1a2 into reg
func (gbcpu *GBCPU) LDr_rr(reg, a1, a2 *byte) {
	val := gbcpu.getValCartAddr(a1, a2)
	*reg = val
}

func (gbcpu *GBCPU) JPaa() {
	jmpAddr := gbcpu.getOperands(2)
	gbcpu.Regs.PC = jmpAddr
}

func (gbcpu *GBCPU) sliceToInt(slice []byte) uint16 {
	return binary.LittleEndian.Uint16(slice)
}

func (gbcpu *GBCPU) getOperands(number uint16) []byte {
	begin := gbcpu.sliceToInt(gbcpu.Regs.PC) + 1
	end := gbcpu.sliceToInt(gbcpu.Regs.PC) + (1 + number)
	
	return GbMMU.Cart.MBC[begin:end]
}

func (gbcpu *GBCPU) getValCartAddr(a1, a2 *byte) byte {
	return GbMMU.Cart.MBC[binary.LittleEndian.Uint16([]byte{*a1, *a2})]
}