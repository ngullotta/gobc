package gb

import (
	"fmt"
	"strings"
)

var OpCycles = []int{
	1, 3, 2, 2, 1, 1, 2, 1, 5, 2, 2, 2, 1, 1, 2, 1, // 0
	0, 3, 2, 2, 1, 1, 2, 1, 3, 2, 2, 2, 1, 1, 2, 1, // 1
	2, 3, 2, 2, 1, 1, 2, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 2
	2, 3, 2, 2, 3, 3, 3, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 3
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 4
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 5
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 6
	2, 2, 2, 2, 2, 2, 0, 2, 1, 1, 1, 1, 1, 1, 2, 1, // 7
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 8
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 9
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // a
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // b
	2, 3, 3, 4, 3, 4, 2, 4, 2, 4, 3, 0, 3, 6, 2, 4, // c
	2, 3, 3, 0, 3, 4, 2, 4, 2, 4, 3, 0, 3, 0, 2, 4, // d
	3, 3, 2, 0, 0, 4, 2, 4, 4, 1, 4, 0, 0, 0, 2, 4, // e
	3, 3, 2, 1, 0, 4, 2, 4, 3, 2, 4, 1, 0, 0, 2, 4, // f
} //0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f

type CPU struct {
	regs Registers
	bus  *MMU

	SP uint16
	PC uint16

	IME bool

	halted bool

	debug bool
}

var (
	DMG = Registers{
		A: 0x01, F: 0xB0,
		B: 0x00, C: 0x13,
		D: 0x00, E: 0xD8,
		H: 0x01, L: 0x4D,
	}

	// Color mode
	CGB = Registers{
		A: 0x11, F: 0x80,
		B: 0x00, C: 0x00,
		D: 0xFF, E: 0x56,
		H: 0x00, L: 0x0D,
	}
)

func NewCPU() *CPU {
	return &CPU{
		regs:   DMG,
		bus:    &MMU{},
		SP:     0xFFFE,
		PC:     0x0100,
		halted: true,
	}
}

func (c *CPU) fetchu8() byte {
	val := c.bus.Read(c.PC)
	c.PC++
	return val
}

func (c *CPU) fetchu16() uint16 {
	lo := uint16(c.fetchu8())
	hi := uint16(c.fetchu8())
	return (hi << 8) | lo
}

func (c *CPU) push16(val uint16) {
	c.SP--
	c.bus.Write(c.SP, byte(val>>8))
	c.SP--
	c.bus.Write(c.SP, byte(val))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.bus.Read(c.SP))
	c.SP++
	hi := uint16(c.bus.Read(c.SP))
	c.SP++
	return (hi << 8) | lo
}

func (c *CPU) inc8(val byte) byte {
	res := val + 1
	c.regs.SetZ(res == 0)
	c.regs.SetN(false)
	c.regs.SetH(val&0xF == 0xF)
	return res
}

func (c *CPU) dec8(val byte) byte {
	res := val - 1
	c.regs.SetZ(res == 0)
	c.regs.SetN(true)
	c.regs.SetH(val&0xF == 0)
	return res
}

func (c *CPU) add8(val byte, addCarry bool) {
	carry := byte(0)
	if addCarry && c.regs.GetC() {
		carry = 1
	}
	res := uint16(c.regs.A) + uint16(val) + uint16(carry)
	half := (c.regs.A & 0xF) + (val & 0xF) + carry

	c.regs.A = byte(res)
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetN(false)
	c.regs.SetH(half > 0xF)
	c.regs.SetC(res > 0xFF)
}

func (c *CPU) sub8(val byte, subCarry bool) {
	carry := byte(0)
	if subCarry && c.regs.GetC() {
		carry = 1
	}
	res := int16(c.regs.A) - int16(val) - int16(carry)
	half := int16(c.regs.A&0xF) - int16(val&0xF) - int16(carry)

	c.regs.A = byte(res)
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetN(true)
	c.regs.SetH(half < 0)
	c.regs.SetC(res < 0)
}

func (c *CPU) and8(val byte) {
	c.regs.A &= val
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetN(false)
	c.regs.SetH(true)
	c.regs.SetC(false)
}

func (c *CPU) or8(val byte) {
	c.regs.A |= val
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetN(false)
	c.regs.SetH(false)
	c.regs.SetC(false)
}

func (c *CPU) xor8(val byte) {
	c.regs.A ^= val
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetN(false)
	c.regs.SetH(false)
	c.regs.SetC(false)
}

