package vic

import (
	"github.com/newhook/6502/c64/memory"
	"log"
)

const (
// Screen dimensions
)

const (
	// VIC-II Timing Constants
	SCREEN_WIDTH       = 403
	VISIBLE_WIDTH      = 320
	FIRST_DISPLAY_LINE = 51
	LAST_DISPLAY_LINE  = 251

	// Sprite Constants
	NUM_SPRITES       = 8
	SPRITE_WIDTH      = 24
	SPRITE_DMA_CYCLES = 2

	// Bad Line Constants
	FIRST_BA_LINE = 0x30
	LAST_BA_LINE  = 0xF7

	CYCLES_PER_LINE    = 63  // CPU cycles per line
	FIRST_VISIBLE_LINE = 51  // First visible raster line
	LAST_VISIBLE_LINE  = 251 // Last visible raster line
	TOTAL_LINES        = 312 // Total raster lines (PAL)

	// Border timing
	LEFT_BORDER_START  = 0
	LEFT_BORDER_END    = 24
	RIGHT_BORDER_START = LEFT_BORDER_END + VISIBLE_WIDTH

	// Memory locations
	SPRITE_POINTER_BASE = 0x07F8
	COLOR_RAM_BASE      = 0xD800
)

// VICEvent represents different VIC-II events
type VICEvent struct {
	Type EventType
	Data interface{}
}

type EventType int

const (
	EventFrameComplete EventType = iota
	EventRasterIRQ
)

// VIC-II registers
const (
	REG_SPRITE0_X     = 0x00
	REG_SPRITE0_Y     = 0x01
	REG_SPRITE_ENABLE = 0x15
	REG_CONTROL1      = 0x11
	REG_CONTROL2      = 0x16
	REG_RASTER        = 0x12
	REG_BACKGROUND    = 0x21
	REG_BORDER        = 0x20
)

const (
	// Control Register 1 ($D011) bits
	CTRL1_RASTER8 uint8 = 0x80 // Bit 7: Bit 8 of raster compare register
	CTRL1_ECM     uint8 = 0x40 // Bit 6: Extended Color Mode
	CTRL1_BMM     uint8 = 0x20 // Bit 5: Bitmap Mode
	CTRL1_DEN     uint8 = 0x10 // Bit 4: Display Enable
	CTRL1_RSEL    uint8 = 0x08 // Bit 3: Row Select (24/25 rows)
	CTRL1_YSCROLL uint8 = 0x07 // Bits 2-0: Vertical Scroll

	// Control Register 2 ($D016) bits
	CTRL2_UNUSED  uint8 = 0xC0 // Bits 7-6: Unused
	CTRL2_RES     uint8 = 0x20 // Bit 5: Reset
	CTRL2_MCM     uint8 = 0x10 // Bit 4: Multicolor Mode
	CTRL2_CSEL    uint8 = 0x08 // Bit 3: Column Select (40/38 columns)
	CTRL2_XSCROLL uint8 = 0x07 // Bits 2-0: Horizontal Scroll
)

type DisplayMode uint8

const (
	MODE_STANDARD_TEXT DisplayMode = iota
	MODE_MULTICOLOR_TEXT
	MODE_STANDARD_BITMAP
	MODE_MULTICOLOR_BITMAP
	MODE_EXTENDED_TEXT
)

type VIC struct {
	mem *memory.Manager

	// Raster beam position
	rasterX     uint16
	rasterY     uint16
	rasterCycle uint16
	frameCount  uint64

	// Display state
	displayMode   DisplayMode
	badLine       bool
	badLineEnable bool
	displayActive bool
	borderActive  bool

	// Display pointers
	videoMatrix uint16
	charGen     uint16
	bitmapBase  uint16

	// Sprite state
	sprites [8]Sprite
	//spriteDMAActive uint8

	// Colors and display buffer
	backgroundColor [4]uint8
	borderColor     uint8
	displayBuffer   []uint8 // Frame buffer for rendering
	colorBuffer     []uint8 // Color data for current line

	// Registers
	registers [47]uint8

	// Interrupt state
	irqLine    bool
	rasterIRQ  uint16
	irqStatus  uint8
	irqEnabled uint8
}

type Sprite struct {
	enabled    bool
	xPos       uint16
	yPos       uint8
	multicolor bool
	expandX    bool
	expandY    bool
	dmaCount   uint8
	dataPtr    uint16
}

func NewVIC(mem *memory.Manager) *VIC {
	return &VIC{
		mem:           mem,
		displayBuffer: make([]uint8, VISIBLE_WIDTH*(LAST_VISIBLE_LINE-FIRST_VISIBLE_LINE)),
		colorBuffer:   make([]uint8, VISIBLE_WIDTH),
		registers:     [47]uint8{},
		sprites:       [8]Sprite{},
	}
}

