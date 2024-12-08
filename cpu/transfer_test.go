package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterTransfers(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name        string
		opcode      uint8
		setup       func(*CPUAndMemory)
		verify      func(*CPUAndMemory) bool
		expectZ     bool
		expectN     bool
		affectFlags bool
	}{
		{
			name:   "TAX - Transfer zero",
			opcode: TAX,
			setup: func(c *CPUAndMemory) {
				c.A = 0x00
				c.X = 0xFF // Pre-set X to ensure it changes
			},
			verify: func(c *CPUAndMemory) bool {
				return c.X == 0x00 && c.A == 0x00
			},
			expectZ:     true,
			expectN:     false,
			affectFlags: true,
		},
		{
			name:   "TAX - Transfer negative",
			opcode: TAX,
			setup: func(c *CPUAndMemory) {
				c.A = 0x80
				c.X = 0x00
			},
			verify: func(c *CPUAndMemory) bool {
				return c.X == 0x80 && c.A == 0x80
			},
			expectZ:     false,
			expectN:     true,
			affectFlags: true,
		},
		{
			name:   "TAY - Transfer positive",
			opcode: TAY,
			setup: func(c *CPUAndMemory) {
				c.A = 0x40
				c.Y = 0x00
			},
			verify: func(c *CPUAndMemory) bool {
				return c.Y == 0x40 && c.A == 0x40
			},
			expectZ:     false,
			expectN:     false,
			affectFlags: true,
		},
		{
			name:   "TXA - Transfer zero",
			opcode: TXA,
			setup: func(c *CPUAndMemory) {
				c.X = 0x00
				c.A = 0xFF
			},
			verify: func(c *CPUAndMemory) bool {
				return c.A == 0x00 && c.X == 0x00
			},
			expectZ:     true,
			expectN:     false,
			affectFlags: true,
		},
		{
			name:   "TYA - Transfer negative",
			opcode: TYA,
			setup: func(c *CPUAndMemory) {
				c.Y = 0xFF
				c.A = 0x00
			},
			verify: func(c *CPUAndMemory) bool {
				return c.A == 0xFF && c.Y == 0xFF
			},
			expectZ:     false,
			expectN:     true,
			affectFlags: true,
		},
		{
			name:   "TSX - Transfer stack pointer",
			opcode: TSX,
			setup: func(c *CPUAndMemory) {
				c.SP = 0x7F
				c.X = 0x00
			},
			verify: func(c *CPUAndMemory) bool {
				return c.X == 0x7F && c.SP == 0x7F
			},
			expectZ:     false,
			expectN:     false,
			affectFlags: true,
		},
		{
			name:   "TXS - Transfer to stack pointer",
			opcode: TXS,
			setup: func(c *CPUAndMemory) {
				c.X = 0xFF
				c.SP = 0x00
				c.P = 0x00 // Clear flags
			},
			verify: func(c *CPUAndMemory) bool {
				return c.SP == 0xFF && c.X == 0xFF
			},
			expectZ:     false,
			expectN:     false,
			affectFlags: false, // TXS doesn't affect flags
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			cpu.P = 0x00 // Clear flags
			test.setup(cpu)

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(uint8(2), cycles, "incorrect cycle count")
			assert.True(test.verify(cpu), "register transfer failed")

			if test.affectFlags {
				assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
				assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
			} else {
				// For TXS, flags should be unchanged
				assert.Equal(uint8(0), cpu.P, "flags should not be affected")
			}
		})
	}
}

func TestTransferEdgeCases(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name   string
		setup  func(*CPUAndMemory)
		verify func(*CPUAndMemory) bool
	}{
		{
			name: "Multiple transfers preserve values",
			setup: func(c *CPUAndMemory) {
				c.A = 0x42
				c.Memory[0x0200] = TAX
				c.Memory[0x0201] = TAY
			},
			verify: func(c *CPUAndMemory) bool {
				return c.A == 0x42 && c.X == 0x42 && c.Y == 0x42
			},
		},
		{
			name: "Stack pointer circular transfer",
			setup: func(c *CPUAndMemory) {
				c.X = 0x55
				c.Memory[0x0200] = TXS
				c.Memory[0x0201] = TSX
			},
			verify: func(c *CPUAndMemory) bool {
				return c.X == 0x55 && c.SP == 0x55
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			test.setup(cpu)

			// Execute multiple instructions
			cpu.Step()
			cpu.Step()

			// Verify final state
			assert.True(test.verify(cpu))
		})
	}
}
