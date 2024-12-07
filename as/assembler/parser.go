package assembler

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser represents the assembly parser
type Parser struct {
	lexer     *Lexer
	assembler *Assembler
	tokens    []Token
	position  int
}

// Line represents a parsed assembly line
type Line struct {
	Label       string
	Instruction string
	Directive   string
	Operand     string
	AddressMode AddressMode
	Value       uint16
	IsRelative  bool
	SymbolName  string
}

func NewParser(lexer *Lexer, assembler *Assembler) *Parser {
	return &Parser{
		lexer:     lexer,
		assembler: assembler,
		tokens:    make([]Token, 0),
		position:  0,
	}
}

// parseOperand combines operand tokens into a single string
func (p *Parser) parseOperand() string {
	var operand strings.Builder
	for p.position < len(p.tokens) {
		operand.WriteString(p.tokens[p.position].Value)
		p.position++
	}
	return operand.String()
}

func (p *Parser) detectAddressMode(line *Line) error {
	operand := strings.TrimSpace(line.Operand)

	// Get instruction entry to check supported modes
	inst, exists := instructionSet[line.Instruction]
	if !exists {
		return fmt.Errorf("unknown instruction: %s", line.Instruction)
	}

	if operand == "" {
		// Check if this instruction can use accumulator mode with no operand
		switch line.Instruction {
		case "LSR", "ASL", "ROL", "ROR":
			if _, supported := inst.Modes[Accumulator]; supported {
				line.AddressMode = Accumulator
				return nil
			}
		}
		if _, supported := inst.Modes[Implicit]; supported {
			line.AddressMode = Implicit
			return nil
		}
		return fmt.Errorf("instruction %s requires an operand", line.Instruction)
	}

	if operand == "A" || operand == "a" {
		if _, supported := inst.Modes[Accumulator]; supported {
			line.AddressMode = Accumulator
			return nil
		}
		return fmt.Errorf("instruction %s does not support accumulator mode", line.Instruction)
	}

	// Remove spaces around commas and parentheses for consistent parsing
	operand = strings.ReplaceAll(operand, " ,", ",")
	operand = strings.ReplaceAll(operand, ", ", ",")
	operand = strings.ReplaceAll(operand, "( ", "(")
	operand = strings.ReplaceAll(operand, " )", ")")

	// Immediate addressing (#$xx or #xx)
	if strings.HasPrefix(operand, "#") {
		if _, supported := inst.Modes[Immediate]; supported {
			line.AddressMode = Immediate
			line.Value = p.parseValue(operand[1:])
			return nil
		}
		return fmt.Errorf("instruction %s does not support immediate mode", line.Instruction)
	}

	// Indirect addressing
	if strings.HasPrefix(operand, "(") {
		if strings.HasSuffix(operand, ",X)") {
			if _, supported := inst.Modes[IndirectX]; supported {
				line.AddressMode = IndirectX
				base := operand[1 : len(operand)-3]
				if !isNumeric(base) {
					line.SymbolName = base
				}
				line.Value = p.parseValue(base)
				return nil
			}
			return fmt.Errorf("instruction %s does not support indirect X mode", line.Instruction)
		}
		if strings.HasSuffix(operand, "),Y") {
			if _, supported := inst.Modes[IndirectY]; supported {
				line.AddressMode = IndirectY
				base := operand[1 : len(operand)-3]
				if !isNumeric(base) {
					line.SymbolName = base
				}
				line.Value = p.parseValue(base)
				return nil
			}
			return fmt.Errorf("instruction %s does not support indirect Y mode", line.Instruction)
		}
		if strings.HasSuffix(operand, ")") {
			if _, supported := inst.Modes[Indirect]; supported {
				line.AddressMode = Indirect
				base := operand[1 : len(operand)-1]
				if !isNumeric(base) {
					line.SymbolName = base
				}
				line.Value = p.parseValue(base)
				return nil
			}
			return fmt.Errorf("instruction %s does not support indirect mode", line.Instruction)
		}
	}

	// X/Y indexing
	if strings.HasSuffix(operand, ",X") {
		base := operand[:len(operand)-2]
		value := p.parseValue(base)

		// Try zero page X if value fits and mode is supported
		if value < 0x100 {
			if _, supported := inst.Modes[ZeroPageX]; supported {
				line.AddressMode = ZeroPageX
				if !isNumeric(base) {
					line.SymbolName = base
				}
				line.Value = value
				return nil
			}
		}

		if _, supported := inst.Modes[AbsoluteX]; supported {
			line.AddressMode = AbsoluteX
			if !isNumeric(base) {
				line.SymbolName = base
			}
			line.Value = value
			return nil
		}

		return fmt.Errorf("instruction %s does not support X-indexed addressing", line.Instruction)
	}

	if strings.HasSuffix(operand, ",Y") {
		base := operand[:len(operand)-2]
		value := p.parseValue(base)

		// Try zero page Y if value fits and mode is supported
		if value < 0x100 {
			if _, supported := inst.Modes[ZeroPageY]; supported {
				line.AddressMode = ZeroPageY
				if !isNumeric(base) {
					line.SymbolName = base
				}
				line.Value = value
				return nil
			}
		}

		if _, supported := inst.Modes[AbsoluteY]; supported {
			line.AddressMode = AbsoluteY
			if !isNumeric(base) {
				line.SymbolName = base
			}
			line.Value = value
			return nil
		}

		return fmt.Errorf("instruction %s does not support Y-indexed addressing", line.Instruction)
	}

	// Non-indexed addressing
	value := p.parseValue(operand)

	// Try zero page if value fits and mode is supported
	if value < 0x100 {
		if _, supported := inst.Modes[ZeroPage]; supported {
			line.AddressMode = ZeroPage
			if !isNumeric(operand) {
				line.SymbolName = operand
			}
			line.Value = value
			return nil
		}
	}

	if _, supported := inst.Modes[Absolute]; supported {
		line.AddressMode = Absolute
		if !isNumeric(operand) {
			line.SymbolName = operand
		}
		line.Value = value
		return nil
	}

	if _, supported := inst.Modes[Relative]; supported {
		line.AddressMode = Relative
		if !isNumeric(operand) {
			line.SymbolName = operand
		}
		line.Value = value
		return nil
	}

	return fmt.Errorf("no valid addressing mode found for instruction %s with operand %s",
		line.Instruction, line.Operand)
}

