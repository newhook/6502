package cpu

import "fmt"

// The naming convention uses the instruction name followed by the addressing mode:
//
// IMM: Immediate
// ZP: Zero Page
// ZPX: Zero Page,X
// ZPY: Zero Page,Y
// ABS: Absolute
// ABX: Absolute,X
// ABY: Absolute,Y
// INX: (Indirect,X)
// INY: (Indirect),Y
// ACC: Accumulator (for shifts)

const (
	// Load/Store Operations
	LDA_IMM = 0xA9
	LDA_ZP  = 0xA5
	LDA_ZPX = 0xB5
	LDA_ABS = 0xAD
	LDA_ABX = 0xBD
	LDA_ABY = 0xB9
	LDA_INX = 0xA1
	LDA_INY = 0xB1

	LDX_IMM = 0xA2
	LDX_ZP  = 0xA6
	LDX_ZPY = 0xB6
	LDX_ABS = 0xAE
	LDX_ABY = 0xBE

	LDY_IMM = 0xA0
	LDY_ZP  = 0xA4
	LDY_ZPX = 0xB4
	LDY_ABS = 0xAC
	LDY_ABX = 0xBC

	STA_ZP  = 0x85
	STA_ZPX = 0x95
	STA_ABS = 0x8D
	STA_ABX = 0x9D
	STA_ABY = 0x99
	STA_INX = 0x81
	STA_INY = 0x91

	STX_ZP  = 0x86
	STX_ZPY = 0x96
	STX_ABS = 0x8E

	STY_ZP  = 0x84
	STY_ZPX = 0x94
	STY_ABS = 0x8C

	// Register Transfers
	TAX = 0xAA
	TAY = 0xA8
	TXA = 0x8A
	TYA = 0x98
	TSX = 0xBA
	TXS = 0x9A

	// Stack Operations
	PHA = 0x48
	PHP = 0x08
	PLA = 0x68
	PLP = 0x28

	// Logical Operations
	AND_IMM = 0x29
	AND_ZP  = 0x25
	AND_ZPX = 0x35
	AND_ABS = 0x2D
	AND_ABX = 0x3D
	AND_ABY = 0x39
	AND_INX = 0x21
	AND_INY = 0x31

	EOR_IMM = 0x49
	EOR_ZP  = 0x45
	EOR_ZPX = 0x55
	EOR_ABS = 0x4D
	EOR_ABX = 0x5D
	EOR_ABY = 0x59
	EOR_INX = 0x41
	EOR_INY = 0x51

	ORA_IMM = 0x09
	ORA_ZP  = 0x05
	ORA_ZPX = 0x15
	ORA_ABS = 0x0D
	ORA_ABX = 0x1D
	ORA_ABY = 0x19
	ORA_INX = 0x01
	ORA_INY = 0x11

	BIT_ZP  = 0x24
	BIT_ABS = 0x2C

	// Arithmetic Operations
	ADC_IMM = 0x69
	ADC_ZP  = 0x65
	ADC_ZPX = 0x75
	ADC_ABS = 0x6D
	ADC_ABX = 0x7D
	ADC_ABY = 0x79
	ADC_INX = 0x61
	ADC_INY = 0x71

	SBC_IMM = 0xE9
	SBC_ZP  = 0xE5
	SBC_ZPX = 0xF5
	SBC_ABS = 0xED
	SBC_ABX = 0xFD
	SBC_ABY = 0xF9
	SBC_INX = 0xE1
	SBC_INY = 0xF1

	CMP_IMM = 0xC9
	CMP_ZP  = 0xC5
	CMP_ZPX = 0xD5
	CMP_ABS = 0xCD
	CMP_ABX = 0xDD
	CMP_ABY = 0xD9
	CMP_INX = 0xC1
	CMP_INY = 0xD1

	CPX_IMM = 0xE0
	CPX_ZP  = 0xE4
	CPX_ABS = 0xEC

	CPY_IMM = 0xC0
	CPY_ZP  = 0xC4
	CPY_ABS = 0xCC

	// Increments & Decrements
	INC_ZP  = 0xE6
	INC_ZPX = 0xF6
	INC_ABS = 0xEE
	INC_ABX = 0xFE

	DEC_ZP  = 0xC6
	DEC_ZPX = 0xD6
	DEC_ABS = 0xCE
	DEC_ABX = 0xDE

	INX = 0xE8
	INY = 0xC8
	DEX = 0xCA
	DEY = 0x88

	// Shifts
	ASL_ACC = 0x0A
	ASL_ZP  = 0x06
	ASL_ZPX = 0x16
	ASL_ABS = 0x0E
	ASL_ABX = 0x1E

	LSR_ACC = 0x4A
	LSR_ZP  = 0x46
	LSR_ZPX = 0x56
	LSR_ABS = 0x4E
	LSR_ABX = 0x5E

	ROL_ACC = 0x2A
	ROL_ZP  = 0x26
	ROL_ZPX = 0x36
	ROL_ABS = 0x2E
	ROL_ABX = 0x3E

	ROR_ACC = 0x6A
	ROR_ZP  = 0x66
	ROR_ZPX = 0x76
	ROR_ABS = 0x6E
	ROR_ABX = 0x7E

	// Jumps & Calls
	JMP_ABS = 0x4C
	JMP_IND = 0x6C
	JSR_ABS = 0x20
	RTS     = 0x60

	// Branches
	BCC = 0x90
	BCS = 0xB0
	BEQ = 0xF0
	BMI = 0x30
	BNE = 0xD0
	BPL = 0x10
	BVC = 0x50
	BVS = 0x70

	// Status Flag Changes
	CLC = 0x18
	CLD = 0xD8
	CLI = 0x58
	CLV = 0xB8
	SEC = 0x38
	SED = 0xF8
	SEI = 0x78

	// System Functions
	BRK = 0x00
	NOP = 0xEA
	RTI = 0x40
)

