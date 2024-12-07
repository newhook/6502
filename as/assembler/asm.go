package assembler

import (
	"fmt"
)

// Symbol represents a label or variable in the assembly
type Symbol struct {
	Name      string
	Value     uint16
	IsDefined bool
}

// Assembler holds the state of our assembler
type Assembler struct {
	symbols     map[string]*Symbol
	currentPass int
	pc          uint16
	output      []byte
	errors      []string
}

// NewAssembler creates a new instance of our assembler
func NewAssembler() *Assembler {
	return &Assembler{
		symbols: make(map[string]*Symbol),
		pc:      0,
		errors:  make([]string, 0),
	}
}

// Helper functions for assembler
func (a *Assembler) Assemble(source string) error {
	a.currentPass = 1
	a.pc = 0
	a.output = make([]byte, 0)

	// First pass: collect symbols
	lexer := NewLexer(source)
	parser := NewParser(lexer, a)

	for {
		line, err := parser.ParseLine()
		if err != nil {
			return err
		}
		if line == nil {
			break
		}

		// Handle labels
		if line.Label != "" {
			a.symbols[line.Label] = &Symbol{
				Name:      line.Label,
				Value:     a.pc,
				IsDefined: true,
			}
			//fmt.Printf("Label: %s PC: %x\n", line.Label, a.pc)
		}
		if line.Directive != "" {
			if handler, exists := directiveHandlers[line.Directive]; exists {
				if err := handler(a, line.Operand); err != nil {
					return err
				}
			}
		}

		// Update PC based on instruction size
		if line.Instruction != "" {
			if inst, exists := instructionSet[line.Instruction]; exists {
				if mode, exists := inst.Modes[line.AddressMode]; exists {
					a.pc += uint16(mode.Size)
					//fmt.Printf("inst: %s Size: %x PC: %x\n", line.Instruction, mode.Size, a.pc)
				}
			}
		}
	}

	// Second pass: generate code
	a.currentPass = 2
	a.pc = 0
	lexer = NewLexer(source)
	parser = NewParser(lexer, a)

	for {
		line, err := parser.ParseLine()
		if err != nil {
			return err
		}
		if line == nil {
			break
		}

		err = a.generateCode(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Assembler) generateCode(line *Line) error {
	// ignore directive handlers here.
	if line.Directive != "" {
		if _, exists := directiveHandlers[line.Directive]; exists {
			if handler, exists := directiveHandlers[line.Directive]; exists {
				if err := handler(a, line.Operand); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if line.Instruction == "" {
		return nil
	}

	inst, exists := instructionSet[line.Instruction]
	if !exists {
		return fmt.Errorf("unknown instruction: %s", line.Instruction)
	}

	// If we have a symbol reference, get its final value
	if line.SymbolName != "" {
		if symbol, exists := a.symbols[line.SymbolName]; exists {
			line.Value = symbol.Value
			// Only try to optimize if the value is in zero page
			if line.Value < 0x100 {
				var optimizedMode AddressMode
				switch line.AddressMode {
				case Absolute:
					optimizedMode = ZeroPage
				case AbsoluteX:
					optimizedMode = ZeroPageX
				case AbsoluteY:
					optimizedMode = ZeroPageY
				}
				// Only optimize if the instruction supports the zero page mode
				if optimizedMode != line.AddressMode {
					if _, supported := inst.Modes[optimizedMode]; supported {
						line.AddressMode = optimizedMode
					}
				}
			}
		}
	}

	mode, exists := inst.Modes[line.AddressMode]
	if !exists {
		return fmt.Errorf("invalid addressing mode for instruction %s", line.Instruction)
	}

	// Output opcode
	a.output = append(a.output, mode.Opcode)

	if mode.AddressMode == Relative {
		// Calculate relative offset
		// PC will be at next instruction when branch is executed
		nextPC := a.pc + 2
		offset := int16(line.Value) - int16(nextPC)

		// Check if branch is in range (-128 to +127)
		if offset < -128 || offset > 127 {
			return fmt.Errorf("branch target out of range (%d bytes)", offset)
		}

		// Output the offset.
		a.output = append(a.output, uint8(offset))
	} else {
		// Output operand bytes
		switch mode.Size {
		case 2:
			a.output = append(a.output, uint8(line.Value))
		case 3:
			a.output = append(a.output, uint8(line.Value))
			a.output = append(a.output, uint8(line.Value>>8))
		}
	}

	a.pc += uint16(mode.Size)
	return nil
}

func (a *Assembler) GetOutput() []byte {
	return a.output
}
