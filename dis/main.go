package main

import (
	"flag"
	"fmt"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/dis/disassembler"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Command line flags
	inputFile := flag.String("i", "", "Input binary file")
	startAddr := flag.String("a", "", "Start address")
	flag.Parse()

	addrStr := *startAddr
	if strings.HasPrefix(addrStr, "$") {
		addrStr = "0x" + addrStr[1:]
	}
	startAddrInt, err := strconv.ParseUint(addrStr, 0, 16)
	if err != nil {
		fmt.Printf("Error parsing start address: %v\n", err)
		return
	}

	// Create and initialize CPU
	c := cpu.NewCPU()
	len, err := LoadAndSetupBinary(c, *inputFile, int(startAddrInt))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(disassembler.DisassembleMemory(c.Memory[:], int(startAddrInt), len))
}

func LoadAndSetupBinary(c *cpu.CPU, filename string, startAddr int) (int, error) {
	// Read the binary file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read binary file: %v", err)
	}

	// Check if the binary will fit in memory
	if int(startAddr)+len(data) > len(c.Memory) {
		return 0, fmt.Errorf("binary file too large for available memory")
	}

	// Copy binary data into CPU memory starting at 0xF000
	for i, b := range data {
		c.Memory[uint16(startAddr)+uint16(i)] = b
	}

	// Set up reset vector at 0xFFFC-0xFFFD to point to 0xF000
	c.Memory[0xFFFC] = 0x00 // Low byte
	c.Memory[0xFFFD] = 0xF0 // High byte

	// Set up IRQ vector at 0xFFFE-0xFFFF to point to 0xF5A4
	c.Memory[0xFFFE] = 0xA4 // Low byte
	c.Memory[0xFFFF] = 0xF5 // High byte

	// Set the Program Counter to the reset vector location
	c.PC = uint16(startAddr)

	return len(data), nil
}