// CPU represents the 6502 processor
type CPU struct {
	// Registers
	A  uint8  // Accumulator
	X  uint8  // X index register
	Y  uint8  // Y index register
	PC uint16 // Program Counter
	SP uint8  // Stack Pointer
	P  uint8  // Status Register (Flags)

	// Memory
	Memory [65536]uint8
}

// Status flag bits
const (
	FlagC uint8 = 0x01 // Carry
	FlagZ uint8 = 0x02 // Zero
	FlagI uint8 = 0x04 // Interrupt Disable
	FlagD uint8 = 0x08 // Decimal Mode
	FlagB uint8 = 0x10 // Break Command
	FlagV uint8 = 0x40 // Overflow
	FlagN uint8 = 0x80 // Negative
)

// NewCPU creates a new 6502 CPU instance
func NewCPU() *CPU {
	return &CPU{
		SP: 0xFF, // Stack pointer starts at top of stack
		P:  0x24, // IRQ disabled by default
	}
}

// Reset resets the CPU to its initial state
func (c *CPU) Reset() {
	// Read reset vector at 0xFFFC-0xFFFD
	lowByte := uint16(c.Memory[0xFFFC])
	highByte := uint16(c.Memory[0xFFFD])
	c.PC = (highByte << 8) | lowByte

	c.SP = 0xFF
	c.P = 0x24
	c.A = 0
	c.X = 0
	c.Y = 0
}

// Step executes one instruction and returns number of cycles used
func (c *CPU) Step() uint8 {
	// Fetch
	opcode := c.Memory[c.PC]
	c.PC++

	// Decode and Execute
	return c.execute(opcode)
}

