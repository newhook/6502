package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestROR(t *testing.T) {
	tests := []struct {
		name     string
		opcode   uint8
		setup    func(*CPU, uint8)
		cycles   uint8
		getValue func(*CPU) uint8
	}{
		{
			name:   "ROR Accumulator",
			opcode: ROR_ACC,
			setup: func(c *CPU, value uint8) {
				c.A = value
			},
			cycles: 2,
			getValue: func(c *CPU) uint8 {
				return c.A
			},
		},
		{
			name:   "ROR Zero Page",
			opcode: ROR_ZP,
			setup: func(c *CPU, value uint8) {
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 5,
			getValue: func(c *CPU) uint8 {
				return c.Memory[0x42]
			},
		},
		{
			name:   "ROR Zero Page,X",
			opcode: ROR_ZPX,
			setup: func(c *CPU, value uint8) {
				c.Memory[1] = 0x42     // Zero page address
				c.X = 0x02             // X offset
				c.Memory[0x44] = value // 0x42 + 0x02
			},
			cycles: 6,
			getValue: func(c *CPU) uint8 {
				return c.Memory[0x44]
			},
		},
		{
			name:   "ROR Absolute",
			opcode: ROR_ABS,
			setup: func(c *CPU, value uint8) {
				c.Memory[1] = 0x80 // Low byte
				c.Memory[2] = 0x12 // High byte
				c.Memory[0x1280] = value
			},
			cycles: 6,
			getValue: func(c *CPU) uint8 {
				return c.Memory[0x1280]
			},
		},
		{
			name:   "ROR Absolute,X",
			opcode: ROR_ABX,
			setup: func(c *CPU, value uint8) {
				c.Memory[1] = 0x80 // Low byte
				c.Memory[2] = 0x12 // High byte
				c.X = 0x02
				c.Memory[0x1282] = value // 0x1280 + 0x02
			},
			cycles: 7,
			getValue: func(c *CPU) uint8 {
				return c.Memory[0x1282]
			},
		},
	}

	testCases := []struct {
		value    uint8
		carryIn  bool
		expected uint8
		expectC  bool
		expectZ  bool
		expectN  bool
		desc     string
	}{
		{0xAA, false, 0x55, false, false, false, "No carry in, regular shift"},
		{0x55, false, 0x2A, true, false, false, "Carry out, positive result"},
		{0x00, true, 0x80, false, false, true, "Carry in to bit 7"},
		{0x01, true, 0x80, true, false, true, "Carry in and out"},
		{0x00, false, 0x00, false, true, false, "Zero result"},
	}

	for _, tt := range tests {
		for _, tc := range testCases {
			t.Run(tt.name+"_"+tc.desc, func(t *testing.T) {
				cpu := NewCPU()
				cpu.PC = 1

				if tc.carryIn {
					cpu.P |= FlagC
				} else {
					cpu.P &= ^FlagC
				}

				// Setup the instruction
				cpu.Memory[0] = tt.opcode
				tt.setup(cpu, tc.value)

				// Execute
				cycles := cpu.execute(tt.opcode)

				// Check cycles
				assert.Equal(t, tt.cycles, cycles, "Incorrect cycle count")

				// Check result
				result := tt.getValue(cpu)
				assert.Equal(t, tc.expected, result, "Incorrect result")

				// Check flags
				assert.Equal(t, tc.expectC, cpu.P&FlagC != 0, "Carry flag mismatch")
				assert.Equal(t, tc.expectZ, cpu.P&FlagZ != 0, "Zero flag mismatch")
				assert.Equal(t, tc.expectN, cpu.P&FlagN != 0, "Negative flag mismatch")
			})
		}
	}
}
