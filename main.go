package main

import (
	"fmt"
	"os"

	"./cpu"
	"./lcd"
	"./mmu"
	"github.com/hajimehoshi/ebiten"
)

// Global variables for component structs
var GbCPU = new(cpu.GBCPU)
var GbMMU = new(mmu.GBMMU)
var GbLCD = new(lcd.GBLCD)

func main() {
	cartPath := os.Args[1]

	cpu.GbMMU = GbMMU
	lcd.GbMMU = GbMMU
	lcd.GbCPU = GbCPU
	GbMMU.InitMMU()
	GbCPU.InitCPU()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	ebiten.Run(GbLCD.Run, 320, 288, 2, "Halken")
}