// execute processes a single opcode
func (c *CPU) execute(opcode uint8) uint8 {
	switch opcode {
	case LDA_IMM:
		c.A = c.readImmediate()
		c.updateZN(c.A)
		return 2

	case LDA_ZP:
		c.A = c.readZeroPage()
		c.updateZN(c.A)
		return 3

	case LDA_ZPX:
		c.A = c.readZeroPageX()
		c.updateZN(c.A)
		return 4

	case LDA_ABS:
		c.A = c.readAbsolute()
		c.updateZN(c.A)
		return 4

	case LDA_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.A = value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case LDA_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.A = value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case LDA_INX:
		c.A = c.readIndirectX()
		c.updateZN(c.A)
		return 6

	case LDA_INY:
		value, pageCrossed := c.readIndirectY()
		c.A = value
		c.updateZN(c.A)
		if pageCrossed {
			return 6
		}
		return 5

	case LDX_IMM:
		c.X = c.readImmediate()
		c.updateZN(c.X)
		return 2

	case LDX_ZP:
		c.X = c.readZeroPage()
		c.updateZN(c.X)
		return 3

	case LDX_ZPY: // Note: LDX uses Y register for indexing!
		zeroPageAddr := c.Memory[c.PC]
		c.PC++
		c.X = c.Memory[(zeroPageAddr+c.Y)&0xFF]
		c.updateZN(c.X)
		return 4

	case LDX_ABS:
		c.X = c.readAbsolute()
		c.updateZN(c.X)
		return 4

	case LDX_ABY: // Note: LDX uses Y register for indexing!
		value, pageCrossed := func() (uint8, bool) {
			lowByte := uint16(c.Memory[c.PC])
			c.PC++
			highByte := uint16(c.Memory[c.PC])
			c.PC++
			addr := (highByte << 8) | lowByte
			finalAddr := addr + uint16(c.Y)

			// Check if page boundary was crossed
			pageCrossed := (addr & 0xFF00) != (finalAddr & 0xFF00)

			return c.Memory[finalAddr], pageCrossed
		}()

		c.X = value
		c.updateZN(c.X)
		if pageCrossed {
			return 5
		}
		return 4

	case LDY_IMM:
		c.Y = c.readImmediate()
		c.updateZN(c.Y)
		return 2

	case LDY_ZP:
		c.Y = c.readZeroPage()
		c.updateZN(c.Y)
		return 3

	case LDY_ZPX: // Note: LDY uses X register for indexing!
		zeroPageAddr := c.Memory[c.PC]
		c.PC++
		c.Y = c.Memory[(zeroPageAddr+c.X)&0xFF]
		c.updateZN(c.Y)
		return 4

	case LDY_ABS:
		c.Y = c.readAbsolute()
		c.updateZN(c.Y)
		return 4

	case LDY_ABX: // Note: LDY uses X register for indexing!
		value, pageCrossed := func() (uint8, bool) {
			lowByte := uint16(c.Memory[c.PC])
			c.PC++
			highByte := uint16(c.Memory[c.PC])
			c.PC++
			addr := (highByte << 8) | lowByte
			finalAddr := addr + uint16(c.X)

			// Check if page boundary was crossed
			pageCrossed := (addr & 0xFF00) != (finalAddr & 0xFF00)

			return c.Memory[finalAddr], pageCrossed
		}()

		c.Y = value
		c.updateZN(c.Y)
		if pageCrossed {
			return 5
		}
		return 4

	case STA_ZP:
		addr := c.readImmediate() // Get zero page address
		c.Memory[addr] = c.A
		return 3

	case STA_ZPX:
		addr := (c.readImmediate() + c.X) & 0xFF
		c.Memory[addr] = c.A
		return 4

	case STA_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.A
		return 4

	case STA_ABX:
		addr := c.readAbsoluteAddress() + uint16(c.X)
		c.Memory[addr] = c.A
		return 5

	case STA_ABY:
		addr := c.readAbsoluteAddress() + uint16(c.Y)
		c.Memory[addr] = c.A
		return 5

	case STA_INX:
		zeroPageAddr := (c.readImmediate() + c.X) & 0xFF
		addr := c.readIndirectAddress(zeroPageAddr)
		c.Memory[addr] = c.A
		return 6

	case STA_INY:
		zeroPageAddr := c.readImmediate()
		addr := c.readIndirectAddress(zeroPageAddr) + uint16(c.Y)
		c.Memory[addr] = c.A
		return 6

	// STX - Store X Register
	case STX_ZP:
		addr := c.readImmediate()
		c.Memory[addr] = c.X
		return 3

	case STX_ZPY:
		addr := (c.readImmediate() + c.Y) & 0xFF
		c.Memory[addr] = c.X
		return 4

	case STX_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.X
		return 4

	// STY - Store Y Register
	case STY_ZP:
		addr := c.readImmediate()
		c.Memory[addr] = c.Y
		return 3

	case STY_ZPX:
		addr := (c.readImmediate() + c.X) & 0xFF
		c.Memory[addr] = c.Y
		return 4

	case STY_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.Y
		return 4

		// Transfer Accumulator to X
	case TAX:
		c.X = c.A
		c.updateZN(c.X)
		return 2

	// Transfer Accumulator to Y
	case TAY:
		c.Y = c.A
		c.updateZN(c.Y)
		return 2

	// Transfer X to Accumulator
	case TXA:
		c.A = c.X
		c.updateZN(c.A)
		return 2

	// Transfer Y to Accumulator
	case TYA:
		c.A = c.Y
		c.updateZN(c.A)
		return 2

	// Transfer Stack Pointer to X
	case TSX:
		c.X = c.SP
		c.updateZN(c.X)
		return 2

	// Transfer X to Stack Pointer
	case TXS:
		c.SP = c.X
		// Note: TXS does not affect status flags
		return 2

		// Push Accumulator to Stack
	case PHA:
		c.push(c.A)
		return 3

	// Push Processor Status to Stack
	case PHP:
		// The B flag is always set in the stored value
		c.push(c.P | FlagB)
		return 3

	// Pull Accumulator from Stack
	case PLA:
		c.A = c.pull()
		c.updateZN(c.A)
		return 4

	// Pull Processor Status from Stack
	case PLP:
		// Keep the B flag unchanged when pulling status
		currentB := c.P & FlagB
		c.P = (c.pull() & ^FlagB) | (currentB & FlagB)
		return 4

	// AND - Logical AND with Accumulator
	case AND_IMM:
		c.A &= c.readImmediate()
		c.updateZN(c.A)
		return 2

	case AND_ZP:
		c.A &= c.readZeroPage()
		c.updateZN(c.A)
		return 3

	case AND_ZPX:
		c.A &= c.readZeroPageX()
		c.updateZN(c.A)
		return 4

	case AND_ABS:
		c.A &= c.readAbsolute()
		c.updateZN(c.A)
		return 4

	case AND_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.A &= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case AND_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.A &= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case AND_INX:
		c.A &= c.readIndirectX()
		c.updateZN(c.A)
		return 6

	case AND_INY:
		value, pageCrossed := c.readIndirectY()
		c.A &= value
		c.updateZN(c.A)
		if pageCrossed {
			return 6
		}
		return 5

		// EOR - Exclusive OR with Accumulator
	case EOR_IMM:
		c.A ^= c.readImmediate()
		c.updateZN(c.A)
		return 2

	case EOR_ZP:
		c.A ^= c.readZeroPage()
		c.updateZN(c.A)
		return 3

	case EOR_ZPX:
		c.A ^= c.readZeroPageX()
		c.updateZN(c.A)
		return 4

	case EOR_ABS:
		c.A ^= c.readAbsolute()
		c.updateZN(c.A)
		return 4

	case EOR_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.A ^= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case EOR_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.A ^= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case EOR_INX:
		c.A ^= c.readIndirectX()
		c.updateZN(c.A)
		return 6

	case EOR_INY:
		value, pageCrossed := c.readIndirectY()
		c.A ^= value
		c.updateZN(c.A)
		if pageCrossed {
			return 6
		}
		return 5

		// ORA - Inclusive OR with Accumulator
	case ORA_IMM:
		c.A |= c.readImmediate()
		c.updateZN(c.A)
		return 2

	case ORA_ZP:
		c.A |= c.readZeroPage()
		c.updateZN(c.A)
		return 3

	case ORA_ZPX:
		c.A |= c.readZeroPageX()
		c.updateZN(c.A)
		return 4

	case ORA_ABS:
		c.A |= c.readAbsolute()
		c.updateZN(c.A)
		return 4

	case ORA_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.A |= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case ORA_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.A |= value
		c.updateZN(c.A)
		if pageCrossed {
			return 5
		}
		return 4

	case ORA_INX:
		c.A |= c.readIndirectX()
		c.updateZN(c.A)
		return 6

	case ORA_INY:
		value, pageCrossed := c.readIndirectY()
		c.A |= value
		c.updateZN(c.A)
		if pageCrossed {
			return 6
		}
		return 5

	case BIT_ZP:
		value := c.readZeroPage()
		result := c.A & value

		// Zero flag - Set if result of AND is zero
		if result == 0 {
			c.P |= FlagZ
		} else {
			c.P &^= FlagZ
		}

		// Negative flag - Set to bit 7 of memory value
		if value&0x80 != 0 {
			c.P |= FlagN
		} else {
			c.P &^= FlagN
		}

		// Overflow flag - Set to bit 6 of memory value
		if value&0x40 != 0 {
			c.P |= FlagV
		} else {
			c.P &^= FlagV
		}

		return 3

	case BIT_ABS:
		value := c.readAbsolute()
		result := c.A & value

		// Zero flag - Set if result of AND is zero
		if result == 0 {
			c.P |= FlagZ
		} else {
			c.P &^= FlagZ
		}

		// Negative flag - Set to bit 7 of memory value
		if value&0x80 != 0 {
			c.P |= FlagN
		} else {
			c.P &^= FlagN
		}

		// Overflow flag - Set to bit 6 of memory value
		if value&0x40 != 0 {
			c.P |= FlagV
		} else {
			c.P &^= FlagV
		}

		return 4
	case ADC_IMM:
		c.adc(c.readImmediate())
		return 2

	case ADC_ZP:
		c.adc(c.readZeroPage())
		return 3

	case ADC_ZPX:
		c.adc(c.readZeroPageX())
		return 4

	case ADC_ABS:
		c.adc(c.readAbsolute())
		return 4

	case ADC_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.adc(value)
		if pageCrossed {
			return 5
		}
		return 4

	case ADC_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.adc(value)
		if pageCrossed {
			return 5
		}
		return 4

	case ADC_INX:
		c.adc(c.readIndirectX())
		return 6

	case ADC_INY:
		value, pageCrossed := c.readIndirectY()
		c.adc(value)
		if pageCrossed {
			return 6
		}
		return 5

	case SBC_IMM:
		c.sbc(c.readImmediate())
		return 2

	case SBC_ZP:
		c.sbc(c.readZeroPage())
		return 3

	case SBC_ZPX:
		c.sbc(c.readZeroPageX())
		return 4

	case SBC_ABS:
		c.sbc(c.readAbsolute())
		return 4

	case SBC_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.sbc(value)
		if pageCrossed {
			return 5
		}
		return 4

	case SBC_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.sbc(value)
		if pageCrossed {
			return 5
		}
		return 4

	case SBC_INX:
		c.sbc(c.readIndirectX())
		return 6

	case SBC_INY:
		value, pageCrossed := c.readIndirectY()
		c.sbc(value)
		if pageCrossed {
			return 6
		}
		return 5

	case CMP_IMM:
		c.cmp(c.readImmediate())
		return 2
	case CMP_ZP:
		c.cmp(c.readZeroPage())
		return 3
	case CMP_ZPX:
		c.cmp(c.readZeroPageX())
		return 4
	case CMP_ABS:
		c.cmp(c.readAbsolute())
		return 4
	case CMP_ABX:
		value, pageCrossed := c.readAbsoluteX()
		c.cmp(value)
		if pageCrossed {
			return 5
		}
		return 4
	case CMP_ABY:
		value, pageCrossed := c.readAbsoluteY()
		c.cmp(value)
		if pageCrossed {
			return 5
		}
		return 4
	case CMP_INX:
		c.cmp(c.readIndirectX())
		return 6
	case CMP_INY:
		value, pageCrossed := c.readIndirectY()
		c.cmp(value)
		if pageCrossed {
			return 6
		}
		return 5

		// CPX cases
	case CPX_IMM:
		c.cpx(c.readImmediate())
		return 2
	case CPX_ZP:
		c.cpx(c.readZeroPage())
		return 3
	case CPX_ABS:
		c.cpx(c.readAbsolute())
		return 4

	// CPY cases
	case CPY_IMM:
		c.cpy(c.readImmediate())
		return 2
	case CPY_ZP:
		c.cpy(c.readZeroPage())
		return 3
	case CPY_ABS:
		c.cpy(c.readAbsolute())
		return 4

	case INC_ZP:
		addr := uint16(c.readImmediate())
		c.inc(addr)
		return 5
	case INC_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.inc(addr)
		return 6
	case INC_ABS:
		addr := c.readAbsoluteAddress()
		c.inc(addr)
		return 6
	case INC_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.inc(addr)
		return 7

	case DEC_ZP:
		addr := uint16(c.readImmediate())
		c.dec(addr)
		return 5
	case DEC_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.dec(addr)
		return 6
	case DEC_ABS:
		addr := c.readAbsoluteAddress()
		c.dec(addr)
		return 6
	case DEC_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.dec(addr)
		return 7

	case INX:
		c.X++
		c.updateZN(c.X)
		return 2
	case INY:
		c.Y++
		c.updateZN(c.Y)
		return 2
	case DEX:
		c.X--
		c.updateZN(c.X)
		return 2
	case DEY:
		c.Y--
		c.updateZN(c.Y)
		return 2

	case ASL_ACC:
		c.A = c.asl(c.A)
		return 2
	case ASL_ZP:
		addr := uint16(c.readImmediate())
		c.Memory[addr] = c.asl(c.Memory[addr])
		return 5
	case ASL_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.Memory[addr] = c.asl(c.Memory[addr])
		return 6
	case ASL_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.asl(c.Memory[addr])
		return 6
	case ASL_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.Memory[addr] = c.asl(c.Memory[addr])
		return 7

	case LSR_ACC:
		c.A = c.lsr(c.A)
		return 2
	case LSR_ZP:
		addr := uint16(c.readImmediate())
		c.Memory[addr] = c.lsr(c.Memory[addr])
		return 5
	case LSR_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.Memory[addr] = c.lsr(c.Memory[addr])
		return 6
	case LSR_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.lsr(c.Memory[addr])
		return 6
	case LSR_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.Memory[addr] = c.lsr(c.Memory[addr])
		return 7

		// ROL cases
	case ROL_ACC:
		c.A = c.rol(c.A)
		return 2
	case ROL_ZP:
		addr := uint16(c.readImmediate())
		c.Memory[addr] = c.rol(c.Memory[addr])
		return 5
	case ROL_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.Memory[addr] = c.rol(c.Memory[addr])
		return 6
	case ROL_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.rol(c.Memory[addr])
		return 6
	case ROL_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.Memory[addr] = c.rol(c.Memory[addr])
		return 7

	// ROR cases
	case ROR_ACC:
		c.A = c.ror(c.A)
		return 2
	case ROR_ZP:
		addr := uint16(c.readImmediate())
		c.Memory[addr] = c.ror(c.Memory[addr])
		return 5
	case ROR_ZPX:
		addr := uint16(c.readImmediate() + c.X)
		c.Memory[addr] = c.ror(c.Memory[addr])
		return 6
	case ROR_ABS:
		addr := c.readAbsoluteAddress()
		c.Memory[addr] = c.ror(c.Memory[addr])
		return 6
	case ROR_ABX:
		base := c.readAbsoluteAddress()
		addr := base + uint16(c.X)
		c.Memory[addr] = c.ror(c.Memory[addr])
		return 7

	case JMP_ABS:
		c.PC = c.readAbsoluteAddress()
		return 3

	case JMP_IND:
		addr := c.readAbsoluteAddress()
		// Handle 6502 indirect jump bug at page boundary
		if addr&0xFF == 0xFF {
			low := uint16(c.Memory[addr])
			high := uint16(c.Memory[addr&0xFF00])
			c.PC = (high << 8) | low
		} else {
			c.PC = uint16(c.Memory[addr]) | uint16(c.Memory[addr+1])<<8
		}
		return 5

	case JSR_ABS:
		addr := c.readAbsoluteAddress()
		// Push address of next instruction minus 1
		c.push16(c.PC - 1)
		c.PC = addr
		return 6

	case RTS:
		c.PC = c.pull16() + 1
		return 6

	case BCC:
		return c.branch(c.P&FlagC == 0)
	case BCS:
		return c.branch(c.P&FlagC != 0)
	case BEQ:
		return c.branch(c.P&FlagZ != 0)
	case BMI:
		return c.branch(c.P&FlagN != 0)
	case BNE:
		return c.branch(c.P&FlagZ == 0)
	case BPL:
		return c.branch(c.P&FlagN == 0)
	case BVC:
		return c.branch(c.P&FlagV == 0)
	case BVS:
		return c.branch(c.P&FlagV != 0)

	case CLC:
		c.P &= ^FlagC
		return 2
	case CLD:
		c.P &= ^FlagD
		return 2
	case CLI:
		c.P &= ^FlagI
		return 2
	case CLV:
		c.P &= ^FlagV
		return 2
	case SEC:
		c.P |= FlagC
		return 2
	case SED:
		c.P |= FlagD
		return 2
	case SEI:
		c.P |= FlagI
		return 2

	case BRK:
		pc := c.PC + 2      // Point to instruction after BRK and padding
		c.push16(pc)        // Push next instruction address
		c.push(c.P | FlagB) // Push status with B flag set
		c.P |= FlagI        // Set interrupt disable flag
		// Load IRQ vector
		c.PC = uint16(c.Memory[0xFFFE]) | uint16(c.Memory[0xFFFF])<<8
		return 7

	case NOP:
		return 2

	case RTI:
		c.P = c.pull() & ^FlagB // Pull status, clear B flag
		c.PC = c.pull16()       // Pull return address
		return 6

	default:
		panic(fmt.Sprintf("Unknown opcode: 0x%02X", opcode))
	}
	return 0
}

