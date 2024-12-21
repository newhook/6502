package cia

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimerAInitialization(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	assert.Equal(uint16(0xFFFF), cia.registers.timerALatch, "Timer A latch should initialize to 0xFFFF")
	assert.Equal(uint16(0xFFFF), cia.registers.timerA, "Timer A counter should initialize to 0xFFFF")
	assert.Equal(uint8(0), cia.registers.cra, "CRA should initialize to 0")
}

func TestTimerALatchLoad(t *testing.T) {
	type testCase struct {
		name     string
		low      uint8
		high     uint8
		expected uint16
	}

	testCases := []testCase{
		{
			name:     "Load 0x1234",
			low:      0x34,
			high:     0x12,
			expected: 0x1234,
		},
		{
			name:     "Load 0xFFFF",
			low:      0xFF,
			high:     0xFF,
			expected: 0xFFFF,
		},
		{
			name:     "Load 0x0000",
			low:      0x00,
			high:     0x00,
			expected: 0x0000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cia := NewCIA()
			assert := assert.New(t)

			cia.WriteRegister(TA_LO, tc.low)
			cia.WriteRegister(TA_HI, tc.high)

			assert.Equal(tc.expected, cia.registers.timerALatch, "Timer A latch should be set correctly")
		})
	}
}

func TestTimerAForceLoad(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	// Set initial latch value
	cia.WriteRegister(TA_LO, 0x34)
	cia.WriteRegister(TA_HI, 0x12)

	// Force load and verify
	cia.WriteRegister(CRA, CRA_FORCE)
	assert.Equal(uint16(0x1234), cia.registers.timerA, "Timer should be force loaded")
	assert.Equal(uint8(0), cia.registers.cra&CRA_FORCE, "Force bit should clear automatically")
}

func TestTimerAContinuousMode(t *testing.T) {
	type testCase struct {
		name          string
		initialValue  uint16
		cycles        uint8
		expectedValue uint16
		expectReload  bool
	}

	testCases := []testCase{
		{
			name:          "Count down without reload",
			initialValue:  0x0003,
			cycles:        2,
			expectedValue: 0x0001,
			expectReload:  false,
		},
		{
			name:          "Count down with reload",
			initialValue:  0x0003,
			cycles:        3,
			expectedValue: 0x0003, // Will reload from latch
			expectReload:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cia := NewCIA()
			assert := assert.New(t)

			// Set timer value and start in continuous mode
			cia.WriteRegister(TA_LO, uint8(tc.initialValue&0xFF))
			cia.WriteRegister(TA_HI, uint8(tc.initialValue>>8))
			cia.WriteRegister(CRA, CRA_START)

			cia.Update(tc.cycles)
			assert.Equal(tc.expectedValue, cia.registers.timerA)
		})
	}
}

func TestTimerAOneShotMode(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	// Set timer value and start in one-shot mode
	cia.WriteRegister(TA_LO, 0x02)
	cia.WriteRegister(TA_HI, 0x00)
	cia.WriteRegister(CRA, CRA_START|CRA_RUNMODE)

	// Run until underflow
	cia.Update(2)
	assert.Equal(uint16(0x0002), cia.registers.timerA, "Timer should count down to 0")

	// Verify timer stopped
	cia.Update(1)
	assert.Equal(uint8(0), cia.registers.cra&CRA_START, "Timer should stop in one-shot mode")
}

func TestTimerAInterrupt(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	// Enable Timer A interrupt and set short count
	cia.WriteRegister(ICR, ICR_SET|ICR_TA)
	cia.WriteRegister(TA_LO, 0x01)
	cia.WriteRegister(TA_HI, 0x00)
	cia.WriteRegister(CRA, CRA_START)

	// Run and check interrupt
	event := cia.Update(1)
	assert.True(event.IRQ, "IRQ should be triggered on underflow")

	// Verify interrupt register
	icr := cia.ReadRegister(ICR)
	assert.Equal(uint8(0x81), icr, "ICR should indicate Timer A interrupt")
}