// Update processes one VIC-II cycle
func (v *VIC) Update(cycle uint8) *VICEvent {
	v.rasterCycle++

	// Check for bad line condition
	v.updateBadLine()

	// Handle display generation
	if v.displayActive {
		if v.rasterCycle >= LEFT_BORDER_END && v.rasterCycle < RIGHT_BORDER_START {
			v.generateDisplayData()
		}
	}

	// Handle sprite DMA and collision detection
	v.updateSprites()

	// Update raster position
	if v.rasterCycle >= CYCLES_PER_LINE {
		v.rasterCycle = 0
		v.rasterY++

		if v.rasterY >= TOTAL_LINES {
			v.rasterY = 0
			v.frameCount++
			return &VICEvent{Type: EventFrameComplete}
		}

		// Check for raster IRQ
		if v.rasterY == v.rasterIRQ && v.irqEnabled&0x01 != 0 {
			v.irqStatus |= 0x01
			return &VICEvent{Type: EventRasterIRQ}
		}
	}

	return nil
}

func (v *VIC) updateBadLine() {
	// Bad line condition:
	// 1. Current raster line is between 0x30-0xf7
	// 2. Lower 3 bits of raster line match lower 3 bits of scroll register
	// 3. Display enable bit is set
	if v.rasterY >= 0x30 && v.rasterY <= 0xf7 {
		if uint8(v.rasterY&0x07) == (v.registers[REG_CONTROL1] & CTRL1_YSCROLL) {
			if v.registers[REG_CONTROL1]&CTRL1_DEN != 0 {
				v.badLine = true
				v.badLineEnable = true
				return
			}
		}
	}
	v.badLine = false
}

func (v *VIC) generateDisplayData() {
	pixelIndex := (v.rasterY-FIRST_VISIBLE_LINE)*VISIBLE_WIDTH +
		(v.rasterCycle - LEFT_BORDER_END)

	if v.borderActive {
		v.displayBuffer[pixelIndex] = v.borderColor
		return
	}

	// Calculate character/pixel position
	xPos := v.rasterCycle - LEFT_BORDER_END
	yPos := v.rasterY - FIRST_VISIBLE_LINE
	charCol := xPos / 8
	charRow := yPos / 8
	charIndex := charRow*40 + charCol

	switch v.displayMode {
	case MODE_STANDARD_TEXT:
		v.generateTextMode(pixelIndex, charIndex, xPos, yPos)
	case MODE_MULTICOLOR_TEXT:
		v.generateMulticolorText(pixelIndex, charIndex, xPos, yPos)
	case MODE_STANDARD_BITMAP:
		v.generateBitmapMode(pixelIndex, charIndex, xPos, yPos)
	case MODE_MULTICOLOR_BITMAP:
		v.generateMulticolorBitmap(pixelIndex, charIndex, xPos, yPos)
	}
}

func (v *VIC) generateTextMode(pixelIndex uint16, charIndex uint16, xPos uint16, yPos uint16) {
	// Get character from video matrix
	charPtr := v.videoMatrix + charIndex
	char := v.mem.Read(charPtr)

	// Get character data from character generator
	charDataPtr := v.charGen + uint16(char)*8 + (yPos % 8)
	charData := v.mem.Read(charDataPtr)

	// Get color data
	colorData := v.mem.Read(COLOR_RAM_BASE + charIndex)

	// Calculate pixel
	bitPos := 7 - (xPos % 8)
	pixel := (charData >> bitPos) & 0x01

	if pixel == 1 {
		v.displayBuffer[pixelIndex] = colorData
	} else {
		v.displayBuffer[pixelIndex] = v.backgroundColor[0]
	}
}

func (v *VIC) generateMulticolorText(pixelIndex uint16, charIndex uint16, xPos uint16, yPos uint16) {
	// Similar to standard text mode but handles multicolor mode
	// Implementation here
}

func (v *VIC) generateBitmapMode(pixelIndex uint16, charIndex uint16, xPos uint16, yPos uint16) {
	// Implementation for standard bitmap mode
}

func (v *VIC) generateMulticolorBitmap(pixelIndex uint16, charIndex uint16, xPos uint16, yPos uint16) {
	// Implementation for multicolor bitmap mode
}

func (v *VIC) updateSprites() {
	// Check sprite-sprite and sprite-background collisions
	// Handle sprite DMA
	// Update sprite positions and data
}

// Register access methods
func (v *VIC) WriteRegister(reg uint8, value uint8) {
	v.registers[reg] = value

	switch reg {
	case REG_CONTROL1:
		v.updateDisplayMode()
		v.updateVideoMatrix()
	case REG_CONTROL2:
		v.updateDisplayMode()
	case REG_RASTER:
		v.rasterIRQ = uint16(value) | ((uint16(v.registers[REG_CONTROL1] & CTRL1_RASTER8)) << 1)
	}
}

