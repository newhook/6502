package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBRKNOPRTI(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*CPUAndMemory)
		execute func(*CPUAndMemory)
		verify  func(*CPUAndMemory, *testing.T)
		desc    string
	}{
		{
			name: "BRK pushes PC+2 and flags, loads IRQ vector",
			setup: func(c *CPUAndMemory) {
				c.PC = 0x1000
				c.P = 0x20  // Some arbitrary flags
				c.SP = 0xFF // Stack pointer at top
				// Set up IRQ vector
				c.Memory[0xFFFE] = 0x34
				c.Memory[0xFFFF] = 0x12 // IRQ handler at 0x1234
			},
			execute: func(c *CPUAndMemory) {
				c.execute(BRK)
			},
			verify: func(c *CPUAndMemory, t *testing.T) {
				// Check PC pushed to stack (should be PC+2)
				highByte := c.Memory[0x01FF]
				lowByte := c.Memory[0x01FE]
				pushedPC := uint16(highByte)<<8 | uint16(lowByte)
				assert.Equal(t, uint16(0x1002), pushedPC, "PC+2 should be pushed to stack")

				// Check status pushed with B flag set
				assert.Equal(t, uint8(0x30), c.Memory[0x01FD], "Status with B flag set should be pushed")

				// Check new PC is from IRQ vector
				assert.Equal(t, uint16(0x1234), c.PC, "PC should be loaded from IRQ vector")

				// Check I flag was set
				assert.True(t, c.P&FlagI != 0, "I flag should be set")

				// Check stack pointer
				assert.Equal(t, uint8(0xFC), c.SP, "Stack pointer should be decremented by 3")
			},
			desc: "Test BRK instruction flow",
		},
		{
			name: "NOP does nothing but takes 2 cycles",
			setup: func(c *CPUAndMemory) {
				c.PC = 0x1000
				c.P = 0x20 // Some arbitrary flags
				c.A = 0x42 // Set some registers
				c.X = 0x24
				c.Y = 0x35
			},
			execute: func(c *CPUAndMemory) {
				cycles := c.execute(NOP)
				assert.Equal(t, uint8(2), cycles, "NOP should take 2 cycles")
			},
			verify: func(c *CPUAndMemory, t *testing.T) {
				// Verify nothing changed except PC
				assert.Equal(t, uint8(0x20), c.P, "Flags should be unchanged")
				assert.Equal(t, uint8(0x42), c.A, "A should be unchanged")
				assert.Equal(t, uint8(0x24), c.X, "X should be unchanged")
				assert.Equal(t, uint8(0x35), c.Y, "Y should be unchanged")
			},
			desc: "Test NOP instruction",
		},
		{
			name: "RTI pulls flags and PC correctly",
			setup: func(c *CPUAndMemory) {
				c.SP = 0xFC
				// Setup stack with status and return address
				c.Memory[0x01FD] = 0x20 // Status (without B flag)
				c.Memory[0x01FE] = 0x34 // PC low byte
				c.Memory[0x01FF] = 0x12 // PC high byte
			},
			execute: func(c *CPUAndMemory) {
				cycles := c.execute(RTI)
				assert.Equal(t, uint8(6), cycles, "RTI should take 6 cycles")
			},
			verify: func(c *CPUAndMemory, t *testing.T) {
				// Check restored PC
				assert.Equal(t, uint16(0x1234), c.PC, "PC should be restored from stack")

				// Check restored status (B flag should be clear)
				assert.Equal(t, uint8(0x20), c.P, "Status should be restored without B flag")

				// Check stack pointer
				assert.Equal(t, uint8(0xFF), c.SP, "Stack pointer should be restored")
			},
			desc: "Test RTI instruction flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := NewCPUAndMemory()
			tt.setup(cpu)
			tt.execute(cpu)
			tt.verify(cpu, t)
		})
	}
}
