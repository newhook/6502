package c64

import (
	"fmt"
	"github.com/newhook/6502/c64/cia"
	"github.com/veandco/go-sdl2/sdl"
	"log/slog"
)

// SDLKeyMapping maps SDL scancodes to C64 matrix positions
var SDLKeyMapping = map[sdl.Scancode]Key{
	sdl.SCANCODE_A: {1, 2, "A"},
	sdl.SCANCODE_B: {3, 4, "B"},
	sdl.SCANCODE_C: {2, 4, "C"},
	sdl.SCANCODE_D: {2, 2, "D"},
	sdl.SCANCODE_E: {1, 6, "E"},
	sdl.SCANCODE_F: {2, 5, "F"},
	sdl.SCANCODE_G: {3, 2, "G"},
	sdl.SCANCODE_H: {3, 5, "H"},
	sdl.SCANCODE_I: {4, 1, "I"},
	sdl.SCANCODE_J: {4, 2, "J"},
	sdl.SCANCODE_K: {4, 5, "K"},
	sdl.SCANCODE_L: {5, 2, "L"},
	sdl.SCANCODE_M: {4, 4, "M"},
	sdl.SCANCODE_N: {4, 7, "N"},
	sdl.SCANCODE_O: {4, 6, "O"},
	sdl.SCANCODE_P: {5, 1, "P"},
	sdl.SCANCODE_Q: {7, 6, "Q"},
	sdl.SCANCODE_R: {2, 1, "R"},
	sdl.SCANCODE_S: {1, 5, "S"},
	sdl.SCANCODE_T: {2, 6, "T"},
	sdl.SCANCODE_U: {3, 6, "U"},
	sdl.SCANCODE_V: {3, 7, "V"},
	sdl.SCANCODE_W: {1, 1, "W"},
	sdl.SCANCODE_X: {2, 7, "X"},
	sdl.SCANCODE_Y: {3, 1, "Y"},
	sdl.SCANCODE_Z: {1, 4, "Z"},

	// Numbers
	sdl.SCANCODE_1: {7, 0, "1"},
	sdl.SCANCODE_2: {7, 3, "2"},
	sdl.SCANCODE_3: {1, 0, "3"},
	sdl.SCANCODE_4: {1, 3, "4"},
	sdl.SCANCODE_5: {2, 0, "5"},
	sdl.SCANCODE_6: {2, 3, "6"},
	sdl.SCANCODE_7: {3, 0, "7"},
	sdl.SCANCODE_8: {3, 3, "8"},
	sdl.SCANCODE_9: {4, 0, "9"},
	sdl.SCANCODE_0: {4, 3, "0"},

	// Special keys
	sdl.SCANCODE_RETURN:    {0, 1, "RETURN"},
	sdl.SCANCODE_SPACE:     {7, 4, "SPACE"},
	sdl.SCANCODE_LSHIFT:    {1, 7, "LEFT SHIFT"},
	sdl.SCANCODE_RSHIFT:    {6, 4, "RIGHT SHIFT"},
	sdl.SCANCODE_LCTRL:     {7, 2, "CTRL"},
	sdl.SCANCODE_BACKSPACE: {0, 0, "DEL"},
	sdl.SCANCODE_HOME:      {6, 3, "CLR/HOME"},
	sdl.SCANCODE_INSERT:    {0, 0, "INST/DEL"}, // Same as backspace
}

// KeyMatrix represents the C64's 8x8 keyboard matrix
type KeyMatrix struct {
	// Matrix stores the current state of each key
	// A bit value of 1 means the key is pressed
	Matrix [8]byte
}

// NewKeyMatrix creates a new keyboard matrix
func NewKeyMatrix() *KeyMatrix {
	return &KeyMatrix{}
}

// KeyPress simulates pressing a key on the C64 keyboard
// row and col correspond to the C64's keyboard matrix positions
func (km *KeyMatrix) KeyPress(row, col int) {
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return
	}
	km.Matrix[row] |= (1 << col)
}

