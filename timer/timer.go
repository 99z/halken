// Package timer contains functions to modify the GB's timer registers
// The timer is required to handle things like how quick tetromino should drop,
// and can provide pseudo-random numbers
// Realized timer implementation was necessary when Tetris would play correctly,
// but only would get square tetromino. This is because it uses the divider
// timer register to get a "random" block based on its value
package timer

import (
	"../mmu"
)

// GBTimer keeps time separately from LC
// Allows us to only modify memory values when we know we have to
type GBTimer struct {
	main int
	sub  int
	div  int
}

// GbMMU injection from main.go
// This provides direct memory access
// May refactor to instead return values and write to memory in mmu.go
var GbMMU *mmu.GBMMU

// Increment increments the timer and writes to the divider
// at a rate of 1/16th the base clock speed
func (gbtimer *GBTimer) Increment(cycles int) {
	// Increment by cycles of last opcode
	gbtimer.sub += cycles

	if gbtimer.sub >= 4 {
		gbtimer.main++
		gbtimer.sub -= 4

		gbtimer.div++
		if gbtimer.div == 16 {
			GbMMU.Memory[0xFF04] = (GbMMU.Memory[0xFF04] + 1) & 255
			gbtimer.div = 0
		}
	}

	// Check if we need to step the timer depending on selected
	// input clock speed
	gbtimer.checkStep()
}

// checkStep checks the Timer Control register
// steps the CPU if the timer > the selected clock speed
// This allows for games to run at different speeds like Tetris blocks falling
func (gbtimer *GBTimer) checkStep() {
	selectedClock := 0
	if GbMMU.Memory[0xFF07]&4 != 0 {
		switch GbMMU.Memory[0xFF07] & 3 {
		case 0:
			selectedClock = 64
		case 1:
			selectedClock = 1
		case 2:
			selectedClock = 4
		case 3:
			selectedClock = 16
		}

		if gbtimer.main >= selectedClock {
			gbtimer.step()
		}
	}
}

// Step increments the counter, sets it to modulo if it overflows,
// and sets the bit to call a timer interrupt if overflow happened
func (gbtimer *GBTimer) step() {
	gbtimer.main = 0
	prevCount := GbMMU.Memory[0xFF05]
	GbMMU.Memory[0xFF05]++

	if GbMMU.Memory[0xFF05] < prevCount {
		GbMMU.Memory[0xFF05] = GbMMU.Memory[0xFF06]

		// Set timer interrupt bit
		// TODO This makes Dr. Mario work
		GbMMU.Memory[0xFF0F] ^= 3

		// But this makes Flipull work??
		// GbMMU.Memory[0xFF0F] |= (1 << 2)
	}
}
