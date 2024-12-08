package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLDXImmediate(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name    string
		value   uint8
		cycles  uint8
		expectZ bool
		expectN bool
	}{
		{
			name:    "Zero flag set when loading zero",
			value:   0x00,
			cycles:  2,
			expectZ: true,
			expectN: false,
		},
		{
			name:    "Neither flag set for positive value",
			value:   0x42,
			cycles:  2,
			expectZ: false,
			expectN: false,
		},
		{
			name:    "Negative flag set for negative value",
			value:   0x80,
			cycles:  2,
			expectZ: false,
			expectN: true,
		},
		{
			name:    "Negative flag set for max value",
			value:   0xFF,
			cycles:  2,
			expectZ: false,
			expectN: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDX_IMM
			cpu.Memory[0x0201] = test.value
			cpu.X = 0x00
			cpu.P = 0x00

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.X, "incorrect X register value")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect Zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect Negative flag")
		})
	}
}

func TestLDXZeroPage(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name         string
		zeroPageAddr uint8
		value        uint8
		cycles       uint8
	}{
		{
			name:         "Load value from zero page",
			zeroPageAddr: 0x42,
			value:        0x37,
			cycles:       3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDX_ZP
			cpu.Memory[0x0201] = test.zeroPageAddr
			cpu.Memory[test.zeroPageAddr] = test.value

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.X, "incorrect X register value")
		})
	}
}

func TestLDXZeroPageY(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name         string
		zeroPageAddr uint8
		yReg         uint8
		value        uint8
		cycles       uint8
	}{
		{
			name:         "Normal zero page Y indexed addressing",
			zeroPageAddr: 0x42,
			yReg:         0x01,
			value:        0x37,
			cycles:       4,
		},
		{
			name:         "Zero page wrap-around",
			zeroPageAddr: 0xFF,
			yReg:         0x02,
			value:        0x55,
			cycles:       4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDX_ZPY
			cpu.Memory[0x0201] = test.zeroPageAddr
			cpu.Y = test.yReg

			effectiveAddr := (test.zeroPageAddr + test.yReg) & 0xFF
			cpu.Memory[effectiveAddr] = test.value

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.X, "incorrect X register value")
		})
	}
}

func TestLDXAbsolute(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name   string
		addr   uint16
		value  uint8
		cycles uint8
	}{
		{
			name:   "Load value from absolute address",
			addr:   0x3742,
			value:  0x55,
			cycles: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDX_ABS
			cpu.Memory[0x0201] = uint8(test.addr & 0xFF)
			cpu.Memory[0x0202] = uint8(test.addr >> 8)
			cpu.Memory[test.addr] = test.value

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.X, "incorrect X register value")
		})
	}
}

func TestLDXAbsoluteY(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPUAndMemory()

	tests := []struct {
		name     string
		baseAddr uint16
		yReg     uint8
		value    uint8
		cycles   uint8
	}{
		{
			name:     "Absolute Y indexed - no page cross",
			baseAddr: 0x3742,
			yReg:     0x01,
			value:    0x55,
			cycles:   4,
		},
		{
			name:     "Absolute Y indexed - with page cross",
			baseAddr: 0x37FF,
			yReg:     0x01,
			value:    0x66,
			cycles:   5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDX_ABY
			cpu.Memory[0x0201] = uint8(test.baseAddr & 0xFF)
			cpu.Memory[0x0202] = uint8(test.baseAddr >> 8)
			cpu.Y = test.yReg

			effectiveAddr := test.baseAddr + uint16(test.yReg)
			cpu.Memory[effectiveAddr] = test.value

			// Execute
			cycles := cpu.Step()

			// Assert
			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.X, "incorrect X register value")
		})
	}
}
