package gb

import (
	"testing"
	"unsafe"
)

type mockMMU struct {
	memory [0x10000]byte
}

func (m *mockMMU) Read(addr uint16) byte {
	return m.memory[addr]
}

func TestCPUOps(t *testing.T) {
	tests := []struct {
		name        string
		bytecode    []byte
		initialRegs Registers
		wantCycles  int
		wantPC      uint16
		wantRegs    Registers
	}{
		{
			name:       "0x00 NOP",
			bytecode:   []byte{0x00},
			wantCycles: 4,
			wantPC:     0x0101,
			wantRegs:   DMG,
		},
		{
			name:       "0x3E LD A, d8",
			bytecode:   []byte{0x3E, 0x42},
			wantCycles: 8,
			wantPC:     0x0102,
			wantRegs:   Registers{A: 0x42, F: 0xB0, C: 0x13, E: 0xD8, H: 0x01, L: 0x4D},
		},
		{
			name:       "0x01 LD BC, d16",
			bytecode:   []byte{0x01, 0x34, 0x12}, // Little-endian: 0x1234
			wantCycles: 12,
			wantPC:     0x0103,
			wantRegs:   Registers{A: 0x01, F: 0xB0, B: 0x12, C: 0x34, E: 0xD8, H: 0x01, L: 0x4D},
		},
		{
			name:       "0x11 LD DE, d16",
			bytecode:   []byte{0x11, 0x78, 0x56}, // Little-endian: 0x5678
			wantCycles: 12,
			wantPC:     0x0103,
			wantRegs:   Registers{A: 0x01, F: 0xB0, C: 0x13, D: 0x56, E: 0x78, H: 0x01, L: 0x4D},
		},
		{
			name:        "0xAF XOR A",
			bytecode:    []byte{0xAF},
			initialRegs: Registers{A: 0xFF, F: 0x00},
			wantCycles:  4,
			wantPC:      0x0101,
			wantRegs:    Registers{A: 0x00, F: 0x80}, // Z flag set
		},
		{
			name:       "0xC3 JP a16",
			bytecode:   []byte{0xC3, 0x50, 0x02}, // Jump to 0x0250
			wantCycles: 16,
			wantPC:     0x0250,
			wantRegs:   DMG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPU()
			cpu.halted = false

			if tt.initialRegs != (Registers{}) {
				cpu.regs = tt.initialRegs
			}

			// Load opcode bytes into memory starting at default PC (0x0100)
			mem := &mockMMU{}
			copy(mem.memory[0x0100:], tt.bytecode)
			cpu.bus = (*MMU)(unsafe.Pointer(mem)) // Ensure bus read works

			cycles := cpu.Step()

			if cycles != tt.wantCycles {
				t.Errorf("cycles = %d, want %d", cycles, tt.wantCycles)
			}
			if cpu.PC != tt.wantPC {
				t.Errorf("PC = 0x%04X, want 0x%04X", cpu.PC, tt.wantPC)
			}
			if cpu.regs != tt.wantRegs {
				t.Errorf("regs = %+v, want %+v", cpu.regs, tt.wantRegs)
			}
		})
	}
}