// branch performs a relative branch if condition is true
func (c *CPU) branch(condition bool) uint8 {
	offset := int8(c.readImmediate())
	if !condition {
		return 2 // Branch not taken
	}

	oldPC := c.PC
	c.PC = uint16(int32(c.PC) + int32(offset))

	// Extra cycle if branch crosses page boundary
	if (oldPC & 0xFF00) != (c.PC & 0xFF00) {
		return 4 // Page boundary crossed
	}
	return 3 // Branch taken, no page boundary cross
}

// rol performs rotate left through carry and returns result
func (c *CPU) rol(value uint8) uint8 {
	oldCarry := c.P & FlagC

	// Set carry flag from bit 7
	if value&0x80 != 0 {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	result := value << 1
	if oldCarry != 0 {
		result |= 0x01
	}

	c.updateZN(result)
	return result
}

// ror performs rotate right through carry and returns result
func (c *CPU) ror(value uint8) uint8 {
	oldCarry := c.P & FlagC

	// Set carry flag from bit 0
	if value&0x01 != 0 {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	result := value >> 1
	if oldCarry != 0 {
		result |= 0x80
	}

	c.updateZN(result)
	return result
}

// lsr performs logical shift right on value and returns result
func (c *CPU) lsr(value uint8) uint8 {
	// Set carry flag from bit 0
	if value&0x01 != 0 {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	result := value >> 1
	c.updateZN(result)
	return result
}

// asl performs arithmetic shift left on value and returns result
func (c *CPU) asl(value uint8) uint8 {
	// Set carry flag from bit 7
	if value&0x80 != 0 {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	result := value << 1
	c.updateZN(result)
	return result
}

// dec decrements the value at the specified memory address
func (c *CPU) dec(addr uint16) {
	value := c.Memory[addr]
	result := value - 1
	c.Memory[addr] = result
	c.updateZN(result)
}

// inc increments the value at the specified memory address
func (c *CPU) inc(addr uint16) {
	value := c.Memory[addr]
	result := value + 1
	c.Memory[addr] = result
	c.updateZN(result)
}

// cpx performs the comparison operation with X register and sets appropriate flags
func (c *CPU) cpx(value uint8) {
	result := c.X - value

	// Set carry flag if X >= value
	if c.X >= value {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	// Update zero and negative flags based on result
	c.updateZN(result)
}

// cpy performs the comparison operation with Y register and sets appropriate flags
func (c *CPU) cpy(value uint8) {
	result := c.Y - value

	// Set carry flag if Y >= value
	if c.Y >= value {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	// Update zero and negative flags based on result
	c.updateZN(result)
}

// cmp performs the comparison operation and sets appropriate flags
func (c *CPU) cmp(value uint8) {
	result := c.A - value

	// Set carry flag if A >= value
	if c.A >= value {
		c.P |= FlagC
	} else {
		c.P &= ^FlagC
	}

	// Update zero and negative flags based on result
	c.updateZN(result)
}

// Helper function for SBC operation
func (c *CPU) sbc(value uint8) {
	// SBC operation is equivalent to ADC of the two's complement
	c.adc(^value) // ^ operator performs NOT operation
}

// Helper function for SBC operation
//func (c *CPU) sbc(value uint8) {
//	carry := uint8(1)
//	if c.P&FlagC != 0 {
//		carry = 0
//	}
//
//	// Save original accumulator value for overflow check
//	orig := c.A
//	result := uint16(c.A) - uint16(value) - uint16(carry)
//
//	// Set carry flag (inverted borrow)
//	if result < 0x100 {
//		c.P |= FlagC
//	} else {
//		c.P &^= FlagC
//	}
//
//	// Set overflow flag
//	// Overflow occurs when the sign of the result is incorrect
//	// This happens when:
//	// 1. Subtracting a positive from a negative gives a positive
//	// 2. Subtracting a negative from a positive gives a negative
//	if ((orig^value)&0x80 != 0) && ((orig^uint8(result))&0x80 != 0) {
//		c.P |= FlagV
//	} else {
//		c.P &^= FlagV
//	}
//
//	// Store result and update N and Z flags
//	c.A = uint8(result)
//	c.updateZN(c.A)
//}

func (c *CPU) adc(value uint8) {
	// Convert to uint16 to handle carry bit
	sum := uint16(c.A) + uint16(value) + uint16(c.P&FlagC)

	// Handle decimal mode if D flag is set
	if c.P&FlagD != 0 {
		// Convert to BCD
		if ((c.A & 0xF) + (value & 0xF) + (c.P & FlagC)) > 9 {
			sum += 0x6
		}
		if sum > 0x99 {
			sum += 0x60
		}
	}

	// Set carry flag
	if sum > 0xFF {
		c.P |= FlagC
	} else {
		c.P &^= FlagC
	}

	// Set overflow flag
	// Overflow occurs if both operands have the same sign bit AND
	// the sign bit of the result differs from the sign bit of the operands
	if ((c.A^value)&0x80) == 0 && ((c.A^uint8(sum))&0x80) != 0 {
		c.P |= FlagV
	} else {
		c.P &^= FlagV
	}

	// Store result and update N,Z flags
	c.A = uint8(sum)
	c.updateZN(c.A)
}

func (c *CPU) readImmediate() uint8 {
	value := c.Memory[c.PC]
	c.PC++
	return value
}

func (c *CPU) readZeroPage() uint8 {
	addr := c.Memory[c.PC]
	c.PC++
	return c.Memory[addr]
}

func (c *CPU) readZeroPageX() uint8 {
	addr := c.Memory[c.PC]
	c.PC++
	return c.Memory[(addr+c.X)&0xFF] // & 0xFF ensures zero-page wrap-around
}

func (c *CPU) readAbsolute() uint8 {
	lowByte := uint16(c.Memory[c.PC])
	c.PC++
	highByte := uint16(c.Memory[c.PC])
	c.PC++
	addr := (highByte << 8) | lowByte
	return c.Memory[addr]
}

func (c *CPU) readAbsoluteX() (uint8, bool) {
	lowByte := uint16(c.Memory[c.PC])
	c.PC++
	highByte := uint16(c.Memory[c.PC])
	c.PC++
	addr := (highByte << 8) | lowByte
	finalAddr := addr + uint16(c.X)

	// Return true if page boundary crossed (extra cycle)
	pageCrossed := (addr & 0xFF00) != (finalAddr & 0xFF00)

	return c.Memory[finalAddr], pageCrossed
}

func (c *CPU) readAbsoluteY() (uint8, bool) {
	lowByte := uint16(c.Memory[c.PC])
	c.PC++
	highByte := uint16(c.Memory[c.PC])
	c.PC++
	addr := (highByte << 8) | lowByte
	finalAddr := addr + uint16(c.Y)

	pageCrossed := (addr & 0xFF00) != (finalAddr & 0xFF00)

	return c.Memory[finalAddr], pageCrossed
}

func (c *CPU) readIndirectX() uint8 {
	zeroPageAddr := c.Memory[c.PC]
	c.PC++

	// Add X register with wrap-around
	effectiveAddr := (zeroPageAddr + c.X) & 0xFF

	// Read effective address from zero page
	lowByte := uint16(c.Memory[effectiveAddr])
	highByte := uint16(c.Memory[(effectiveAddr+1)&0xFF])

	addr := (highByte << 8) | lowByte
	return c.Memory[addr]
}

func (c *CPU) readIndirectY() (uint8, bool) {
	zeroPageAddr := c.Memory[c.PC]
	c.PC++

	// Read address from zero page
	lowByte := uint16(c.Memory[zeroPageAddr])
	highByte := uint16(c.Memory[(zeroPageAddr+1)&0xFF])

	baseAddr := (highByte << 8) | lowByte
	finalAddr := baseAddr + uint16(c.Y)

	pageCrossed := (baseAddr & 0xFF00) != (finalAddr & 0xFF00)

	return c.Memory[finalAddr], pageCrossed
}

func (c *CPU) readAbsoluteAddress() uint16 {
	lowByte := uint16(c.Memory[c.PC])
	c.PC++
	highByte := uint16(c.Memory[c.PC])
	c.PC++
	return (highByte << 8) | lowByte
}

// Helper function to read indirect address
func (c *CPU) readIndirectAddress(zeroPageAddr uint8) uint16 {
	lowByte := uint16(c.Memory[zeroPageAddr])
	highByte := uint16(c.Memory[(zeroPageAddr+1)&0xFF])
	return (highByte << 8) | lowByte
}

// Add helper functions for stack operations
func (c *CPU) push(value uint8) {
	c.Memory[0x0100|uint16(c.SP)] = value
	c.SP--
}

// push16 pushes a 16-bit value onto the stack
func (c *CPU) push16(value uint16) {
	high := uint8(value >> 8)
	low := uint8(value & 0xFF)
	c.push(high)
	c.push(low)
}

// pull16 pulls a 16-bit value from the stack
func (c *CPU) pull16() uint16 {
	low := uint16(c.pull())
	high := uint16(c.pull())
	return (high << 8) | low
}

func (c *CPU) pull() uint8 {
	c.SP++
	return c.Memory[0x0100|uint16(c.SP)]
}

// updateZN updates Zero and Negative flags based on value
func (c *CPU) updateZN(value uint8) {
	if value == 0 {
		c.P |= FlagZ
	} else {
		c.P &^= FlagZ
	}

	if value&0x80 != 0 {
		c.P |= FlagN
	} else {
		c.P &^= FlagN
	}
}
