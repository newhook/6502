package cia

// Register offsets from CIA base address
const (
	PRA       = 0x00 // Peripheral Data Register A
	PRB       = 0x01 // Peripheral Data Register B
	DDRA      = 0x02 // Data Direction Register A
	DDRB      = 0x03 // Data Direction Register B
	TA_LO     = 0x04 // Timer A Low Byte
	TA_HI     = 0x05 // Timer A High Byte
	TB_LO     = 0x06 // Timer B Low Byte
	TB_HI     = 0x07 // Timer B High Byte
	TOD_10THS = 0x08 // Time of Day Tenths
	TOD_SEC   = 0x09 // Time of Day Seconds
	TOD_MIN   = 0x0A // Time of Day Minutes
	TOD_HR    = 0x0B // Time of Day Hours
	SDR       = 0x0C // Serial Data Register
	ICR       = 0x0D // Interrupt Control Register
	CRA       = 0x0E // Control Register A
	CRB       = 0x0F // Control Register B
)

// Control Register A bits
const (
	CRA_START   uint8 = 0x01 // Start timer A (1 = Start, 0 = Stop)
	CRA_PBON    uint8 = 0x02 // Timer A output on PB6 (1 = Output, 0 = No output)
	CRA_OUTMODE uint8 = 0x04 // Timer A output mode (1 = Toggle, 0 = Pulse)
	CRA_RUNMODE uint8 = 0x08 // Timer A run mode (1 = One-shot, 0 = Continuous)
	CRA_FORCE   uint8 = 0x10 // Force Timer A load (1 = Force load)
	CRA_INMODE  uint8 = 0x20 // Timer A input mode (1 = CNT, 0 = Clock)
	CRA_SPMODE  uint8 = 0x40 // Serial Port mode (1 = Output, 0 = Input)
	CRA_TODIN   uint8 = 0x80 // Time of Day frequency (1 = 50Hz, 0 = 60Hz)
)

//Bit 0: Start/Stop Timer A (1 = Start, 0 = Stop)
//Bit 1: Output mode to PB6 (1 = Toggle, 0 = Pulse)
//Bit 2: Timer mode (1 = Continuous, 0 = One-shot)
//Bit 3: Force Load (1 = Force load strobe, reloads from latch)
//Bit 4: Input mode (1 = Count CNT transitions, 0 = Count system clock)
//Bit 5: Serial port mode (1 = Output, 0 = Input)
//Bit 6: Direction of PB6 (1 = Output, 0 = Input)
//Bit 7: Direction of PA (TOD) (1 = Output, 0 = Input)

// Control Register B bits - similar to CRA but for Timer B
const (
	CRB_START   uint8 = 0x01 // Start timer B (1 = Start, 0 = Stop)
	CRB_PBON    uint8 = 0x02 // Timer B output on PB7 (1 = Output, 0 = No output)
	CRB_OUTMODE uint8 = 0x04 // Timer B output mode (1 = Toggle, 0 = Pulse)
	CRB_RUNMODE uint8 = 0x08 // Timer B run mode (1 = One-shot, 0 = Continuous)
	CRB_FORCE   uint8 = 0x10 // Force Timer B load (1 = Force load)
	CRB_INMODE  uint8 = 0x60 // Timer B input mode (00 = Clock, 01 = CNT, 10 = Timer A underflow, 11 = Timer A underflow with CNT)
	CRB_ALARM   uint8 = 0x80 // Time of Day alarm (1 = Alarm, 0 = Clock)
)

// Interrupt Control Register bits
const (
	ICR_TA   uint8 = 0x01 // Timer A interrupt (1 = Enable/Set)
	ICR_TB   uint8 = 0x02 // Timer B interrupt (1 = Enable/Set)
	ICR_TOD  uint8 = 0x04 // Time of Day alarm interrupt (1 = Enable/Set)
	ICR_SDR  uint8 = 0x08 // Serial Port interrupt (1 = Enable/Set)
	ICR_FLAG uint8 = 0x10 // FLAG line interrupt (1 = Enable/Set)
	ICR_SET  uint8 = 0x80 // Set/Clear flag (1 = Set, 0 = Clear)
)

