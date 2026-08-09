package main

import (
	"fmt"
	"gobc/internal/gb"
	"os"
)

func main() {
	gb, err := gb.NewGameboy(os.Args[1])
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
	gb.Play()
}
