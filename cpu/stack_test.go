package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStackOperations(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name         string
		opcode       uint8
		setup        func(*CPU)
		verify       func(*CPU) bool
		cycles       uint8
		expectZ      bool
		expectN      bool
		affectsFlags bool
	}{
		{
			name:   "PHA - Push zero accumulator",
			opcode: PHA,
			setup: func(c *CPU) {
				c.A = 0x00
				c.SP = 0xFF
			},
			verify: func(c *CPU) bool {
				return c.Memory[0x01FF] == 0x00 && c.SP == 0xFE
			},
			cycles:       3,
			affectsFlags: false,
		},
		{
			name:   "PHA - Push non-zero accumulator",
			opcode: PHA,
			setup: func(c *CPU) {
				c.A = 0x42
				c.SP = 0xFF
			},
			verify: func(c *CPU) bool {
				return c.Memory[0x01FF] == 0x42 && c.SP == 0xFE
			},
			cycles:       3,
			affectsFlags: false,
		},
		{
			name:   "PHP - Push processor status",
			opcode: PHP,
			setup: func(c *CPU) {
				c.P = FlagC | FlagZ // Set some flags
				c.SP = 0xFF
			},
			verify: func(c *CPU) bool {
				// PHP always sets the B flag in the pushed value
				return c.Memory[0x01FF] == (FlagC|FlagZ|FlagB) && c.SP == 0xFE
			},
			cycles:       3,
			affectsFlags: false,
		},
		{
			name:   "PLA - Pull zero value",
			opcode: PLA,
			setup: func(c *CPU) {
				c.SP = 0xFE
				c.Memory[0x01FF] = 0x00
			},
			verify: func(c *CPU) bool {
				return c.A == 0x00 && c.SP == 0xFF
			},
			cycles:       4,
			expectZ:      true,
			expectN:      false,
			affectsFlags: true,
		},
		{
			name:   "PLA - Pull negative value",
			opcode: PLA,
			setup: func(c *CPU) {
				c.SP = 0xFE
				c.Memory[0x01FF] = 0x80
			},
			verify: func(c *CPU) bool {
				return c.A == 0x80 && c.SP == 0xFF
			},
			cycles:       4,
			expectZ:      false,
			expectN:      true,
			affectsFlags: true,
		},
		{
			name:   "PLP - Pull processor status",
			opcode: PLP,
			setup: func(c *CPU) {
				c.SP = 0xFE
				c.Memory[0x01FF] = FlagC | FlagZ // Value on stack
				c.P = FlagB | FlagN              // Current flags with B set
			},
			verify: func(c *CPU) bool {
				// PLP should preserve the current B flag status
				return (c.P & ^uint8(FlagB)) == (FlagC|FlagZ) &&
					(c.P&FlagB) != 0 &&
					c.SP == 0xFF
			},
			cycles:       4,
			affectsFlags: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = test.opcode
			test.setup(cpu)

			// Save initial flags if needed
			initialFlags := cpu.P

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.True(test.verify(cpu), "stack operation failed")

			if test.affectsFlags {
				if test.opcode != PLP {
					assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
					assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
				}
			} else {
				assert.Equal(initialFlags, cpu.P, "flags should not be affected")
			}
		})
	}
}

func TestStackEdgeCases(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name   string
		setup  func(*CPU)
		verify func(*CPU) bool
	}{
		{
			name: "Stack wrap-around",
			setup: func(c *CPU) {
				c.SP = 0x00
				c.A = 0x42
				c.Memory[0x0200] = PHA
				c.Memory[0x0201] = PLA
			},
			verify: func(c *CPU) bool {
				return c.SP == 0x00 && c.A == 0x42
			},
		},
		{
			name: "Push and pull preserves value",
			setup: func(c *CPU) {
				c.SP = 0xFF
				c.A = 0x42
				c.Memory[0x0200] = PHA
				c.Memory[0x0201] = PLA
			},
			verify: func(c *CPU) bool {
				return c.A == 0x42 && c.SP == 0xFF
			},
		},
		{
			name: "PHP/PLP preserves B flag",
			setup: func(c *CPU) {
				c.SP = 0xFF
				c.P = FlagB | FlagC
				c.Memory[0x0200] = PHP
				c.Memory[0x0201] = PLP
			},
			verify: func(c *CPU) bool {
				return (c.P&FlagB) != 0 && (c.P&FlagC) != 0
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