type Registers struct {
	// Port registers
	portA uint8
	portB uint8
	ddrA  uint8
	ddrB  uint8

	// Timer registers and state
	timerALatch uint16
	timerBLatch uint16
	timerA      uint16
	timerB      uint16

	// TOD registers
	todTenths uint8
	todSec    uint8
	todMin    uint8
	todHr     uint8

	// Other registers
	sdr uint8

	// The Interrupt Control Register consists
	// of a write-only MASK register and a read-only
	// DATA register. Any interrupt will set the
	// corresponding bit in the DATA register. Any
	// interrupt which is enabled by the MASK register will
	// set the IR bit (MSB) of the DATA register and bring
	// the /IRQ pin low
	icrMask uint8 // interrupt control mask (ICR)
	icrData uint8 // interrupt control data (ICR)
	cra     uint8
	crb     uint8
}

// CIA represents a complete 6526 CIA chip
type CIA struct {
	registers Registers
	cycles    uint64

	todFrequency uint8
	todMode      uint8
	todCycles    uint16

	timerAOutput    bool
	timerBOutput    bool
	timerAUnderflow bool
	timerBUnderflow bool

	irq      bool
	isNMI    bool
	todAlarm [4]uint8

	// CNT pin state tracking
	cntPrevious bool // Previous CNT pin state
	cntCurrent  bool // Current CNT pin state
	cntPos      bool // Positive edge detected
	cntHigh     bool // Current CNT level (used by Timer B)
}

const (
	TOD_CLOCK = 0
	TOD_ALARM = 1
)

func NewCIA() *CIA {
	return &CIA{
		registers: Registers{
			timerALatch: 0xFFFF,
			timerA:      0xFFFF,
			timerBLatch: 0xFFFF,
			timerB:      0xFFFF,
		},
	}
}

// Call this whenever the CNT pin state changes
func (c *CIA) setCNT(level bool) {
	// For the CIA1 (DC00), the CNT pin was connected to the cassette motor
	// For the CIA2 (DD00), the CNT pin was connected to the serial bus clock line (CLK)
	c.cntPrevious = c.cntCurrent
	c.cntCurrent = level
	c.cntHigh = level

	// Detect positive edge (transition from low to high)
	c.cntPos = !c.cntPrevious && c.cntCurrent
}

// Update advances the CIA state by the specified number of cycles
func (c *CIA) Update(cycles uint8) *CIAEvent {
	event := &CIAEvent{}

	// Handle timers for each cycle
	for i := uint8(0); i < cycles; i++ {
		// Update Timer A if it's counting system clock
		if c.registers.cra&CRA_START != 0 && c.registers.cra&CRA_INMODE == 0 {
			c.updateTimerA()
		}

		// Update Timer B if it's counting system clock
		if c.registers.crb&CRB_START != 0 {
			// Check Timer B input mode
			switch c.registers.crb & CRB_INMODE {
			case 0x00: // System clock
				c.updateTimerB()
			case 0x40: // Count Timer A underflows
				if c.timerAUnderflow {
					c.updateTimerB()
				}
			}
		}

		// Clear underflow flags
		c.timerAUnderflow = false
		c.timerBUnderflow = false

		// Update TOD if needed
		c.todCycles++
		if c.todCycles >= c.todPeriod() {
			c.updateTOD()
			c.todCycles = 0
		}
	}

	// Check for interrupts
	if c.registers.icrData != 0 {
		// If any enabled interrupt occurred
		if (c.registers.icrData & c.registers.icrMask & 0x1F) != 0 {
			// Set interrupt output if not already set
			if !c.irq {
				c.irq = true
				// Signal interrupt based on CIA type
				if c.isNMI {
					event.NMI = true
				} else {
					event.IRQ = true
				}
			}
		}
	}

	return event
}

