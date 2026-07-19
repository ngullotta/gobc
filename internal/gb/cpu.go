package gb

type CPU struct {
	registers Registers

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
		registers: DMG,
		SP:        0xFFFE,
		PC:        0x0100,
		halted:    true,
	}
}
