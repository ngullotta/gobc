package gb

import (
	"fmt"
	"os"
	"strings"
)

type Cart struct {
	Title string
}

type Gameboy struct {
	cpu     *CPU
	mmu     *MMU
	cart    *Cart
	running bool
	div     byte
	timer   int
}

func (gb *Gameboy) Update() int {
	if !gb.running {
		return 0
	}

	cycles := 0
	cyclesOp := 4
	cyclesOp += gb.cpu.Step()
	cycles += cyclesOp

	gb.updateTimers(cyclesOp)
	gb.handleInterrupts()

	return cycles
}

func (gb *Gameboy) GetCart() *Cart { return gb.cart }

func (gb *Gameboy) Play() {
	for gb.running {
		gb.Update()
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

	cpu := NewCPU(mmu)

	gb := &Gameboy{
		cpu:     cpu,
		mmu:     mmu,
		cart:    cart,
		running: true,
		div:     0,
		timer:   0,
	}

	return gb, nil
}

func (gb *Gameboy) updateDiv(cycles int) {
	gb.div += byte(cycles)
	// Manually inc DIV in memory (write here resets to 0)
	gb.mmu.IO[DIV-0xFF00]++
}

// https://gbdev.io/pandocs/Timer_and_Divider_Registers.html
func (gb *Gameboy) updateTimers(cycles int) {
	gb.updateDiv(cycles)

	tac := gb.mmu.Read(TAC)
	clkEnable := tac & 0x3
	if clkEnable != 0 {
		gb.timer += cycles
		freq := [4]int{1024, 16, 64, 256}[clkEnable]
		for gb.timer >= freq {
			gb.timer -= freq
			tima := gb.mmu.Read(TIMA)
			if tima == 0xFF {
				gb.mmu.Write(TIMA, gb.mmu.Read(TMA))
			} else {
				gb.mmu.Write(TIMA, tima+1)
			}
		}
	}
}

func (gb *Gameboy) handleInterrupts() int {
	ime := gb.mmu.IME
	req := gb.mmu.Read(IF)
	if req > 0 {
		for i := range 5 {
			enabled := (ime>>i)&1 == 1
			requested := (req>>i)&1 == 1
			if enabled && requested {
				fmt.Fprintf(os.Stderr, "Unhandled Interrupt")
				return 20
			}
		}
	}
	return 0
}
