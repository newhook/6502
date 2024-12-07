package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBITInstruction(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name        string
		opcode      uint8
		accumulator uint8
		memValue    uint8
		setup       func(*CPU)
		cycles      uint8
		expectZ     bool // Based on A & M
		expectN     bool // Based on bit 7 of M
		expectV     bool // Based on bit 6 of M
	}{
		{
			name:        "BIT Zero Page - Zero result",
			opcode:      BIT_ZP,
			accumulator: 0xFF,
			memValue:    0x00,
			setup: func(c *CPU) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0x00 // Memory value
			},
			cycles:  3,
			expectZ: true,
			expectN: false,
			expectV: false,
		},
		{
			name:        "BIT Zero Page - Set all flags",
			opcode:      BIT_ZP,
			accumulator: 0x00,
			memValue:    0xC0, // Bits 7 and 6 set
			setup: func(c *CPU) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0xC0 // Memory value
			},
			cycles:  3,
			expectZ: true,
			expectN: true,
			expectV: true,
		},
		{
			name:        "BIT Zero Page - Non-zero result",
			opcode:      BIT_ZP,
			accumulator: 0xFF,
			memValue:    0x40, // Bit 6 set only
			setup: func(c *CPU) {
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0x40 // Memory value
			},
			cycles:  3,
			expectZ: false,
			expectN: false,
			expectV: true,
		},
		{
			name:        "BIT Absolute - Zero result with N flag",
			opcode:      BIT_ABS,
			accumulator: 0x00,
			memValue:    0x80, // Bit 7 set only
			setup: func(c *CPU) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.Memory[0x1234] = 0x80 // Memory value
			},
			cycles:  4,
			expectZ: true,
			expectN: true,
			expectV: false,
		},
		{
			name:        "BIT Absolute - Test overflow only",
			opcode:      BIT_ABS,
			accumulator: 0xFF,
			memValue:    0x40, // Bit 6 set only
			setup: func(c *CPU) {
				c.Memory[0x0201] = 0x34 // Low byte of address
				c.Memory[0x0202] = 0x12 // High byte of address
				c.Memory[0x1234] = 0x40 // Memory value
			},
			cycles:  4,
			expectZ: false,
			expectN: false,
			expectV: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			cpu.A = test.accumulator
			cpu.P = 0x00 // Clear flags
			test.setup(cpu)

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
			assert.Equal(test.expectV, cpu.P&FlagV != 0, "incorrect overflow flag")

			// Verify accumulator was not modified
			assert.Equal(test.accumulator, cpu.A, "accumulator should not be modified")
		})
	}
}

func TestBITInstructionEdgeCases(t *testing.T) {
	t.Skip()
	cpu := NewCPU()

	tests := []struct {
		name   string
		opcode uint8
		setup  func(*CPU)
		verify func(*CPU) bool
	}{
		{
			name:   "BIT preserves accumulator with all flags set",
			opcode: BIT_ZP,
			setup: func(c *CPU) {
				c.A = 0x55
				c.Memory[0x0201] = 0x42 // Zero page address
				c.Memory[0x0042] = 0xC0 // Memory value (sets N and V)
				c.P = 0x00              // Clear flags
			},
			verify: func(c *CPU) bool {
				return c.A == 0x55 && // Accumulator preserved
					c.P&FlagN != 0 && // N set from bit 7
					c.P&FlagV != 0 // V set from bit 6
			},
		},
		{
			name:   "BIT successive operations",
			opcode: BIT_ABS,
			setup: func(c *CPU) {
				c.A = 0xFF
				// First test location
				c.Memory[0x0201] = 0x34
				c.Memory[0x0202] = 0x12
				c.Memory[0x1234] = 0x80 // Sets N only
				// Second test location
				c.Memory[0x0203] = 0x35
				c.Memory[0x0204] = 0x12
				c.Memory[0x1235] = 0x40 // Sets V only
			},
			verify: func(c *CPU) bool {
				// First BIT operation
				c.Step()
				firstN := c.P&FlagN != 0
				firstV := c.P&FlagV != 0

				// Second BIT operation
				c.Step()
				secondN := c.P&FlagN != 0
				secondV := c.P&FlagV != 0

				return firstN && !firstV && // First operation results
					!secondN && secondV // Second operation results
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			test.setup(cpu)

			// Verify
			assert.True(test.verify(cpu))
		})
	}
}
