package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"./cpu"
	"./mmu"
)

// Global variables for component structs
var GbCPU = new(cpu.GBCPU)
var GbMMU = new(mmu.GBMMU)

func main() {
	cartPath := os.Args[1]

	cpu.GbMMU = GbMMU
	GbMMU.InitMMU()
	GbCPU.InitCPU()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Println("main: %s", err)
		os.Exit(1)
	}

	// fmt.Printf("Title: %s\nCGBFlag: %v\nType: %v\nROM: %v\nRAM: %v\n",
	// 	GbMMU.Cart.Title, GbMMU.Cart.CGBFlag, GbMMU.Cart.Type, GbMMU.Cart.ROMSize, GbMMU.Cart.RAMSize)

	for i := 0; i < 250; i++ {
		GbCPU.Jumped = false
		opcode := GbCPU.Regs.PC[:]
		opcodeInt := binary.LittleEndian.Uint16(opcode)

		operation := GbMMU.Memory[opcodeInt]

		fmt.Printf("%02X:%02X\t%02X\t%v\n", opcode[1], opcode[0], operation, GbCPU.Instrs[operation])
		GbCPU.Instrs[operation].Executor()

		if GbCPU.Jumped {
			continue
		} else {
			nextInstr := binary.LittleEndian.Uint16(GbCPU.Regs.PC) + GbCPU.Instrs[operation].NumOperands
			binary.LittleEndian.PutUint16(GbCPU.Regs.PC, nextInstr)
		}
	}
}
