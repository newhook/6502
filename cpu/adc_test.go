package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestADC(t *testing.T) {
	defaultFlags := uint8(0x24) // Break and Unused flags are set by default

	tests := []struct {
		name   string
		setup  func(*CPU)
		opcode uint8
		want   uint8
		wantA  uint8
		wantP  uint8
		cycles uint8
	}{
		{
			name: "ADC_IMM simple addition",
			setup: func(c *CPU) {
				c.A = 0x20
				c.Memory[0] = 0x10
			},
			opcode: ADC_IMM,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 2,
		},
		{
			name: "ADC_IMM with carry flag set",
			setup: func(c *CPU) {
				c.A = 0x20
				c.P = defaultFlags | FlagC
				c.Memory[0] = 0x10
			},
			opcode: ADC_IMM,
			wantA:  0x31,
			wantP:  defaultFlags,
			cycles: 2,
		},
		{
			name: "ADC_IMM with overflow",
			setup: func(c *CPU) {
				c.A = 0x50
				c.Memory[0] = 0x50
			},
			opcode: ADC_IMM,
			wantA:  0xA0,
			wantP:  defaultFlags | FlagN | FlagV,
			cycles: 2,
		},
		{
			name: "ADC_IMM resulting in zero",
			setup: func(c *CPU) {
				c.A = 0xFF
				c.Memory[0] = 0x01
			},
			opcode: ADC_IMM,
			wantA:  0x00,
			wantP:  defaultFlags | FlagZ | FlagC,
			cycles: 2,
		},
		{
			name: "ADC_ZP",
			setup: func(c *CPU) {
				c.A = 0x10
				c.Memory[0] = 0x42    // ZP address
				c.Memory[0x42] = 0x20 // Value at ZP
			},
			opcode: ADC_ZP,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 3,
		},
		{
			name: "ADC_ZPX",
			setup: func(c *CPU) {
				c.A = 0x10
				c.X = 0x05
				c.Memory[0] = 0x42    // ZP address
				c.Memory[0x47] = 0x20 // Value at ZP+X
			},
			opcode: ADC_ZPX,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 4,
		},
		{
			name: "ADC_ABS",
			setup: func(c *CPU) {
				c.A = 0x10
				c.Memory[0] = 0x80 // Low byte of address
				c.Memory[1] = 0x12 // High byte of address
				c.Memory[0x1280] = 0x20
			},
			opcode: ADC_ABS,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 4,
		},
		{
			name: "ADC_ABX no page cross",
			setup: func(c *CPU) {
				c.A = 0x10
				c.X = 0x05
				c.Memory[0] = 0x80 // Low byte of address
				c.Memory[1] = 0x12 // High byte of address
				c.Memory[0x1285] = 0x20
			},
			opcode: ADC_ABX,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 4,
		},
		{
			name: "ADC_ABX with page cross",
			setup: func(c *CPU) {
				c.A = 0x10
				c.X = 0xFF
				c.Memory[0] = 0x80 // Low byte of address
				c.Memory[1] = 0x12 // High byte of address
				c.Memory[0x137F] = 0x20
			},
			opcode: ADC_ABX,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 5,
		},
		{
			name: "ADC_ABY no page cross",
			setup: func(c *CPU) {
				c.A = 0x10
				c.Y = 0x05
				c.Memory[0] = 0x80 // Low byte of address
				c.Memory[1] = 0x12 // High byte of address
				c.Memory[0x1285] = 0x20
			},
			opcode: ADC_ABY,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 4,
		},
		{
			name: "ADC_INX",
			setup: func(c *CPU) {
				c.A = 0x10
				c.X = 0x05
				c.Memory[0] = 0x80      // ZP address
				c.Memory[0x85] = 0x00   // Low byte of indirect address
				c.Memory[0x86] = 0x12   // High byte of indirect address
				c.Memory[0x1200] = 0x20 // Final value
			},
			opcode: ADC_INX,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 6,
		},
		{
			name: "ADC_INY no page cross",
			setup: func(c *CPU) {
				c.A = 0x10
				c.Y = 0x05
				c.Memory[0] = 0x80      // ZP address
				c.Memory[0x80] = 0x00   // Low byte of indirect address
				c.Memory[0x81] = 0x12   // High byte of indirect address
				c.Memory[0x1205] = 0x20 // Final value
			},
			opcode: ADC_INY,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 5,
		},
		{
			name: "ADC_INY with page cross",
			setup: func(c *CPU) {
				c.A = 0x10
				c.Y = 0xFF
				c.Memory[0] = 0x80      // ZP address
				c.Memory[0x80] = 0x02   // Low byte of indirect address
				c.Memory[0x81] = 0x12   // High byte of indirect address
				c.Memory[0x1301] = 0x20 // Final value
			},
			opcode: ADC_INY,
			wantA:  0x30,
			wantP:  defaultFlags,
			cycles: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCPU()
			tt.setup(c)

			cycles := c.execute(tt.opcode)

			assert.Equal(t, tt.wantA, c.A, "accumulator value")
			assert.Equal(t, tt.wantP, c.P, "status register")
			assert.Equal(t, tt.cycles, cycles, "cycles used")
		})
	}
}

