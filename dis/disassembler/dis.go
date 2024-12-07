package disassembler

import (
	"fmt"
	"strings"
)

// Decode takes an opcode and returns the corresponding instruction
func Decode(opcode byte) (Instruction, bool) {
	instruction, exists := instructionSet[opcode]
	return instruction, exists
}

// DisassembleInstruction formats a complete instruction with its operands
func DisassembleInstruction(inst Instruction, operandBytes []byte, pc int) string {
	operand := inst.Mode.FormatOperand(operandBytes)
	if operand == "" {
		return inst.Name
	}

	// Special case for relative addressing - update target address based on PC
	if inst.Mode == Relative {
		offset := int8(operandBytes[0])
		target := pc + 2 + int(offset)
		return fmt.Sprintf("%s $%04X", inst.Name, target)
	}

	return fmt.Sprintf("%s %s", inst.Name, operand)
}

// DisassembleMemory disassembles a range of memory starting at the given address
func DisassembleMemory(memory []byte, startAddr int, length int) string {
	var out strings.Builder
	pc := startAddr
	endAddr := startAddr + length

	for pc < endAddr {
		// Get opcode
		opcode := memory[pc]

		// Decode instruction
		inst, exists := instructionSet[opcode]
		if !exists {
			// Handle invalid opcode
			fmt.Fprintf(&out, "$%04X: db $%02X        ; Invalid opcode\n", pc, opcode)
			pc++
			continue
		}

		// Get operand bytes based on addressing mode
		operandCount := inst.Mode.GetOperandBytes()
		var operandBytes []byte

		// Bounds check
		if pc+operandCount >= len(memory) {
			fmt.Fprintf(&out, "$%04X: db $%02X        ; Incomplete instruction\n", pc, opcode)
			break
		}

		// Extract operand bytes
		if operandCount > 0 {
			operandBytes = memory[pc+1 : pc+1+operandCount]
		}

		// Format the instruction with its operands
		asmInst := DisassembleInstruction(inst, operandBytes, pc)

		// Format the hex dump
		var hexDump string
		if operandCount == 0 {
			hexDump = fmt.Sprintf("%02X", opcode)
		} else if operandCount == 1 {
			hexDump = fmt.Sprintf("%02X %02X", opcode, operandBytes[0])
		} else {
			hexDump = fmt.Sprintf("%02X %02X %02X", opcode, operandBytes[0], operandBytes[1])
		}

		// Output formatted line
		fmt.Fprintf(&out, "$%04X: %-8s  %s\n", pc, hexDump, asmInst)

		// Advance PC
		pc += 1 + operandCount
	}

	return out.String()
}

// DisassembleBytes is a convenience function for disassembling a slice of bytes
func DisassembleBytes(bytes []byte) string {
	return DisassembleMemory(bytes, 0, len(bytes))
}
