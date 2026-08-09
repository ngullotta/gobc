package gb

import (
	"fmt"
	"os"
)

// Friendly guide for my reference
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
// IME    0xFFFF  0xFFFF (Interrupt Master Enable)
type MMU struct {
	ROM   [0x8000]byte
	VRAM  [0x2000]byte
	EXRAM [0x2000]byte
	WRAM  [0x2000]byte
	OAM   [0xA0]byte
	IO    [0x80]byte
	HRAM  [0x80]byte
	IME   byte
}

// Hardware registers
const (
	JOYP = 0xFF00
	SB   = 0xFF01
	SC   = 0xFF02
	DIV  = 0xFF03
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
	IF   = 0xFF0F
)

func (mmu *MMU) writeIO(addr uint16, val byte) {
	switch addr {
	case SB:
		mmu.IO[addr-0xFF00] = val
	case SC:
		if val == 0x81 {
			fmt.Fprintf(os.Stderr, "%c", mmu.Read(SB))
		}
	case DIV:
		mmu.IO[DIV-0xFF00] = val
	case TIMA:
		mmu.IO[TIMA-0xFF00] = val
	case TMA:
		mmu.IO[TMA-0xFF00] = val
	case TAC:
		mmu.IO[TAC-0xFF00] = val
	case IF:
		mmu.IO[IF-0xFF00] = val
	default:
		mmu.IO[addr-0xFF00] = val
		fmt.Fprintf(os.Stderr, "Unhandled Write [IO]: 0x%X -> 0x%X\n", addr, val)
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
	default:
		fmt.Fprintf(os.Stderr, "Unhandled Read: 0x%X\n", addr)
		return 0xFF
	}
}

func (mmu *MMU) Write(addr uint16, val byte) {
	switch {
	case addr < 0x8000: // ROM
		return
	case addr >= 0x8000 && addr <= 0x9FFF: // VRAM
		mmu.VRAM[addr-0x8000] = val
		return
	case addr >= 0xA000 && addr <= 0xBFFF: // External RAM
		mmu.EXRAM[addr-0xA000] = val
		return
	case addr >= 0xC000 && addr <= 0xDFFF: // WRAM
		mmu.WRAM[addr-0xC000] = val
		return
	case addr >= 0xFE00 && addr <= 0xFE9F: // OAM
		mmu.OAM[addr-0xFE00] = val
		return
	case addr >= 0xFF00 && addr <= 0xFF7F: // IO
		mmu.writeIO(addr, val)
		return
	case addr >= 0xFF80 && addr <= 0xFFFE: // HRAM
		mmu.HRAM[addr-0xFF80] = val
		return
	case addr == 0xFFFF: // IME
		mmu.IME = val
	default:
		fmt.Fprintf(os.Stderr, "Unhandled Write: 0x%X -> 0x%X\n", addr, val)
	}
}