func (c *CPU) cp8(val byte) {
	res := int16(c.regs.A) - int16(val)
	half := int16(c.regs.A&0xF) - int16(val&0xF)
	c.regs.SetZ(byte(res) == 0)
	c.regs.SetN(true)
	c.regs.SetH(half < 0)
	c.regs.SetC(res < 0)
}

func (c *CPU) addHL(val uint16) {
	hl := c.regs.GetHL()
	res := uint32(hl) + uint32(val)
	c.regs.SetN(false)
	c.regs.SetH((hl&0xFFF)+(val&0xFFF) > 0xFFF)
	c.regs.SetC(res > 0xFFFF)
	c.regs.SetDE(uint16(res))
}

func (c *CPU) daa() {
	a := int16(c.regs.A)
	if !c.regs.GetN() {
		if c.regs.GetC() || a > 0x99 {
			a += 0x60
			c.regs.SetC(true)
		}
		if c.regs.GetH() || (a&0x0F) > 0x09 {
			a += 0x06
		}
	} else {
		if c.regs.GetC() {
			a -= 0x60
		}
		if c.regs.GetH() {
			a -= 0x06
		}
	}
	c.regs.A = byte(a)
	c.regs.SetZ(c.regs.A == 0)
	c.regs.SetH(false)
}

// Helpers for CB prefix reg targeting (0=B, 1=C, 2=D, 3=E, 4=H, 5=L, 6=(HL), 7=A)
func (c *CPU) getReg8(idx byte) byte {
	switch idx {
	case 0:
		return c.regs.B
	case 1:
		return c.regs.C
	case 2:
		return c.regs.D
	case 3:
		return c.regs.E
	case 4:
		return c.regs.H
	case 5:
		return c.regs.L
	case 6:
		return c.bus.Read(c.regs.GetHL())
	case 7:
		return c.regs.A
	}
	return 0
}

func (c *CPU) setReg8(idx byte, val byte) {
	switch idx {
	case 0:
		c.regs.B = val
	case 1:
		c.regs.C = val
	case 2:
		c.regs.D = val
	case 3:
		c.regs.E = val
	case 4:
		c.regs.H = val
	case 5:
		c.regs.L = val
	case 6:
		c.bus.Write(c.regs.GetHL(), val)
	case 7:
		c.regs.A = val
	}
}

func (c *CPU) executeCB() {
	op := c.fetchu8()
	x := op >> 6
	y := (op >> 3) & 7
	z := op & 7

	val := c.getReg8(z)
	switch x {
	case 0: // Shifts and Rotates
		switch y {
		case 0: // RLC
			c.regs.SetC(val&0x80 != 0)
			val = (val << 1) | (val >> 7)
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 1: // RRC
			c.regs.SetC(val&0x01 != 0)
			val = (val >> 1) | (val << 7)
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 2: // RL
			carry := byte(0)
			if c.regs.GetC() {
				carry = 1
			}
			c.regs.SetC(val&0x80 != 0)
			val = (val << 1) | carry
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 3: // RR
			carry := byte(0)
			if c.regs.GetC() {
				carry = 0x80
			}
			c.regs.SetC(val&0x01 != 0)
			val = (val >> 1) | carry
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 4: // SLA
			c.regs.SetC(val&0x80 != 0)
			val <<= 1
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 5: // SRA
			c.regs.SetC(val&0x01 != 0)
			val = (val >> 1) | (val & 0x80)
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		case 6: // SWAP
			val = (val >> 4) | (val << 4)
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
			c.regs.SetC(false)
		case 7: // SRL
			c.regs.SetC(val&0x01 != 0)
			val >>= 1
			c.regs.SetZ(val == 0)
			c.regs.SetN(false)
			c.regs.SetH(false)
		}
		c.setReg8(z, val)
	case 1: // BIT test
		c.regs.SetZ(val&(1<<y) == 0)
		c.regs.SetN(false)
		c.regs.SetH(true)
	case 2: // RES (Reset Bit)
		c.setReg8(z, val&^(1<<y))
	case 3: // SET (Set Bit)
		c.setReg8(z, val|(1<<y))
	}
}

