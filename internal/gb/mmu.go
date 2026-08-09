package gb

import (
	"errors"
	"fmt"
	"os"
)

// "Friendly" guide for my reference
// Source: https://gbdev.io/pandocs/Memory_Map.html#memory-map
//
// Label Start	 End
// ROM   0x0000  0x3FFF (Bank 00 only)
// ROM   0x4000  0x7FFF (Banks 01-NN)
// VRAM  0x8000  0x9FFF (Switchable bank 0/1 (CGB mode only))
// EXRAM 0xA000  0xBFFF (External RAM)
// WRAM  0xC000  0xCFFF (Work RAM section 1)
// WRAM  0xD000  0xDFFF (Work RAM section 2)
// ERAM  0xE000  0xFDFF (Echo RAM (Unusable))
// OAM   0xFE00  0xFE9F (Object attribute memory)
// INOP  0xFEA0  0xFEFF (Unusable)
// IO    0xFF00  0xFF7F (Hardware registers)
// HRAM  0xFF80  0xFFFE (High RAM)
// IE    0xFFFF  0xFFFF (Interrupt Flag)
type MMU struct {
	ROM   [0x8000]byte
	VRAM  [0x2000]byte
	EXRAM [0x2000]byte
	WRAM  [0x2000]byte
	OAM   [0xA0]byte
	IO    [0x80]byte
	HRAM  [0x80]byte
	IF    bool
}

type Mode int

const (
	DMG0 Mode = iota
)

var initIODMG = [0x80]byte{
	// timer / interrupt flags
	0x04: 0x1E, // DIV (0xFF04)
	0x05: 0x00, // TIMA (0xFF05)
	0x06: 0x00, // TMA (0xFF06)
	0x07: 0xF8, // TAC (0xFF07)
	0x0F: 0xE1, // IF  (0xFF0F)

	// sound registers (0xFF10 - 0xFF26)
	0x10: 0x80,
	0x11: 0xBF,
	0x12: 0xF3,
	0x14: 0xBF,
	0x16: 0x3F,
	0x17: 0x00,
	0x19: 0xBF,
	0x1A: 0x7F,
	0x1B: 0xFF,
	0x1C: 0x9F,
	0x1E: 0xBF,
	0x20: 0xFF,
	0x21: 0x00,
	0x22: 0x00,
	0x23: 0xBF,
	0x24: 0x77,
	0x25: 0xF3,
	0x26: 0xF1,

	// ??? (0xFF40 - 0xFF4B) LCD / PPU and unused sound / other defaults
	0x40: 0x91, // LCDC (0xFF40) - typical power-on value
	0x41: 0x85, // STAT (0xFF41) - typical power-on value
	0x42: 0x00, // SCY
	0x43: 0x00, // SCX
	0x45: 0x00, // LYC
	0x47: 0xFC, // BGP (0xFF47) - BG palette (DMG)
	0x48: 0xFF, // OBP0
	0x49: 0xFF, // OBP1
	0x4A: 0x00, // WY
	0x4B: 0x00, // WX
}

func (mmu *MMU) Init(mode Mode) error {
	// Force erase here in case new ROM is loaded
	empty := make([]byte, 0, 0xFFFF)
	copy(mmu.ROM[:], empty[:0x8000])
	copy(mmu.VRAM[:], empty[:0x2000])
	copy(mmu.EXRAM[:], empty[:0x2000])
	copy(mmu.WRAM[:], empty[:0x2000])
	copy(mmu.OAM[:], empty[:0x0A])
	mmu.IF = false

	switch mode {
	case DMG0:
		copy(mmu.IO[:], initIODMG[:])
	default:
		return fmt.Errorf("unsupported Mode: %d", mode)
	}
	return nil
}

const (
	SB   = 0xFF01
	SC   = 0xFF02
	DIV  = 0xFF03
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
)

func (mmu *MMU) Write(addr uint16, val byte) {
	switch {
	case addr < 0x8000: // ROM
		return
	case addr >= 0x8000 && addr <= 0x9FFF: // VRAM
		mmu.VRAM[addr-0x8000] = val
	case addr >= 0xA000 && addr <= 0xBFFF: // External RAM
		mmu.EXRAM[addr-0xA000] = val
	case addr >= 0xC000 && addr <= 0xDFFF: // WRAM
		mmu.WRAM[addr-0xC000] = val
	case addr >= 0xFE00 && addr <= 0xFE9F: // OAM
		mmu.OAM[addr-0xFE00] = val
	case addr >= 0xFF00 && addr <= 0xFF7F: // IO
		switch addr {
		case SB:
			mmu.IO[SC-0xFF00] |= 0x80
		case SC:
			if val == 0x81 {
				fmt.Fprintf(os.Stderr, "%c", mmu.Read(SB))
			}
		case DIV:
			mmu.IO[DIV-0xFF00] = 0 // RESET DIV
		case TIMA:
			mmu.IO[TIMA-0xFF00] = val
		case TMA:
			mmu.IO[TMA-0xFF00] = val
		case TAC:
			mmu.IO[TAC-0xFF00] = val | 0xF8
		default:
			mmu.IO[addr-0xFF00] = val
		}
	case addr >= 0xFF80 && addr <= 0xFFFE: // HRAM
		mmu.HRAM[addr-0xFF80] = val
	}
}

func (mmu *MMU) Read(addr uint16) byte {
	switch {
	case addr < 0x8000: // ROM
		return mmu.ROM[addr]
	case addr >= 0x8000 && addr <= 0x9FFF: // VRAM
		return mmu.VRAM[addr-0x8000]
	case addr >= 0xA000 && addr <= 0xBFFF: // External RAM
		return mmu.EXRAM[addr-0xA000]
	case addr >= 0xC000 && addr <= 0xDFFF: // WRAM
		return mmu.WRAM[addr-0xC000]
	case addr >= 0xFE00 && addr <= 0xFE9F: // OAM
		return mmu.OAM[addr-0xFE00]
	case addr >= 0xFF00 && addr <= 0xFF7F: // IO
		return mmu.IO[addr-0xFF00]
	case addr >= 0xFF80 && addr <= 0xFFFE: // HRAM
		return mmu.HRAM[addr-0xFF80]
	case addr == 0xFFFF: // Interrupt Enable
		if mmu.IF {
			return 1
		} else {
			return 0
		}
	default:
		return 0xFF
	}
}

func (m *MMU) LoadROM(data []byte) error {
	if len(data) > len(m.ROM) {
		return errors.New("ROM data exceeds 32kb")
	}

	copy(m.ROM[:], data)

	return nil
}
