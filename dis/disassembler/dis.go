package disassembler

import (
	"fmt"
	"github.com/newhook/6502/cpu"
	"strings"
)

const maxMemory = 0xffff

type Location struct {
	PC           uint16
	Value        uint8
	OperandBytes []byte
	Inst         *Instruction
}

func (l Location) instruction() string {
	if l.Inst == nil {
		return fmt.Sprintf("$%04X: db $%02X        ; Invalid opcode\n", l.PC, l.Value)
	}
	operand := l.Inst.Mode.FormatOperand(l.OperandBytes)
	if operand == "" {
		return l.Inst.Name
	}

	// Special case for relative addressing - update target address based on PC
	if l.Inst.Mode == Relative {
		offset := int8(l.OperandBytes[0])
		target := l.PC + 2 + uint16(offset)
		return fmt.Sprintf("%s $%04X", l.Inst.Name, target)
	}

	return fmt.Sprintf("%s %s", l.Inst.Name, operand)
}

func (l Location) Size() int {
	if l.Inst == nil {
		return 1
	}
	return 1 + l.Inst.Mode.GetOperandBytes()
}

func (l Location) String() string {
	var operandCount int
	if l.Inst != nil {
		operandCount = l.Inst.Mode.GetOperandBytes()
	}

	// Format the hex dump
	var hexDump string
	if operandCount == 0 {
		hexDump = fmt.Sprintf("%02X", l.Value)
	} else if operandCount == 1 {
		hexDump = fmt.Sprintf("%02X %02X", l.Value, l.OperandBytes[0])
	} else {
		hexDump = fmt.Sprintf("%02X %02X %02X", l.Value, l.OperandBytes[0], l.OperandBytes[1])
	}

	return fmt.Sprintf("$%04X: %-8s  %s", l.PC, hexDump, l.instruction())
}

// Decode takes an opcode and returns the corresponding instruction
func Decode(opcode byte) (Instruction, bool) {
	instruction, exists := instructionSet[opcode]
	return instruction, exists
}

func DisassembleInstructions(memory cpu.MemoryBus) []Location {
	pc := 0
	endAddr := maxMemory

	var rows []Location
	for pc < endAddr {
		loc := disassembleLocation(memory, pc)
		rows = append(rows, loc)
		pc += loc.Size()
	}

	return rows
}

// DisassembleMemory disassembles a range of memory starting at the given address
func DisassembleMemory(memory cpu.MemoryBus, startAddr int, length int) string {
	var out strings.Builder
	pc := startAddr
	endAddr := startAddr + length

	for pc < endAddr {
		loc := disassembleLocation(memory, pc)
		out.WriteString(loc.String())
		out.WriteString("\n")
		pc += loc.Size()
	}

	return out.String()
}

func disassembleLocation(memory cpu.MemoryBus, pc int) Location {
	// Get opcode
	opcode := memory.Read(uint16(pc))
	l := Location{PC: uint16(pc), Value: opcode}

	// Decode instruction
	inst, exists := instructionSet[opcode]
	if !exists {
		// Handle invalid opcode
		return l
	}

	// Get operand bytes based on addressing mode
	operandCount := inst.Mode.GetOperandBytes()

	// Bounds check
	if pc+operandCount >= maxMemory {
		return l
		//row := fmt.Sprintf("$%04X: db $%02X        ; Incomplete instruction\n", pc, opcode)
		//return pc, row
	}
	l.Inst = &inst

	// Extract operand bytes
	if operandCount > 0 {
		for i := 0; i < operandCount; i++ {
			//l.OperandBytes = memory[pc+1 : pc+1+operandCount]
			l.OperandBytes[i] = memory.Read(uint16(pc + 1 + i))
		}
	}

	return l
}

// DisassembleBytes is a convenience function for disassembling a slice of bytes
func DisassembleBytes(bytes cpu.MemoryBus) string {
	return DisassembleMemory(bytes, 0, maxMemory)
}
