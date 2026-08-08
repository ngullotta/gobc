package main

import (
	"fmt"
	"gobc/internal/gb"
	"os"
	"strings"
)

type Cart struct {
	Title string
}

type Gameboy struct {
	cpu  *gb.CPU
	mmu  *gb.MMU
	cart *Cart
	mode gb.Mode
}

func NewGameboy(romPath string, mode gb.Mode) (*Gameboy, error) {
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

	mmu := &gb.MMU{}
	mmu.Init(mode)
	mmu.LoadROM(data)

	gb := &Gameboy{
		cpu:  gb.NewCPU(),
		mmu:  mmu,
		cart: cart,
		mode: mode,
	}

	return gb, nil
}

func (gb *Gameboy) GetCart() *Cart { return gb.cart }

func main() {
	gb, err := NewGameboy(os.Args[1], gb.DMG0)
	if err != nil {
		panic(err)
	}

	if gb == nil {
		panic("Failed to create gb")
	}

	cart := gb.GetCart()
	if cart == nil {
		panic("WHAT")
	}

	fmt.Printf("Loaded Cart: %q\n", gb.GetCart().Title)

	// cpu.Play()
	// for range int(math.Pow10(7)) {
	// cpu.Step()
	// }
}
