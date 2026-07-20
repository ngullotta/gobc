package gb

import "fmt"

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

func (cpu *CPU) Exec(op byte) int {
	// Reference: https://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
	// Execute `op` and ret num cpu cycles
	switch op {
	case 0x00: // NOP
		return 4

	case 0x3E: // LD A, d8
		cpu.regs.A = cpu.fetchu8()
		return 8

	case 0x01: // LD BC, d16
		cpu.regs.SetBC(cpu.fetchu16())
		return 12

	case 0x11: // LD DE, d16
		cpu.regs.SetDE(cpu.fetchu16())
		return 12

	case 0xAF: // XOR A
		cpu.regs.A ^= cpu.regs.A
		// Update flags (Z=1, N=0, H=0, C=0)
		cpu.regs.F = 0x80
		return 4

	case 0xC3: // JP a16
		cpu.PC = cpu.fetchu16()
		return 16

	default:
		panic(fmt.Sprintf("Unhandled opcode: 0x%02X at PC: 0x%04X", op, cpu.PC-1))
	}
}

func (cpu *CPU) Step() int {
	if cpu.halted {
		return 0
	}

	cpu.PC++

	op := cpu.bus.Read(cpu.PC)

	return cpu.Exec(op)
}
