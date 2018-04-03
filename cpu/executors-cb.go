package cpu

import (
	"encoding/binary"
)

// Good reference with some info Z80 heaven doesn't contain:
// http://www.devrs.com/gb/files/GBCPU_Instr.html

// RLCr -> e.g. RLC B
// Performs 8-bit rotation to the left
// Rotated bit is copied to carry
// Flags: Z00C
func (gbcpu *GBCPU) RLCr(reg *byte) {
	carry := *reg << 1
	*reg = (carry | carry>>7)

	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if carry != 0x0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
}

// RLCHL -> e.g. RLC (HL)
// Performs 8-bit rotation to the left of value at address (HL)
// Rotated bit is copied to carry
// Flags: Z00C
func (gbcpu *GBCPU) RLCHL() {
	gbcpu.RLCr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// RRCr -> e.g. RRC B
// Performs 8-bit rotation to the right
// Rotated bit is copied to carry
// Flags: Z00C
func (gbcpu *GBCPU) RRCr(reg *byte) {
	carry := *reg >> 1
	*reg = (carry | carry<<7)

	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if carry != 0x0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
}

// RRCHL -> e.g. RRC (HL)
// Performs 8-bit rotation to the right of value at address (HL)
// Rotated bit is copied to carry
// Flags: Z00C
func (gbcpu *GBCPU) RRCHL() {
	gbcpu.RRCr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// RLr -> e.g. RL B
// Rotates register left through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: Z00C
func (gbcpu *GBCPU) RLr(reg *byte) {
	carry := (*reg >> 7)
	oldCarry := gbcpu.Regs.getCarry()
	*reg = ((*reg << 1) | oldCarry)

	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if carry != 0x0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()
	gbcpu.Regs.clearZero()
}

// RLHL -> e.g. RL (HL)
// Rotates value at address (HL) left through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: Z00C
func (gbcpu *GBCPU) RLHL() {
	gbcpu.RLr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// RRr -> e.g. RR B
// Rotates register right through CF
// Flags: Z00C
func (gbcpu *GBCPU) RRr(reg *byte) {
	msBit := 0
	if *reg&0x01 == 0x01 {
		msBit = 1
	}

	*reg = *reg >> 1

	if *reg == 0x0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if gbcpu.Regs.getCarry() != 0x0 {
		*reg ^= 0x80
	}

	if msBit != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.clearHalfCarry()

	gbcpu.Regs.Dump()
}

// RRHL -> e.g. RR (HL)
// Rotates value at address (HL) right through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: Z00C
func (gbcpu *GBCPU) RRHL() {
	gbcpu.RRr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// SLAr -> e.g. SLA B
// Shift reg left into carry
// Least significant bit of reg set to 0
// Flags: Z00C
func (gbcpu *GBCPU) SLAr(reg *byte) {
	lsBit := 0

	if *reg&0x80 == 0x80 {
		lsBit = 1
	}

	*reg = *reg << 1

	if *reg == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if lsBit != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}
}

// SLAHL -> e.g. SLA (HL)
// Shift value at addr (HL) left into carry
// Least significant bit of reg set to 0
// Flags: Z00C
func (gbcpu *GBCPU) SLAHL() {
	gbcpu.SLAr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// SRAr -> e.g. SRA B
// Shift reg right into carry
// Most significant bit of reg is unaffected
// Flags: Z000
func (gbcpu *GBCPU) SRAr(reg *byte) {
	*reg = (*reg >> 1) | (*reg & 0x80)

	if *reg == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearCarry()
}

// SRAHL -> e.g. SRA (HL)
// Shift value at addr (HL) right into carry
// Most significant bit of reg is unaffected
// Flags: Z000
func (gbcpu *GBCPU) SRAHL() {
	gbcpu.SRAr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// SWAPr -> e.g. SWAP B
// Swap nibbles of reg
// Flags: Z000
func (gbcpu *GBCPU) SWAPr(reg *byte) {
	*reg = (*reg << 4) | (*reg >> 4)
}

// SWAPHL -> e.g. SWAP (HL)
// Swap nibbles of value at addr (HL)
// Flags: Z000
func (gbcpu *GBCPU) SWAPHL() {
	gbcpu.SWAPr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// SRLr -> e.g. SRL B
// Shift reg right into carry
// Most significant bit of reg is set to 0
// Flags: Z00C
func (gbcpu *GBCPU) SRLr(reg *byte) {
	msBit := 0

	if *reg&0x01 == 0x01 {
		msBit = 1
	}

	*reg = *reg >> 1

	if *reg == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	if msBit != 0 {
		gbcpu.Regs.setCarry()
	} else {
		gbcpu.Regs.clearCarry()
	}

	gbcpu.Regs.Dump()
}

// SRLHL -> e.g. SRL (HL)
// Shift value at addr (HL) right into carry
// Most significant bit of reg is set to 0
// Flags: Z00C
func (gbcpu *GBCPU) SRLHL() {
	gbcpu.SRLr(&GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// BITnr -> e.g. BIT 0,B
// Test bit at position in reg
// Flags: Z01-
func (gbcpu *GBCPU) BITnr(pos uint8, reg *byte) {
	bitVal := *reg & (1 << pos)

	if bitVal == 0 {
		gbcpu.Regs.setZero()
	} else {
		gbcpu.Regs.clearZero()
	}

	gbcpu.Regs.clearSubtract()
	gbcpu.Regs.setHalfCarry()
}

// BITHL -> e.g. BIT 0,(HL)
// Test bit at position in value at addr (HL)
// Flags: Z01-
func (gbcpu *GBCPU) BITHL(pos uint8) {
	gbcpu.BITnr(pos, &GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// RESnr -> e.g. RES 0,B
// Reset bit in register
// Flags: ----
func (gbcpu *GBCPU) RESnr(pos uint8, reg *byte) {
	*reg &^= (1 << pos)
}

// RESHL -> e.g. RES 0,(HL)
// Reset bit in value at addr (HL)
// Flags: ----
func (gbcpu *GBCPU) RESHL(pos uint8) {
	gbcpu.RESnr(pos, &GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}

// SETnr -> e.g. SET 0,B
// Set bit in register
// Flags: ----
func (gbcpu *GBCPU) SETnr(pos uint8, reg *byte) {
	*reg |= (1 << pos)
}

// SETHL -> e.g. SET 0,(HL)
// Set bit in value at addr (HL)
// Flags: ----
func (gbcpu *GBCPU) SETHL(pos uint8) {
	gbcpu.SETnr(pos, &GbMMU.Memory[binary.LittleEndian.Uint16([]byte{gbcpu.Regs.h, gbcpu.Regs.l})])
}
