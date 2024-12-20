package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEORInstructions(t *testing.T) {
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
			name:        "EOR Immediate - Basic XOR operation",
			opcode:      EOR_IMM,
			accumulator: 0xFF,
			operand:     0x0F,
			expected:    0xF0,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x0F
			},
			cycles:  2,
			expectZ: false,
			expectN: true,
		},
		{
			name:        "EOR Immediate - Result zero",
			opcode:      EOR_IMM,
			accumulator: 0xFF,
			operand:     0xFF,
			expected:    0x00,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0xFF
			},
			cycles:  2,
			expectZ: true,
			expectN: false,
		},
		{
			name:        "EOR Zero Page",
			opcode:      EOR_ZP,
			accumulator: 0xAA,
			operand:     0x55,
			expected:    0xFF,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0x55 // Operand
			},
			cycles:  3,
			expectZ: false,
			expectN: true,
		},
		{
			name:        "EOR Zero Page,X",
			opcode:      EOR_ZPX,
			accumulator: 0xFF,
			operand:     0xFF,
			expected:    0x00,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.X = 0x02              // X offset
				c.Memory[0x0044] = 0xFF // Operand at (0x42 + 0x02)
			},
			cycles:  4,
			expectZ: true,
			expectN: false,
		},
		{
			name:        "EOR Absolute",
			opcode:      EOR_ABS,
			accumulator: 0x0F,
			operand:     0xF0,
			expected:    0xFF,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.Memory[0x1234] = 0xF0 // Operand
			},
			cycles:  4,
			expectZ: false,
			expectN: true,
		},
		{
			name:        "EOR Absolute,X without page cross",
			opcode:      EOR_ABX,
			accumulator: 0x55,
			operand:     0xAA,
			expected:    0xFF,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.X = 0x01              // X offset
				c.Memory[0x1235] = 0xAA // Operand at (0x1234 + 0x01)
			},
			cycles:  4,
			expectZ: false,
			expectN: true,
		},
		{
			name:        "EOR Absolute,X with page cross",
			opcode:      EOR_ABX,
			accumulator: 0x55,
			operand:     0xAA,
			expected:    0xFF,
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0201] = 0xFF // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.X = 0x01              // X offset causing page cross
				c.Memory[0x1300] = 0xAA // Operand at (0x12FF + 0x01)
			},
			cycles:  5,
			expectZ: false,
			expectN: true,
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
			assert.Equal(test.expected, cpu.A, "incorrect EOR result")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
		})
	}
}

func TestEORIndirectModes(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name     string
		opcode   uint8
		setup    func(*CPUAndMemory)
		expected uint8
		cycles   uint8
		expectZ  bool
		expectN  bool
	}{
		{
			name:   "EOR Indirect,X",
			opcode: EOR_INX,
			setup: func(c *CPUAndMemory) {
				c.A = 0x55
				c.X = 0x02
				c.Memory[0x0201] = 0x20 // Zero page address
				// Effective address: 0x20 + 0x02 = 0x22
				c.Memory[0x0022] = 0x34 // Low byte of indirect address
				c.Memory[0x0023] = 0x12 // High byte of indirect address
				c.Memory[0x1234] = 0xAA // Operand
			},
			expected: 0xFF,
			cycles:   6,
			expectZ:  false,
			expectN:  true,
		},
		{
			name:   "EOR Indirect,Y without page cross",
			opcode: EOR_INY,
			setup: func(c *CPUAndMemory) {
				c.A = 0xFF
				c.Y = 0x02
				c.Memory[0x0201] = 0x20 // Zero page address
				c.Memory[0x0020] = 0x34 // Low byte of indirect address
				c.Memory[0x0021] = 0x12 // High byte of indirect address
				// Effective address: 0x1234 + 0x02 = 0x1236
				c.Memory[0x1236] = 0xFF // Operand
			},
			expected: 0x00,
			cycles:   5,
			expectZ:  true,
			expectN:  false,
		},
		{
			name:   "EOR Indirect,Y with page cross",
			opcode: EOR_INY,
			setup: func(c *CPUAndMemory) {
				c.A = 0x55
				c.Y = 0xFF              // Will cause page cross
				c.Memory[0x0201] = 0x20 // Zero page address
				c.Memory[0x0020] = 0x34 // Low byte of indirect address
				c.Memory[0x0021] = 0x12 // High byte of indirect address
				// Effective address: 0x1234 + 0xFF = 0x1333
				c.Memory[0x1333] = 0xAA // Operand
			},
			expected: 0xFF,
			cycles:   6,
			expectZ:  false,
			expectN:  true,
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
			assert.Equal(test.expected, cpu.A, "incorrect EOR result")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
		})
	}
}