// Reference: https://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
var instructions = [0x100]func(*CPU){
	// 0x00 => 0x0F
	0x00: func(c *CPU) {},                                        // NOP
	0x01: func(c *CPU) { c.regs.SetBC(c.fetchu16()) },            // LD BC, d16
	0x02: func(c *CPU) { c.bus.Write(c.regs.GetBC(), c.regs.A) }, // LD (BC), A
	0x03: func(c *CPU) { c.regs.SetBC(c.regs.GetBC() + 1) },      // INC BC
	0x04: func(c *CPU) { c.regs.B = c.inc8(c.regs.B) },           // INC B
	0x05: func(c *CPU) { c.regs.B = c.dec8(c.regs.B) },           // DEC B
	0x06: func(c *CPU) { c.regs.B = c.fetchu8() },                // LD B, d8
	0x07: func(c *CPU) {
		c.regs.SetC(c.regs.A&0x80 != 0)
		c.regs.A = (c.regs.A << 1) | (c.regs.A >> 7)
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH(false)
	}, // RLCA
	0x08: func(c *CPU) { addr := c.fetchu16(); c.bus.Write(addr, byte(c.SP)); c.bus.Write(addr+1, byte(c.SP>>8)) }, // LD (a16), SP
	0x09: func(c *CPU) { c.addHL(c.regs.GetBC()) },                                                                 // ADD HL, BC
	0x0A: func(c *CPU) { c.regs.A = c.bus.Read(c.regs.GetBC()) },                                                   // LD A, (BC)
	0x0B: func(c *CPU) { c.regs.SetBC(c.regs.GetBC() - 1) },                                                        // DEC BC
	0x0C: func(c *CPU) { c.regs.C = c.inc8(c.regs.C) },                                                             // INC C
	0x0D: func(c *CPU) { c.regs.C = c.dec8(c.regs.C) },                                                             // DEC C
	0x0E: func(c *CPU) { c.regs.C = c.fetchu8() },                                                                  // LD C, d8
	0x0F: func(c *CPU) {
		c.regs.SetC(c.regs.A&0x01 != 0)
		c.regs.A = (c.regs.A >> 1) | (c.regs.A << 7)
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH(false)
	}, // RRCA

	// 0x10 => 0x1F
	0x10: func(c *CPU) { c.fetchu8() /* STOP - skip next byte */ },
	0x11: func(c *CPU) { c.regs.SetDE(c.fetchu16()) },            // LD DE, d16
	0x12: func(c *CPU) { c.bus.Write(c.regs.GetDE(), c.regs.A) }, // LD (DE), A
	0x13: func(c *CPU) { c.regs.SetDE(c.regs.GetDE() + 1) },      // INC DE
	0x14: func(c *CPU) { c.regs.D = c.inc8(c.regs.D) },           // INC D
	0x15: func(c *CPU) { c.regs.D = c.dec8(c.regs.D) },           // DEC D
	0x16: func(c *CPU) { c.regs.D = c.fetchu8() },                // LD D, d8
	0x17: func(c *CPU) {
		carry := byte(0)
		if c.regs.GetC() {
			carry = 1
		}
		c.regs.SetC(c.regs.A&0x80 != 0)
		c.regs.A = (c.regs.A << 1) | carry
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH(false)
	}, // RLA
	0x18: func(c *CPU) { offset := int8(c.fetchu8()); c.PC = uint16(int32(c.PC) + int32(offset)) }, // JR r8
	0x19: func(c *CPU) { c.addHL(c.regs.GetDE()) },                                                 // ADD HL, DE
	0x1A: func(c *CPU) { c.regs.A = c.bus.Read(c.regs.GetDE()) },                                   // LD A, (DE)
	0x1B: func(c *CPU) { c.regs.SetDE(c.regs.GetDE() - 1) },                                        // DEC DE
	0x1C: func(c *CPU) { c.regs.E = c.inc8(c.regs.E) },                                             // INC E
	0x1D: func(c *CPU) { c.regs.E = c.dec8(c.regs.E) },                                             // DEC E
	0x1E: func(c *CPU) { c.regs.E = c.fetchu8() },                                                  // LD E, d8
	0x1F: func(c *CPU) {
		carry := byte(0)
		if c.regs.GetC() {
			carry = 0x80
		}
		c.regs.SetC(c.regs.A&0x01 != 0)
		c.regs.A = (c.regs.A >> 1) | carry
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH(false)
	}, // RRA

	// 0x20 => 0x2F
	0x20: func(c *CPU) {
		offset := int8(c.fetchu8())
		if !c.regs.GetZ() {
			c.PC = uint16(int32(c.PC) + int32(offset))
		}
	}, // JR NZ, r8
	0x21: func(c *CPU) { c.regs.SetHL(c.fetchu16()) },                                              // LD HL, d16
	0x22: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.A); c.regs.SetHL(c.regs.GetHL() + 1) }, // LDI (HL), A
	0x23: func(c *CPU) { c.regs.SetHL(c.regs.GetHL() + 1) },                                        // INC HL
	0x24: func(c *CPU) { c.regs.H = c.inc8(c.regs.H) },                                             // INC H
	0x25: func(c *CPU) { c.regs.H = c.dec8(c.regs.H) },                                             // DEC H
	0x26: func(c *CPU) { c.regs.H = c.fetchu8() },                                                  // LD H, d8
	0x27: func(c *CPU) { c.daa() },                                                                 // DAA
	0x28: func(c *CPU) {
		offset := int8(c.fetchu8())
		if c.regs.GetZ() {
			c.PC = uint16(int32(c.PC) + int32(offset))
		}
	}, // JR Z, r8
	0x29: func(c *CPU) { c.addHL(c.regs.GetHL()) },                                                 // ADD HL, HL
	0x2A: func(c *CPU) { c.regs.A = c.bus.Read(c.regs.GetHL()); c.regs.SetHL(c.regs.GetHL() + 1) }, // LDI A, (HL)
	0x2B: func(c *CPU) { c.regs.SetHL(c.regs.GetHL() - 1) },                                        // DEC HL
	0x2C: func(c *CPU) { c.regs.L = c.inc8(c.regs.L) },                                             // INC L
	0x2D: func(c *CPU) { c.regs.L = c.dec8(c.regs.L) },                                             // DEC L
	0x2E: func(c *CPU) { c.regs.L = c.fetchu8() },                                                  // LD L, d8
	0x2F: func(c *CPU) { c.regs.A = ^c.regs.A; c.regs.SetN(true); c.regs.SetH(true) },              // CPL

	// 0x30 => 0x3F
	0x30: func(c *CPU) {
		offset := int8(c.fetchu8())
		if !c.regs.GetC() {
			c.PC = uint16(int32(c.PC) + int32(offset))
		}
	}, // JR NC, r8
	0x31: func(c *CPU) { c.SP = c.fetchu16() },                                                     // LD SP, d16
	0x32: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.A); c.regs.SetHL(c.regs.GetHL() - 1) }, // LDD (HL), A
	0x33: func(c *CPU) { c.SP++ },                                                                  // INC SP
	0x34: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.inc8(c.bus.Read(c.regs.GetHL()))) },         // INC (HL)
	0x35: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.dec8(c.bus.Read(c.regs.GetHL()))) },         // DEC (HL)
	0x36: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.fetchu8()) },                                // LD (HL), d8
	0x37: func(c *CPU) { c.regs.SetN(false); c.regs.SetH(false); c.regs.SetC(true) },               // SCF
	0x38: func(c *CPU) {
		offset := int8(c.fetchu8())
		if c.regs.GetC() {
			c.PC = uint16(int32(c.PC) + int32(offset))
		}
	}, // JR C, r8
	0x39: func(c *CPU) { c.addHL(c.SP) },                                                           // ADD HL, SP
	0x3A: func(c *CPU) { c.regs.A = c.bus.Read(c.regs.GetHL()); c.regs.SetHL(c.regs.GetHL() - 1) }, // LDD A, (HL)
	0x3B: func(c *CPU) { c.SP-- },                                                                  // DEC SP
	0x3C: func(c *CPU) { c.regs.A = c.inc8(c.regs.A) },                                             // INC A
	0x3D: func(c *CPU) { c.regs.A = c.dec8(c.regs.A) },                                             // DEC A
	0x3E: func(c *CPU) { c.regs.A = c.fetchu8() },                                                  // LD A, d8
	0x3F: func(c *CPU) { c.regs.SetN(false); c.regs.SetH(false); c.regs.SetC(!c.regs.GetC()) },     // CCF

	// 0x40 => 0x4F
	0x40: func(c *CPU) {},
	0x41: func(c *CPU) { c.regs.B = c.regs.C },
	0x42: func(c *CPU) { c.regs.B = c.regs.D },
	0x43: func(c *CPU) { c.regs.B = c.regs.E },
	0x44: func(c *CPU) { c.regs.B = c.regs.H },
	0x45: func(c *CPU) { c.regs.B = c.regs.L },
	0x46: func(c *CPU) { c.regs.B = c.bus.Read(c.regs.GetHL()) },
	0x47: func(c *CPU) { c.regs.B = c.regs.A },
	0x48: func(c *CPU) { c.regs.C = c.regs.B },
	0x49: func(c *CPU) {},
	0x4A: func(c *CPU) { c.regs.C = c.regs.D },
	0x4B: func(c *CPU) { c.regs.C = c.regs.E },
	0x4C: func(c *CPU) { c.regs.C = c.regs.H },
	0x4D: func(c *CPU) { c.regs.C = c.regs.L },
	0x4E: func(c *CPU) { c.regs.C = c.bus.Read(c.regs.GetHL()) },
	0x4F: func(c *CPU) { c.regs.C = c.regs.A },

	// 0x50 => 0x5F
	0x50: func(c *CPU) { c.regs.D = c.regs.B },
	0x51: func(c *CPU) { c.regs.D = c.regs.C },
	0x52: func(c *CPU) {},
	0x53: func(c *CPU) { c.regs.D = c.regs.E },
	0x54: func(c *CPU) { c.regs.D = c.regs.H },
	0x55: func(c *CPU) { c.regs.D = c.regs.L },
	0x56: func(c *CPU) { c.regs.D = c.bus.Read(c.regs.GetHL()) },
	0x57: func(c *CPU) { c.regs.D = c.regs.A },
	0x58: func(c *CPU) { c.regs.E = c.regs.B },
	0x59: func(c *CPU) { c.regs.E = c.regs.C },
	0x5A: func(c *CPU) { c.regs.E = c.regs.D },
	0x5B: func(c *CPU) {},
	0x5C: func(c *CPU) { c.regs.E = c.regs.H },
	0x5D: func(c *CPU) { c.regs.E = c.regs.L },
	0x5E: func(c *CPU) { c.regs.E = c.bus.Read(c.regs.GetHL()) },
	0x5F: func(c *CPU) { c.regs.E = c.regs.A },

	// 0x60 => 0x6F
	0x60: func(c *CPU) { c.regs.H = c.regs.B },
	0x61: func(c *CPU) { c.regs.H = c.regs.C },
	0x62: func(c *CPU) { c.regs.H = c.regs.D },
	0x63: func(c *CPU) { c.regs.H = c.regs.E },
	0x64: func(c *CPU) {},
	0x65: func(c *CPU) { c.regs.H = c.regs.L },
	0x66: func(c *CPU) { c.regs.H = c.bus.Read(c.regs.GetHL()) },
	0x67: func(c *CPU) { c.regs.H = c.regs.A },
	0x68: func(c *CPU) { c.regs.L = c.regs.B },
	0x69: func(c *CPU) { c.regs.L = c.regs.C },
	0x6A: func(c *CPU) { c.regs.L = c.regs.D },
	0x6B: func(c *CPU) { c.regs.L = c.regs.E },
	0x6C: func(c *CPU) { c.regs.L = c.regs.H },
	0x6D: func(c *CPU) {},
	0x6E: func(c *CPU) { c.regs.L = c.bus.Read(c.regs.GetHL()) },
	0x6F: func(c *CPU) { c.regs.L = c.regs.A },

	// 0x70 => 0x7F
	0x70: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.B) },
	0x71: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.C) },
	0x72: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.D) },
	0x73: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.E) },
	0x74: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.H) },
	0x75: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.L) },
	0x76: func(c *CPU) { c.halted = true }, // HALT
	0x77: func(c *CPU) { c.bus.Write(c.regs.GetHL(), c.regs.A) },
	0x78: func(c *CPU) { c.regs.A = c.regs.B },
	0x79: func(c *CPU) { c.regs.A = c.regs.C },
	0x7A: func(c *CPU) { c.regs.A = c.regs.D },
	0x7B: func(c *CPU) { c.regs.A = c.regs.E },
	0x7C: func(c *CPU) { c.regs.A = c.regs.H },
	0x7D: func(c *CPU) { c.regs.A = c.regs.L },
	0x7E: func(c *CPU) { c.regs.A = c.bus.Read(c.regs.GetHL()) },
	0x7F: func(c *CPU) {},

	// ADD
	0x80: func(c *CPU) { c.add8(c.regs.B, false) },
	0x81: func(c *CPU) { c.add8(c.regs.C, false) },
	0x82: func(c *CPU) { c.add8(c.regs.D, false) },
	0x83: func(c *CPU) { c.add8(c.regs.E, false) },
	0x84: func(c *CPU) { c.add8(c.regs.H, false) },
	0x85: func(c *CPU) { c.add8(c.regs.L, false) },
	0x86: func(c *CPU) { c.add8(c.bus.Read(c.regs.GetHL()), false) },
	0x87: func(c *CPU) { c.add8(c.regs.A, false) },
	// ADC
	0x88: func(c *CPU) { c.add8(c.regs.B, true) },
	0x89: func(c *CPU) { c.add8(c.regs.C, true) },
	0x8A: func(c *CPU) { c.add8(c.regs.D, true) },
	0x8B: func(c *CPU) { c.add8(c.regs.E, true) },
	0x8C: func(c *CPU) { c.add8(c.regs.H, true) },
	0x8D: func(c *CPU) { c.add8(c.regs.L, true) },
	0x8E: func(c *CPU) { c.add8(c.bus.Read(c.regs.GetHL()), true) },
	0x8F: func(c *CPU) { c.add8(c.regs.A, true) },
	// SUB
	0x90: func(c *CPU) { c.sub8(c.regs.B, false) },
	0x91: func(c *CPU) { c.sub8(c.regs.C, false) },
	0x92: func(c *CPU) { c.sub8(c.regs.D, false) },
	0x93: func(c *CPU) { c.sub8(c.regs.E, false) },
	0x94: func(c *CPU) { c.sub8(c.regs.H, false) },
	0x95: func(c *CPU) { c.sub8(c.regs.L, false) },
	0x96: func(c *CPU) { c.sub8(c.bus.Read(c.regs.GetHL()), false) },
	0x97: func(c *CPU) { c.sub8(c.regs.A, false) },
	// SBC
	0x98: func(c *CPU) { c.sub8(c.regs.B, true) },
	0x99: func(c *CPU) { c.sub8(c.regs.C, true) },
	0x9A: func(c *CPU) { c.sub8(c.regs.D, true) },
	0x9B: func(c *CPU) { c.sub8(c.regs.E, true) },
	0x9C: func(c *CPU) { c.sub8(c.regs.H, true) },
	0x9D: func(c *CPU) { c.sub8(c.regs.L, true) },
	0x9E: func(c *CPU) { c.sub8(c.bus.Read(c.regs.GetHL()), true) },
	0x9F: func(c *CPU) { c.sub8(c.regs.A, true) },
	// AND
	0xA0: func(c *CPU) { c.and8(c.regs.B) },
	0xA1: func(c *CPU) { c.and8(c.regs.C) },
	0xA2: func(c *CPU) { c.and8(c.regs.D) },
	0xA3: func(c *CPU) { c.and8(c.regs.E) },
	0xA4: func(c *CPU) { c.and8(c.regs.H) },
	0xA5: func(c *CPU) { c.and8(c.regs.L) },
	0xA6: func(c *CPU) { c.and8(c.bus.Read(c.regs.GetHL())) },
	0xA7: func(c *CPU) { c.and8(c.regs.A) },
	// XOR
	0xA8: func(c *CPU) { c.xor8(c.regs.B) },
	0xA9: func(c *CPU) { c.xor8(c.regs.C) },
	0xAA: func(c *CPU) { c.xor8(c.regs.D) },
	0xAB: func(c *CPU) { c.xor8(c.regs.E) },
	0xAC: func(c *CPU) { c.xor8(c.regs.H) },
	0xAD: func(c *CPU) { c.xor8(c.regs.L) },
	0xAE: func(c *CPU) { c.xor8(c.bus.Read(c.regs.GetHL())) },
	0xAF: func(c *CPU) { c.xor8(c.regs.A) },
	// OR
	0xB0: func(c *CPU) { c.or8(c.regs.B) },
	0xB1: func(c *CPU) { c.or8(c.regs.C) },
	0xB2: func(c *CPU) { c.or8(c.regs.D) },
	0xB3: func(c *CPU) { c.or8(c.regs.E) },
	0xB4: func(c *CPU) { c.or8(c.regs.H) },
	0xB5: func(c *CPU) { c.or8(c.regs.L) },
	0xB6: func(c *CPU) { c.or8(c.bus.Read(c.regs.GetHL())) },
	0xB7: func(c *CPU) { c.or8(c.regs.A) },
	// CP
	0xB8: func(c *CPU) { c.cp8(c.regs.B) },
	0xB9: func(c *CPU) { c.cp8(c.regs.C) },
	0xBA: func(c *CPU) { c.cp8(c.regs.D) },
	0xBB: func(c *CPU) { c.cp8(c.regs.E) },
	0xBC: func(c *CPU) { c.cp8(c.regs.H) },
	0xBD: func(c *CPU) { c.cp8(c.regs.L) },
	0xBE: func(c *CPU) { c.cp8(c.bus.Read(c.regs.GetHL())) },
	0xBF: func(c *CPU) { c.cp8(c.regs.A) },

	// 0xCx
	0xC0: func(c *CPU) {
		if !c.regs.GetZ() {
			c.PC = c.pop16()
		}
	}, // RET NZ
	0xC1: func(c *CPU) { c.regs.SetBC(c.pop16()) }, // POP BC
	0xC2: func(c *CPU) {
		addr := c.fetchu16()
		if !c.regs.GetZ() {
			c.PC = addr
		}
	}, // JP NZ, a16
	0xC3: func(c *CPU) { c.PC = c.fetchu16() }, // JP a16
	0xC4: func(c *CPU) {
		addr := c.fetchu16()
		if !c.regs.GetZ() {
			c.push16(c.PC)
			c.PC = addr
		}
	}, // CALL NZ, a16
	0xC5: func(c *CPU) { c.push16(c.regs.GetBC()) },      // PUSH BC
	0xC6: func(c *CPU) { c.add8(c.fetchu8(), false) },    // ADD A, d8
	0xC7: func(c *CPU) { c.push16(c.PC); c.PC = 0x0000 }, // RST 00H
	0xC8: func(c *CPU) {
		if c.regs.GetZ() {
			c.PC = c.pop16()
		}
	}, // RET Z
	0xC9: func(c *CPU) { c.PC = c.pop16() }, // RET
	0xCA: func(c *CPU) {
		addr := c.fetchu16()
		if c.regs.GetZ() {
			c.PC = addr
		}
	}, // JP Z, a16
	0xCB: func(c *CPU) { c.executeCB() }, // PREFIX CB
	0xCC: func(c *CPU) {
		addr := c.fetchu16()
		if c.regs.GetZ() {
			c.push16(c.PC)
			c.PC = addr
		}
	}, // CALL Z, a16
	0xCD: func(c *CPU) { addr := c.fetchu16(); c.push16(c.PC); c.PC = addr }, // CALL a16
	0xCE: func(c *CPU) { c.add8(c.fetchu8(), true) },                         // ADC A, d8
	0xCF: func(c *CPU) { c.push16(c.PC); c.PC = 0x0008 },                     // RST 08H

	// 0xDx
	0xD0: func(c *CPU) {
		if !c.regs.GetC() {
			c.PC = c.pop16()
		}
	}, // RET NC
	0xD1: func(c *CPU) { c.regs.SetDE(c.pop16()) }, // POP DE
	0xD2: func(c *CPU) {
		addr := c.fetchu16()
		if !c.regs.GetC() {
			c.PC = addr
		}
	}, // JP NC, a16
	0xD4: func(c *CPU) {
		addr := c.fetchu16()
		if !c.regs.GetC() {
			c.push16(c.PC)
			c.PC = addr
		}
	}, // CALL NC, a16
	0xD5: func(c *CPU) { c.push16(c.regs.GetDE()) },      // PUSH DE
	0xD6: func(c *CPU) { c.sub8(c.fetchu8(), false) },    // SUB d8
	0xD7: func(c *CPU) { c.push16(c.PC); c.PC = 0x0010 }, // RST 10H
	0xD8: func(c *CPU) {
		if c.regs.GetC() {
			c.PC = c.pop16()
		}
	}, // RET C
	0xD9: func(c *CPU) { c.PC = c.pop16(); c.IME = true }, // RETI
	0xDA: func(c *CPU) {
		addr := c.fetchu16()
		if c.regs.GetC() {
			c.PC = addr
		}
	}, // JP C, a16
	0xDC: func(c *CPU) {
		addr := c.fetchu16()
		if c.regs.GetC() {
			c.push16(c.PC)
			c.PC = addr
		}
	}, // CALL C, a16
	0xDE: func(c *CPU) { c.sub8(c.fetchu8(), true) },     // SBC A, d8
	0xDF: func(c *CPU) { c.push16(c.PC); c.PC = 0x0018 }, // RST 18H

	// 0xEx
	0xE0: func(c *CPU) { c.bus.Write(0xFF00+uint16(c.fetchu8()), c.regs.A) }, // LDH (a8), A
	0xE1: func(c *CPU) { c.regs.SetHL(c.pop16()) },                           // POP HL
	0xE2: func(c *CPU) { c.bus.Write(0xFF00+uint16(c.regs.C), c.regs.A) },    // LD (C), A
	0xE5: func(c *CPU) { c.push16(c.regs.GetHL()) },                          // PUSH HL
	0xE6: func(c *CPU) { c.and8(c.fetchu8()) },                               // AND d8
	0xE7: func(c *CPU) { c.push16(c.PC); c.PC = 0x0020 },                     // RST 20H
	0xE8: func(c *CPU) {
		offset := int8(c.fetchu8())
		sp := c.SP
		c.SP = uint16(int32(sp) + int32(offset))
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH((sp&0xF)+(uint16(byte(offset))&0xF) > 0xF)
		c.regs.SetC((sp&0xFF)+(uint16(byte(offset))&0xFF) > 0xFF)
	}, // ADD SP, r8
	0xE9: func(c *CPU) { c.PC = c.regs.GetHL() },               // JP HL
	0xEA: func(c *CPU) { c.bus.Write(c.fetchu16(), c.regs.A) }, // LD (a16), A
	0xEE: func(c *CPU) { c.xor8(c.fetchu8()) },                 // XOR d8
	0xEF: func(c *CPU) { c.push16(c.PC); c.PC = 0x0028 },       // RST 28H

	// 0xF0 => 0xFF
	0xF0: func(c *CPU) { c.regs.A = c.bus.Read(0xFF00 + uint16(c.fetchu8())) }, // LDH A, (a8)
	0xF1: func(c *CPU) { c.regs.SetAF(c.pop16()) },                             // POP AF
	0xF2: func(c *CPU) { c.regs.A = c.bus.Read(0xFF00 + uint16(c.regs.C)) },    // LD A, (C)
	0xF3: func(c *CPU) { c.IME = false },                                       // DI
	0xF5: func(c *CPU) { c.push16(c.regs.GetAF()) },                            // PUSH AF
	0xF6: func(c *CPU) { c.or8(c.fetchu8()) },                                  // OR d8
	0xF7: func(c *CPU) { c.push16(c.PC); c.PC = 0x0030 },                       // RST 30H
	0xF8: func(c *CPU) {
		offset := int8(c.fetchu8())
		sp := c.SP
		c.regs.SetHL(uint16(int32(sp) + int32(offset)))
		c.regs.SetZ(false)
		c.regs.SetN(false)
		c.regs.SetH((sp&0xF)+(uint16(byte(offset))&0xF) > 0xF)
		c.regs.SetC((sp&0xFF)+(uint16(byte(offset))&0xFF) > 0xFF)
	}, // LD HL, SP+r8
	0xF9: func(c *CPU) { c.SP = c.regs.GetHL() },               // LD SP, HL
	0xFA: func(c *CPU) { c.regs.A = c.bus.Read(c.fetchu16()) }, // LD A, (a16)
	0xFB: func(c *CPU) { c.IME = true },                        // EI
	0xFE: func(c *CPU) { c.cp8(c.fetchu8()) },                  // CP d8
	0xFF: func(c *CPU) { c.push16(c.PC); c.PC = 0x0038 },       // RST 38H
}

func (cpu *CPU) Exec(op byte) int {
	if cpu.halted {
		return 0
	}

	instructions[op](cpu)

	if cpu.debug {
		fmt.Printf("[*] PC = 0x%04X, OP = 0x%02X\033[0m\n", cpu.PC, op)
	}

	return OpCycles[op]
}

func (cpu *CPU) Step() int {
	if cpu.halted {
		return 0
	}

	return cpu.Exec(cpu.fetchu8())
}

func (cpu *CPU) Play() {
	cpu.halted = !cpu.halted
}

func (cpu *CPU) Debug() {
	cpu.debug = !cpu.debug
}

// Not a permanent place for these, just need to expose them for main.go testing
func (cpu *CPU) LoadROM(data []byte) error {
	return cpu.bus.LoadROM(data)
}

func (cpu *CPU) GetCartName() string {
	rawTitle := string(cpu.bus.ROM[0x134:0x142])
	return strings.Trim(rawTitle, "\x00")
}

func InitCPUInstructions() {
	for k, v := range instructions {
		if v == nil {
			instructions[k] = func(*CPU) {
				fmt.Printf("\033[31;1;4m")
			}
		}
	}
}