// isNumeric checks if the string represents a number (hex, binary, or decimal)
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "$") || strings.HasPrefix(s, "%") {
		return true
	}
	_, err := strconv.ParseUint(s, 10, 16)
	return err == nil
}

// parseValue converts a string value to uint16
func (p *Parser) parseValue(s string) uint16 {
	s = strings.TrimSpace(s)

	// Check for hex value ($)
	if strings.HasPrefix(s, "$") {
		val, err := strconv.ParseUint(s[1:], 16, 16)
		if err == nil {
			return uint16(val)
		}
	}

	// Check for binary value (%)
	if strings.HasPrefix(s, "%") {
		val, err := strconv.ParseUint(s[1:], 2, 16)
		if err == nil {
			return uint16(val)
		}
	}

	// Check if it's a symbol
	if p.assembler.symbols != nil {
		if symbol, exists := p.assembler.symbols[s]; exists {
			return symbol.Value
		}
	}

	// Try decimal
	val, err := strconv.ParseUint(s, 10, 16)
	if err == nil {
		return uint16(val)
	}

	return 0
}

func (p *Parser) ParseLine() (*Line, error) {
	p.tokens = make([]Token, 0)

	// Collect all tokens until EOL
	for {
		token := p.lexer.NextToken()
		if token.Type == EOF {
			if len(p.tokens) == 0 {
				return nil, nil
			}
			break
		}
		if token.Type == EOL {
			break
		}
		if token.Type != COMMENT {
			p.tokens = append(p.tokens, token)
		}
	}

	line := &Line{}
	if len(p.tokens) == 0 {
		return line, nil
	}
	p.position = 0

	if p.position < len(p.tokens) {
		token := p.tokens[p.position]
		if token.Type == LABEL {
			line.Label = token.Value
			p.position++
			if p.position < len(p.tokens) {
				if p.tokens[p.position].Type == OPERAND {
					p.position++
				}
			}
		}
	}

	if p.position < len(p.tokens) {
		token := p.tokens[p.position]
		if token.Type == DIRECTIVE {
			line.Directive = strings.ToLower(token.Value)
			p.position++
			line.Operand = p.parseOperand()
		} else if token.Type == INSTRUCTION {
			line.Instruction = strings.ToUpper(token.Value)
			p.position++
			line.Operand = p.parseOperand()
			if err := p.detectAddressMode(line); err != nil {
				return nil, err
			}
		}
	}

	return line, nil
}

