package main

import (
	"./mmu"
	"./cpu"
	"fmt"
	"os"
)

func main() {
	CartPath := os.Args[1]
	
	gbmmu := new(mmu.GBMMU)
	gbmmu.InitMMU()
	
	gbcpu := new(cpu.GBCPU)
	gbcpu.InitCPU()
	
	err := gbmmu.LoadCart(CartPath)
	if err != nil {
		fmt.Println("main: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Title: %s\nCGBFlag: %v\nType: %v\nROM: %v\nRAM: %v\n",
		gbmmu.Cart.Title, gbmmu.Cart.CGBFlag, gbmmu.Cart.Type, gbmmu.Cart.ROMSize, gbmmu.Cart.RAMSize)
	
	for i := 0; i < 10; i++ {
		operation := gbmmu.Cart.MBC[i]
		fmt.Printf("%02X\t%v\n", operation, gbcpu.Instrs[operation])
	}
}
