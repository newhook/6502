package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDEC(t *testing.T) {
	tests := []struct {
		name     string
		opcode   uint8
		setupMem func(*CPU, uint8)
		cycles   uint8
		memCheck func(*CPU, uint8) uint16 // Returns address to check
	}{
		{
			name:   "DEC Zero Page",
			opcode: DEC_ZP,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = DEC_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 5,
			memCheck: func(c *CPU, _ uint8) uint16 {
				return 0x42
			},
		},
		{
			name:   "DEC Zero Page,X",
			opcode: DEC_ZPX,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = DEC_ZPX
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
			name:   "DEC Absolute",
			opcode: DEC_ABS,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = DEC_ABS
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
			name:   "DEC Absolute,X",
			opcode: DEC_ABX,
			setupMem: func(c *CPU, value uint8) {
				c.Memory[0] = DEC_ABX
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
		expectedZ bool
		expectedN bool
	}{
		{0x01, 0x00, true, false},  // 1 -> 0
		{0x00, 0xFF, false, true},  // 0 -> 255 (underflow)
		{0x80, 0x7F, false, false}, // 128 -> 127 (sign flip)
		{0xFF, 0xFE, false, true},  // 255 -> 254
		{0x45, 0x44, false, false}, // Regular decrement
	}

	for _, tt := range tests {
		for _, tv := range testValues {
			testName := tt.name + "_" +
				string(tv.initial) + "_to_" +
				string(tv.expected)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPU()
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
