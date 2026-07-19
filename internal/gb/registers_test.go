package gb

import (
	"testing"
)

func TestRegisterPairs(t *testing.T) {
	tests := []struct {
		name     string
		setPair  func(r *Registers, val uint16)
		getPair  func(r *Registers) uint16
		highReg  func(r *Registers) uint8
		lowReg   func(r *Registers) uint8
		input    uint16
		wantHigh uint8
		wantLow  uint8
		wantPair uint16
	}{
		{
			name:     "BC basic",
			setPair:  (*Registers).SetBC,
			getPair:  (*Registers).GetBC,
			highReg:  func(r *Registers) uint8 { return r.B },
			lowReg:   func(r *Registers) uint8 { return r.C },
			input:    0xCAFE,
			wantHigh: 0xCA,
			wantLow:  0xFE,
			wantPair: 0xCAFE,
		},
		{
			name:     "BC zero",
			setPair:  (*Registers).SetBC,
			getPair:  (*Registers).GetBC,
			highReg:  func(r *Registers) uint8 { return r.B },
			lowReg:   func(r *Registers) uint8 { return r.C },
			input:    0x0000,
			wantHigh: 0x00,
			wantLow:  0x00,
			wantPair: 0x0000,
		},
		{
			name:     "DE basic",
			setPair:  (*Registers).SetDE,
			getPair:  (*Registers).GetDE,
			highReg:  func(r *Registers) uint8 { return r.D },
			lowReg:   func(r *Registers) uint8 { return r.E },
			input:    0xBEB0,
			wantHigh: 0xBE,
			wantLow:  0xB0,
			wantPair: 0xBEB0,
		},
		{
			name:     "DE zero",
			setPair:  (*Registers).SetDE,
			getPair:  (*Registers).GetDE,
			highReg:  func(r *Registers) uint8 { return r.D },
			lowReg:   func(r *Registers) uint8 { return r.E },
			input:    0x0000,
			wantHigh: 0x00,
			wantLow:  0x00,
			wantPair: 0x0000,
		},
		{
			name:     "HL basic",
			setPair:  (*Registers).SetHL,
			getPair:  (*Registers).GetHL,
			highReg:  func(r *Registers) uint8 { return r.H },
			lowReg:   func(r *Registers) uint8 { return r.L },
			input:    0xC00F,
			wantHigh: 0xC0,
			wantLow:  0x0F,
			wantPair: 0xC00F,
		},
		{
			name:     "HL zero",
			setPair:  (*Registers).SetHL,
			getPair:  (*Registers).GetHL,
			highReg:  func(r *Registers) uint8 { return r.H },
			lowReg:   func(r *Registers) uint8 { return r.L },
			input:    0x0000,
			wantHigh: 0x00,
			wantLow:  0x00,
			wantPair: 0x0000,
		},
		{
			name:     "AF basic",
			setPair:  (*Registers).SetAF,
			getPair:  (*Registers).GetAF,
			highReg:  func(r *Registers) uint8 { return r.A },
			lowReg:   func(r *Registers) uint8 { return r.F },
			input:    0x5A5A,
			wantHigh: 0x5A,
			wantLow:  0x50, // The lower 4 bits are always zero, balme nintendo
			wantPair: 0x5A50,
		},
		{
			name:     "AF zero",
			setPair:  (*Registers).SetAF,
			getPair:  (*Registers).GetAF,
			highReg:  func(r *Registers) uint8 { return r.A },
			lowReg:   func(r *Registers) uint8 { return r.F },
			input:    0x0000,
			wantHigh: 0x00,
			wantLow:  0x00,
			wantPair: 0x0000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Registers

			tt.setPair(&r, tt.input)

			if got := tt.highReg(&r); got != tt.wantHigh {
				t.Errorf(
					"High register mismatch: got 0x%02X, want 0x%02X",
					got,
					tt.wantHigh,
				)
			}

			if got := tt.lowReg(&r); got != tt.wantLow {
				t.Errorf(
					"Low register mismatch: got 0x%02X, want 0x%02X",
					got,
					tt.wantLow,
				)
			}

			if got := tt.getPair(&r); got != tt.wantPair {
				t.Errorf(
					"Combined 16-bit pair mismatch: got 0x%04X, want 0x%04X",
					got,
					tt.wantPair,
				)
			}
		})
	}
}
