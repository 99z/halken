package cartcon

type Cartridge struct {
	Title	string
	CGBFlag	int
	Type	int
	ROMSize	int
	RAMSize	int
	MBC		[]byte
}
