package mmu

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// Reference http://gameboy.mongenel.com/dmg/asmmemmap.html
type GBMMU struct {
	// Array of bytes for contiguous memory access
	Memory [0xFFFF]byte
	areas  map[string][]uint16
}

// Initial register values
// Reference: http://bgb.bircd.org/pandocs.htm#powerupsequence
func (gbmmu *GBMMU) InitMMU() {
	gbmmu.areas = make(map[string][]uint16)
	gbmmu.areas["vectors"] = []uint16{0x0000, 0x00FF}
	gbmmu.areas["cartHeader"] = []uint16{0x0100, 0x014F}
	gbmmu.areas["cartBankFixed"] = []uint16{0x0150, 0x3FFF}
	gbmmu.areas["cartBankSwitchable"] = []uint16{0x4000, 0x7FFF}
	gbmmu.areas["characterRAM"] = []uint16{0x8000, 0x97FF}
	gbmmu.areas["bgMapData1"] = []uint16{0x9800, 0x9BFF}
	gbmmu.areas["bgMapData2"] = []uint16{0x9C00, 0x9FFF}
	gbmmu.areas["cartRAM"] = []uint16{0xA000, 0xBFFF}
	gbmmu.areas["workRAM"] = []uint16{0xC000, 0xDFFF}
	gbmmu.areas["echoRAM"] = []uint16{0xE000, 0xFDFF}
	gbmmu.areas["OAM"] = []uint16{0xFE00, 0xFE9F}
	gbmmu.areas["unused"] = []uint16{0xFEA0, 0xFEFF}
	gbmmu.areas["hardwareIO"] = []uint16{0xFF00, 0xFF7F}
	gbmmu.areas["zeroPage"] = []uint16{0xFF80, 0xFFFE}
	gbmmu.areas["interruptFlag"] = []uint16{0xFFFF}

	// I/O register initial values after boot ROM
	// TODO Might want to just execute boot ROM instead, since
	// documentation online is sketchy about these
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

func (gbmmu *GBMMU) WriteByte(addr uint16, data byte) {
	if addr >= 65535 {
		addr--
	}

	gbmmu.Memory[addr] = data
}

func (gbmmu *GBMMU) ReadByte(addr []byte) byte {
	memLoc := binary.LittleEndian.Uint16(addr)
	if memLoc >= 65535 {
		memLoc--
	}

	return gbmmu.Memory[memLoc]
}

// Reads cartridge ROM into Memory
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
