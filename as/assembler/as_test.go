package assembler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleInstructions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "LDA immediate",
			input:    "LDA #$FF",
			expected: []byte{0xA9, 0xFF},
		},
		{
			name:     "LDA zero page",
			input:    "LDA $12",
			expected: []byte{0xA5, 0x12},
		},
		{
			name:     "LDA absolute",
			input:    "LDA $1234",
			expected: []byte{0xAD, 0x34, 0x12},
		},
		{
			name:     "STA absolute",
			input:    "STA $0081",
			expected: []byte{0x85, 0x81}, // Should use zero page
		},
		{
			name:     "LSR accumulator implicit",
			input:    "LSR",
			expected: []byte{0x4A},
		},
		{
			name:     "LSR accumulator explicit",
			input:    "LSR A",
			expected: []byte{0x4A},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAssembler()
			err := asm.Assemble(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, asm.output)
		})
	}
}

func TestBranchInstructions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name: "forward branch",
			input: `
				BEQ target
				NOP
				NOP
			target:
				RTS`,
			expected: []byte{0xF0, 0x02, 0xEA, 0xEA, 0x60},
		},
		{
			name: "backward branch",
			input: `
			start:
				NOP
				BEQ start
				RTS`,
			expected: []byte{0xEA, 0xF0, 0xFD, 0x60},
		},
		{
			name: "branch too far",
			input: `
				BEQ target
				.org $1000
			target:
				RTS`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAssembler()
			err := asm.Assemble(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, asm.output)
		})
	}
}

func TestDirectives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name: "org directive",
			input: `
				.org $1000
				LDA #$00`,
			expected: []byte{0xA9, 0x00},
		},
		{
			name: "multiple org directives",
			input: `
				.org $1000
				LDA #$00
				.org $1010
				LDA #$01`,
			expected: []byte{0xA9, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xA9, 0x01},
		},
		{
			name:     "byte directive",
			input:    `.byte $01, $02, $03`,
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "word directive",
			input:    `.word $1234, $5678`,
			expected: []byte{0x34, 0x12, 0x78, 0x56},
		},
		{
			name:     "byte string directive",
			input:    `.byte "Hello"`,
			expected: []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAssembler()
			err := asm.Assemble(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, asm.output)
		})
	}
}

func TestSymbols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name: "forward reference",
			input: `
				JMP target
			target:
				RTS`,
			expected: []byte{0x4C, 0x03, 0x00, 0x60},
		},
		{
			name: "backward reference",
			input: `
			start:
				JMP start`,
			expected: []byte{0x4C, 0x00, 0x00},
		},
		{
			name: "zero page reference",
			input: `
			data: .byte $12
				  LDA data`,
			expected: []byte{0x12, 0xA5, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := NewAssembler()
			err := asm.Assemble(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, asm.output)
		})
	}
}