func (c *CIA) updateTimerA() {
	// If timer is not started, return immediately
	if c.registers.cra&CRA_START == 0 {
		return
	}

	// Handle input mode
	shouldDecrement := false
	if c.registers.cra&CRA_INMODE != 0 {
		// CNT mode
		shouldDecrement = c.cntPos
	} else {
		// CPU cycle mode
		shouldDecrement = true
	}

	if !shouldDecrement {
		return
	}

	// Decrement timer
	c.registers.timerA--

	// Check for timer underflow
	if c.registers.timerA == 0 {
		// Set interrupt flag
		if c.registers.icrMask&ICR_TA != 0 {
			c.registers.icrData |= ICR_TA
		}

		// Handle PB6 output if enabled
		if c.registers.cra&CRA_PBON != 0 {
			if c.registers.cra&CRA_OUTMODE != 0 {
				// Toggle mode
				c.registers.portB ^= 0x40 // Toggle bit 6
			} else {
				// Pulse mode - set high for one cycle
				c.registers.portB |= 0x40
			}
		}

		// Check run mode
		if c.registers.cra&CRA_RUNMODE != 0 {
			// One-shot mode: stop timer
			c.registers.cra &= ^CRA_START
		}

		// Reload timer from latch
		c.registers.timerA = c.registers.timerALatch
	} else if c.registers.cra&CRA_PBON != 0 &&
		c.registers.cra&CRA_OUTMODE == 0 {
		// In pulse mode, clear PB6 after one cycle
		c.registers.portB &= ^uint8(0x40)
	}

	// Handle forced load
	if c.registers.cra&CRA_FORCE != 0 {
		c.registers.timerA = c.registers.timerALatch
		c.registers.cra &= ^CRA_FORCE // Clear force load bit
	}
}

func (c *CIA) updateTimerB() {
	// If timer is not started, return immediately
	if c.registers.crb&CRB_START == 0 {
		return
	}

	// Handle different input modes
	shouldDecrement := false
	inmode := (c.registers.crb & CRB_INMODE) >> 5 // Extract input mode bits

	switch inmode {
	case 0: // Count CPU cycles
		shouldDecrement = true
	case 1: // Count positive CNT transitions
		shouldDecrement = c.cntPos // This would be set by CNT pin handler
	case 2: // Count Timer A underflow
		shouldDecrement = c.timerAUnderflow // This would be set in updateTimerA
	case 3: // Count Timer A underflow while CNT is high
		shouldDecrement = c.timerAUnderflow && c.cntHigh
	}

	// Only decrement if conditions are met
	if !shouldDecrement {
		return
	}

	// Decrement timer
	c.registers.timerB--

	// Check for timer underflow
	if c.registers.timerB == 0 {
		// Set interrupt flag
		if c.registers.icrMask&ICR_TB != 0 {
			c.registers.icrData |= ICR_TB
		}

		// Handle PB7 output if enabled
		if c.registers.crb&CRB_PBON != 0 {
			if c.registers.crb&CRB_OUTMODE != 0 {
				// Toggle mode
				c.registers.portB ^= 0x80 // Toggle bit 7
			} else {
				// Pulse mode - set high for one cycle
				c.registers.portB |= 0x80
			}
		}

		// Check run mode
		if c.registers.crb&CRB_RUNMODE != 0 {
			// One-shot mode: stop timer
			c.registers.crb &= ^CRB_START
		}

		// Reload timer from latch
		c.registers.timerB = c.registers.timerBLatch
	} else if c.registers.crb&CRB_PBON != 0 &&
		c.registers.crb&CRB_OUTMODE == 0 {
		// In pulse mode, clear PB7 after one cycle
		c.registers.portB &= ^uint8(0x80)
	}

	// Handle forced load
	if c.registers.crb&CRB_FORCE != 0 {
		c.registers.timerB = c.registers.timerBLatch
		c.registers.crb &= ^CRB_FORCE // Clear force load bit
	}
}

