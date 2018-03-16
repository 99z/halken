package main

import (
	"./mmu"
	"./cpu"
	"fmt"
	"os"
	"encoding/binary"
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
	
	for i := 0; i < 10; i++ {
		opcode := GbCPU.Regs.PC[:]
		operation := GbMMU.Cart.MBC[binary.LittleEndian.Uint16(opcode)]
		fmt.Printf("%02X\t%v\n", operation, GbCPU.Instrs[operation])
		GbCPU.Instrs[operation].Executor()
	}
}
