package main

import (
	"fmt"
	"os"

	"./cpu"
	"./io"
	"./lcd"
	"./mmu"
	"./timer"
	"github.com/hajimehoshi/ebiten"
)

// Global variables for component structs
var GbCPU = new(cpu.GBCPU)
var GbTimer = new(timer.GBTimer)
var GbMMU = new(mmu.GBMMU)
var GbLCD = new(lcd.GBLCD)
var GbIO = new(io.GBIO)

func main() {
	cartPath := os.Args[1]

	cpu.GbMMU = GbMMU
	lcd.GbMMU = GbMMU
	timer.GbMMU = GbMMU
	mmu.GbIO = GbIO
	lcd.GbIO = GbIO
	lcd.GbCPU = GbCPU
	lcd.GbTimer = GbTimer
	GbMMU.InitMMU()
	GbCPU.InitCPU()
	GbIO.InitIO()
	GbLCD.InitLCD()

	err := GbMMU.LoadCart(cartPath)
	if err != nil {
		fmt.Printf("main: %s\n", err)
		os.Exit(1)
	}

	ebiten.Run(GbLCD.Run, 160, 144, 4, "Halken")
}