func (c *CIA) todPeriod() uint16 {
	if c.registers.cra&CRA_TODIN != 0 {
		return 20000 // 50Hz = 20ms
	}
	return 16667 // 60Hz = 16.67ms
}

func (c *CIA) updateTOD() {
	// Add 1 to tenths in BCD
	c.registers.todTenths = (c.registers.todTenths + 1) & 0x0F
	if c.registers.todTenths > 0x09 {
		c.registers.todTenths = 0x00

		// Add 1 to seconds in BCD
		if (c.registers.todSec & 0x0F) == 0x09 {
			c.registers.todSec = c.registers.todSec + 0x10 - 0x09
		} else {
			c.registers.todSec = c.registers.todSec + 0x01
		}

		if c.registers.todSec > 0x59 {
			c.registers.todSec = 0x00

			// Add 1 to minutes in BCD
			if (c.registers.todMin & 0x0F) == 0x09 {
				c.registers.todMin = c.registers.todMin + 0x10 - 0x09
			} else {
				c.registers.todMin = c.registers.todMin + 0x01
			}

			if c.registers.todMin > 0x59 {
				c.registers.todMin = 0x00

				// Hours are special (1-12 with PM bit)
				hours := c.registers.todHr & 0x1F
				pmBit := c.registers.todHr & 0x80

				if hours == 0x11 {
					// Going from 11 to 12
					hours = 0x12
					pmBit ^= 0x80 // Toggle PM bit
				} else if hours == 0x12 {
					// Going from 12 to 1
					hours = 0x01
				} else if (hours & 0x0F) == 0x09 {
					// Going from 9 to 10
					hours = 0x10
				} else {
					// Normal increment
					hours = hours + 0x01
				}

				c.registers.todHr = hours | pmBit
			}
		}
	}

	// Check for alarm match
	if c.registers.todTenths == c.todAlarm[0] &&
		c.registers.todSec == c.todAlarm[1] &&
		c.registers.todMin == c.todAlarm[2] &&
		c.registers.todHr == c.todAlarm[3] {
		if c.registers.icrMask&ICR_TOD != 0 {
			c.registers.icrData |= ICR_TOD
		}
	}
}

// Register access methods
func (c *CIA) WriteRegister(reg uint8, val uint8) {
	switch reg {
	case PRA:
		c.registers.portA = val
	case PRB:
		c.registers.portB = val
	case DDRA:
		c.registers.ddrA = val
	case DDRB:
		c.registers.ddrB = val
	case TA_LO:
		c.registers.timerALatch = (c.registers.timerALatch & 0xFF00) | uint16(val)
	case TA_HI:
		c.registers.timerALatch = (c.registers.timerALatch & 0x00FF) | (uint16(val) << 8)
		c.registers.timerA = c.registers.timerALatch
	case TB_LO:
		c.registers.timerBLatch = (c.registers.timerBLatch & 0xFF00) | uint16(val)
	case TB_HI:
		c.registers.timerBLatch = (c.registers.timerBLatch & 0x00FF) | (uint16(val) << 8)
		c.registers.timerB = c.registers.timerBLatch
	case TOD_10THS:
		if c.registers.crb&CRB_ALARM != 0 {
			c.todAlarm[0] = val & 0x0F // Only lower 4 bits valid
		} else {
			c.registers.todTenths = val & 0x0F
		}
	case TOD_SEC:
		if c.registers.crb&CRB_ALARM != 0 {
			c.todAlarm[1] = val & 0x7F // Only 7 bits valid
		} else {
			c.registers.todSec = val & 0x7F
		}
	case TOD_MIN:
		if c.registers.crb&CRB_ALARM != 0 {
			c.todAlarm[2] = val & 0x7F // Only 7 bits valid
		} else {
			c.registers.todMin = val & 0x7F
		}
	case TOD_HR:
		// Convert 0 to 12
		hours := val & 0x1F
		if hours == 0 {
			hours = 0x12
		}
		if c.registers.crb&CRB_ALARM != 0 {
			c.todAlarm[3] = hours | (val & 0x80) // Keep AM/PM bit
		} else {
			c.registers.todHr = hours | (val & 0x80)
		}
	case SDR:
		c.registers.sdr = val
	case ICR:
		c.writeICR(val)
	case CRA:
		c.writeCRA(val)
	case CRB:
		c.writeCRB(val)
	}
}

