package c64

import (
	"errors"
	"fmt"
	"github.com/newhook/6502/c64/cia"
	"github.com/veandco/go-sdl2/sdl"
	"log/slog"
	"unicode"
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

	sdl.SCANCODE_COMMA:        {5, 7, "COMMA"},
	sdl.SCANCODE_PERIOD:       {5, 4, "PERIOD"},
	sdl.SCANCODE_SEMICOLON:    {6, 2, "SEMICOLON"},
	sdl.SCANCODE_SLASH:        {6, 7, "SLASH"},
	sdl.SCANCODE_EQUALS:       {6, 5, "EQUALS"},
	sdl.SCANCODE_MINUS:        {5, 3, "MINUS"},
	sdl.SCANCODE_APOSTROPHE:   {7, 3, "QUOTE"},
	sdl.SCANCODE_BACKSLASH:    {6, 6, "LEFT ARROW"},
	sdl.SCANCODE_LEFTBRACKET:  {5, 6, "AT"},       // @ symbol
	sdl.SCANCODE_RIGHTBRACKET: {6, 1, "ASTERISK"}, // *
	sdl.SCANCODE_GRAVE:        {7, 1, "POUND"},    // £ symbol
	sdl.SCANCODE_UP:           {6, 6, "UP ARROW"}, // ↑

	// Function keys
	sdl.SCANCODE_F1: {0, 4, "F1"},
	sdl.SCANCODE_F3: {0, 5, "F3"},
	sdl.SCANCODE_F5: {0, 6, "F5"},
	sdl.SCANCODE_F7: {0, 3, "F7"},

	// F2, F4, F6, F8 are SHIFT + F1, F3, F5, F7

	// Additional special keys
	sdl.SCANCODE_LALT:   {7, 5, "COMMODORE"}, // Using Left Alt as Commodore key
	sdl.SCANCODE_ESCAPE: {7, 7, "RUN/STOP"},  // Using ESC as RUN/STOP
	// Note: RESTORE key is typically handled specially as it's connected to the NMI line
	//sdl.SCANCODE_PLUS:  {5, 0, "PLUS"},
	//sdl.SCANCODE_COLON: {5, 5, "COLON"},
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

	pasting     bool
	lastTick    uint64
	pastebuffer string
	keydown     *Key
	shift       bool
}

// NewKeyboard creates a new C64 keyboard
func NewKeyboard() *Keyboard {
	return &Keyboard{
		Matrix: NewKeyMatrix(),
	}
}

func (k *Keyboard) HandleSDLEvent(event sdl.Event) {
	// Handle paste hotkey
	if e, ok := event.(*sdl.KeyboardEvent); ok && e.Keysym.Scancode == sdl.SCANCODE_F10 && e.Type == sdl.KEYDOWN {
		if err := k.PasteFromClipboard(); err != nil {
			slog.Error("Failed to paste from clipboard", "error", err)
		}
		return
	}

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
		}
	}
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

