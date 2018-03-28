package cpu

import (
	"../mmu"
)

// Good reference with some info Z80 heaven doesn't contain:
// http://www.devrs.com/gb/files/GBCPU_Instr.html

// GbMMU variable injection from main.go
// Prevents having to set MMU pointer as a field on the CPU struct
var GbMMU *mmu.GBMMU

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

// RRr -> e.g. RR B
// Rotates register right through CF
// Store old 0th bit into carry
// Old carry becomes new 7th bit
// Flags: Z00C
func (gbcpu *GBCPU) RRr(reg *byte) {
	carry := (*reg << 7)
	oldCarry := gbcpu.Regs.getCarry()
	*reg = ((*reg >> 1) | oldCarry)

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

// SWAPr -> e.g. SWAP B
// Swap nibbles of reg
// Flags: Z000
func (gbcpu *GBCPU) SWAPr(reg *byte) {
	*reg = (*reg << 4) | (*reg >> 4)
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

// RESnr -> e.g. RES 0,B
// Reset bit in register
// Flags: ----
func (gbcpu *GBCPU) RESnr(pos uint8, reg *byte) {
	*reg &^= (1 << pos)
}

// SETnr -> e.g. SET 0,B
// Set bit in register
// Flags: ----
func (gbcpu *GBCPU) SETnr(pos uint8, reg *byte) {
	*reg |= (1 << pos)
}
