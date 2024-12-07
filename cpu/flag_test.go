package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlagInstructions(t *testing.T) {
	tests := []struct {
		name      string
		opcode    uint8
		initFlags uint8
		checkFlag uint8
		expectSet bool
		desc      string
	}{
		{
			name:      "CLC clears carry flag",
			opcode:    CLC,
			initFlags: FlagC,
			checkFlag: FlagC,
			expectSet: false,
			desc:      "Carry should be cleared",
		},
		{
			name:      "CLC with carry already clear",
			opcode:    CLC,
			initFlags: 0,
			checkFlag: FlagC,
			expectSet: false,
			desc:      "Carry should remain clear",
		},
		{
			name:      "CLD clears decimal flag",
			opcode:    CLD,
			initFlags: FlagD,
			checkFlag: FlagD,
			expectSet: false,
			desc:      "Decimal should be cleared",
		},
		{
			name:      "CLI clears interrupt flag",
			opcode:    CLI,
			initFlags: FlagI,
			checkFlag: FlagI,
			expectSet: false,
			desc:      "Interrupt should be cleared",
		},
		{
			name:      "CLV clears overflow flag",
			opcode:    CLV,
			initFlags: FlagV,
			checkFlag: FlagV,
			expectSet: false,
			desc:      "Overflow should be cleared",
		},
		{
			name:      "SEC sets carry flag",
			opcode:    SEC,
			initFlags: 0,
			checkFlag: FlagC,
			expectSet: true,
			desc:      "Carry should be set",
		},
		{
			name:      "SEC with carry already set",
			opcode:    SEC,
			initFlags: FlagC,
			checkFlag: FlagC,
			expectSet: true,
			desc:      "Carry should remain set",
		},
		{
			name:      "SED sets decimal flag",
			opcode:    SED,
			initFlags: 0,
			checkFlag: FlagD,
			expectSet: true,
			desc:      "Decimal should be set",
		},
		{
			name:      "SEI sets interrupt flag",
			opcode:    SEI,
			initFlags: 0,
			checkFlag: FlagI,
			expectSet: true,
			desc:      "Interrupt should be set",
		},
		// Test that other flags are unaffected
		{
			name:      "CLC only affects carry",
			opcode:    CLC,
			initFlags: FlagD | FlagI | FlagV | FlagC,
			checkFlag: FlagD | FlagI | FlagV,
			expectSet: true,
			desc:      "Other flags should remain unchanged",
		},
		{
			name:      "SEC only affects carry",
			opcode:    SEC,
			initFlags: FlagD | FlagI | FlagV,
			checkFlag: FlagD | FlagI | FlagV,
			expectSet: true,
			desc:      "Other flags should remain unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPU()

			// Setup
			cpu.P = tt.initFlags
			cpu.Memory[0] = tt.opcode

			// Execute
			cycles := cpu.execute(tt.opcode)

			// Verify cycles
			assert.Equal(t, uint8(2), cycles,
				"All flag instructions should take 2 cycles")

			// Verify flag state
			if tt.expectSet {
				assert.Equal(t, tt.checkFlag, cpu.P&tt.checkFlag,
					"Expected flags not set: %s", tt.desc)
			} else {
				assert.Equal(t, uint8(0), cpu.P&tt.checkFlag,
					"Expected flags not cleared: %s", tt.desc)
			}
		})
	}
}
