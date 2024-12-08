package cpu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestINC(t *testing.T) {
	tests := []struct {
		name     string
		opcode   uint8
		setupMem func(*CPUAndMemory, uint8)
		cycles   uint8
		memCheck func(*CPUAndMemory, uint8) uint16 // Returns address to check
	}{
		{
			name:   "INC Zero Page",
			opcode: INC_ZP,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = INC_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 5,
			memCheck: func(c *CPUAndMemory, _ uint8) uint16 {
				return 0x42
			},
		},
		{
			name:   "INC Zero Page,X",
			opcode: INC_ZPX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = INC_ZPX
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
			name:   "INC Absolute",
			opcode: INC_ABS,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = INC_ABS
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
			name:   "INC Absolute,X",
			opcode: INC_ABX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = INC_ABX
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
		expectedZ bool
		expectedN bool
	}{
		{0x00, 0x01, false, false}, // 0 -> 1
		{0x7F, 0x80, false, true},  // 127 -> 128 (sign flip)
		{0xFE, 0xFF, false, true},  // 254 -> 255
		{0xFF, 0x00, true, false},  // 255 -> 0 (overflow)
		{0x44, 0x45, false, false}, // Regular increment
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

				// Check memory value
				addr := tt.memCheck(cpu, tv.initial)
				assert.Equal(t, tv.expected, cpu.Memory[addr],
					"Memory value incorrect")

				// Check flags
				assert.Equal(t, tv.expectedZ, (cpu.P&FlagZ) != 0,
					"Zero flag mismatch")
				assert.Equal(t, tv.expectedN, (cpu.P&FlagN) != 0,
					"Negative flag mismatch")
			})
		}
	}
}
