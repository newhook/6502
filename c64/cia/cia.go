package cia

type TimerMode uint8

const (
	// Timer control bits
	TIMER_START    uint8 = 0x01
	TIMER_PBON    uint8 = 0x02  // Port B output
	TIMER_OUTMODE uint8 = 0x04  // Toggle/Pulse
	TIMER_RUNMODE uint8 = 0x08  // One-shot/Continuous
	TIMER_FORCE   uint8 = 0x10  // Force latched value
	TIMER_INMODE  uint8 = 0x20  // Count CPU cycles/CNT transitions
	TIMER_SERIAL  uint8 = 0x40  // Serial port input/output
	TIMER_50HZ    uint8 = 0x80  // 50/60Hz real-time clock
)

// Timer represents one CIA timer (A or B)
type Timer struct {
	counter   uint16 // Current counter value
	latch     uint16 // Latched value to load
	control   uint8  // Control register
	running   bool
	underflow bool   // Set when timer underflows
}

// CIA represents a complete 6526 CIA chip
type CIA struct {
	// Timers
	TimerA Timer
	TimerB Timer

	// I/O ports
	PortA      uint8
	PortB      uint8
	DirA       uint8 // Data direction for port A
	DirB       uint8 // Data direction for port B

	// Time of Day clock
	TOD_Hours   uint8
	TOD_Minutes uint8
	TOD_Seconds uint8
	TOD_Tenths  uint8
	TOD_Latch   [4]uint8
	TOD_Alarm   [4]uint8
	TOD_Running bool
	TOD_50Hz    bool

	// Serial port
	Serial     uint8
	SerialCnt  uint8

	// Interrupt control
	InterruptMask  uint8
	InterruptData  uint8
	InterruptState bool

	// Cycle counting
	cycles uint64
}

func NewCIA() *CIA {
	return &CIA{
		TimerA: Timer{latch: 0xFFFF},
		TimerB: Timer{latch: 0xFFFF},
		TOD_Running: true,
	}
}

// Update advances the CIA state by the specified number of cycles
func (c *CIA) Update(cycles uint8) *CIAEvent {
	c.cycles += uint64(cycles)
	var event *CIAEvent

	// Update timers
	if timerEvent := c.updateTimers(cycles); timerEvent != nil {
		event = timerEvent
	}

	// Update Time of Day clock
	c.updateTOD()

	// Update serial port if active
	if c.TimerA.control&TIMER_SERIAL != 0 {
		c.updateSerial()
	}

	return event
}

func (c *CIA) updateTimers(cycles uint8) *CIAEvent {
	// Update Timer A
	if c.TimerA.running {
		for i := uint8(0); i < cycles; i++ {
			c.TimerA.counter--
			if c.TimerA.counter == 0 {
				c.TimerA.underflow = true

				// Check if Timer B is counting Timer A underflows
				if c.TimerB.control&TIMER_INMODE != 0 {
					c.TimerB.counter--
				}

				// Reload from latch if continuous mode
				if c.TimerA.control&TIMER_RUNMODE == 0 {
					c.TimerA.counter = c.TimerA.latch
				} else {
					c.TimerA.running = false
				}

				// Generate interrupt if enabled
				if c.InterruptMask&0x01 != 0 {
					c.InterruptData |= 0x01
					c.InterruptState = true
					return &CIAEvent{Type: EventIRQ}
				}
			}
		}
	}

	// Update Timer B
	if c.TimerB.running {
		for i := uint8(0); i < cycles; i++ {
			if c.TimerB.control&TIMER_INMODE == 0 {
				c.TimerB.counter--
			}
			if c.TimerB.counter == 0 {
				c.TimerB.underflow = true

				// Reload from latch if continuous mode
				if c.TimerB.control&TIMER_RUNMODE == 0 {
					c.TimerB.counter = c.TimerB.latch
				} else {
					c.TimerB.running = false
				}

				// Generate interrupt if enabled
				if c.InterruptMask&0x02 != 0 {
					c.InterruptData |= 0x02
					c.InterruptState = true
					return &CIAEvent{Type: EventIRQ}
				}
			}
		}
	}

	return nil
}

