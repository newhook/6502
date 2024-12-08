package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJumpInstructions(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*CPUAndMemory)
		opcode  uint8
		execute func(*CPUAndMemory)
		verify  func(*CPUAndMemory, *testing.T)
		cycles  uint8
	}{
		{
			name: "JMP Absolute",
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0000] = JMP_ABS
				c.Memory[0x0001] = 0x34
				c.Memory[0x0002] = 0x12 // Jump target: 0x1234
			},
			opcode: JMP_ABS,
			verify: func(c *CPUAndMemory, t *testing.T) {
				assert.Equal(t, uint16(0x1234), c.PC, "PC should be 0x1234")
			},
			cycles: 3,
		},
		{
			name: "JMP Indirect",
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0000] = JMP_IND
				c.Memory[0x0001] = 0x34
				c.Memory[0x0002] = 0x12 // Indirect address: 0x1234
				c.Memory[0x1234] = 0x78
				c.Memory[0x1235] = 0x56 // Jump target: 0x5678
			},
			opcode: JMP_IND,
			verify: func(c *CPUAndMemory, t *testing.T) {
				assert.Equal(t, uint16(0x5678), c.PC, "PC should be 0x5678")
			},
			cycles: 5,
		},
		{
			name: "JMP Indirect Page Boundary Bug",
			setup: func(c *CPUAndMemory) {
				c.Memory[0x0000] = JMP_IND
				c.Memory[0x0001] = 0xFF
				c.Memory[0x0002] = 0x12 // Indirect address: 0x12FF
				c.Memory[0x12FF] = 0x78
				c.Memory[0x1200] = 0x56 // Should read from 0x1200, not 0x1300
			},
			opcode: JMP_IND,
			verify: func(c *CPUAndMemory, t *testing.T) {
				assert.Equal(t, uint16(0x5678), c.PC, "PC should be 0x5678")
			},
			cycles: 5,
		},
		{
			name: "JSR Absolute",
			setup: func(c *CPUAndMemory) {
				c.PC = 0x0000
				c.SP = 0xFF
				c.Memory[0x0000] = JSR_ABS
				c.Memory[0x0001] = 0x34
				c.Memory[0x0002] = 0x12 // Subroutine address: 0x1234
			},
			opcode: JSR_ABS,
			verify: func(c *CPUAndMemory, t *testing.T) {
				assert.Equal(t, uint16(0x1234), c.PC, "PC should be 0x1234")
				// Check stack contains return address - 1
				returnAddr := uint16(c.Memory[0x01FF])<<8 | uint16(c.Memory[0x01FE])
				assert.Equal(t, uint16(0x0002), returnAddr, "Return address should be 0x0002")
				assert.Equal(t, uint8(0xFD), c.SP, "Stack pointer should be 0xFD")
			},
			cycles: 6,
		},
		{
			name: "RTS",
			setup: func(c *CPUAndMemory) {
				c.SP = 0xFD
				// Setup stack with return address - 1
				c.Memory[0x01FE] = 0x34 // Low byte
				c.Memory[0x01FF] = 0x12 // High byte
				c.Memory[0x0000] = RTS
			},
			opcode: RTS,
			verify: func(c *CPUAndMemory, t *testing.T) {
				assert.Equal(t, uint16(0x1235), c.PC, "PC should be 0x1235 (return address + 1)")
				assert.Equal(t, uint8(0xFF), c.SP, "Stack pointer should be 0xFF")
			},
			cycles: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPUAndMemory()
			tt.setup(cpu)
			cpu.PC = 1

			cycles := cpu.execute(tt.opcode)
			tt.verify(cpu, t)

			if tt.cycles > 0 {
				assert.Equal(t, tt.cycles, cycles, "Incorrect cycle count")
			}
		})
	}
}