func (v *VIC) ReadRegister(reg uint8) uint8 {
	switch reg {
	case REG_RASTER:
		return uint8(v.rasterY & 0xFF)
	default:
		return v.registers[reg]
	}
}

func (v *VIC) updateDisplayMode() {
	// Update display mode based on control registers
	ctrl1 := v.registers[REG_CONTROL1]
	ctrl2 := v.registers[REG_CONTROL2]

	if ctrl1&CTRL1_BMM != 0 {
		// Bitmap mode
		if ctrl2&CTRL2_MCM != 0 {
			v.displayMode = MODE_MULTICOLOR_BITMAP
		} else {
			v.displayMode = MODE_STANDARD_BITMAP
		}
	} else {
		// Text mode
		if ctrl2&CTRL2_MCM != 0 {
			v.displayMode = MODE_MULTICOLOR_TEXT
		} else {
			v.displayMode = MODE_STANDARD_TEXT
		}
	}
}

func (v *VIC) GetDisplayBuffer() []uint8 {
	return v.displayBuffer
}

func (v *VIC) IsBadLine() bool {
	return v.badLine
}

func (v *VIC) GetRasterPosition() (uint16, uint16) {
	return v.rasterX, v.rasterY
}

// Memory bank selection bits in CIA2 Port A (0xDD00)
const (
	BANK_0 = 0x03 // Bank 0: 0x0000-0x3FFF
	BANK_1 = 0x02 // Bank 1: 0x4000-0x7FFF
	BANK_2 = 0x01 // Bank 2: 0x8000-0xBFFF
	BANK_3 = 0x00 // Bank 3: 0xC000-0xFFFF
)

// Video matrix and character generator base addresses within selected bank
const (
	VIDEO_MATRIX_SIZE = 0x0400 // 1K video matrix
	CHAR_ROM_SIZE     = 0x1000 // 4K character ROM
)

func (v *VIC) updateVideoMatrix() {
	// Get memory control register ($D018)
	memControl := v.registers[0x18]

	// Get bank selection from CIA2 Port A (top 2 bits)
	bankSelect := v.mem.Read(0xDD00) & 0x03
	bankBase := uint16(^bankSelect&0x03) << 14 // Convert to actual base address

	// Video Matrix Base Address (VM13-VM10)
	// Bits 4-7 of $D018 specify video matrix base address within selected bank
	videoBase := uint16(memControl&0xF0) << 6
	v.videoMatrix = bankBase | videoBase

	// Character Generator/Bitmap Base
	// Bit 3 of $D018 selects bitmap base in bitmap modes
	// Bits 1-2 select character generator base in text modes
	if v.registers[0x11]&0x20 != 0 { // Bitmap mode
		v.bitmapBase = bankBase
		if memControl&0x08 != 0 {
			v.bitmapBase |= 0x2000 // Set to 8192 if bit 3 is set
		}
	} else { // Text mode
		// Character base is either ROM or RAM depending on bank
		if bankBase >= 0xC000 {
			// Use character ROM when in bank 3
			v.charGen = 0xD000
		} else {
			// In RAM banks, use specified base address
			charBase := uint16(memControl&0x0E) << 10
			v.charGen = bankBase | charBase
		}
	}

	// Debug output
	v.logMemoryLayout()
}

// Helper function to output memory layout for debugging
func (v *VIC) logMemoryLayout() {
	mode := "text"
	if v.registers[0x11]&0x20 != 0 {
		mode = "bitmap"
	}

	log.Printf("VIC-II Memory Layout:")
	log.Printf("Mode: %s", mode)
	log.Printf("Bank Base: $%04X", v.videoMatrix&0xC000)
	log.Printf("Video Matrix: $%04X", v.videoMatrix)
	if mode == "text" {
		log.Printf("Character Data: $%04X", v.charGen)
	} else {
		log.Printf("Bitmap Base: $%04X", v.bitmapBase)
	}
}

// Helper method to get current video matrix pointer
func (v *VIC) getCurrentVideoAddress(charPos uint16) uint16 {
	return v.videoMatrix + charPos
}

// Helper method to get current character/bitmap data pointer
func (v *VIC) getCurrentCharacterAddress(charCode uint8, rowInChar uint8) uint16 {
	if v.registers[0x11]&0x20 != 0 { // Bitmap mode
		// In bitmap mode, address is based on pixel position
		return v.bitmapBase + uint16(charCode)*8 + uint16(rowInChar)
	} else {
		// In text mode, address is based on character code
		return v.charGen + uint16(charCode)*8 + uint16(rowInChar)
	}
}
