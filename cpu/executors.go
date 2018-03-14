// Contains functions to execute assembly instructions
package cpu

func (cpu *GBCPU) LDrr_n(reg1, reg2 *byte) {
	// get 16 bit operand

	// Little Endian, so reversed
	*reg1 = operand[1]
	*reg2 = operand[0]
}

func (cpu *GBCPU) LDrr_r(high, low, op *byte) {
	// Should this be filling first byte with zeros?
	*high = *op
	*low = 0x00
}

func (cpu *GBCPU) INCrr(high, low *byte) {
	rr := cpu.JoinBytes(high, low)
	rr++

	newHigh := ((rr >> 8) & 0xFF)
	newLow := rr & 0xff
}

func (cpu *GBCPU) JoinBytes(high, low byte) {
	return (high << 8) | (low & 0xFF)
}
