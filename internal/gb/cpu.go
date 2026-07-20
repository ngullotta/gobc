package gb

import "fmt"

var OpcodeCycles = []int{
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

	halted bool
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

var instructions = [0x100]func(*CPU){
	// 0x0x
	0x00: func(c *CPU) {},                             // NOP
	0x01: func(c *CPU) { c.regs.SetBC(c.fetchu16()) }, // LD BC, d16
	0x06: func(c *CPU) { c.regs.B = c.fetchu8() },     // LD B, d8
	0x0E: func(c *CPU) { c.regs.C = c.fetchu8() },     // LD C, d8

	// 0x1x
	0x11: func(c *CPU) { c.regs.SetDE(c.fetchu16()) }, // LD DE, d16
	0x16: func(c *CPU) { c.regs.D = c.fetchu8() },     // LD D, d8
	0x1E: func(c *CPU) { c.regs.E = c.fetchu8() },     // LD E, d8

	// 0x2x
	0x21: func(c *CPU) { c.regs.SetHL(c.fetchu16()) }, // LD HL, d16
	0x26: func(c *CPU) { c.regs.H = c.fetchu8() },     // LD H, d8
	0x2E: func(c *CPU) { c.regs.L = c.fetchu8() },     // LD L, d8

	// 0x3x
	0x31: func(c *CPU) { c.SP = c.fetchu16() },                 // LD SP, d16
	0x37: func(c *CPU) { c.regs.F = (c.regs.F & 0x80) | 0x10 }, // SCF (Set Carry Flag: N=0, H=0, C=1)
	0x3E: func(c *CPU) { c.regs.A = c.fetchu8() },              // LD A, d8
	0x3F: func(c *CPU) { c.regs.F = (c.regs.F & 0x80) ^ 0x10 }, // CCF (Complement Carry Flag: N=0, H=0, C=~C)

	// 0x4x (8-bit LD B, r / C, r)
	0x40: func(c *CPU) {},
	0x41: func(c *CPU) { c.regs.B = c.regs.C },
	0x42: func(c *CPU) { c.regs.B = c.regs.D },
	0x43: func(c *CPU) { c.regs.B = c.regs.E },
	0x44: func(c *CPU) { c.regs.B = c.regs.H },
	0x45: func(c *CPU) { c.regs.B = c.regs.L },
	0x47: func(c *CPU) { c.regs.B = c.regs.A },
	0x48: func(c *CPU) { c.regs.C = c.regs.B },
	0x49: func(c *CPU) {},
	0x4A: func(c *CPU) { c.regs.C = c.regs.D },
	0x4B: func(c *CPU) { c.regs.C = c.regs.E },
	0x4C: func(c *CPU) { c.regs.C = c.regs.H },
	0x4D: func(c *CPU) { c.regs.C = c.regs.L },
	0x4F: func(c *CPU) { c.regs.C = c.regs.A },

	// 0x5x (8-bit LD D, r / E, r)
	0x50: func(c *CPU) { c.regs.D = c.regs.B },
	0x51: func(c *CPU) { c.regs.D = c.regs.C },
	0x52: func(c *CPU) {},
	0x53: func(c *CPU) { c.regs.D = c.regs.E },
	0x54: func(c *CPU) { c.regs.D = c.regs.H },
	0x55: func(c *CPU) { c.regs.D = c.regs.L },
	0x57: func(c *CPU) { c.regs.D = c.regs.A },
	0x58: func(c *CPU) { c.regs.E = c.regs.B },
	0x59: func(c *CPU) { c.regs.E = c.regs.C },
	0x5A: func(c *CPU) { c.regs.E = c.regs.D },
	0x5B: func(c *CPU) {},
	0x5C: func(c *CPU) { c.regs.E = c.regs.H },
	0x5D: func(c *CPU) { c.regs.E = c.regs.L },
	0x5F: func(c *CPU) { c.regs.E = c.regs.A },

	// 0x6x (8-bit LD H, r / L, r)
	0x60: func(c *CPU) { c.regs.H = c.regs.B },
	0x61: func(c *CPU) { c.regs.H = c.regs.C },
	0x62: func(c *CPU) { c.regs.H = c.regs.D },
	0x63: func(c *CPU) { c.regs.H = c.regs.E },
	0x64: func(c *CPU) {},
	0x65: func(c *CPU) { c.regs.H = c.regs.L },
	0x67: func(c *CPU) { c.regs.H = c.regs.A },
	0x68: func(c *CPU) { c.regs.L = c.regs.B },
	0x69: func(c *CPU) { c.regs.L = c.regs.C },
	0x6A: func(c *CPU) { c.regs.L = c.regs.D },
	0x6B: func(c *CPU) { c.regs.L = c.regs.E },
	0x6C: func(c *CPU) { c.regs.L = c.regs.H },
	0x6D: func(c *CPU) {},
	0x6F: func(c *CPU) { c.regs.L = c.regs.A },

	// 0x7x (8-bit LD A, r)
	0x78: func(c *CPU) { c.regs.A = c.regs.B },
	0x79: func(c *CPU) { c.regs.A = c.regs.C },
	0x7A: func(c *CPU) { c.regs.A = c.regs.D },
	0x7B: func(c *CPU) { c.regs.A = c.regs.E },
	0x7C: func(c *CPU) { c.regs.A = c.regs.H },
	0x7D: func(c *CPU) { c.regs.A = c.regs.L },
	0x7F: func(c *CPU) {},

	// 0xAx (XOR r)
	0xA8: func(c *CPU) {
		c.regs.A ^= c.regs.B
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xA9: func(c *CPU) {
		c.regs.A ^= c.regs.C
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xAA: func(c *CPU) {
		c.regs.A ^= c.regs.D
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xAB: func(c *CPU) {
		c.regs.A ^= c.regs.E
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xAC: func(c *CPU) {
		c.regs.A ^= c.regs.H
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xAD: func(c *CPU) {
		c.regs.A ^= c.regs.L
		if c.regs.A == 0 {
			c.regs.F = 0x80
		} else {
			c.regs.F = 0
		}
	},
	0xAF: func(c *CPU) { c.regs.A = 0; c.regs.F = 0x80 }, // XOR A

	// 0xCx
	0xC3: func(c *CPU) { c.PC = c.fetchu16() }, // JP a16
}

func (cpu *CPU) Exec(op byte) int {
	// Reference: https://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
	// Execute `op` and ret num cpu cycles
	inst := instructions[op]
	if inst == nil {
		panic(fmt.Sprintf("Unhandled op code: 0x%x\n", op))
	}
	inst(cpu)
	return OpcodeCycles[op]
}

func (cpu *CPU) Step() int {
	if cpu.halted {
		return 0
	}

	return cpu.Exec(cpu.fetchu8())
}

func (cpu *CPU) Start() {
	cpu.halted = false
}

func (cpu *CPU) Stop() {
	cpu.halted = true
}

// Not a permanent place for these, just need to expose them for main.go testing
func (cpu *CPU) LoadROM(data []byte) error {
	return cpu.bus.LoadROM(data)
}

func (cpu *CPU) GetCartName() string {
	title := ""
	for i := uint16(0x134); i < 0x142; i++ {
		chr := cpu.bus.Read(i)
		if chr != 0x00 {
			title += string(chr)
		}
	}
	return title
}
