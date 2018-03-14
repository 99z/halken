package main

import (
	"./mmu"
	"fmt"
	"os"
)

func main() {
	cartPath := os.Args[1]
	loadedCart, err := mmu.LoadCart(cartPath)
	if err != nil {
		fmt.Println("main: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %s\n", string(loadedCart[0x0134:0x0143]))
}