// KeyRelease simulates releasing a key
func (km *KeyMatrix) KeyRelease(row, col int) {
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return
	}
	km.Matrix[row] &^= (1 << col)
}

// ReadRow returns the state of a specific row in the matrix
func (km *KeyMatrix) ReadRow(row int) byte {
	if row < 0 || row > 7 {
		return 0
	}
	return km.Matrix[row]
}

// Key represents a C64 key with its matrix position
type Key struct {
	Row    int
	Col    int
	Symbol string
}

// Keyboard represents the complete C64 keyboard
type Keyboard struct {
	Matrix *KeyMatrix
	CIA    *cia.CIA
}

// NewKeyboard creates a new C64 keyboard
func NewKeyboard() *Keyboard {
	return &Keyboard{
		Matrix: NewKeyMatrix(),
	}
}

func (k *Keyboard) HandleSDLEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if mapping, exists := SDLKeyMapping[e.Keysym.Scancode]; exists {
			if e.Type == sdl.KEYDOWN {
				slog.Info(fmt.Sprintf("keydown %s %x %x", mapping.Symbol, mapping.Row, mapping.Col))
				k.Matrix.KeyPress(mapping.Row, mapping.Col)
			} else if e.Type == sdl.KEYUP {
				slog.Info(fmt.Sprintf("keyup %s %x %x", mapping.Symbol, mapping.Row, mapping.Col))
				k.Matrix.KeyRelease(mapping.Row, mapping.Col)
			}

			// Trigger a keyboard scan after state change
			k.ScanKeyboard()
		}
	}
}

// ScanKeyboard performs a keyboard matrix scan
// This simulates how the C64 ROM would scan the keyboard
func (k *Keyboard) ScanKeyboard() byte {
	return 0
	// The C64 sets a row low by writing to CIA1 Port A
	// Then reads the column states from CIA1 Port B

	// Get the current row selection from Port A
	// Inverted because 0 selects a row
	portA := k.CIA.ReadRegister(cia.PRA)
	slog.Info(fmt.Sprintf("port a %x\n", portA))
	rowSelect := ^k.CIA.ReadRegister(cia.PRA)
	slog.Info(fmt.Sprintln("row select", rowSelect))

	var result byte = 0xFF

	// Check each row that is selected (0 bit in rowSelect)
	for row := 0; row < 8; row++ {
		if rowSelect&(1<<row) != 0 {
			// This row is selected (low)
			// Get the key states for this row
			rowState := k.Matrix.ReadRow(row)

			// Combine with result
			// A 0 bit means a key is pressed
			result &= ^rowState
		}
	}

	slog.Info(fmt.Sprintln("write portb", result))

	// Update CIA Port B with the result
	k.CIA.WriteRegister(cia.PRB, result)

	return result
}

// WritePortA writes to CIA1 Port A
// This is called when the CPU writes to $DC00
func (k *Keyboard) WritePortA(value byte) {
	k.CIA.WriteRegister(cia.PRA, value)
	// Trigger a keyboard scan when Port A changes
	k.ScanKeyboard()
}

// ReadPortB reads from CIA1 Port B
// This is called when the CPU reads from $DC01
func (k *Keyboard) ReadPortB() byte {
	return ^k.CIA.ReadRegister(cia.PRB)
}

func (k *Keyboard) GetState(selectedRows uint8) uint8 {
	value := uint8(0xFF) // Pull-up resistors

	// For each selected row (0 bit in selectedRows)
	for row := uint8(0); row < 8; row++ {
		if selectedRows&(1<<row) == 0 {
			// Get the state of all columns for this row
			colState := k.Matrix.Matrix[row]
			// If any keys are pressed in this row (1 bits)
			// pull those columns low (0 bits in result)
			value &^= colState
		}
	}

	return value
}