func (c *CIA) updateTOD() {
	if !c.TOD_Running {
		return
	}

	// Update every 1/10th second
	todCycles := c.TOD_50Hz ? 19656 : 19968 // PAL/NTSC cycles for 1/10th second
	if c.cycles >= uint64(todCycles) {
		c.cycles -= uint64(todCycles)

		// Update TOD registers
		c.TOD_Tenths++
		if c.TOD_Tenths >= 10 {
			c.TOD_Tenths = 0
			c.TOD_Seconds++
			if c.TOD_Seconds >= 60 {
				c.TOD_Seconds = 0
				c.TOD_Minutes++
				if c.TOD_Minutes >= 60 {
					c.TOD_Minutes = 0
					c.TOD_Hours++
					if c.TOD_Hours >= 12 {
						c.TOD_Hours = 0
					}
				}
			}
		}

		// Check alarm
		if c.checkAlarm() {
			c.InterruptData |= 0x04
			c.InterruptState = true
		}
	}
}

func (c *CIA) checkAlarm() bool {
	return c.TOD_Hours == c.TOD_Alarm[0] &&
		c.TOD_Minutes == c.TOD_Alarm[1] &&
		c.TOD_Seconds == c.TOD_Alarm[2] &&
		c.TOD_Tenths == c.TOD_Alarm[3]
}

// Register access methods
func (c *CIA) WriteRegister(reg uint8, value uint8) {
	switch reg {
	case 0x00: // Port A data
		c.PortA = (c.PortA & ^c.DirA) | (value & c.DirA)
	case 0x01: // Port B data
		c.PortB = (c.PortB & ^c.DirB) | (value & c.DirB)
	case 0x02: // Port A direction
		c.DirA = value
	case 0x03: // Port B direction
		c.DirB = value
	case 0x04, 0x05: // Timer A latch
		if reg == 0x04 {
			c.TimerA.latch = (c.TimerA.latch & 0xFF00) | uint16(value)
		} else {
			c.TimerA.latch = (c.TimerA.latch & 0x00FF) | (uint16(value) << 8)
		}
	case 0x06, 0x07: // Timer B latch
		if reg == 0x06 {
			c.TimerB.latch = (c.TimerB.latch & 0xFF00) | uint16(value)
		} else {
			c.TimerB.latch = (c.TimerB.latch & 0x00FF) | (uint16(value) << 8)
		}
	case 0x0E: // Control Register A
		c.TimerA.control = value
		if value&TIMER_START != 0 {
			c.TimerA.counter = c.TimerA.latch
			c.TimerA.running = true
		}
	case 0x0F: // Control Register B
		c.TimerB.control = value
		if value&TIMER_START != 0 {
			c.TimerB.counter = c.TimerB.latch
			c.TimerB.running = true
		}
	}
}

func (c *CIA) ReadRegister(reg uint8) uint8 {
	switch reg {
	case 0x00: // Port A
		return (c.PortA & c.DirA) | (0xFF & ^c.DirA)
	case 0x01: // Port B
		return (c.PortB & c.DirB) | (0xFF & ^c.DirB)
	case 0x02: // Port A direction
		return c.DirA
	case 0x03: // Port B direction
		return c.DirB
	case 0x04: // Timer A low
		return uint8(c.TimerA.counter & 0xFF)
	case 0x05: // Timer A high
		return uint8(c.TimerA.counter >> 8)
	case 0x06: // Timer B low
		return uint8(c.TimerB.counter & 0xFF)
	case 0x07: // Timer B high
		return uint8(c.TimerB.counter >> 8)
	case 0x0D: // Interrupt Data
		value := c.InterruptData
		c.InterruptData = 0
		c.InterruptState = false
		return value
	}
	return 0
}

// Helper methods for C64 core
func (c *CIA) IsIRQActive() bool {
	return c.InterruptState
}

type CIAEvent struct {
	Type EventType
	Data interface{}
}

type EventType int

const (
	EventIRQ EventType = iota
	EventNMI
	EventSerial
)