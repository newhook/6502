package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBranchInstructions(t *testing.T) {
	tests := []struct {
		name      string
		opcode    uint8
		offset    int8
		initPC    uint16
		initFlags uint8
		expectPC  uint16
		cycles    uint8
		desc      string
	}{
		// BCC tests
		{
			name:      "BCC taken",
			opcode:    BCC,
			offset:    10,
			initPC:    0x0200,
			initFlags: 0,
			expectPC:  0x020C,
			cycles:    3,
			desc:      "Branch forward, carry clear",
		},
		{
			name:      "BCC not taken",
			opcode:    BCC,
			offset:    10,
			initPC:    0x0200,
			initFlags: FlagC,
			expectPC:  0x0202,
			cycles:    2,
			desc:      "Don't branch, carry set",
		},
		{
			name:      "BCC page cross",
			opcode:    BCC,
			offset:    127,
			initPC:    0x02F0,
			initFlags: 0,
			expectPC:  0x0371,
			cycles:    4,
			desc:      "Branch crosses page boundary",
		},
		// BCS tests
		{
			name:      "BCS taken",
			opcode:    BCS,
			offset:    -10,
			initPC:    0x0200,
			initFlags: FlagC,
			expectPC:  0x01F8,
			cycles:    4,
			desc:      "Branch backward, carry set",
		},
		// BEQ tests
		{
			name:      "BEQ taken",
			opcode:    BEQ,
			offset:    5,
			initPC:    0x0200,
			initFlags: FlagZ,
			expectPC:  0x0207,
			cycles:    3,
			desc:      "Branch forward, zero set",
		},
		// BMI tests
		{
			name:      "BMI taken",
			opcode:    BMI,
			offset:    -5,
			initPC:    0x0200,
			initFlags: FlagN,
			expectPC:  0x01FD,
			cycles:    4,
			desc:      "Branch backward, negative set",
		},
		// BNE tests
		{
			name:      "BNE taken",
			opcode:    BNE,
			offset:    15,
			initPC:    0x0200,
			initFlags: 0,
			expectPC:  0x0211,
			cycles:    3,
			desc:      "Branch forward, zero clear",
		},
		// BPL tests
		{
			name:      "BPL taken",
			opcode:    BPL,
			offset:    -15,
			initPC:    0x0200,
			initFlags: 0,
			expectPC:  0x01F3,
			cycles:    4,
			desc:      "Branch backward, negative clear",
		},
		// BVC tests
		{
			name:      "BVC taken",
			opcode:    BVC,
			offset:    20,
			initPC:    0x0200,
			initFlags: 0,
			expectPC:  0x0216,
			cycles:    3,
			desc:      "Branch forward, overflow clear",
		},
		// BVS tests
		{
			name:      "BVS taken",
			opcode:    BVS,
			offset:    -20,
			initPC:    0x0200,
			initFlags: FlagV,
			expectPC:  0x01EE,
			cycles:    4,
			desc:      "Branch backward, overflow set",
		},
		// Page boundary crossing examples for each
		{
			name:      "BEQ page cross forward",
			opcode:    BEQ,
			offset:    127,
			initPC:    0x02F0,
			initFlags: FlagZ,
			expectPC:  0x0371,
			cycles:    4,
			desc:      "Branch crosses page boundary forward",
		},
		{
			name:      "BMI page cross backward",
			opcode:    BMI,
			offset:    -128,
			initPC:    0x0280,
			initFlags: FlagN,
			expectPC:  0x0202,
			cycles:    3,
			desc:      "Branch crosses page boundary backward",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPUAndMemory()

			// Setup initial state
			cpu.P = tt.initFlags

			// Setup memory
			cpu.Memory[tt.initPC] = tt.opcode
			cpu.Memory[tt.initPC+1] = uint8(tt.offset)
			cpu.PC = tt.initPC + 1

			// Execute
			cycles := cpu.execute(tt.opcode)

			// Verify
			assert.Equal(t, tt.expectPC, cpu.PC,
				"PC incorrect: %s", tt.desc)
			assert.Equal(t, tt.cycles, cycles,
				"Cycle count incorrect: %s", tt.desc)
		})
	}
}
