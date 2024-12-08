package cpu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestCMP contains all test cases for the CMP instruction
func TestCMP(t *testing.T) {
	tests := []struct {
		name        string
		accumulator uint8
		value       uint8
		expectedC   bool
		expectedZ   bool
		expectedN   bool
	}{
		{
			name:        "Equal values",
			accumulator: 0x42,
			value:       0x42,
			expectedC:   true,
			expectedZ:   true,
			expectedN:   false,
		},
		{
			name:        "A greater than value",
			accumulator: 0x50,
			value:       0x30,
			expectedC:   true,
			expectedZ:   false,
			expectedN:   false,
		},
		{
			name:        "A less than value",
			accumulator: 0x30,
			value:       0x50,
			expectedC:   false,
			expectedZ:   false,
			expectedN:   true,
		},
		{
			name:        "Negative result",
			accumulator: 0x00,
			value:       0x01,
			expectedC:   false,
			expectedZ:   false,
			expectedN:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPUAndMemory()
			cpu.A = tt.accumulator

			// Test immediate mode
			cpu.Memory[0] = tt.value
			cycles := cpu.execute(CMP_IMM)

			assert.Equal(t, uint8(2), cycles, "CMP_IMM should take 2 cycles")
			assert.Equal(t, tt.expectedC, (cpu.P&FlagC) != 0, "Carry flag mismatch")
			assert.Equal(t, tt.expectedZ, (cpu.P&FlagZ) != 0, "Zero flag mismatch")
			assert.Equal(t, tt.expectedN, (cpu.P&FlagN) != 0, "Negative flag mismatch")
		})
	}
}

func TestCMPAddressingModes(t *testing.T) {
	tests := []struct {
		name       string
		opcode     uint8
		setupMem   func(*CPUAndMemory, uint8)
		cycles     uint8
		extraCycle bool
	}{
		{
			name:   "CMP Immediate",
			opcode: CMP_IMM,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_IMM
				c.Memory[1] = value
			},
			cycles: 2,
		},
		{
			name:   "CMP Zero Page",
			opcode: CMP_ZP,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ZP
				c.Memory[1] = 0x42 // Zero page address
				c.Memory[0x42] = value
			},
			cycles: 3,
		},
		{
			name:   "CMP Zero Page,X",
			opcode: CMP_ZPX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ZPX
				c.Memory[1] = 0x42     // Zero page address
				c.X = 0x01             // X offset
				c.Memory[0x43] = value // 0x42 + 0x01 = 0x43
			},
			cycles: 4,
		},
		{
			name:   "CMP Absolute",
			opcode: CMP_ABS,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ABS
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 4,
		},
		{
			name:   "CMP Absolute,X (no page cross)",
			opcode: CMP_ABX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ABX
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.X = 0x01
				c.Memory[0x1281] = value // 0x1280 + 0x01
			},
			cycles: 4,
		},
		{
			name:   "CMP Absolute,X (with page cross)",
			opcode: CMP_ABX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ABX
				c.Memory[1] = 0xFF // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.X = 0x01
				c.Memory[0x1300] = value // 0x12FF + 0x01
			},
			cycles:     4,
			extraCycle: true,
		},
		{
			name:   "CMP Absolute,Y (no page cross)",
			opcode: CMP_ABY,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ABY
				c.Memory[1] = 0x80 // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Y = 0x01
				c.Memory[0x1281] = value // 0x1280 + 0x01
			},
			cycles: 4,
		},
		{
			name:   "CMP Absolute,Y (with page cross)",
			opcode: CMP_ABY,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_ABY
				c.Memory[1] = 0xFF // Low byte of address
				c.Memory[2] = 0x12 // High byte of address
				c.Y = 0x01
				c.Memory[0x1300] = value // 0x12FF + 0x01
			},
			cycles:     4,
			extraCycle: true,
		},
		{
			name:   "CMP Indirect,X",
			opcode: CMP_INX,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_INX
				c.Memory[1] = 0x20 // Zero page address
				c.X = 0x01
				// Indirect address stored at 0x21 (0x20 + 0x01)
				c.Memory[0x21] = 0x80 // Low byte of address
				c.Memory[0x22] = 0x12 // High byte of address
				c.Memory[0x1280] = value
			},
			cycles: 6,
		},
		{
			name:   "CMP Indirect,Y (no page cross)",
			opcode: CMP_INY,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_INY
				c.Memory[1] = 0x20 // Zero page address
				// Indirect address stored at 0x20
				c.Memory[0x20] = 0x80 // Low byte of address
				c.Memory[0x21] = 0x12 // High byte of address
				c.Y = 0x01
				c.Memory[0x1281] = value // 0x1280 + 0x01
			},
			cycles: 5,
		},
		{
			name:   "CMP Indirect,Y (with page cross)",
			opcode: CMP_INY,
			setupMem: func(c *CPUAndMemory, value uint8) {
				c.Memory[0] = CMP_INY
				c.Memory[1] = 0x20 // Zero page address
				// Indirect address stored at 0x20
				c.Memory[0x20] = 0xFF // Low byte of address
				c.Memory[0x21] = 0x12 // High byte of address
				c.Y = 0x01
				c.Memory[0x1300] = value // 0x12FF + 0x01
			},
			cycles:     5,
			extraCycle: true,
		},
	}

	// Test values to compare against
	compareValues := []struct {
		accumulator uint8
		value       uint8
		expectedC   bool
		expectedZ   bool
		expectedN   bool
	}{
		{0x42, 0x42, true, true, false},   // Equal values
		{0x50, 0x30, true, false, false},  // A greater than value
		{0x30, 0x50, false, false, true},  // A less than value
		{0x00, 0x01, false, false, true},  // Negative result
		{0xFF, 0x01, true, false, true},   // A much greater than value
		{0x01, 0xFF, false, false, false}, // A much less than value
	}

	for _, tt := range tests {
		for _, cv := range compareValues {
			testName := tt.name + "_" +
				"A" + fmt.Sprintf("%x", cv.accumulator) + "_" +
				"M" + fmt.Sprintf("%x", cv.value)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPUAndMemory()
				cpu.A = cv.accumulator
				cpu.PC = 1
				tt.setupMem(cpu, cv.value)

				cycles := cpu.execute(tt.opcode)

				expectedCycles := tt.cycles
				if tt.extraCycle {
					expectedCycles++
				}

				assert.Equal(t, expectedCycles, cycles,
					"Unexpected number of cycles")
				assert.Equal(t, cv.expectedC, (cpu.P&FlagC) != 0,
					"Carry flag mismatch")
				assert.Equal(t, cv.expectedZ, (cpu.P&FlagZ) != 0,
					"Zero flag mismatch")
				assert.Equal(t, cv.expectedN, (cpu.P&FlagN) != 0,
					"Negative flag mismatch")

				// Ensure accumulator wasn't modified
				assert.Equal(t, cv.accumulator, cpu.A,
					"Accumulator should not be modified")
			})
		}
	}
}
