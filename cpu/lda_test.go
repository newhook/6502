package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLDAImmediate(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name    string
		value   uint8
		cycles  uint8
		expectZ bool
		expectN bool
	}{
		{
			name:    "Load zero - sets zero flag",
			value:   0x00,
			cycles:  2,
			expectZ: true,
			expectN: false,
		},
		{
			name:    "Load positive value - no flags",
			value:   0x42,
			cycles:  2,
			expectZ: false,
			expectN: false,
		},
		{
			name:    "Load negative value - sets negative flag",
			value:   0x80,
			cycles:  2,
			expectZ: false,
			expectN: true,
		},
		{
			name:    "Load max value - sets negative flag",
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
			cpu.Memory[0x0200] = LDA_IMM
			cpu.Memory[0x0201] = test.value
			cpu.A = 0x00 // Reset accumulator
			cpu.P = 0x00 // Reset flags

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
			assert.Equal(test.expectZ, cpu.P&FlagZ != 0, "incorrect zero flag")
			assert.Equal(test.expectN, cpu.P&FlagN != 0, "incorrect negative flag")
		})
	}
}

func TestLDAZeroPage(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name   string
		zpAddr uint8
		value  uint8
		cycles uint8
	}{
		{
			name:   "Load from zero page",
			zpAddr: 0x42,
			value:  0x37,
			cycles: 3,
		},
		{
			name:   "Load from zero page boundary",
			zpAddr: 0xFF,
			value:  0x55,
			cycles: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_ZP
			cpu.Memory[0x0201] = test.zpAddr
			cpu.Memory[test.zpAddr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAZeroPageX(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name   string
		zpAddr uint8
		xReg   uint8
		value  uint8
		cycles uint8
	}{
		{
			name:   "Basic zero page X indexed",
			zpAddr: 0x42,
			xReg:   0x01,
			value:  0x37,
			cycles: 4,
		},
		{
			name:   "Zero page X with wrap",
			zpAddr: 0xFF,
			xReg:   0x02,
			value:  0x55,
			cycles: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_ZPX
			cpu.Memory[0x0201] = test.zpAddr
			cpu.X = test.xReg

			effectiveAddr := (test.zpAddr + test.xReg) & 0xFF
			cpu.Memory[effectiveAddr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAAbsolute(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name   string
		addr   uint16
		value  uint8
		cycles uint8
	}{
		{
			name:   "Load from absolute address",
			addr:   0x1234,
			value:  0x42,
			cycles: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_ABS
			cpu.Memory[0x0201] = uint8(test.addr & 0xFF)
			cpu.Memory[0x0202] = uint8(test.addr >> 8)
			cpu.Memory[test.addr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAAbsoluteX(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name     string
		baseAddr uint16
		xReg     uint8
		value    uint8
		cycles   uint8
	}{
		{
			name:     "No page cross",
			baseAddr: 0x1234,
			xReg:     0x01,
			value:    0x42,
			cycles:   4,
		},
		{
			name:     "Page cross",
			baseAddr: 0x12FF,
			xReg:     0x01,
			value:    0x42,
			cycles:   5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_ABX
			cpu.Memory[0x0201] = uint8(test.baseAddr & 0xFF)
			cpu.Memory[0x0202] = uint8(test.baseAddr >> 8)
			cpu.X = test.xReg

			effectiveAddr := test.baseAddr + uint16(test.xReg)
			cpu.Memory[effectiveAddr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAAbsoluteY(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name     string
		baseAddr uint16
		yReg     uint8
		value    uint8
		cycles   uint8
	}{
		{
			name:     "No page cross",
			baseAddr: 0x1234,
			yReg:     0x01,
			value:    0x42,
			cycles:   4,
		},
		{
			name:     "Page cross",
			baseAddr: 0x12FF,
			yReg:     0x01,
			value:    0x42,
			cycles:   5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_ABY
			cpu.Memory[0x0201] = uint8(test.baseAddr & 0xFF)
			cpu.Memory[0x0202] = uint8(test.baseAddr >> 8)
			cpu.Y = test.yReg

			effectiveAddr := test.baseAddr + uint16(test.yReg)
			cpu.Memory[effectiveAddr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAIndirectX(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name   string
		zpAddr uint8
		xReg   uint8
		target uint16
		value  uint8
		cycles uint8
	}{
		{
			name:   "Basic indirect X",
			zpAddr: 0x20,
			xReg:   0x04,
			target: 0x1234,
			value:  0x42,
			cycles: 6,
		},
		{
			name:   "Indirect X with wrap",
			zpAddr: 0xFF,
			xReg:   0x01,
			target: 0x1234,
			value:  0x42,
			cycles: 6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_INX
			cpu.Memory[0x0201] = test.zpAddr
			cpu.X = test.xReg

			// Setup indirect address
			effectiveZP := (test.zpAddr + test.xReg) & 0xFF
			cpu.Memory[effectiveZP] = uint8(test.target & 0xFF)
			cpu.Memory[(effectiveZP+1)&0xFF] = uint8(test.target >> 8)

			// Setup target value
			cpu.Memory[test.target] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}

func TestLDAIndirectY(t *testing.T) {
	assert := assert.New(t)
	cpu := NewCPU()

	tests := []struct {
		name      string
		zpAddr    uint8
		yReg      uint8
		target    uint16
		value     uint8
		cycles    uint8
		pageCross bool
	}{
		{
			name:      "No page cross",
			zpAddr:    0x20,
			yReg:      0x04,
			target:    0x1234,
			value:     0x42,
			cycles:    5,
			pageCross: false,
		},
		{
			name:      "With page cross",
			zpAddr:    0x20,
			yReg:      0xFF,
			target:    0x1234,
			value:     0x42,
			cycles:    6,
			pageCross: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu.PC = 0x0200
			cpu.Memory[0x0200] = LDA_INY
			cpu.Memory[0x0201] = test.zpAddr
			cpu.Y = test.yReg

			// Setup indirect address
			cpu.Memory[test.zpAddr] = uint8(test.target & 0xFF)
			cpu.Memory[(test.zpAddr+1)&0xFF] = uint8(test.target >> 8)

			// Setup target value
			finalAddr := test.target + uint16(test.yReg)
			cpu.Memory[finalAddr] = test.value

			cycles := cpu.Step()

			assert.Equal(test.cycles, cycles, "incorrect cycle count")
			assert.Equal(test.value, cpu.A, "incorrect accumulator value")
		})
	}
}
