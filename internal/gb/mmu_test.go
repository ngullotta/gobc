package gb

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestRW(t *testing.T) {
	mmu := &MMU{}

	rom := make([]byte, 0x8000)
	rand.Read(mmu.ROM[:])

	if err := mmu.LoadROM(rom); err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	if !bytes.Equal(rom, mmu.ROM[:]) {
		t.Fatalf("Invalid MMU ROM")
	}

	memTests := []struct {
		name string
		addr uint16
		val  byte
	}{
		{"VRAM Start", 0x8000, 0xAA},
		{"VRAM End", 0x9FFF, 0xBB},
		{"WRAM Start", 0xC000, 0xCC},
		{"WRAM End", 0xDFFF, 0xDD},
		{"HRAM Start", 0xFF80, 0xEE},
		{"HRAM End", 0xFFFE, 0xFF},
	}

	for _, tt := range memTests {
		t.Run(tt.name, func(t *testing.T) {
			mmu.Write(tt.addr, tt.val)
			if got := mmu.Read(tt.addr); got != tt.val {
				t.Errorf(
					"Address 0x%04X: got 0x%02X, want 0x%02X",
					tt.addr,
					got,
					tt.val,
				)
			}
		})
	}
}
