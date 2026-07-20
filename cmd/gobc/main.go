package main

import (
	"fmt"
	"gobc/internal/gb"
	"os"
)

func main() {
	file, err := os.Open(os.Args[1])

	if err != nil {
		panic(err)
	}
	defer file.Close()

	data := make([]byte, 0x8000)
	file.Read(data)

	cpu := gb.NewCPU()
	cpu.LoadROM(data)

	fmt.Printf("Loaded Cart: %q\n", cpu.GetCartName())

	// Debug stuff
	gb.InitCPUInstructions()
	cpu.Debug()

	cpu.Play()
	for range 32 {
		cpu.Step()
	}
}
