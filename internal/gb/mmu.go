package gb

import (
	"errors"
	"fmt"
	"os"
)

// Special registers
const (
	SB   = 0xFF01
	SC   = 0xFF02
	DIV  = 0xFF04
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
)

type MMU struct {
	ROM  [0x8000]byte
	VRAM [0x2000]byte
	WRAM [0x2000]byte
	IO   [0x80]byte
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
	case addr >= 0xFF00 && addr <= 0xFF7F: // IO
		m.IO[addr-0xFF00] = val
		switch addr {
		case SC:
			// Transfer requested
			if val == 0x81 {
				fmt.Fprintf(os.Stderr, "%c", m.Read(SB))
				m.IO[SC-0xFF00] |= 0x80 // ack receipt by setting bit 7
			}
		case DIV:
			m.IO[DIV-0xFF00] = 0 // RESET DIV
		}

		if addr == 0xFF04 {
			m.IO[addr-0xFF00] = 0 // RESET DIV
		}
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
	case addr >= 0xFF00 && addr <= 0xFF7F: // IO
		return m.IO[addr-0xFF00]
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
