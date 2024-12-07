package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterIncDec(t *testing.T) {
	tests := []struct {
		name   string
		opcode uint8
		isX    bool // true for X register operations, false for Y
		isInc  bool // true for increment, false for decrement
	}{
		{"INX", INX, true, true},
		{"INY", INY, false, true},
		{"DEX", DEX, true, false},
		{"DEY", DEY, false, false},
	}

	testValues := []struct {
		initial   uint8
		expectedZ bool
		expectedN bool
	}{
		{0x00, false, false}, // 0->1 or 0->255
		{0x7F, false, true},  // 127->128 or 127->126
		{0xFE, false, true},  // 254->255 or 254->253
		{0xFF, true, false},  // 255->0 or 255->254
		{0x80, false, false}, // 128->129 or 128->127
	}

	for _, tt := range tests {
		for _, tv := range testValues {
			initial := tv.initial
			var expected uint8
			var expectedZ, expectedN bool

			if tt.isInc {
				expected = initial + 1
				// Update expected flags for increment
				expectedZ = expected == 0
				expectedN = expected&0x80 != 0
			} else {
				expected = initial - 1
				// Update expected flags for decrement
				expectedZ = expected == 0
				expectedN = expected&0x80 != 0
			}

			testName := tt.name + "_" +
				string(initial) + "_to_" +
				string(expected)

			t.Run(testName, func(t *testing.T) {
				cpu := NewCPU()

				// Set initial register value
				if tt.isX {
					cpu.X = initial
				} else {
					cpu.Y = initial
				}

				cpu.Memory[0] = tt.opcode
				cycles := cpu.execute(tt.opcode)

				// Check cycles
				assert.Equal(t, uint8(2), cycles,
					"Incorrect number of cycles")

				// Check register value
				if tt.isX {
					assert.Equal(t, expected, cpu.X,
						"X register value incorrect")
				} else {
					assert.Equal(t, expected, cpu.Y,
						"Y register value incorrect")
				}

				// Check flags
				assert.Equal(t, expectedZ, (cpu.P&FlagZ) != 0,
					"Zero flag mismatch")
				assert.Equal(t, expectedN, (cpu.P&FlagN) != 0,
					"Negative flag mismatch")

				// Check that the other register wasn't modified
				if tt.isX {
					assert.Equal(t, uint8(0), cpu.Y,
						"Y register should not be modified")
				} else {
					assert.Equal(t, uint8(0), cpu.X,
						"X register should not be modified")
				}
			})
		}
	}
}
