package disassembler

import (
	"fmt"
	"github.com/newhook/6502/cpu"
)

// Instruction represents a decoded 6502 instruction
type Instruction struct {
	Name   string
	Mode   AddressingMode
	Bytes  int
	OpCode byte
}

// AddressingMode represents the different 6502 addressing modes
type AddressingMode int

const (
	Implicit AddressingMode = iota
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

// FormatOperand formats the operand bytes according to the addressing mode
func (mode AddressingMode) FormatOperand(bytes []byte) string {
	switch mode {
	case Implicit:
		return ""
	case Accumulator:
		return "A"
	case Immediate:
		return fmt.Sprintf("#$%02X", bytes[0])
	case ZeroPage:
		return fmt.Sprintf("$%02X", bytes[0])
	case ZeroPageX:
		return fmt.Sprintf("$%02X,X", bytes[0])
	case ZeroPageY:
		return fmt.Sprintf("$%02X,Y", bytes[0])
	case Absolute:
		return fmt.Sprintf("$%02X%02X", bytes[1], bytes[0])
	case AbsoluteX:
		return fmt.Sprintf("$%02X%02X,X", bytes[1], bytes[0])
	case AbsoluteY:
		return fmt.Sprintf("$%02X%02X,Y", bytes[1], bytes[0])
	case Indirect:
		return fmt.Sprintf("($%02X%02X)", bytes[1], bytes[0])
	case IndirectX:
		return fmt.Sprintf("($%02X,X)", bytes[0])
	case IndirectY:
		return fmt.Sprintf("($%02X),Y", bytes[0])
	case Relative:
		// Handle relative addressing for branch instructions
		offset := int8(bytes[0])
		// PC is assumed to be the address after the branch instruction (2 bytes)
		target := uint16(2) + uint16(offset)
		return fmt.Sprintf("$%04X", target)
	default:
		return "???"
	}
}

// GetOperandBytes returns the number of operand bytes for a given addressing mode
func (mode AddressingMode) GetOperandBytes() int {
	switch mode {
	case Implicit, Accumulator:
		return 0
	case Immediate, ZeroPage, ZeroPageX, ZeroPageY, IndirectX, IndirectY, Relative:
		return 1
	case Absolute, AbsoluteX, AbsoluteY, Indirect:
		return 2
	default:
		return 0
	}
}

// String returns a string representation of the addressing mode
func (mode AddressingMode) String() string {
	switch mode {
	case Implicit:
		return "Implicit"
	case Accumulator:
		return "Accumulator"
	case Immediate:
		return "Immediate"
	case ZeroPage:
		return "Zero Page"
	case ZeroPageX:
		return "Zero Page,X"
	case ZeroPageY:
		return "Zero Page,Y"
	case Absolute:
		return "Absolute"
	case AbsoluteX:
		return "Absolute,X"
	case AbsoluteY:
		return "Absolute,Y"
	case Indirect:
		return "Indirect"
	case IndirectX:
		return "Indirect,X"
	case IndirectY:
		return "Indirect,Y"
	case Relative:
		return "Relative"
	default:
		return "Unknown"
	}
}

// instructionSet maps opcodes to their corresponding instructions
var instructionSet = map[byte]Instruction{
	// Load/Store Operations
	cpu.LDA_IMM: {"LDA", Immediate, 2, cpu.LDA_IMM},
	cpu.LDA_ZP:  {"LDA", ZeroPage, 2, cpu.LDA_ZP},
	cpu.LDA_ZPX: {"LDA", ZeroPageX, 2, cpu.LDA_ZPX},
	cpu.LDA_ABS: {"LDA", Absolute, 3, cpu.LDA_ABS},
	cpu.LDA_ABX: {"LDA", AbsoluteX, 3, cpu.LDA_ABX},
	cpu.LDA_ABY: {"LDA", AbsoluteY, 3, cpu.LDA_ABY},
	cpu.LDA_INX: {"LDA", IndirectX, 2, cpu.LDA_INX},
	cpu.LDA_INY: {"LDA", IndirectY, 2, cpu.LDA_INY},

	cpu.LDX_IMM: {"LDX", Immediate, 2, cpu.LDX_IMM},
	cpu.LDX_ZP:  {"LDX", ZeroPage, 2, cpu.LDX_ZP},
	cpu.LDX_ZPY: {"LDX", ZeroPageY, 2, cpu.LDX_ZPY},
	cpu.LDX_ABS: {"LDX", Absolute, 3, cpu.LDX_ABS},
	cpu.LDX_ABY: {"LDX", AbsoluteY, 3, cpu.LDX_ABY},

	cpu.LDY_IMM: {"LDY", Immediate, 2, cpu.LDY_IMM},
	cpu.LDY_ZP:  {"LDY", ZeroPage, 2, cpu.LDY_ZP},
	cpu.LDY_ZPX: {"LDY", ZeroPageX, 2, cpu.LDY_ZPX},
	cpu.LDY_ABS: {"LDY", Absolute, 3, cpu.LDY_ABS},
	cpu.LDY_ABX: {"LDY", AbsoluteX, 3, cpu.LDY_ABX},

	cpu.STA_ZP:  {"STA", ZeroPage, 2, cpu.STA_ZP},
	cpu.STA_ZPX: {"STA", ZeroPageX, 2, cpu.STA_ZPX},
	cpu.STA_ABS: {"STA", Absolute, 3, cpu.STA_ABS},
	cpu.STA_ABX: {"STA", AbsoluteX, 3, cpu.STA_ABX},
	cpu.STA_ABY: {"STA", AbsoluteY, 3, cpu.STA_ABY},
	cpu.STA_INX: {"STA", IndirectX, 2, cpu.STA_INX},
	cpu.STA_INY: {"STA", IndirectY, 2, cpu.STA_INY},

	cpu.STX_ZP:  {"STX", ZeroPage, 2, cpu.STX_ZP},
	cpu.STX_ZPY: {"STX", ZeroPageY, 2, cpu.STX_ZPY},
	cpu.STX_ABS: {"STX", Absolute, 3, cpu.STX_ABS},

	cpu.STY_ZP:  {"STY", ZeroPage, 2, cpu.STY_ZP},
	cpu.STY_ZPX: {"STY", ZeroPageX, 2, cpu.STY_ZPX},
	cpu.STY_ABS: {"STY", Absolute, 3, cpu.STY_ABS},

	// Register Instructions
	cpu.TAX: {"TAX", Implicit, 1, cpu.TAX},
	cpu.TXA: {"TXA", Implicit, 1, cpu.TXA},
	cpu.TAY: {"TAY", Implicit, 1, cpu.TAY},
	cpu.TYA: {"TYA", Implicit, 1, cpu.TYA},
	cpu.TSX: {"TSX", Implicit, 1, cpu.TSX},
	cpu.TXS: {"TXS", Implicit, 1, cpu.TXS},

	// Stack Operations
	cpu.PHA: {"PHA", Implicit, 1, cpu.PHA},
	cpu.PLA: {"PLA", Implicit, 1, cpu.PLA},
	cpu.PHP: {"PHP", Implicit, 1, cpu.PHP},
	cpu.PLP: {"PLP", Implicit, 1, cpu.PLP},

	// Logical Operations
	cpu.AND_IMM: {"AND", Immediate, 2, cpu.AND_IMM},
	cpu.AND_ZP:  {"AND", ZeroPage, 2, cpu.AND_ZP},
	cpu.AND_ZPX: {"AND", ZeroPageX, 2, cpu.AND_ZPX},
	cpu.AND_ABS: {"AND", Absolute, 3, cpu.AND_ABS},
	cpu.AND_ABX: {"AND", AbsoluteX, 3, cpu.AND_ABX},
	cpu.AND_ABY: {"AND", AbsoluteY, 3, cpu.AND_ABY},
	cpu.AND_INX: {"AND", IndirectX, 2, cpu.AND_INX},
	cpu.AND_INY: {"AND", IndirectY, 2, cpu.AND_INY},

	cpu.EOR_IMM: {"EOR", Immediate, 2, cpu.EOR_IMM},
	cpu.EOR_ZP:  {"EOR", ZeroPage, 2, cpu.EOR_ZP},
	cpu.EOR_ZPX: {"EOR", ZeroPageX, 2, cpu.EOR_ZPX},
	cpu.EOR_ABS: {"EOR", Absolute, 3, cpu.EOR_ABS},
	cpu.EOR_ABX: {"EOR", AbsoluteX, 3, cpu.EOR_ABX},
	cpu.EOR_ABY: {"EOR", AbsoluteY, 3, cpu.EOR_ABY},
	cpu.EOR_INX: {"EOR", IndirectX, 2, cpu.EOR_INX},
	cpu.EOR_INY: {"EOR", IndirectY, 2, cpu.EOR_INY},

	cpu.ORA_IMM: {"ORA", Immediate, 2, cpu.ORA_IMM},
	cpu.ORA_ZP:  {"ORA", ZeroPage, 2, cpu.ORA_ZP},
	cpu.ORA_ZPX: {"ORA", ZeroPageX, 2, cpu.ORA_ZPX},
	cpu.ORA_ABS: {"ORA", Absolute, 3, cpu.ORA_ABS},
	cpu.ORA_ABX: {"ORA", AbsoluteX, 3, cpu.ORA_ABX},
	cpu.ORA_ABY: {"ORA", AbsoluteY, 3, cpu.ORA_ABY},
	cpu.ORA_INX: {"ORA", IndirectX, 2, cpu.ORA_INX},
	cpu.ORA_INY: {"ORA", IndirectY, 2, cpu.ORA_INY},

	cpu.BIT_ZP:  {"BIT", ZeroPage, 2, cpu.BIT_ZP},
	cpu.BIT_ABS: {"BIT", Absolute, 3, cpu.BIT_ABS},

	// Arithmetic Operations
	cpu.ADC_IMM: {"ADC", Immediate, 2, cpu.ADC_IMM},
	cpu.ADC_ZP:  {"ADC", ZeroPage, 2, cpu.ADC_ZP},
	cpu.ADC_ZPX: {"ADC", ZeroPageX, 2, cpu.ADC_ZPX},
	cpu.ADC_ABS: {"ADC", Absolute, 3, cpu.ADC_ABS},
	cpu.ADC_ABX: {"ADC", AbsoluteX, 3, cpu.ADC_ABX},
	cpu.ADC_ABY: {"ADC", AbsoluteY, 3, cpu.ADC_ABY},
	cpu.ADC_INX: {"ADC", IndirectX, 2, cpu.ADC_INX},
	cpu.ADC_INY: {"ADC", IndirectY, 2, cpu.ADC_INY},

	cpu.SBC_IMM: {"SBC", Immediate, 2, cpu.SBC_IMM},
	cpu.SBC_ZP:  {"SBC", ZeroPage, 2, cpu.SBC_ZP},
	cpu.SBC_ZPX: {"SBC", ZeroPageX, 2, cpu.SBC_ZPX},
	cpu.SBC_ABS: {"SBC", Absolute, 3, cpu.SBC_ABS},
	cpu.SBC_ABX: {"SBC", AbsoluteX, 3, cpu.SBC_ABX},
	cpu.SBC_ABY: {"SBC", AbsoluteY, 3, cpu.SBC_ABY},
	cpu.SBC_INX: {"SBC", IndirectX, 2, cpu.SBC_INX},
	cpu.SBC_INY: {"SBC", IndirectY, 2, cpu.SBC_INY},

	cpu.CMP_IMM: {"CMP", Immediate, 2, cpu.CMP_IMM},
	cpu.CMP_ZP:  {"CMP", ZeroPage, 2, cpu.CMP_ZP},
	cpu.CMP_ZPX: {"CMP", ZeroPageX, 2, cpu.CMP_ZPX},
	cpu.CMP_ABS: {"CMP", Absolute, 3, cpu.CMP_ABS},
	cpu.CMP_ABX: {"CMP", AbsoluteX, 3, cpu.CMP_ABX},
	cpu.CMP_ABY: {"CMP", AbsoluteY, 3, cpu.CMP_ABY},
	cpu.CMP_INX: {"CMP", IndirectX, 2, cpu.CMP_INX},
	cpu.CMP_INY: {"CMP", IndirectY, 2, cpu.CMP_INY},

	cpu.CPX_IMM: {"CPX", Immediate, 2, cpu.CPX_IMM},
	cpu.CPX_ZP:  {"CPX", ZeroPage, 2, cpu.CPX_ZP},
	cpu.CPX_ABS: {"CPX", Absolute, 3, cpu.CPX_ABS},

	cpu.CPY_IMM: {"CPY", Immediate, 2, cpu.CPY_IMM},
	cpu.CPY_ZP:  {"CPY", ZeroPage, 2, cpu.CPY_ZP},
	cpu.CPY_ABS: {"CPY", Absolute, 3, cpu.CPY_ABS},

	// Increments & Decrements
	cpu.INC_ZP:  {"INC", ZeroPage, 2, cpu.INC_ZP},
	cpu.INC_ZPX: {"INC", ZeroPageX, 2, cpu.INC_ZPX},
	cpu.INC_ABS: {"INC", Absolute, 3, cpu.INC_ABS},
	cpu.INC_ABX: {"INC", AbsoluteX, 3, cpu.INC_ABX},

	cpu.INX: {"INX", Implicit, 1, cpu.INX},
	cpu.INY: {"INY", Implicit, 1, cpu.INY},

	cpu.DEC_ZP:  {"DEC", ZeroPage, 2, cpu.DEC_ZP},
	cpu.DEC_ZPX: {"DEC", ZeroPageX, 2, cpu.DEC_ZPX},
	cpu.DEC_ABS: {"DEC", Absolute, 3, cpu.DEC_ABS},
	cpu.DEC_ABX: {"DEC", AbsoluteX, 3, cpu.DEC_ABX},

	cpu.DEX: {"DEX", Implicit, 1, cpu.DEX},
	cpu.DEY: {"DEY", Implicit, 1, cpu.DEY},

	// Shifts & Rotates
	cpu.ASL_ACC: {"ASL", Accumulator, 1, cpu.ASL_ACC},
	cpu.ASL_ZP:  {"ASL", ZeroPage, 2, cpu.ASL_ZP},
	cpu.ASL_ZPX: {"ASL", ZeroPageX, 2, cpu.ASL_ZPX},
	cpu.ASL_ABS: {"ASL", Absolute, 3, cpu.ASL_ABS},
	cpu.ASL_ABX: {"ASL", AbsoluteX, 3, cpu.ASL_ABX},

	cpu.LSR_ACC: {"LSR", Accumulator, 1, cpu.LSR_ACC},
	cpu.LSR_ZP:  {"LSR", ZeroPage, 2, cpu.LSR_ZP},
	cpu.LSR_ZPX: {"LSR", ZeroPageX, 2, cpu.LSR_ZPX},
	cpu.LSR_ABS: {"LSR", Absolute, 3, cpu.LSR_ABS},
	cpu.LSR_ABX: {"LSR", AbsoluteX, 3, cpu.LSR_ABX},

	cpu.ROL_ACC: {"ROL", Accumulator, 1, cpu.ROL_ACC},
	cpu.ROL_ZP:  {"ROL", ZeroPage, 2, cpu.ROL_ZP},
	cpu.ROL_ZPX: {"ROL", ZeroPageX, 2, cpu.ROL_ZPX},
	cpu.ROL_ABS: {"ROL", Absolute, 3, cpu.ROL_ABS},
	cpu.ROL_ABX: {"ROL", AbsoluteX, 3, cpu.ROL_ABX},

	cpu.ROR_ACC: {"ROR", Accumulator, 1, cpu.ROR_ACC},
	cpu.ROR_ZP:  {"ROR", ZeroPage, 2, cpu.ROR_ZP},
	cpu.ROR_ZPX: {"ROR", ZeroPageX, 2, cpu.ROR_ZPX},
	cpu.ROR_ABS: {"ROR", Absolute, 3, cpu.ROR_ABS},
	cpu.ROR_ABX: {"ROR", AbsoluteX, 3, cpu.ROR_ABX},

	// Jumps & Calls
	cpu.JMP_ABS: {"JMP", Absolute, 3, cpu.JMP_ABS},
	cpu.JMP_IND: {"JMP", Indirect, 3, cpu.JMP_IND},
	cpu.JSR_ABS: {"JSR", Absolute, 3, cpu.JSR_ABS},
	cpu.RTS:     {"RTS", Implicit, 1, cpu.RTS},

	// Branches
	cpu.BCC: {"BCC", Relative, 2, cpu.BCC},
	cpu.BCS: {"BCS", Relative, 2, cpu.BCS},
	cpu.BEQ: {"BEQ", Relative, 2, cpu.BEQ},
	cpu.BMI: {"BMI", Relative, 2, cpu.BMI},
	cpu.BNE: {"BNE", Relative, 2, cpu.BNE},
	cpu.BPL: {"BPL", Relative, 2, cpu.BPL},
	cpu.BVC: {"BVC", Relative, 2, cpu.BVC},
	cpu.BVS: {"BVS", Relative, 2, cpu.BVS},

	// Status Flag Changes
	cpu.CLC: {"CLC", Implicit, 1, cpu.CLC},
	cpu.SEC: {"SEC", Implicit, 1, cpu.SEC},
	cpu.CLI: {"CLI", Implicit, 1, cpu.CLI},
	cpu.SEI: {"SEI", Implicit, 1, cpu.SEI},
	cpu.CLV: {"CLV", Implicit, 1, cpu.CLV},
	cpu.CLD: {"CLD", Implicit, 1, cpu.CLD},
	cpu.SED: {"SED", Implicit, 1, cpu.SED},

	// System Functions
	cpu.BRK: {"BRK", Implicit, 1, cpu.BRK},
	cpu.RTI: {"RTI", Implicit, 1, cpu.RTI},
	cpu.NOP: {"NOP", Implicit, 1, cpu.NOP},
}
