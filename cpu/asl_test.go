package cpu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestASL(t *testing.T) {
	tests := []struct {
		name        string
		opcode      uint8
		setupMem    func(*CPU, uint8)
		cycles      uint8
		accumulator bool
		memCheck    func(*CPU, uint8) uint16 // Returns address to check
	}{
		{
			name:   "ASL Accumulator",
			opcode: ASL_ACC,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = ASL_ACC
				c.A = value
			},
			cycles:      2,
			accumulator: true,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0 // Not used for accumulator mode
			},
		},
		{
			name:   "ASL Zero Page",
			opcode: ASL_ZP,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = ASL_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 5,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0x42
			},
		},
		{
			name:   "ASL Zero Page,X",
			opcode: ASL_ZPX,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = ASL_ZPX
				c.Memory[1] = 0x42     // Zero page address
				c.X = 0x01             // X offset
				c.Memory[0x43] = value // 0x42 + 0x01 = 0x43
			},
			cycles: 6,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0x43
			},
		},
		{
			name:   "ASL Absolute",
			opcode: ASL_ABS,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = ASL_ABS
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 6,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0x1280
			},
		},
		{
			name:   "ASL Absolute,X",
			opcode: ASL_ABX,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = ASL_ABX
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.X = 0x01
				c.Memory[0x1281] = value // 0x1280 + 0x01
			},
			cycles: 7,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0x1281
			},
		},
	}

	testValues := []struct {
		initial   uint8
		expected  uint8
		expectedC bool
		expectedZ bool
		expectedN bool
	}{
		{0x01, 0x02, false, false, false}, // No flags set
		{0x40, 0x80, false, false, true},  // Sets negative flag
		{0x80, 0x00, true, true, false},   // Sets carry and zero flags
		{0xFF, 0xFE, true, false, true},   // Sets carry and negative flags
		{0x00, 0x00, false, true, false},  // Sets zero flag
		{0x55, 0xAA, false, false, true},  // Complex bit pattern
	}

	for _, tt := range tests {
		for _, tv := range testValues {
			testName := tt.name + "_" +
				fmt.Sprintf("%x", tv.initial) + "_to_" +
				fmt.Sprintf("%x", tv.expected)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPU()
				cpu.PC = 1

				tt.setupMem(cpu, tv.initial)
				cycles := cpu.execute(tt.opcode)

				// Check cycles
				assert.Equal(t, tt.cycles, cycles,
					"Unexpected number of cycles")

				// Check result value
				if tt.accumulator {
					assert.Equal(t, tv.expected, cpu.A,
						"Accumulator value incorrect")
				} else {
					addr := tt.memCheck(cpu, tv.initial)
					assert.Equal(t, tv.expected, cpu.Memory[addr],
						"Memory value incorrect")
				}

				// Check flags
				assert.Equal(t, tv.expectedC, (cpu.P&FlagC) != 0,
					"Carry flag mismatch")
				assert.Equal(t, tv.expectedZ, (cpu.P&FlagZ) != 0,
					"Zero flag mismatch")
				assert.Equal(t, tv.expectedN, (cpu.P&FlagN) != 0,
					"Negative flag mismatch")
			})
		}
	}
}
