package gb

import "errors"

type MMU struct {
	ROM  [0x8000]byte
	VRAM [0x2000]byte
	WRAM [0x2000]byte
	HRAM [0x80]byte
}

func (m *MMU) Write(addr uint16, val byte) {
	switch {
	case addr < 0x8000: // ROM
		return
	case addr >= 0x8000 && addr <= 0x9FFF: // VRAM
		m.VRAM[addr-0x8000] = val
	case addr >= 0xC000 && addr <= 0xDFFF: // WRAM
		m.WRAM[addr-0xC000] = val
	case addr >= 0xFF80 && addr <= 0xFFFE: // HRAM
		m.HRAM[addr-0xFF80] = val
	}
}

func (m *MMU) Read(addr uint16) byte {
	switch {
	case addr < 0x8000: // ROM
		return m.ROM[addr]
	case addr >= 0x8000 && addr <= 0x9FFF: // VRAM
		return m.VRAM[addr-0x8000]
	case addr >= 0xC000 && addr <= 0xDFFF: // WRAM
		return m.WRAM[addr-0xC000]
	case addr >= 0xFF80 && addr <= 0xFFFE: // HRAM
		return m.HRAM[addr-0xFF80]
	}
	return 0
}

func (m *MMU) LoadROM(data []byte) error {
	if len(data) > len(m.ROM) {
		return errors.New("ROM data exceeds 32kb")
	}

	copy(m.ROM[:], data)

	return nil
}
