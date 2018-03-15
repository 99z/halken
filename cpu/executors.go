// Contains functions to execute assembly instructions
package cpu

func (cpu *GBCPU) LDrr_n(reg1, reg2 *byte) {
	// get 16 bit operand

	// Little Endian, so reversed
	//*reg1 = operand[1]
	//*reg2 = operand[0]
}

func (cpu *GBCPU) LDrr_r(high, low, op *byte) {
	// Should this be filling first byte with zeros?
	*high = *op
	*low = 0x00
}

func (cpu *GBCPU) INCrr(high, low *byte) {
	rr := cpu.JoinBytes(*high, *low)
	rr++

	newHigh := byte(((rr >> 8) & 0xFF))
	newLow := byte(rr & 0xFF)
	
	// Set value of reg pointers to new values
	high = &newHigh
	low = &newLow
}

func (cpu *GBCPU) JPaa(pc *[2]byte) {
	//high := cart[pc+sizeof(byte)]
	//low := cart[pc+(sizeof(byte)*2)]
}

// Pull out into utilities file?
func (cpu *GBCPU) JoinBytes(high, low byte) uint16 {
	return uint16((high << 8) | (low & 0xFF))
}