func TestADCFlagBehavior(t *testing.T) {
	defaultFlags := uint8(0x24)

	tests := []struct {
		name  string
		a     uint8
		value uint8
		initP uint8
		wantA uint8
		wantP uint8
	}{
		{
			name:  "Negative flag set",
			a:     0x00,
			value: 0x80,
			initP: defaultFlags,
			wantA: 0x80,
			wantP: defaultFlags | FlagN,
		},
		{
			name:  "Zero flag set with overflow",
			a:     0x80,
			value: 0x80,
			initP: defaultFlags,
			wantA: 0x00,
			wantP: defaultFlags | FlagZ | FlagC | FlagV,
		},
		{
			name:  "Zero flag only",
			a:     0x00,
			value: 0x00,
			initP: defaultFlags,
			wantA: 0x00,
			wantP: defaultFlags | FlagZ,
		},
		{
			name:  "Carry flag set",
			a:     0xFF,
			value: 0x01,
			initP: defaultFlags,
			wantA: 0x00,
			wantP: defaultFlags | FlagZ | FlagC,
		},
		{
			name:  "Overflow flag set (positive to negative)",
			a:     0x7F,
			value: 0x01,
			initP: defaultFlags,
			wantA: 0x80,
			wantP: defaultFlags | FlagN | FlagV,
		},
		{
			name:  "Initial carry considered",
			a:     0x01,
			value: 0x01,
			initP: defaultFlags | FlagC,
			wantA: 0x03,
			wantP: defaultFlags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCPU()
			c.A = tt.a
			c.P = tt.initP
			c.Memory[0] = tt.value
			c.execute(ADC_IMM)

			assert.Equal(t, tt.wantA, c.A, "accumulator value")
			assert.Equal(t, tt.wantP, c.P, "status register")
		})
	}
}

func TestADCDecimalMode(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name        string
		accumulator uint8
		operand     uint8
		carryIn     bool
		expected    uint8
		expectC     bool
	}{
		{
			name:        "BCD: 12 + 34 = 46",
			accumulator: 0x12,
			operand:     0x34,
			carryIn:     false,
			expected:    0x46,
			expectC:     false,
		},
		{
			name:        "BCD: 15 + 26 = 41",
			accumulator: 0x15,
			operand:     0x26,
			carryIn:     false,
			expected:    0x41,
			expectC:     false,
		},
		{
			name:        "BCD: 51 + 51 = 02 (with carry)",
			accumulator: 0x51,
			operand:     0x51,
			carryIn:     false,
			expected:    0x02,
			expectC:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = ADC_IMM
			cpu.Memory[0x0201] = test.operand
			cpu.A = test.accumulator
			cpu.P = FlagD // Set decimal mode
			if test.carryIn {
				cpu.P |= FlagC
			}

			// Execute
			cpu.Step()

			// Assert
			assert.Equal(test.expected, cpu.A, "incorrect BCD result")
			assert.Equal(test.expectC, cpu.P&FlagC != 0, "incorrect carry flag")
		})
	}
}
