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

	fmt.Printf("Title: %s\nCGBFlag: %v\nType: %v\nROM: %v\nRAM: %v\n",
		GbMMU.Cart.Title, GbMMU.Cart.CGBFlag, GbMMU.Cart.Type, GbMMU.Cart.ROMSize, GbMMU.Cart.RAMSize)

	for {
		GbCPU.Jumped = false
		opcode := GbCPU.Regs.PC[:]
		operation := GbMMU.Cart.MBC[binary.LittleEndian.Uint16(opcode)]
		fmt.Printf("%02X\t%v\n", operation, GbCPU.Instrs[operation])
		GbCPU.Instrs[operation].Executor()

		if GbCPU.Jumped {
			continue
		} else {
			nextInstr := binary.LittleEndian.Uint16(GbCPU.Regs.PC) + GbCPU.Instrs[operation].NumOperands
			binary.LittleEndian.PutUint16(GbCPU.Regs.PC, nextInstr)
		}
	}
}
