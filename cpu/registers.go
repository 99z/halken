package cpu

// 8 bit registers
type Register struct {
	a byte // Accumulator
	f byte // Flags

	b byte
	c byte

	d byte
	e byte

	h byte
	l byte

	sp [2]byte // Stack pointer
	pc [2]byte // Program counter
}

func (r *Register) writeAF(data [2]byte) {
	r.a = data[0]
	r.f = data[1]
}

func (r *Register) readAF() [2]byte {
	return [2]byte{r.a, r.f}
}

func (r *Register) writeBC(data [2]byte) {
	r.b = data[0]
	r.c = data[1]
}

func (r *Register) readBC() [2]byte {
	return [2]byte{r.b, r.c}
}

func (r *Register) writeDE(data [2]byte) {
	r.d = data[0]
	r.e = data[1]
}

func (r *Register) readDE() [2]byte {
	return [2]byte{r.d, r.e}
}

func (r *Register) writeHL(data [2]byte) {
	r.h = data[0]
	r.l = data[1]
}

func (r *Register) readHL() [2]byte {
	return [2]byte{r.h, r.l}
}