func TestTimerAPB6Output2(t *testing.T) {
	type testCase struct {
		name        string
		toggleMode  bool
		cycles      []uint8 // Series of cycle counts to test sequence
		expectedPB6 []uint8 // Expected PB6 state after each cycle count
	}

	testCases := []testCase{
		{
			name:        "Toggle mode sequence",
			toggleMode:  true,
			cycles:      []uint8{1, 2, 2, 2},             // Initial, underflow, next underflow, another underflow
			expectedPB6: []uint8{0x00, 0x40, 0x00, 0x40}, // Low, high, low, high
		},
		{
			name:        "Pulse mode sequence",
			toggleMode:  false,
			cycles:      []uint8{1, 1, 1, 1},             // Initial, underflow, after pulse, next underflow
			expectedPB6: []uint8{0x00, 0x40, 0x00, 0x40}, // Low, high, low, high
		},
		{
			name:        "PB6 output disabled",
			toggleMode:  false,
			cycles:      []uint8{2},    // Just test underflow
			expectedPB6: []uint8{0x00}, // Should stay low
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cia := NewCIA()
			assert := assert.New(t)

			// Set PB6 as output
			cia.WriteRegister(DDRB, 0x40)

			// Configure PB6 output mode
			cra := CRA_START
			if tc.name != "PB6 output disabled" {
				cra |= CRA_PBON
			}
			if tc.toggleMode {
				cra |= CRA_OUTMODE
			}
			cia.WriteRegister(CRA, cra)

			// Set timer value
			cia.WriteRegister(TA_LO, 0x02)
			cia.WriteRegister(TA_HI, 0x00)

			// Run through the sequence of cycles and check PB6 state
			for i, cycleCount := range tc.cycles {
				cia.Update(cycleCount)
				pb := cia.ReadRegister(PRB)
				assert.Equal(tc.expectedPB6[i], pb&0x40,
					fmt.Sprintf("PB6 state incorrect after cycle sequence %d", i))
			}
		})
	}
}

func TestTimerAPB6Output(t *testing.T) {
	type testCase struct {
		name        string
		toggleMode  bool
		cycles      uint8
		expectedPB6 uint8
		nextCycle   uint8
		nextPB6     uint8
	}

	testCases := []testCase{
		{
			name:        "Toggle mode",
			toggleMode:  true,
			cycles:      2,
			expectedPB6: 0x40,
			nextCycle:   1,
			nextPB6:     0x40, // Stays high until next underflow
		},
		{
			name:        "Pulse mode",
			toggleMode:  false,
			cycles:      2,
			expectedPB6: 0x40,
			nextCycle:   1,
			nextPB6:     0x00, // Returns low after one cycle
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cia := NewCIA()
			assert := assert.New(t)

			// Set PB6 as output
			cia.WriteRegister(DDRB, 0x40)

			// Configure PB6 output mode
			cra := CRA_START | CRA_PBON
			if tc.toggleMode {
				cra |= CRA_OUTMODE
			}
			cia.WriteRegister(CRA, cra)

			// Set timer value
			cia.WriteRegister(TA_LO, 0x02)
			cia.WriteRegister(TA_HI, 0x00)

			// Run until underflow
			cia.Update(tc.cycles)
			pb := cia.ReadRegister(PRB)
			assert.Equal(tc.expectedPB6, pb&0x40, "PB6 initial state incorrect")

			// Run additional cycle and check
			cia.Update(tc.nextCycle)
			pb = cia.ReadRegister(PRB)
			assert.Equal(tc.nextPB6, pb&0x40, "PB6 subsequent state incorrect")
		})
	}
}

func TestTimerAStop(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	// Start timer with a value
	cia.WriteRegister(TA_LO, 0x05)
	cia.WriteRegister(TA_HI, 0x00)
	cia.WriteRegister(CRA, CRA_START)

	// Run for 2 cycles
	cia.Update(2)
	initialValue := cia.registers.timerA

	// Stop timer
	cia.WriteRegister(CRA, 0)

	// Run more cycles
	cia.Update(2)
	assert.Equal(initialValue, cia.registers.timerA, "Timer should not count when stopped")
}

func TestTimerAReload(t *testing.T) {
	cia := NewCIA()
	assert := assert.New(t)

	// Set initial latch value
	cia.WriteRegister(TA_LO, 0x03)
	cia.WriteRegister(TA_HI, 0x00)

	// Start timer and run to underflow
	cia.WriteRegister(CRA, CRA_START)
	cia.Update(3)

	// Verify reload from latch
	assert.Equal(uint16(0x0003), cia.registers.timerA, "Timer should reload from latch after underflow")
}

func TestTimerAReadRegister(t *testing.T) {
	type testCase struct {
		name    string
		value   uint16
		regLow  uint8
		regHigh uint8
	}

	testCases := []testCase{
		{
			name:    "Read 0x1234",
			value:   0x1234,
			regLow:  0x34,
			regHigh: 0x12,
		},
		{
			name:    "Read 0xFFFF",
			value:   0xFFFF,
			regLow:  0xFF,
			regHigh: 0xFF,
		},
		{
			name:    "Read 0x0000",
			value:   0x0000,
			regLow:  0x00,
			regHigh: 0x00,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cia := NewCIA()
			assert := assert.New(t)

			cia.registers.timerA = tc.value

			lowByte := cia.ReadRegister(TA_LO)
			highByte := cia.ReadRegister(TA_HI)

			assert.Equal(tc.regLow, lowByte, "Timer A low byte read incorrect")
			assert.Equal(tc.regHigh, highByte, "Timer A high byte read incorrect")
		})
	}
}
