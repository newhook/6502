package cpu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLSR(t *testing.T) {
	tests := []struct {
		name        string
		opcode      uint8
		setupMem    func(*CPUAndMemory, uint8)
		cycles      uint8
		accumulator bool
		memCheck    func(*CPUAndMemory, uint8) uint16 // Returns address to check
	}{
		{
			name:   "LSR Accumulator",
			opcode: LSR_ACC,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = LSR_ACC
				c.A = value
			},
			cycles:      2,
			accumulator: true,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
				return 0 // Not used for accumulator mode
			},
		},
		{
			name:   "LSR Zero Page",
			opcode: LSR_ZP,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = LSR_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 5,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
				return 0x42
			},
		},
		{
			name:   "LSR Zero Page,X",
			opcode: LSR_ZPX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = LSR_ZPX
				c.Memory[1] = 0x42     // Zero page address
				c.X = 0x01             // X offset
				c.Memory[0x43] = value // 0x42 + 0x01 = 0x43
			},
			cycles: 6,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
				return 0x43
			},
		},
		{
			name:   "LSR Absolute",
			opcode: LSR_ABS,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = LSR_ABS
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 6,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
				return 0x1280
			},
		},
		{
			name:   "LSR Absolute,X",
			opcode: LSR_ABX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = LSR_ABX
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.X = 0x01
				c.Memory[0x1281] = value // 0x1280 + 0x01
			},
			cycles: 7,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
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
		{0x02, 0x01, false, false, false}, // Simple shift
		{0x01, 0x00, true, true, false},   // Sets carry and zero flags
		{0x80, 0x40, false, false, false}, // Highest bit shifted out
		{0xFF, 0x7F, true, false, false},  // Sets carry flag
		{0x00, 0x00, false, true, false},  // Zero input/output
		{0xAA, 0x55, false, false, false}, // Complex bit pattern
	}

	for _, tt := range tests {
		for _, tv := range testValues {
			testName := tt.name + "_" +
				fmt.Sprintf("%x", tv.initial) + "_to_" +
				fmt.Sprintf("%x", tv.expected)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPUAndMemory()
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
