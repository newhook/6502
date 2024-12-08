package cpu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCPXCPYAddressingModes(t *testing.T) {
	tests := []struct {
		name     string
		opcode   uint8
		setupMem func(*CPUAndMemory, uint8)
		cycles   uint8
		isX      bool // true for CPX, false for CPY
	}{
		// CPX addressing modes
		{
			name:   "CPX Immediate",
			opcode: CPX_IMM,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPX_IMM
				c.Memory[1] = value
			},
			cycles: 2,
			isX:    true,
		},
		{
			name:   "CPX Zero Page",
			opcode: CPX_ZP,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPX_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 3,
			isX:    true,
		},
		{
			name:   "CPX Absolute",
			opcode: CPX_ABS,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPX_ABS
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 4,
			isX:    true,
		},

		// CPY addressing modes
		{
			name:   "CPY Immediate",
			opcode: CPY_IMM,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPY_IMM
				c.Memory[1] = value
			},
			cycles: 2,
			isX:    false,
		},
		{
			name:   "CPY Zero Page",
			opcode: CPY_ZP,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPY_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 3,
			isX:    false,
		},
		{
			name:   "CPY Absolute",
			opcode: CPY_ABS,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CPY_ABS
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 4,
			isX:    false,
		},
	}

	compareValues := []struct {
		register  uint8
		value     uint8
		expectedC bool
		expectedZ bool
		expectedN bool
	}{
		{0x42, 0x42, true, true, false},   // Equal values
		{0x50, 0x30, true, false, false},  // Register greater than value
		{0x30, 0x50, false, false, true},  // Register less than value
		{0x00, 0x01, false, false, true},  // Negative result
		{0xFF, 0x01, true, false, true},   // Register much greater than value
		{0x01, 0xFF, false, false, false}, // Register much less than value
	}

	for _, tt := range tests {
		for _, cv := range compareValues {
			testName := tt.name + "_" +
				"Reg" + fmt.Sprintf("%x", cv.register) + "_" +
				"M" + fmt.Sprintf("%x", cv.value)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPUAndMemory()
				cpu.PC = 1

				// Set the appropriate register
				if tt.isX {
					cpu.X = cv.register
				} else {
					cpu.Y = cv.register
				}

				tt.setupMem(cpu, cv.value)
				cycles := cpu.execute(tt.opcode)

				assert.Equal(t, tt.cycles, cycles,
					"Unexpected number of cycles")
				assert.Equal(t, cv.expectedC, (cpu.P&FlagC) != 0,
					"Carry flag mismatch")
				assert.Equal(t, cv.expectedZ, (cpu.P&FlagZ) != 0,
					"Zero flag mismatch")
				assert.Equal(t, cv.expectedN, (cpu.P&FlagN) != 0,
					"Negative flag mismatch")

				// Ensure registers weren't modified
				if tt.isX {
					assert.Equal(t, cv.register, cpu.X,
						"X register should not be modified")
				} else {
					assert.Equal(t, cv.register, cpu.Y,
						"Y register should not be modified")
				}
			})
		}
	}
}
