package assembler

import "strings"

// Token represents the smallest unit of code in our assembly
type Token struct {
	Type    TokenType
	Value   string
	LineNum int
}

// TokenType identifies different types of tokens
type TokenType int

const (
	LABEL TokenType = iota
	INSTRUCTION
	DIRECTIVE
	OPERAND
	COMMENT
	EOL
	EOF = 6
)

// Lexer breaks source code into tokens
type Lexer struct {
	input     string
	position  int
	lineNum   int
	lastToken Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		position: 0,
		lineNum:  1,
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.position >= len(l.input) {
		return Token{Type: EOF, LineNum: l.lineNum}
	}

	char := l.input[l.position]

	switch {
	case isLetter(char):
		return l.readIdentifier()
	case isDigit(char) || char == '$' || char == '%':
		return l.readNumber()
	case char == ';':
		return l.readComment()
	case char == ':':
		l.position++
		if l.lastToken.Type == INSTRUCTION {
			// Convert the last instruction token to a label
			l.lastToken.Type = LABEL
			return l.lastToken
		}
		return Token{Type: OPERAND, Value: ":", LineNum: l.lineNum}
	case char == '\n':
		l.lineNum++
		l.position++
		return Token{Type: EOL, LineNum: l.lineNum - 1}
	default:
		token := Token{
			Type:    OPERAND,
			Value:   string(char),
			LineNum: l.lineNum,
		}
		l.position++
		return token
	}
}

func (l *Lexer) readIdentifier() Token {
	position := l.position
	for l.position < len(l.input) && (isLetter(l.input[l.position]) || isDigit(l.input[l.position])) {
		l.position++
	}

	value := l.input[position:l.position]
	var tokenType TokenType

	// Check if it's an instruction
	if _, exists := instructionSet[strings.ToUpper(value)]; exists {
		tokenType = INSTRUCTION
	} else if strings.HasPrefix(value, ".") {
		tokenType = DIRECTIVE
	} else {
		tokenType = LABEL
	}

	token := Token{
		Type:    tokenType,
		Value:   value,
		LineNum: l.lineNum,
	}
	l.lastToken = token
	return token
}

func (l *Lexer) readNumber() Token {
	position := l.position
	base := 10

	// Handle hex ($) and binary (%) prefixes
	if l.input[position] == '$' {
		base = 16
		l.position++
	} else if l.input[position] == '%' {
		base = 2
		l.position++
	}

	for l.position < len(l.input) && (isDigit(l.input[l.position]) ||
		(base == 16 && isHexDigit(l.input[l.position]))) {
		l.position++
	}

	return Token{
		Type:    OPERAND,
		Value:   l.input[position:l.position],
		LineNum: l.lineNum,
	}
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) && (l.input[l.position] == ' ' || l.input[l.position] == '\t' || l.input[l.position] == '\r') {
		l.position++
	}
}

func (l *Lexer) readComment() Token {
	position := l.position
	for l.position < len(l.input) && l.input[l.position] != '\n' {
		l.position++
	}
	return Token{
		Type:    COMMENT,
		Value:   l.input[position:l.position],
		LineNum: l.lineNum,
	}
}

// Helper functions
func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_' || ch == '.'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}
