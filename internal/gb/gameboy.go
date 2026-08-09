package gb

import (
	"math"
	"os"
	"strings"
)

type Cart struct {
	Title string
}

type Gameboy struct {
	cpu   *CPU
	mmu   *MMU
	cart  *Cart
	div   byte
	timer int
}

func (gb *Gameboy) GetCart() *Cart { return gb.cart }

func (gb *Gameboy) Play() {
	for range int(math.Pow10(9)) {
		ncycles := gb.cpu.Step()
		gb.updateTimers(ncycles)
	}
}

func NewGameboy(romPath string) (*Gameboy, error) {
	file, err := os.Open(romPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]byte, 0x8000)

	n, err := file.Read(data)
	if err != nil || n == 0 {
		return nil, err
	}

	title := string(data[0x134:0x142])
	cart := &Cart{
		Title: strings.Trim(title, "\x00"),
	}

	mmu := &MMU{}
	copy(mmu.ROM[:], data)

	cpu := NewCPU()
	cpu.Play()

	gb := &Gameboy{
		cpu:  cpu,
		mmu:  mmu,
		cart: cart,
	}

	return gb, nil
}

func (gb *Gameboy) updateTimers(cycles int) {
	gb.div += byte(cycles)
	gb.mmu.IO[0x3]++
	clkEnable := (gb.mmu.IO[0x7] & 0x3) == 0
	if clkEnable {
		gb.timer += cycles
		freq := int(gb.mmu.IO[0x07] & 0x3)
		cpt := [4]int{1024, 16, 64, 256}[freq]
		for gb.timer >= cpt {
			gb.timer -= cpt
			tima := gb.mmu.IO[0x05]
			if tima == 0xFF {
				gb.mmu.Write(0xFF05, gb.mmu.Read(0xFF06))
			} else {
				gb.mmu.Write(0xFF05, tima+1)
			}
		}
	}
}