// DirectiveHandler defines a function type for directive processing
type DirectiveHandler func(a *Assembler, operand string) error

// Map of directives to their handlers
var directiveHandlers = map[string]DirectiveHandler{
	".org":  handleOrg,
	".byte": handleByte,
	".word": handleWord,
}

// handleOrg processes the .org directive
func handleOrg(a *Assembler, operand string) error {
	value := parseNumber(operand)
	if a.currentPass == 1 {
		a.pc = value
	} else {
		// On pass 2, pad output to reach org address if needed,
		// but if the .org directive is the first instruction.
		if len(a.output) > 0 {
			for count := value - a.pc; count > 0; count-- {
				a.output = append(a.output, 0)
			}
		}
		a.pc = value
	}
	return nil
}

// handleByte processes the .byte directive
func handleByte(a *Assembler, operand string) error {
	values := parseByteList(operand)
	if a.currentPass == 2 {
		for _, v := range values {
			a.output = append(a.output, v)
		}
	}
	a.pc += uint16(len(values))
	return nil
}

// handleWord processes the .word directive
func handleWord(a *Assembler, operand string) error {
	values := parseWordList(operand)
	if a.currentPass == 2 {
		for _, v := range values {
			a.output = append(a.output, uint8(v&0xFF))
			a.output = append(a.output, uint8(v>>8))
		}
	}
	a.pc += uint16(len(values) * 2)
	return nil
}

// parseByteList splits a comma-separated list of values and parses each one
func parseByteList(operand string) []uint8 {
	parts := strings.Split(operand, ",")
	values := make([]uint8, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Handle string literals
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			str := part[1 : len(part)-1]
			for _, ch := range str {
				values = append(values, uint8(ch))
			}
		} else {
			value := parseNumber(part)
			values = append(values, uint8(value))
		}
	}
	return values
}

// parseWordList splits a comma-separated list of values and parses each one
func parseWordList(operand string) []uint16 {
	parts := strings.Split(operand, ",")
	values := make([]uint16, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		value := parseNumber(part)
		values = append(values, uint16(value))
	}
	return values
}

// parseNumber handles different number formats (hex, binary, decimal)
func parseNumber(s string) uint16 {
	s = strings.TrimSpace(s)

	// Handle hex ($)
	if strings.HasPrefix(s, "$") {
		val, err := strconv.ParseUint(s[1:], 16, 16)
		if err == nil {
			return uint16(val)
		}
	}

	// Handle binary (%)
	if strings.HasPrefix(s, "%") {
		val, err := strconv.ParseUint(s[1:], 2, 16)
		if err == nil {
			return uint16(val)
		}
	}

	// Handle decimal
	val, err := strconv.ParseUint(s, 10, 16)
	if err == nil {
		return uint16(val)
	}

	return 0
}