func (k *Keyboard) SimulateTextInput(text string) {
	for _, char := range text {
		// Find matching key mapping
		var key Key
		found := false

		// Convert character to uppercase since C64 keyboard is caps
		upperChar := unicode.ToUpper(char)

		// Search through mappings
		for _, mapping := range SDLKeyMapping {
			if mapping.Symbol == string(upperChar) {
				key = mapping
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Simulate key press
		k.Matrix.KeyPress(key.Row, key.Col)

		// Small delay to simulate typing
		sdl.Delay(50)

		// Release key
		k.Matrix.KeyRelease(key.Row, key.Col)
	}
}

const pasteText = `
10 v=53248:pokev+21,1:poke 2040,192:fort=12288to12350:poket,255:next
20 pokev+39,1
25 x=0:y=50
30 pokev,x:pokev+1,y
35 wait53265,128
38 x=x+1:ifx>255thenend
40 goto 30
`

func (k *Keyboard) PasteFromClipboard() error {
	k.pasting = true
	k.pastebuffer = pasteText
	return nil

	if !sdl.HasClipboardText() {
		return errors.New("clipboard is empty")
	}

	text, err := sdl.GetClipboardText()
	if err != nil {
		return fmt.Errorf("failed to get clipboard text: %v", err)
	}
	k.pasting = true
	k.pastebuffer = text
	return nil
}

const TickDelta = 5

func (k *Keyboard) Tick() {
	if !k.pasting {
		return
	}

	currentTick := sdl.GetTicks64()
	delta := currentTick - k.lastTick
	if delta < TickDelta {
		return
	}

	k.lastTick = currentTick

	if k.keydown != nil {
		k.Matrix.KeyRelease(k.keydown.Row, k.keydown.Col)
		if k.shift {
			k.Matrix.KeyRelease(1, 7) // left shift.
		}
		k.keydown = nil
		k.shift = false
	}

	if len(k.pastebuffer) == 0 {
		k.pasting = false
		return
	}

	char := rune(k.pastebuffer[0])
	k.pastebuffer = k.pastebuffer[1:]

	// Convert character to uppercase since C64 keyboard is caps
	upperChar := unicode.ToUpper(char)
	key, ok := ASCIIToKeyMap[upperChar]
	if !ok {
		k.pasting = false
		return
	}

	// Simulate key press
	k.keydown = &key
	k.Matrix.KeyPress(k.keydown.Row, k.keydown.Col)

	if ShiftKeyMap[upperChar] {
		k.Matrix.KeyPress(1, 7) // left shift.
		k.shift = true
	}
}

var ASCIIToKeyMap = map[rune]Key{
	'A':  {1, 2, "A"},
	'B':  {3, 4, "B"},
	'C':  {2, 4, "C"},
	'D':  {2, 2, "D"},
	'E':  {1, 6, "E"},
	'F':  {2, 5, "F"},
	'G':  {3, 2, "G"},
	'H':  {3, 5, "H"},
	'I':  {4, 1, "I"},
	'J':  {4, 2, "J"},
	'K':  {4, 5, "K"},
	'L':  {5, 2, "L"},
	'M':  {4, 4, "M"},
	'N':  {4, 7, "N"},
	'O':  {4, 6, "O"},
	'P':  {5, 1, "P"},
	'Q':  {7, 6, "Q"},
	'R':  {2, 1, "R"},
	'S':  {1, 5, "S"},
	'T':  {2, 6, "T"},
	'U':  {3, 6, "U"},
	'V':  {3, 7, "V"},
	'W':  {1, 1, "W"},
	'X':  {2, 7, "X"},
	'Y':  {3, 1, "Y"},
	'Z':  {1, 4, "Z"},
	'1':  {7, 0, "1"},
	'2':  {7, 3, "2"},
	'3':  {1, 0, "3"},
	'4':  {1, 3, "4"},
	'5':  {2, 0, "5"},
	'6':  {2, 3, "6"},
	'7':  {3, 0, "7"},
	'8':  {3, 3, "8"},
	'9':  {4, 0, "9"},
	'0':  {4, 3, "0"},
	' ':  {7, 4, "SPACE"},
	'\n': {0, 1, "RETURN"},
	',':  {5, 7, "COMMA"},
	'.':  {5, 4, "PERIOD"},
	';':  {6, 2, "SEMICOLON"},
	':':  {5, 5, "COLON"},
	'/':  {6, 7, "SLASH"},
	'=':  {6, 5, "EQUALS"},
	'-':  {5, 3, "MINUS"},
	'\'': {7, 3, "QUOTE"},
	'\\': {6, 6, "LEFT ARROW"},
	'@':  {5, 6, "AT"},
	'*':  {6, 1, "ASTERISK"},
	'£':  {7, 1, "POUND"},
	'+':  {5, 0, "PLUS"},
	'<':  {5, 7, "COMMA"},  // SHIFT + ,
	'>':  {5, 4, "PERIOD"}, // SHIFT + .
}

var ShiftKeyMap = map[rune]bool{
	'<': true,
	'>': true,
}
