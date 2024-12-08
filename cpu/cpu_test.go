package cpu_test

import (
	"github.com/newhook/6502/cpu"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Memory [65536]uint8

func (c *Memory) Read(address uint16) uint8 {
	return c[address]
}
func (c *Memory) Write(address uint16, value uint8) {
	c[address] = value
}

func TestCPUMemoryIntegration(t *testing.T) {
	mem := &Memory{}
	c := cpu.NewCPU(mem)

	// Write a simple program to memory
	mem.Write(0x0200, 0xA9) // LDA #$42
	mem.Write(0x0201, 0x42)
	mem.Write(0x0202, 0x00) // BRK

	// Set PC to start of program
	c.PC = 0x0200

	// Execute instruction
	c.Step()

	// Verify results
	assert.Equal(t, uint8(0x42), c.A)
}