func (c *CIA) writeICR(val uint8) {
	if val&ICR_SET != 0 {
		// Set interrupt mask bits
		c.registers.icrMask |= val & 0x1F
	} else {
		// Clear interrupt mask bits
		c.registers.icrMask &= ^(val & 0x1F)
	}
}

func (c *CIA) writeCRA(val uint8) {
	oldStart := c.registers.cra & CRA_START
	c.registers.cra = val

	// Handle timer force load
	if val&CRA_FORCE != 0 {
		c.registers.timerA = c.registers.timerALatch
		c.registers.cra &= ^CRA_FORCE // Clear force load bit
	}

	// Handle timer start/stop
	if oldStart == 0 && (val&CRA_START != 0) {
		// Timer is being started - load initial value if it's 0
		if c.registers.timerA == 0 {
			c.registers.timerA = c.registers.timerALatch
		}
	}

	// Update TOD frequency if changed
	if val&CRA_TODIN != 0 {
		// Set TOD to 50Hz
		c.todFrequency = 50
	} else {
		// Set TOD to 60Hz
		c.todFrequency = 60
	}
}

func (c *CIA) writeCRB(val uint8) {
	oldStart := c.registers.crb & CRB_START
	c.registers.crb = val

	// Handle timer force load
	if val&CRB_FORCE != 0 {
		c.registers.timerB = c.registers.timerBLatch
		c.registers.crb &= ^CRB_FORCE // Clear force load bit
	}

	// Handle timer start/stop
	if oldStart == 0 && (val&CRB_START != 0) {
		// Timer is being started - load initial value if it's 0
		if c.registers.timerB == 0 {
			c.registers.timerB = c.registers.timerBLatch
		}
	}

	// Handle TOD alarm/clock mode
	if val&CRB_ALARM != 0 {
		// Writing to TOD registers sets alarm time
		c.todMode = TOD_ALARM
	} else {
		// Writing to TOD registers sets clock time
		c.todMode = TOD_CLOCK
	}
}

func (c *CIA) ReadRegister(reg uint8) uint8 {
	switch reg {
	case PRA:
		return c.readPortA()
	case PRB:
		return c.readPortB()
	case DDRA:
		return c.registers.ddrA
	case DDRB:
		return c.registers.ddrB
	case TA_LO:
		return uint8(c.registers.timerA & 0xFF)
	case TA_HI:
		return uint8(c.registers.timerA >> 8)
	case TB_LO:
		return uint8(c.registers.timerB & 0xFF)
	case TB_HI:
		return uint8(c.registers.timerB >> 8)
	case TOD_10THS:
		return c.registers.todTenths
	case TOD_SEC:
		return c.registers.todSec
	case TOD_MIN:
		return c.registers.todMin
	case TOD_HR:
		return c.registers.todHr
	case SDR:
		return c.registers.sdr
	case ICR:
		return c.readICR()
	case CRA:
		return c.registers.cra
	case CRB:
		return c.registers.crb
	}
	return 0
}

func (c *CIA) readPortA() uint8 {
	// For each bit:
	// - If DDR bit is 1 (output), use value from port register
	// - If DDR bit is 0 (input), use value from external device

	// First get the current state of external input lines
	inputValues := c.getPortAInput() // This would be different for CIA1 vs CIA2

	// For output bits, use port register value, for input bits use external value
	return (c.registers.portA & c.registers.ddrA) | (inputValues & ^c.registers.ddrA)
}

func (c *CIA) getPortAInput() uint8 {
	// cia1 - keyboard
	// cia2 - rs232, bank selection.
	// VIC bank bits (0-1) are special - they're always readable
	// regardless of DDRA, and they're inverted
	vicBankBits := c.registers.portA & 0x03
	return ^vicBankBits & 0x03
}

