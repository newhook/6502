package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSTXInstructions(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name   string
		opcode uint8
		setup  func(*CPUAndMemory)
		addr   uint16
		cycles uint8
	}{
		{
			name:   "STX Zero Page",
			opcode: STX_ZP,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.X = 0x37              // Value to store
			},
			addr:   0x42,
			cycles: 3,
		},
		{
			name:   "STX Zero Page,Y",
			opcode: STX_ZPY,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Y = 0x02              // Y offset
				c.X = 0x37              // Value to store
			},
			addr:   0x44, // 0x42 + 0x02
			cycles: 4,
		},
		{
			name:   "STX Absolute",
			opcode: STX_ABS,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte
				c.Memory[0x0202] = 0x12 // High byte
				c.X = 0x37              // Value to store
			},
			addr:   0x1234,
			cycles: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			test.setup(cpu)

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(cpu.X, cpu.Memory[test.addr], "value not stored correctly")
		})
	}
}
