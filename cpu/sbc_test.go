package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSBC(t *testing.T) {
	tests := []struct {
		name         string
		initial      uint8 // Initial value in A
		subtract     uint8 // Value to subtract
		initialCarry bool  // Initial carry flag state
		expectedA    uint8 // Expected value in A
		expectedZ    bool  // Expected Zero flag
		expectedN    bool  // Expected Negative flag
		expectedV    bool  // Expected Overflow flag
		expectedC    bool  // Expected Carry flag
	}{
		{
			name:         "Basic subtraction",
			initial:      0x50,
			subtract:     0x30,
			initialCarry: true,
			expectedA:    0x20,
			expectedZ:    false,
			expectedN:    false,
			expectedV:    false,
			expectedC:    true,
		},
		{
			name:         "Subtraction with borrow",
			initial:      0x50,
			subtract:     0x70,
			initialCarry: true,
			expectedA:    0xE0,
			expectedZ:    false,
			expectedN:    true,
			expectedV:    false,
			expectedC:    false,
		},
		{
			name:         "Zero result",
			initial:      0x50,
			subtract:     0x50,
			initialCarry: true,
			expectedA:    0x00,
			expectedZ:    true,
			expectedN:    false,
			expectedV:    false,
			expectedC:    true,
		},
		{
			name:         "Overflow case",
			initial:      0x80,
			subtract:     0x01,
			initialCarry: true,
			expectedA:    0x7F,
			expectedZ:    false,
			expectedN:    false,
			expectedV:    true,
			expectedC:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPU()
			cpu.A = tt.initial
			if tt.initialCarry {
				cpu.P |= FlagC
			} else {
				cpu.P &^= FlagC
			}

			// Test immediate mode
			cpu.Memory[0] = tt.subtract
			cycles := cpu.execute(SBC_IMM)

			assert.Equal(t, uint8(2), cycles, "Cycles not correct")
			assert.Equal(t, tt.expectedA, cpu.A, "Accumulator value incorrect")
			assert.Equal(t, tt.expectedZ, cpu.P&FlagZ != 0, "Zero flag incorrect")
			assert.Equal(t, tt.expectedN, cpu.P&FlagN != 0, "Negative flag incorrect")
			assert.Equal(t, tt.expectedV, cpu.P&FlagV != 0, "Overflow flag incorrect")
			assert.Equal(t, tt.expectedC, cpu.P&FlagC != 0, "Carry flag incorrect")
		})
	}
}

func TestSBCAddressingModes(t *testing.T) {
	tests := []struct {
		name     string
		opcode   uint8
		setup    func(*CPU)
		cycles   uint8
		expected uint8
	}{
		{
			name:   "Zero Page",
			opcode: SBC_ZP,
			setup: func(c *CPU) {
				c.Memory[0] = 0x42    // Zero page address
				c.Memory[0x42] = 0x10 // Value to subtract
				c.A = 0x50            // Initial value
				c.P |= FlagC          // Set carry flag
			},
			cycles:   3,
			expected: 0x40,
		},
		{
			name:   "Zero Page,X",
			opcode: SBC_ZPX,
			setup: func(c *CPU) {
				c.Memory[0] = 0x42    // Zero page address
				c.X = 0x02            // X offset
				c.Memory[0x44] = 0x10 // Value to subtract at (0x42 + 0x02)
				c.A = 0x50            // Initial value
				c.P |= FlagC          // Set carry flag
			},
			cycles:   4,
			expected: 0x40,
		},
		// Add more addressing mode tests as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPU()
			tt.setup(cpu)

			cycles := cpu.execute(tt.opcode)

			assert.Equal(t, tt.cycles, cycles, "Cycles not correct")
			assert.Equal(t, tt.expected, cpu.A, "Accumulator value incorrect")
		})
	}
}
