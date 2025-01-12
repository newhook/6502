package main

import (
	"flag"
	"fmt"
	"github.com/newhook/6502/c64/t64"
	"github.com/newhook/6502/dis/disassembler"
	"github.com/newhook/6502/dis/flow"
	"os"
	"strconv"
	"strings"
)

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
	traceFlow := flag.Bool("flow", false, "")
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

	if *traceFlow {
		if err := doflow(*inputFile, int(startAddrInt)); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	// Create and initialize CPU
	memory := &Memory{}
	len, err := LoadAndSetupBinary(memory, *inputFile, int(startAddrInt))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(disassembler.DisassembleMemory(memory, int(startAddrInt), len))
}

func doflow(name string, startAddr int) error {
	loadAddr, programData, err := t64.LoadProgramFromT64(name, 0)
	if err != nil {
		return err
	}
	memory := &Memory{}
	for i, b := range programData {
		memory.Write(loadAddr+uint16(i), b)
	}

	// Now copy the program data into the memory
	//$0900-$4900 → $8000-$C000
	src := uint16(0x900)
	end := src + 16*1024
	dst := uint16(0x8000)
	for src < end {
		memory.Write(dst, memory.Read(src))
		src++
		dst++
	}

	tracer := flow.NewFlowTracer(memory)
	tracer.AddEntryPoint(uint16(startAddr))
	instructions := tracer.TraceFlow()

	var out strings.Builder
	for _, loc := range instructions {
		out.WriteString(loc.String())
		out.WriteString("\n")
	}

	fmt.Print(out.String())
	return nil
}

func LoadAndSetupBinary(mem *Memory, filename string, startAddr int) (int, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".t64") {
		loadAddr, programData, err := t64.LoadProgramFromT64(filename, 0)
		if err != nil {
			return 0, err
		}
		for i, b := range programData {
			mem.Write(loadAddr+uint16(i), b)
		}
		fmt.Printf("loaded %d bytes at $%04X\n", len(programData), loadAddr)
		return len(programData), nil
	}

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

	return len(data), nil
}
