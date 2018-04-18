package main

import (
	"fmt"
	"os"

	"./cpu"
	"./io"
	"./lcd"
	"./mmu"
	"github.com/hajimehoshi/ebiten"
)

// Global variables for component structs
var GbCPU = new(cpu.GBCPU)
var GbMMU = new(mmu.GBMMU)
var GbLCD = new(lcd.GBLCD)
var GbIO = new(io.GBIO)

func main() {
	cartPath := os.Args[1]

	cpu.GbMMU = GbMMU
	lcd.GbMMU = GbMMU
	mmu.GbIO = GbIO
	lcd.GbIO = GbIO
	lcd.GbCPU = GbCPU
	GbMMU.InitMMU()
	GbCPU.InitCPU()
	GbIO.InitIO()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	ebiten.Run(GbLCD.Run, 160, 144, 4, "Halken")
}
