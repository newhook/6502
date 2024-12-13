package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/mon/monitor"
	"os"
	"strconv"
	"strings"
)

func LoadAndSetupBinary(c *cpu.CPU, mem *Memory, filename string, startAddr int) (int, error) {
	// Read the binary file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read binary file: %v", err)
	}

	// Check if the binary will fit in memory
	if int(startAddr)+len(data) > len(mem) {
		return 0, fmt.Errorf("binary file too large for available memory")
	}

	// Copy binary data into CPU memory starting at 0xF000
	for i, b := range data {
		mem[uint16(startAddr)+uint16(i)] = b
	}

	// Set up reset vector at 0xFFFC-0xFFFD to point to 0xF000
	mem[0xFFFC] = 0x00 // Low byte
	mem[0xFFFD] = 0xF0 // High byte

	// Set up IRQ vector at 0xFFFE-0xFFFF to point to 0xF5A4
	mem[0xFFFE] = 0xA4 // Low byte
	mem[0xFFFF] = 0xF5 // High byte

	// Set the Program Counter to the reset vector location
	c.PC = uint16(startAddr)

	return len(data), nil
}

type Memory [65536]uint8

func (c *Memory) Read(address uint16) uint8 {
	return c[address]
}
func (c *Memory) Write(address uint16, value uint8) {
	c[address] = value
}

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
	memory := &Memory{}
	c := cpu.NewCPU(memory)
	_, err = LoadAndSetupBinary(c, memory, *inputFile, int(startAddrInt))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	p := tea.NewProgram(monitor.NewMonitor(c, c, memory))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
