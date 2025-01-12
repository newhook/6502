package main

import (
	"fmt"
	"github.com/newhook/6502/c64/d64"
)

func main() {
	d, err := d64.LoadD64("/Users/matthew/6502/6502/c64emu/roms/LodeRunner.d64")
	if err != nil {
		panic(err)
	}

	entries, err := d.ReadDirectory()
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		fmt.Printf("%+v\n", e)
	}
}
