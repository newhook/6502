package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSTAInstructions(t *testing.T) {
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
			name:   "STA Zero Page",
			opcode: STA_ZP,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.A = 0x37              // Value to store
			},
			addr:   0x42,
			cycles: 3,
		},
		{
			name:   "STA Zero Page,X",
			opcode: STA_ZPX,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.X = 0x02              // X offset
				c.A = 0x37              // Value to store
			},
			addr:   0x44, // 0x42 + 0x02
			cycles: 4,
		},
		{
			name:   "STA Absolute",
			opcode: STA_ABS,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte
				c.Memory[0x0202] = 0x12 // High byte
				c.A = 0x37              // Value to store
			},
			addr:   0x1234,
			cycles: 4,
		},
		{
			name:   "STA Absolute,X",
			opcode: STA_ABX,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte
				c.Memory[0x0202] = 0x12 // High byte
				c.X = 0x02              // X offset
				c.A = 0x37              // Value to store
			},
			addr:   0x1236, // 0x1234 + 0x02
			cycles: 5,
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
			if test.opcode != STA_INX && test.opcode != STA_INY {
				assert.Equal(cpu.A, cpu.Memory[test.addr], "value not stored correctly")
			}
		})
	}
}
