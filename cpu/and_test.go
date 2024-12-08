package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestANDInstructions(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name        string
		opcode      uint8
		accumulator uint8
		operand     uint8
		expected    uint8
		setup       func(*CPUAndMemory)
		cycles      uint8
		expectZ     bool
		expectN     bool
	}{
		{
			name:        "AND Immediate - Basic AND operation",
			opcode:      AND_IMM,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x0F
			},
			cycles:  2,
			expectZ: false,
			expectN: false,
		},
		{
			name:        "AND Immediate - Result zero",
			opcode:      AND_IMM,
			accumulator: 0xFF,
			operand:     0x00,
			expected:    0x00,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x00
			},
			cycles:  2,
			expectZ: true,
			expectN: false,
		},
		{
			name:        "AND Immediate - Result negative",
			opcode:      AND_IMM,
			accumulator: 0xFF,
			operand:     0x80,
			expected:    0x80,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x80
			},
			cycles:  2,
			expectZ: false,
			expectN: true,
		},
		{
			name:        "AND Zero Page",
			opcode:      AND_ZP,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0x0F // Operand
			},
			cycles:  3,
			expectZ: false,
			expectN: false,
		},
		{
			name:        "AND Zero Page,X",
			opcode:      AND_ZPX,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.X = 0x02              // X offset
				c.Memory[0x0044] = 0x0F // Operand at (0x42 + 0x02)
			},
			cycles:  4,
			expectZ: false,
			expectN: false,
		},
		{
			name:        "AND Absolute",
			opcode:      AND_ABS,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.Memory[0x1234] = 0x0F // Operand
			},
			cycles:  4,
			expectZ: false,
			expectN: false,
		},
		{
			name:        "AND Absolute,X without page cross",
			opcode:      AND_ABX,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.X = 0x01              // X offset
				c.Memory[0x1235] = 0x0F // Operand at (0x1234 + 0x01)
			},
			cycles:  4,
			expectZ: false,
			expectN: false,
		},
		{
			name:        "AND Absolute,X with page cross",
			opcode:      AND_ABX,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0x0F,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0xFF // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.X = 0x01              // X offset causing page cross
				c.Memory[0x1300] = 0x0F // Operand at (0x12FF + 0x01)
			},
			cycles:  5,
			expectZ: false,
			expectN: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			cpu.A = test.accumulator
			test.setup(cpu)

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.expected, cpu.A, "incorrect AND result")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
		})
	}
}

func TestANDIndirectModes(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name      string
		opcode    uint8
		setup     func(*CPUAndMemory)
		expected  uint8
		cycles    uint8
		pageCross bool
	}{
		{
			name:   "AND Indirect,X",
			opcode: AND_INX,
			setup: func(c *CPUAndMemory) {
				c.A = 0xFF
				c.X = 0x02
				c.Memory[0x0201] = 0x20 // Zero page address
				// Effective address: 0x20 + 0x02 = 0x22
				c.Memory[0x0022] = 0x34 // Low byte of indirect address
				c.Memory[0x0023] = 0x12 // High byte of indirect address
				c.Memory[0x1234] = 0x0F // Operand
			},
			expected: 0x0F,
			cycles:   6,
		},
		{
			name:   "AND Indirect,Y without page cross",
			opcode: AND_INY,
			setup: func(c *CPUAndMemory) {
				c.A = 0xFF
				c.Y = 0x02
				c.Memory[0x0201] = 0x20 // Zero page address
				c.Memory[0x0020] = 0x34 // Low byte of indirect address
				c.Memory[0x0021] = 0x12 // High byte of indirect address
				// Effective address: 0x1234 + 0x02 = 0x1236
				c.Memory[0x1236] = 0x0F // Operand
			},
			expected: 0x0F,
			cycles:   5,
		},
		{
			name:   "AND Indirect,Y with page cross",
			opcode: AND_INY,
			setup: func(c *CPUAndMemory) {
				c.A = 0xFF
				c.Y = 0xFF              // Will cause page cross
				c.Memory[0x0201] = 0x20 // Zero page address
				c.Memory[0x0020] = 0x34 // Low byte of indirect address
				c.Memory[0x0021] = 0x12 // High byte of indirect address
				// Effective address: 0x1234 + 0xFF = 0x1333
				c.Memory[0x1333] = 0x0F // Operand
			},
			expected:  0x0F,
			cycles:    6,
			pageCross: true,
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
			assert.Equal(test.expected, cpu.A, "incorrect AND result")
		})
	}
}
