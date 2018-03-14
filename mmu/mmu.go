package mmu

import (
	"fmt"
	"io/ioutil"
)

// Reads cartridge ROM into memory
// Returns ROM as byte slice
func LoadCart(path string) ([]byte, error) {
	cart, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("MMU: loadCart(%s) failed: %s", path, err)
	}

	return cart, nil
}
