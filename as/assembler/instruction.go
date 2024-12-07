package assembler

// AddressMode represents different 6502 addressing modes
type AddressMode int

const (
	Implicit AddressMode = iota
	Accumulator
	Immediate
	ZeroPage
	ZeroPageX
	ZeroPageY
	Absolute
	AbsoluteX
	AbsoluteY
	Indirect
	IndirectX
	IndirectY
	Relative
)

// Instruction represents a 6502 assembly instruction
type Instruction struct {
	Opcode      byte
	Size        int
	Cycles      int
	AddressMode AddressMode
}

// InstructionEntry represents an entry in our instruction lookup table
type InstructionEntry struct {
	BaseOpcode byte
	Modes      map[AddressMode]Instruction
}

// Create instruction set lookup table
var instructionSet = map[string]InstructionEntry{
	"ADC": {
		BaseOpcode: 0x69,
		Modes: map[AddressMode]Instruction{
			Immediate: {0x69, 2, 2, Immediate},
			ZeroPage:  {0x65, 2, 3, ZeroPage},
			ZeroPageX: {0x75, 2, 4, ZeroPageX},
			Absolute:  {0x6D, 3, 4, Absolute},
			AbsoluteX: {0x7D, 3, 4, AbsoluteX},
			AbsoluteY: {0x79, 3, 4, AbsoluteY},
			IndirectX: {0x61, 2, 6, IndirectX},
			IndirectY: {0x71, 2, 5, IndirectY},
		},
	},
	"AND": {
		BaseOpcode: 0x29,
		Modes: map[AddressMode]Instruction{
			Immediate: {0x29, 2, 2, Immediate},
			ZeroPage:  {0x25, 2, 3, ZeroPage},
			ZeroPageX: {0x35, 2, 4, ZeroPageX},
			Absolute:  {0x2D, 3, 4, Absolute},
			AbsoluteX: {0x3D, 3, 4, AbsoluteX},
			AbsoluteY: {0x39, 3, 4, AbsoluteY},
			IndirectX: {0x21, 2, 6, IndirectX},
			IndirectY: {0x31, 2, 5, IndirectY},
		},
	},
	"ASL": {
		BaseOpcode: 0x0A,
		Modes: map[AddressMode]Instruction{
			Accumulator: {0x0A, 1, 2, Accumulator},
			ZeroPage:    {0x06, 2, 5, ZeroPage},
			ZeroPageX:   {0x16, 2, 6, ZeroPageX},
			Absolute:    {0x0E, 3, 6, Absolute},
			AbsoluteX:   {0x1E, 3, 7, AbsoluteX},
		},
	},
	"BIT": {
		BaseOpcode: 0x24,
		Modes: map[AddressMode]Instruction{
			ZeroPage: {0x24, 2, 3, ZeroPage},
			Absolute: {0x2C, 3, 4, Absolute},
		},
	},
	"BPL": {BaseOpcode: 0x10, Modes: map[AddressMode]Instruction{Relative: {0x10, 2, 2, Relative}}},
	"BMI": {BaseOpcode: 0x30, Modes: map[AddressMode]Instruction{Relative: {0x30, 2, 2, Relative}}},
	"BVC": {BaseOpcode: 0x50, Modes: map[AddressMode]Instruction{Relative: {0x50, 2, 2, Relative}}},
	"BVS": {BaseOpcode: 0x70, Modes: map[AddressMode]Instruction{Relative: {0x70, 2, 2, Relative}}},
	"BCC": {BaseOpcode: 0x90, Modes: map[AddressMode]Instruction{Relative: {0x90, 2, 2, Relative}}},
	"BCS": {BaseOpcode: 0xB0, Modes: map[AddressMode]Instruction{Relative: {0xB0, 2, 2, Relative}}},
	"BNE": {BaseOpcode: 0xD0, Modes: map[AddressMode]Instruction{Relative: {0xD0, 2, 2, Relative}}},
	"BEQ": {BaseOpcode: 0xF0, Modes: map[AddressMode]Instruction{Relative: {0xF0, 2, 2, Relative}}},
	"BRK": {BaseOpcode: 0x00, Modes: map[AddressMode]Instruction{Relative: {0x00, 1, 7, Relative}}},
	"CMP": {
		BaseOpcode: 0xC9,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xC9, 2, 2, Immediate},
			ZeroPage:  {0xC5, 2, 3, ZeroPage},
			ZeroPageX: {0xD5, 2, 4, ZeroPageX},
			Absolute:  {0xCD, 3, 4, Absolute},
			AbsoluteX: {0xDD, 3, 4, AbsoluteX},
			AbsoluteY: {0xD9, 3, 4, AbsoluteY},
			IndirectX: {0xC1, 2, 6, IndirectX},
			IndirectY: {0xD1, 2, 5, IndirectY},
		},
	},
	"CPX": {
		BaseOpcode: 0xE0,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xE0, 2, 2, Immediate},
			ZeroPage:  {0xE4, 2, 3, ZeroPage},
			Absolute:  {0xEC, 3, 4, Absolute},
		},
	},
	"CPY": {
		BaseOpcode: 0xC0,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xC0, 2, 2, Immediate},
			ZeroPage:  {0xC4, 2, 3, ZeroPage},
			Absolute:  {0xCC, 3, 4, Absolute},
		},
	},
	"DEC": {
		BaseOpcode: 0xC6,
		Modes: map[AddressMode]Instruction{
			ZeroPage:  {0xC6, 2, 5, ZeroPage},
			ZeroPageX: {0xD6, 2, 6, ZeroPageX},
			Absolute:  {0xCE, 3, 6, Absolute},
			AbsoluteX: {0xDE, 3, 7, AbsoluteX},
		},
	},
	"EOR": {
		BaseOpcode: 0x49,
		Modes: map[AddressMode]Instruction{
			Immediate: {0x49, 2, 2, Immediate},
			ZeroPage:  {0x45, 2, 3, ZeroPage},
			ZeroPageX: {0x55, 2, 4, ZeroPageX},
			Absolute:  {0x4D, 3, 4, Absolute},
			AbsoluteX: {0x5D, 3, 4, AbsoluteX},
			AbsoluteY: {0x59, 3, 4, AbsoluteY},
			IndirectX: {0x41, 2, 6, IndirectX},
			IndirectY: {0x51, 2, 5, IndirectY},
		},
	},
	"CLC": {BaseOpcode: 0x18, Modes: map[AddressMode]Instruction{Implicit: {0x18, 1, 2, Implicit}}},
	"SEC": {BaseOpcode: 0x38, Modes: map[AddressMode]Instruction{Implicit: {0x38, 1, 2, Implicit}}},
	"CLI": {BaseOpcode: 0x58, Modes: map[AddressMode]Instruction{Implicit: {0x58, 1, 2, Implicit}}},
	"SEI": {BaseOpcode: 0x78, Modes: map[AddressMode]Instruction{Implicit: {0x78, 1, 2, Implicit}}},
	"CLV": {BaseOpcode: 0xB8, Modes: map[AddressMode]Instruction{Implicit: {0xB8, 1, 2, Implicit}}},
	"CLD": {BaseOpcode: 0xD8, Modes: map[AddressMode]Instruction{Implicit: {0xD8, 1, 2, Implicit}}},
	"SED": {BaseOpcode: 0xF8, Modes: map[AddressMode]Instruction{Implicit: {0xF8, 1, 2, Implicit}}},
	"INC": {
		BaseOpcode: 0xE6,
		Modes: map[AddressMode]Instruction{
			ZeroPage:  {0xE6, 2, 5, ZeroPage},
			ZeroPageX: {0xF6, 2, 6, ZeroPageX},
			Absolute:  {0xEE, 3, 6, Absolute},
			AbsoluteX: {0xFE, 3, 7, AbsoluteX},
		},
	},
	"JMP": {
		BaseOpcode: 0x4C,
		Modes: map[AddressMode]Instruction{
			Absolute: {0x4C, 3, 3, Absolute},
			Indirect: {0x6C, 3, 5, Indirect},
		},
	},
	"JSR": {BaseOpcode: 0x20, Modes: map[AddressMode]Instruction{Absolute: {0x20, 3, 6, Absolute}}},
	"LDA": {
		BaseOpcode: 0xA9,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xA9, 2, 2, Immediate},
			ZeroPage:  {0xA5, 2, 3, ZeroPage},
			ZeroPageX: {0xB5, 2, 4, ZeroPageX},
			Absolute:  {0xAD, 3, 4, Absolute},
			AbsoluteX: {0xBD, 3, 4, AbsoluteX},
			AbsoluteY: {0xB9, 3, 4, AbsoluteY},
			IndirectX: {0xA1, 2, 6, IndirectX},
			IndirectY: {0xB1, 2, 5, IndirectY},
		},
	},
	"LDX": {
		BaseOpcode: 0xA2,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xA2, 2, 2, Immediate},
			ZeroPage:  {0xA6, 2, 3, ZeroPage},
			ZeroPageY: {0xB6, 2, 4, ZeroPageY},
			Absolute:  {0xAE, 3, 4, Absolute},
			AbsoluteY: {0xBE, 3, 4, AbsoluteY},
		},
	},
	"LDY": {
		BaseOpcode: 0xA0,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xA0, 2, 2, Immediate},
			ZeroPage:  {0xA4, 2, 3, ZeroPage},
			ZeroPageX: {0xB4, 2, 4, ZeroPageX},
			Absolute:  {0xAC, 3, 4, Absolute},
			AbsoluteX: {0xBC, 3, 4, AbsoluteX},
		},
	},
	"LSR": {
		BaseOpcode: 0x4A,
		Modes: map[AddressMode]Instruction{
			Accumulator: {0x4A, 1, 2, Accumulator},
			ZeroPage:    {0x46, 2, 5, ZeroPage},
			ZeroPageX:   {0x56, 2, 6, ZeroPageX},
			Absolute:    {0x4E, 3, 6, Absolute},
			AbsoluteX:   {0x5E, 3, 7, AbsoluteX},
		},
	},
	"NOP": {BaseOpcode: 0xEA, Modes: map[AddressMode]Instruction{Implicit: {0xEA, 1, 2, Implicit}}},
	"ORA": {
		BaseOpcode: 0x09,
		Modes: map[AddressMode]Instruction{
			Immediate: {0x09, 2, 2, Immediate},
			ZeroPage:  {0x05, 2, 3, ZeroPage},
			ZeroPageX: {0x15, 2, 4, ZeroPageX},
			Absolute:  {0x0D, 3, 4, Absolute},
			AbsoluteX: {0x1D, 3, 4, AbsoluteX},
			AbsoluteY: {0x19, 3, 4, AbsoluteY},
			IndirectX: {0x01, 2, 6, IndirectX},
			IndirectY: {0x11, 2, 5, IndirectY},
		},
	},
	"PHA": {BaseOpcode: 0x48, Modes: map[AddressMode]Instruction{Implicit: {0x48, 1, 3, Implicit}}},
	"PHP": {BaseOpcode: 0x08, Modes: map[AddressMode]Instruction{Implicit: {0x08, 1, 3, Implicit}}},
	"PLA": {BaseOpcode: 0x68, Modes: map[AddressMode]Instruction{Implicit: {0x68, 1, 4, Implicit}}},
	"PLP": {BaseOpcode: 0x28, Modes: map[AddressMode]Instruction{Implicit: {0x28, 1, 4, Implicit}}},
	"ROL": {
		BaseOpcode: 0x2A,
		Modes: map[AddressMode]Instruction{
			Accumulator: {0x2A, 1, 2, Accumulator},
			ZeroPage:    {0x26, 2, 5, ZeroPage},
			ZeroPageX:   {0x36, 2, 6, ZeroPageX},
			Absolute:    {0x2E, 3, 6, Absolute},
			AbsoluteX:   {0x3E, 3, 7, AbsoluteX},
		},
	},
	"ROR": {
		BaseOpcode: 0x6A,
		Modes: map[AddressMode]Instruction{
			Accumulator: {0x6A, 1, 2, Accumulator},
			ZeroPage:    {0x66, 2, 5, ZeroPage},
			ZeroPageX:   {0x76, 2, 6, ZeroPageX},
			Absolute:    {0x6E, 3, 6, Absolute},
			AbsoluteX:   {0x7E, 3, 7, AbsoluteX},
		},
	},
	"RTI": {BaseOpcode: 0x40, Modes: map[AddressMode]Instruction{Implicit: {0x40, 1, 6, Implicit}}},
	"RTS": {BaseOpcode: 0x60, Modes: map[AddressMode]Instruction{Implicit: {0x60, 1, 6, Implicit}}},
	"SBC": {
		BaseOpcode: 0xE9,
		Modes: map[AddressMode]Instruction{
			Immediate: {0xE9, 2, 2, Immediate},
			ZeroPage:  {0xE5, 2, 3, ZeroPage},
			ZeroPageX: {0xF5, 2, 4, ZeroPageX},
			Absolute:  {0xED, 3, 4, Absolute},
			AbsoluteX: {0xFD, 3, 4, AbsoluteX},
			AbsoluteY: {0xF9, 3, 4, AbsoluteY},
			IndirectX: {0xE1, 2, 6, IndirectX},
			IndirectY: {0xF1, 2, 5, IndirectY},
		},
	},
	"STA": {
		BaseOpcode: 0x85,
		Modes: map[AddressMode]Instruction{
			ZeroPage:  {0x85, 2, 3, ZeroPage},
			ZeroPageX: {0x95, 2, 4, ZeroPageX},
			Absolute:  {0x8D, 3, 4, Absolute},
			AbsoluteX: {0x9D, 3, 5, AbsoluteX},
			AbsoluteY: {0x99, 3, 5, AbsoluteY},
			IndirectX: {0x81, 2, 6, IndirectX},
			IndirectY: {0x91, 2, 6, IndirectY},
		},
	},
	"STX": {
		BaseOpcode: 0x86,
		Modes: map[AddressMode]Instruction{
			ZeroPage:  {0x86, 2, 3, ZeroPage},
			ZeroPageY: {0x96, 2, 4, ZeroPageY},
			Absolute:  {0x8E, 3, 4, Absolute},
		},
	},
	"STY": {
		BaseOpcode: 0x84,
		Modes: map[AddressMode]Instruction{
			ZeroPage:  {0x84, 2, 3, ZeroPage},
			ZeroPageX: {0x94, 2, 4, ZeroPageX},
			Absolute:  {0x8C, 3, 4, Absolute},
		},
	},
	"TAX": {BaseOpcode: 0xAA, Modes: map[AddressMode]Instruction{Implicit: {0xAA, 1, 2, Implicit}}},
	"TXA": {BaseOpcode: 0x8A, Modes: map[AddressMode]Instruction{Implicit: {0x8A, 1, 2, Implicit}}},
	"TAY": {BaseOpcode: 0xA8, Modes: map[AddressMode]Instruction{Implicit: {0xA8, 1, 2, Implicit}}},
	"TYA": {BaseOpcode: 0x98, Modes: map[AddressMode]Instruction{Implicit: {0x98, 1, 2, Implicit}}},
	"TSX": {BaseOpcode: 0xBA, Modes: map[AddressMode]Instruction{Implicit: {0xBA, 1, 2, Implicit}}},
	"TXS": {BaseOpcode: 0x9A, Modes: map[AddressMode]Instruction{Implicit: {0x9A, 1, 2, Implicit}}},
	"DEX": {BaseOpcode: 0xCA, Modes: map[AddressMode]Instruction{Implicit: {0xCA, 1, 2, Implicit}}},
	"DEY": {BaseOpcode: 0x88, Modes: map[AddressMode]Instruction{Implicit: {0x88, 1, 2, Implicit}}},
	"INX": {BaseOpcode: 0xE8, Modes: map[AddressMode]Instruction{Implicit: {0xE8, 1, 2, Implicit}}},
	"INY": {BaseOpcode: 0xC8, Modes: map[AddressMode]Instruction{Implicit: {0xC8, 1, 2, Implicit}}},
}
