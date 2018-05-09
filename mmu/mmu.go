// Package mmu contains the GBMMU struct which represents the Game Boy's memory
// and contains methods related to modifying memory
// TODO The WriteData and ReadData methods are pretty lean because I am
// modeling memory as a flat array of bytes. Need to double-check the special
// case read/writes, and determine if I/O should alter memory directly,
// since it is sort of an outlier in that it maintains an internal model
package mmu

import (
	"fmt"
	"io/ioutil"

	"../io"
)

// GBMMU represents the Game Boy's memory
// Reference http://gameboy.mongenel.com/dmg/asmmemmap.html
type GBMMU struct {
	// Array of bytes for contiguous memory access
	Memory [65536]byte
}

// GbIO variable injection from main.go
// Gives us access to instantiated IO struct's methods
var GbIO *io.GBIO

// InitMMU sets initial memory values
// These are actually populated by the Game Boy's bootstrap ROM
// Reference: http://bgb.bircd.org/pandocs.htm#powerupsequence
// TODO load/execute bootstrap ROM instead?
func (gbmmu *GBMMU) InitMMU() {
	// I/O register initial values after boot ROM
	// TODO Might want to just execute boot ROM instead, since
	// documentation online is sketchy about these
	gbmmu.Memory[0xFF0F] = 0xE1
	gbmmu.Memory[0xFF07] = 0xF8
	gbmmu.Memory[0xFF10] = 0x80
	gbmmu.Memory[0xFF11] = 0xBF
	gbmmu.Memory[0xFF12] = 0xF3
	gbmmu.Memory[0xFF14] = 0xBF
	gbmmu.Memory[0xFF16] = 0x3F
	gbmmu.Memory[0xFF19] = 0xBF
	gbmmu.Memory[0xFF1A] = 0x7F
	gbmmu.Memory[0xFF1B] = 0xFF
	gbmmu.Memory[0xFF1C] = 0x9F
	gbmmu.Memory[0xFF1E] = 0xBF
	gbmmu.Memory[0xFF20] = 0xFF
	gbmmu.Memory[0xFF23] = 0xBF
	gbmmu.Memory[0xFF24] = 0x77
	gbmmu.Memory[0xFF25] = 0xF3
	gbmmu.Memory[0xFF26] = 0xF1
	gbmmu.Memory[0xFF40] = 0x91
	// Not in pandocs
	gbmmu.Memory[0xFF41] = 0x85
	gbmmu.Memory[0xFF47] = 0xFC
	gbmmu.Memory[0xFF48] = 0xFF
	gbmmu.Memory[0xFF49] = 0xFF
}

// WriteData handles writing values to memory addresses
// We do this instead of directly setting the value at an index because
// some addresses are handled differently than others
// For example, when a ROM writes to 0xFF00, it is really telling our IO
// handler to select either buttons or dpad
func (gbmmu *GBMMU) WriteData(addr uint16, data byte) {
	if addr == 0xFF00 {
		GbIO.SetCol(data)
	} else if addr == 0xFF0F {
		// TODO What do writes here really do? Ignored or bit set?
		// gbmmu.Memory[0xFF0F] |= (1 << 0)
	} else if addr == 0xFF41 {
		// TODO Same as above
	} else if addr >= 0x0000 && addr <= 0x150 {
		// Don't allow writes to invalid locations
	} else if addr == 0xFF46 {
		spriteAddr := int(data) * 256
		for i := range gbmmu.Memory[0xFE00:0xFEA0] {
			gbmmu.Memory[0xFE00+i] = gbmmu.Memory[spriteAddr+i]
		}
	} else if addr == 0xFF07 {
		gbmmu.Memory[addr] += data
	} else {
		gbmmu.Memory[addr] = data
	}
}

// ReadData handles reading values from memory addresses
// Like WriteData, this is necessary because we may want to specify different
// things to return
// Again, for example, reading 0xFF00 should return the last input, which is
// not stored in any memory directly in my implementation
func (gbmmu *GBMMU) ReadData(addr uint16) byte {
	if addr == 0xFF00 {
		return GbIO.GetInput()
	}

	return gbmmu.Memory[addr]
}

// LoadCart reads cartridge ROM into memory
// Returns ROM as byte slice
func (gbmmu *GBMMU) LoadCart(path string) error {
	cartData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("MMU: loadCart(%s) failed: %s", path, err)
	}

	// Cartridge header layout
	// http://gbdev.gg8.se/wiki/articles/The_Cartridge_Header
	for i, v := range cartData[0x0134:0x0143] {
		gbmmu.Memory[0x0134+i] = v
	}
	gbmmu.Memory[0x0143] = cartData[0x0143]
	gbmmu.Memory[0x0147] = cartData[0x0147]
	gbmmu.Memory[0x0148] = cartData[0x0148]
	gbmmu.Memory[0x0149] = cartData[0x0149]
	for i, v := range cartData[0x0000:] {
		gbmmu.Memory[0x0000+i] = v
	}
	return nil
}
