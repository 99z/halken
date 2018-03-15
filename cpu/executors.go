// Contains functions to execute assembly instructions
package cpu

import (
	"../mmu"
	"fmt"
	"encoding/binary"
)

// Variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

func (gbcpu *GBCPU) LDrr_n(reg1, reg2 *byte) {
	// get 16 bit operand

	// Little Endian, so reversed
	//*reg1 = operand[1]
	//*reg2 = operand[0]
}

func (gbcpu *GBCPU) LDrr_r(high, low, op *byte) {
	// Should this be filling first byte with zeros?
	*high = *op
	*low = 0x00
}

func (gbcpu *GBCPU) INCrr(high, low *byte) {
	rr := gbcpu.JoinBytes(*high, *low)
	rr++

	newHigh := byte(((rr >> 8) & 0xFF))
	newLow := byte(rr & 0xFF)
	
	// Set value of reg pointers to new values
	high = &newHigh
	low = &newLow
}

func (gbcpu *GBCPU) JPaa() {
	jmpAddr := GbMMU.Cart.MBC[gbcpu.regs.pc+1:gbcpu.regs.pc+3]
	gbcpu.regs.pc = binary.LittleEndian.Uint16(jmpAddr)
}

// Pull out into utilities file?
func (gbcpu *GBCPU) JoinBytes(high, low byte) uint16 {
	return uint16((high << 8) | (low & 0xFF))
}
