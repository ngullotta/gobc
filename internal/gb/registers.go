package gb

type Registers struct {
	A, F uint8
	B, C uint8
	D, E uint8
	H, L uint8
}

func (r *Registers) GetAF() uint16 {
	return uint16(r.F) | uint16(r.A)<<8
}

func (r *Registers) SetAF(val uint16) {
	val &= 0xFFF0 // The lower 4 bits are always zero, balme nintendo
	r.F = uint8(val & 0xff)
	r.A = uint8(val >> 8)
}

func (r *Registers) GetBC() uint16 {
	return uint16(r.C) | uint16(r.B)<<8
}

func (r *Registers) SetBC(val uint16) {
	r.C = uint8(val & 0xff)
	r.B = uint8(val >> 8)
}

func (r *Registers) GetDE() uint16 {
	return uint16(r.E) | uint16(r.D)<<8
}

func (r *Registers) SetDE(val uint16) {
	r.E = uint8(val & 0xff)
	r.D = uint8(val >> 8)
}

func (r *Registers) GetHL() uint16 {
	return uint16(r.L) | uint16(r.H)<<8
}

func (r *Registers) SetHL(val uint16) {
	r.L = uint8(val & 0xff)
	r.H = uint8(val >> 8)
}