func (c *CIA) readPortB() uint8 {
	inputValues := c.getPortBInput() // Get external values (joystick, etc)

	// Handle timer outputs on PB6 (Timer A) and PB7 (Timer B)
	var timerOutputs uint8 = 0

	// If Timer A output enabled, handle PB6
	if c.registers.cra&CRA_PBON != 0 {
		if c.registers.cra&CRA_OUTMODE != 0 {
			// Toggle mode - use current toggle state
			if c.timerAOutput {
				timerOutputs |= 0x40
			}
		} else {
			// Pulse mode - only set during underflow
			if c.timerAUnderflow {
				timerOutputs |= 0x40
			}
		}
	}

	// If Timer B output enabled, handle PB7
	if c.registers.crb&CRB_PBON != 0 {
		if c.registers.crb&CRB_OUTMODE != 0 {
			// Toggle mode - use current toggle state
			if c.timerBOutput {
				timerOutputs |= 0x80
			}
		} else {
			// Pulse mode - only set during underflow
			if c.timerBUnderflow {
				timerOutputs |= 0x80
			}
		}
	}

	// Combine:
	// - Port register values for output bits (masked by DDRB)
	// - Input values for input bits (masked by inverted DDRB)
	// - Timer outputs (overriding bits 6-7 if enabled)
	return (c.registers.portB & c.registers.ddrB) | (inputValues & ^c.registers.ddrB) | timerOutputs
}

// CIA1 Port B is used for:
// - Bits 0-7: Keyboard matrix row selection
// - Bits 6-7: Paddles (when enabled)
// - Bits 0-4: Joystick 1 (when reading)
//func (c *CIA1) getPortBInput() uint8 {
//	var result uint8 = 0xFF // Default to all lines high
//
//	// Joystick 1 input (active low):
//	// Bit 0: Up
//	// Bit 1: Down
//	// Bit 2: Left
//	// Bit 3: Right
//	// Bit 4: Fire
//	result &= c.joystick1 | 0xE0 // Preserve bits 5-7
//
//	// Paddle input (if enabled) on bits 6-7
//	if c.paddlesEnabled {
//		result = (result & 0x3F) | (c.paddleValues & 0xC0)
//	}
//
//	return result
//}

// CIA2 Port B is used for:
// - Bits 0-7: Serial bus and RS-232
// - Bits 0-4: Joystick 2
// - Bits 6-7: Paddles (when enabled)
func (c *CIA) getPortBInput() uint8 {
	//var result uint8 = 0xFF // Default to all lines high
	//
	//// Joystick 2 input (active low):
	//// Same mapping as joystick 1
	//result &= c.joystick2 | 0xE0 // Preserve bits 5-7
	//
	//// RS-232 input lines (if enabled)
	//if c.rs232Enabled {
	//	result &= c.rs232Lines
	//}
	//
	//// Paddle input (if enabled) on bits 6-7
	//if c.paddlesEnabled {
	//	result = (result & 0x3F) | (c.paddleValues & 0xC0)
	//}
	//
	//return result
	return 0xff
}

func (c *CIA) readICR() uint8 {
	// Reading ICR returns interrupt flags and clears them
	value := c.registers.icrData

	// Bit 7 indicates if any enabled interrupt occurred
	if (c.registers.icrData & c.registers.icrMask & 0x1F) != 0 {
		value |= 0x80
	}

	// Clear all interrupt flags after reading
	c.registers.icrData = 0

	// Clear interrupt output if no more pending interrupts
	if (value & 0x80) == 0 {
		c.irq = false
	}

	return value
}

// Helper methods for C64 core
func (c *CIA) IsIRQActive() bool {
	//return c.InterruptState
	return false
}

// CIAEvent represents events that can occur during CIA update
type CIAEvent struct {
	IRQ bool // Interrupt request occurred
	NMI bool // Non-maskable interrupt occurred
}
