package cia

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTODClock(t *testing.T) {
	t.Run("basic clock counting", func(t *testing.T) {
		cia := NewCIA()

		// Write initial time: 1:59:59.9 AM
		cia.WriteRegister(TOD_HR, 0x01)    // 1 AM (BCD)
		cia.WriteRegister(TOD_MIN, 0x59)   // 59 in BCD = 0x59
		cia.WriteRegister(TOD_SEC, 0x59)   // 59 in BCD = 0x59
		cia.WriteRegister(TOD_10THS, 0x09) // 9 in BCD = 0x09

		// Tick one tenth
		cia.updateTOD()

		// Should now be 2:00:00.0
		assert.Equal(t, uint8(0x02), cia.Registers.todHr&0x1F) // Hour = 2
		assert.Equal(t, uint8(0x00), cia.Registers.todMin)
		assert.Equal(t, uint8(0x00), cia.Registers.todSec)
		assert.Equal(t, uint8(0x00), cia.Registers.todTenths)
	})

	t.Run("AM/PM transition", func(t *testing.T) {
		cia := NewCIA()

		// Set 11:59:59.9 AM
		cia.WriteRegister(TOD_HR, 0x11) // 11 AM in BCD
		cia.WriteRegister(TOD_MIN, 0x59)
		cia.WriteRegister(TOD_SEC, 0x59)
		cia.WriteRegister(TOD_10THS, 0x09)

		// Tick to 12 PM
		cia.updateTOD()

		// Should be 12:00:00.0 PM
		assert.Equal(t, uint8(0x92), cia.Registers.todHr) // 12 PM in BCD (0x80 for PM | 0x12)
		assert.Equal(t, uint8(0x00), cia.Registers.todMin)
		assert.Equal(t, uint8(0x00), cia.Registers.todSec)
		assert.Equal(t, uint8(0x00), cia.Registers.todTenths)
	})

	t.Run("12 hour rollover", func(t *testing.T) {
		cia := NewCIA()

		// Set 12:59:59.9 PM
		cia.WriteRegister(TOD_HR, 0x92)    // 12 PM in BCD (0x80 for PM | 0x12 for 12)
		cia.WriteRegister(TOD_MIN, 0x59)   // 59 in BCD
		cia.WriteRegister(TOD_SEC, 0x59)   // 59 in BCD
		cia.WriteRegister(TOD_10THS, 0x09) // 9 in BCD

		// Tick to 1 PM
		cia.updateTOD()

		// Should be 1:00:00.0 PM
		assert.Equal(t, uint8(0x81), cia.Registers.todHr) // 1 with PM bit set
		assert.Equal(t, uint8(0x00), cia.Registers.todMin)
		assert.Equal(t, uint8(0x00), cia.Registers.todSec)
		assert.Equal(t, uint8(0x00), cia.Registers.todTenths)
	})
}

func TestTODAlarm(t *testing.T) {
	t.Run("basic alarm trigger", func(t *testing.T) {
		cia := NewCIA()

		// Set current time to 1:59:59.9 AM
		cia.WriteRegister(TOD_HR, 0x01)
		cia.WriteRegister(TOD_MIN, 0x59)
		cia.WriteRegister(TOD_SEC, 0x59)
		cia.WriteRegister(TOD_10THS, 0x09)

		// Set alarm for 2:00:00.0 AM
		cia.WriteRegister(CRB, cia.Registers.crb|CRB_ALARM) // Enable alarm set
		cia.WriteRegister(TOD_HR, 0x02)
		cia.WriteRegister(TOD_MIN, 0x00)
		cia.WriteRegister(TOD_SEC, 0x00)
		cia.WriteRegister(TOD_10THS, 0x00)
		cia.WriteRegister(CRB, cia.Registers.crb&^CRB_ALARM) // Disable alarm set

		// Enable TOD interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TOD)

		// Tick the clock
		cia.updateTOD()

		// Check if alarm triggered
		assert.True(t, (cia.Registers.icrData&ICR_TOD) != 0)
	})

	t.Run("alarm with PM bit", func(t *testing.T) {
		cia := NewCIA()

		// Set current time to 11:59:59.9 AM
		cia.WriteRegister(TOD_HR, 0x11) // 11 in BCD
		cia.WriteRegister(TOD_MIN, 0x59)
		cia.WriteRegister(TOD_SEC, 0x59)
		cia.WriteRegister(TOD_10THS, 0x09)

		// Set alarm for 12:00:00.0 PM
		cia.WriteRegister(CRB, cia.Registers.crb|CRB_ALARM)
		cia.WriteRegister(TOD_HR, 0x92) // 12 PM in BCD (0x80 for PM | 0x12)
		cia.WriteRegister(TOD_MIN, 0x00)
		cia.WriteRegister(TOD_SEC, 0x00)
		cia.WriteRegister(TOD_10THS, 0x00)
		cia.WriteRegister(CRB, cia.Registers.crb&^CRB_ALARM)

		// Enable TOD interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TOD)

		// Tick the clock
		cia.updateTOD()

		// Check if alarm triggered
		assert.True(t, (cia.Registers.icrData&ICR_TOD) != 0)
	})

	t.Run("TOD frequency", func(t *testing.T) {
		cia := NewCIA()

		// Test 60Hz (default)
		assert.Equal(t, uint16(16667), cia.todPeriod())

		// Test 50Hz
		cia.WriteRegister(CRA, CRA_TODIN)
		assert.Equal(t, uint16(20000), cia.todPeriod())
	})

	t.Run("invalid hour handling", func(t *testing.T) {
		cia := NewCIA()

		// Try to set hour to 0 (should become 12)
		cia.WriteRegister(TOD_HR, 0x00)
		assert.Equal(t, uint8(0x12), cia.Registers.todHr) // 12 in BCD

		// Try to set alarm hour to 0 (should become 12)
		cia.WriteRegister(CRB, cia.Registers.crb|CRB_ALARM)
		cia.WriteRegister(TOD_HR, 0x00)
		cia.WriteRegister(CRB, cia.Registers.crb&^CRB_ALARM)
		assert.Equal(t, uint8(0x12), cia.todAlarm[3]) // 12 in BCD
	})
}
