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
}

func (gbmmu *GBMMU) WriteByte(addr []byte, data byte) {
	addrInt := binary.LittleEndian.Uint16(addr)
	gbmmu.Memory[addrInt] = data
	// memLoc := addrInt & 0x0FFF
}

func (gbmmu *GBMMU) ReadByte(addr []byte) byte {
	memLoc := binary.LittleEndian.Uint16(addr)
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
	for i, v := range cartData[0x0100:] {
		gbmmu.Memory[0x0100+i] = v
	}
	return nil
}
