package cia

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterrupts(t *testing.T) {
	t.Run("timer A interrupt", func(t *testing.T) {
		cia := NewCIA()

		// Enable Timer A interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TA)

		// Set Timer A to 2 (write low byte first)
		cia.WriteRegister(TA_LO, 0x02)
		cia.WriteRegister(TA_HI, 0x00)

		// Verify timer is loaded with correct value
		assert.Equal(t, uint16(0x0002), cia.registers.timerA)

		// Start Timer A in one-shot mode
		cia.WriteRegister(CRA, CRA_START|CRA_RUNMODE)

		// First cycle - should decrement to 1
		cia.Update(1)
		assert.Equal(t, uint16(0x0001), cia.registers.timerA)
		assert.Equal(t, uint8(0), cia.registers.icrData&ICR_TA)

		// Second cycle - should decrement to 0 and trigger interrupt
		cia.Update(1)
		assert.Equal(t, uint16(0x0002), cia.registers.timerA)
		assert.Equal(t, ICR_TA, cia.registers.icrData&ICR_TA)

		// Verify timer stopped (one-shot mode)
		assert.Equal(t, uint8(0), cia.registers.cra&CRA_START)

		// Reading ICR should clear interrupts and return correct flags
		irqBefore := cia.ReadRegister(ICR)
		assert.Equal(t, uint8(0x80|ICR_TA), irqBefore)

		// Second read should show cleared state
		irqAfter := cia.ReadRegister(ICR)
		assert.Equal(t, uint8(0), irqAfter)
	})

	t.Run("multiple interrupts", func(t *testing.T) {
		cia := NewCIA()

		// Enable Timer A and B interrupts
		cia.WriteRegister(ICR, ICR_SET|ICR_TA|ICR_TB)

		// Set both timers to 1
		cia.WriteRegister(TA_LO, 0x01)
		cia.WriteRegister(TA_HI, 0x00)
		cia.WriteRegister(TB_LO, 0x01)
		cia.WriteRegister(TB_HI, 0x00)

		// Start both timers
		cia.WriteRegister(CRA, CRA_START)
		cia.WriteRegister(CRB, CRB_START)

		// One cycle should trigger both interrupts
		cia.Update(1)

		// Check both interrupts are set
		irq := cia.ReadRegister(ICR)
		assert.Equal(t, 0x80|ICR_TA|ICR_TB, irq)
	})

	t.Run("interrupt masking", func(t *testing.T) {
		cia := NewCIA()

		// Enable only Timer A interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TA)

		// Set both timers to 1
		cia.WriteRegister(TA_LO, 0x01)
		cia.WriteRegister(TA_HI, 0x00)
		cia.WriteRegister(TB_LO, 0x01)
		cia.WriteRegister(TB_HI, 0x00)

		// Start both timers
		cia.WriteRegister(CRA, CRA_START)
		cia.WriteRegister(CRB, CRB_START)

		// One cycle should trigger only Timer A interrupt
		cia.Update(1)

		irq := cia.ReadRegister(ICR)
		assert.Equal(t, uint8(0x80|ICR_TA), irq)
	})

	t.Run("interrupt clear", func(t *testing.T) {
		cia := NewCIA()

		// Enable Timer A interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TA)

		// Set and start Timer A
		cia.WriteRegister(TA_LO, 0x01)
		cia.WriteRegister(CRA, CRA_START)

		// Trigger interrupt
		cia.Update(1)

		// Clear Timer A interrupt
		cia.WriteRegister(ICR, ICR_TA) // Write without SET bit clears

		irq := cia.ReadRegister(ICR)
		assert.Equal(t, uint8(0), irq)
	})

	t.Run("continuous vs one-shot interrupts", func(t *testing.T) {
		cia := NewCIA()

		// Enable Timer A interrupt
		cia.WriteRegister(ICR, ICR_SET|ICR_TA)

		// Set Timer A to 1
		cia.WriteRegister(TA_LO, 0x01)
		cia.WriteRegister(TA_HI, 0x00)

		t.Run("continuous mode", func(t *testing.T) {
			// Start Timer A in continuous mode
			cia.WriteRegister(CRA, CRA_START)

			// Should get interrupt every cycle
			for i := 0; i < 3; i++ {
				cia.Update(1)
				irq := cia.ReadRegister(ICR)
				assert.Equal(t, uint8(0x80|ICR_TA), irq)
			}
		})

		t.Run("one-shot mode", func(t *testing.T) {
			// Start Timer A in one-shot mode
			cia.WriteRegister(CRA, CRA_START|CRA_RUNMODE)

			// First cycle should trigger interrupt
			cia.Update(1)
			irq := cia.ReadRegister(ICR)
			assert.Equal(t, uint8(0x80|ICR_TA), irq)

			// Subsequent cycles should not
			cia.Update(1)
			irq = cia.ReadRegister(ICR)
			assert.Equal(t, uint8(0), irq)
		})
	})
}
