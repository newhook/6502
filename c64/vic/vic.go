package vic

import (
	"fmt"
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
	TOTAL_LINES        = 312 // Total raster lines (PAL)
	FIRST_VISIBLE_LINE = 14
	LAST_VISIBLE_LINE  = 298

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

const ()

// Base address for VIC-II registers
const VICBase uint16 = 0xD000

// Sprite position registers
const (
	RegSprite0X = 0x00 // $D000
	RegSprite0Y = 0x01 // $D001
	RegSprite1X = 0x02 // $D002
	RegSprite1Y = 0x03 // $D003
	RegSprite2X = 0x04 // $D004
	RegSprite2Y = 0x05 // $D005
	RegSprite3X = 0x06 // $D006
	RegSprite3Y = 0x07 // $D007
	RegSprite4X = 0x08 // $D008
	RegSprite4Y = 0x09 // $D009
	RegSprite5X = 0x0A // $D00A
	RegSprite5Y = 0x0B // $D00B
	RegSprite6X = 0x0C // $D00C
	RegSprite6Y = 0x0D // $D00D
	RegSprite7X = 0x0E // $D00E
	RegSprite7Y = 0x0F // $D00F
)

// Sprite and screen control registers
const (
	RegSpriteXMSB     = 0x10 // $D010 - Sprite X MSB
	RegScreenControl1 = 0x11 // $D011
	RegRaster         = 0x12 // $D012
	RegLightPenX      = 0x13 // $D013
	RegLightPenY      = 0x14 // $D014
	RegSpriteEnable   = 0x15 // $D015
	RegScreenControl2 = 0x16 // $D016
	RegSpriteYExpand  = 0x17 // $D017
	RegMemPointers    = 0x18 // $D018
)

// Interrupt registers
const (
	RegInterrupt       = 0x19 // $D019
	RegInterruptEnable = 0x1A // $D01A
)

// Sprite control registers
const (
	RegSpritePriority    = 0x1B // $D01B
	RegSpriteMulticolor  = 0x1C // $D01C
	RegSpriteXExpand     = 0x1D // $D01D
	RegSpriteCollision   = 0x1E // $D01E
	RegSpriteBgCollision = 0x1F // $D01F
)

// Color registers
const (
	RegBorderColor  = 0x20 // $D020
	RegBgColor0     = 0x21 // $D021
	RegBgColor1     = 0x22 // $D022
	RegBgColor2     = 0x23 // $D023
	RegBgColor3     = 0x24 // $D024
	RegSpriteMulti0 = 0x25 // $D025
	RegSpriteMulti1 = 0x26 // $D026
	RegSprite0Color = 0x27 // $D027
	RegSprite1Color = 0x28 // $D028
	RegSprite2Color = 0x29 // $D029
	RegSprite3Color = 0x2A // $D02A
	RegSprite4Color = 0x2B // $D02B
	RegSprite5Color = 0x2C // $D02C
	RegSprite6Color = 0x2D // $D02D
	RegSprite7Color = 0x2E // $D02E
)

// Screen Control 1 (0xD011) bit masks
const (
	ScreenControl1Raster8 = 0x80 // Bit 7: Bit 8 of raster compare register
	ScreenControl1ECM     = 0x40 // Bit 6: Extended Color Mode
	ScreenControl1BMM     = 0x20 // Bit 5: Bitmap Mode
	ScreenControl1DEN     = 0x10 // Bit 4: Display Enable
	ScreenControl1RSEL    = 0x08 // Bit 3: Row Select (24/25 rows)
	ScreenControl1YSCROLL = 0x07 // Bits 2-0: Vertical Scroll
)

// Control Register 2 ($D016) bits
const (
	CTRL2_UNUSED  uint8 = 0xC0 // Bits 7-6: Unused
	CTRL2_RES     uint8 = 0x20 // Bit 5: Reset
	CTRL2_MCM     uint8 = 0x10 // Bit 4: Multicolor Mode
	CTRL2_CSEL    uint8 = 0x08 // Bit 3: Column Select (40/38 columns)
	CTRL2_XSCROLL uint8 = 0x07 // Bits 2-0: Horizontal Scroll
)

// XXX: ^^^ fix vvv

// Screen Control 2 (0xD016) bit masks
const (
	ScreenControl2Reset            = 0x20
	ScreenControl2MultiColor       = 0x10
	ScreenControl2Column40         = 0x08
	ScreenControl2HorizontalScroll = 0x07
)

// Memory pointer (0xD018) bit masks
const (
	MemPointersScreenMask  = 0xF0
	MemPointersCharMask    = 0x0E
	MemPointersScreenShift = 4
	MemPointersCharShift   = 1
)

// Interrupt (0xD019) bit masks
const (
	InterruptRaster       = 0x01
	InterruptSpriteBg     = 0x02
	InterruptSpriteSprite = 0x04
	InterruptLightPen     = 0x08
	InterruptIRQFlag      = 0x80
)

type DisplayMode uint8

const (
	MODE_STANDARD_TEXT DisplayMode = iota
	MODE_MULTICOLOR_TEXT
	MODE_STANDARD_BITMAP
	MODE_MULTICOLOR_BITMAP
	MODE_EXTENDED_TEXT
)

type Registers struct {
	sprites [8]Sprite
	//spriteDMAActive uint8
	spriteCollision   uint8
	spriteBgCollision uint8

	colors [15]uint8

	// Colors and display buffer
	backgroundColor [4]uint8
	borderColor     uint8

	sc1             uint8 // Screen control 1
	sc2             uint8 // Screen control 2
	interruptEnable uint8
	interrupt       uint8
	penX            uint8
	penY            uint8
	memPtr          uint8
}

type VIC struct {
	mem *memory.Manager

	// Raster beam position
	rasterCounter uint16 // y position (0 - 311) pal.
	rasterCycle   uint8  // x position (0 - 63).
	frameCount    uint64

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

	displayBuffer []uint8 // Frame buffer for rendering
	colorBuffer   []uint8 // Color data for current line

	// Interrupt state
	irqLine                bool
	rasterIRQ              uint16 // the raster line at which an interrupt should occur.
	irqStatus              uint8
	spritePriorityRegister uint8

	registers Registers
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
		displayBuffer: make([]uint8, 320*200),
		colorBuffer:   make([]uint8, VISIBLE_WIDTH),
		registers:     Registers{},
	}
}

// Update processes one VIC-II cycle
func (v *VIC) Update(cycle uint8) *VICEvent {
	v.rasterCycle += cycle

	// Check for bad line condition
	v.updateBadLine()

	//// Handle display generation
	if v.rasterCounter >= FIRST_VISIBLE_LINE && v.rasterCounter < LAST_VISIBLE_LINE {
		v.generateDisplayData()
	}

	// Handle sprite DMA and collision detection
	v.updateSprites()

	// Update raster position
	if v.rasterCycle >= CYCLES_PER_LINE {
		v.rasterCycle = 0
		v.rasterCounter++

		if v.rasterCounter >= TOTAL_LINES {
			v.rasterCounter = 0
			v.frameCount++
			return &VICEvent{Type: EventFrameComplete}
		}

		// Check for raster IRQ
		if v.rasterCounter == v.rasterIRQ && v.registers.interruptEnable&0x01 != 0 {
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
	if v.rasterCounter >= 0x30 && v.rasterCounter <= 0xf7 {
		if uint8(v.rasterCounter&0x07) == (v.registers.sc1 & ScreenControl1YSCROLL) {
			if v.registers.sc1&ScreenControl1DEN != 0 {
				v.badLine = true
				v.badLineEnable = true
				return
			}
		}
	}
	v.badLine = false
}

func (v *VIC) generateDisplayData() {
	rasterCounter := v.rasterCounter
	rasterCycle := v.rasterCycle

	// Only render during visible area
	if rasterCounter < 56 || rasterCounter > 255 || rasterCycle < 13 || rasterCycle >= 53 {
		return
	}

	// Calculate which character row and column we're rendering
	charRow := (rasterCounter - 56) / 8           // Which row of characters
	charCol := rasterCycle - 13                   // Which column in the current row
	charIndex := (charRow * 40) + uint16(charCol) // Character position in screen RAM

	// Calculate which line of the character we're drawing (0-7)
	charLine := (rasterCounter - 56) % 8

	// Get character from screen RAM (screen matrix)
	// Screen RAM location is determined by memory pointers register
	screenAddr := v.videoMatrix + charIndex
	//fmt.Printf("%x\n", screenAddr)
	char := v.mem.Read(screenAddr)

	// Get character color from color RAM ($D800-$DBFF)
	colorAddr := 0xD800 + uint16(charIndex)
	charColor := v.mem.Read(colorAddr)

	// Get character data from character ROM/RAM
	// Character memory location determined by memory pointers register
	// XXX: cia.
	//charDataAddr := v.charGen + (uint16(char) * 8) + uint16(charLine)
	//charData := v.mem.Read(charDataAddr)
	charData := v.mem.ReadChar(uint16(char)*8 + uint16(charLine))

	//if charRow == 0 && charCol == 0 {
	//	fmt.Printf("\n--------------\n")
	//} else if charCol == 0 {
	//	fmt.Printf("\n")
	//}
	////fmt.Printf("%02d/%02d/02x:", charRow, charCol, charData)
	////fmt.Printf("%02d", char)
	//if char == 0 {
	//	fmt.Printf(" ")
	//} else {
	//	fmt.Printf("%c", char)
	//}

	// Calculate where in display buffer to put the pixels
	bufferIndex := v.getCurrentPixelIndex(uint16(rasterCycle), rasterCounter)

	// Render all 8 pixels for this character line
	for bit := uint8(0); bit < 8; bit++ {
		pixel := (charData >> (7 - bit)) & 1
		if pixel == 1 {
			v.displayBuffer[bufferIndex+int(bit)] = charColor
		} else {
			v.displayBuffer[bufferIndex+int(bit)] = v.registers.colors[RegBgColor0-RegBorderColor] // Background color
		}
	}

	/*
		switch v.displayMode {
		case MODE_STANDARD_TEXT:
			v.generateTextMode(pixelIndex, charIndex, xPos, yPos)
		case MODE_MULTICOLOR_TEXT:
			v.generateMulticolorText(pixelIndex, charIndex, xPos, yPos)
		case MODE_STANDARD_BITMAP:
			v.generateBitmapMode(pixelIndex, charIndex, xPos, yPos)
		case MODE_MULTICOLOR_BITMAP:
			v.generateMulticolorBitmap(pixelIndex, charIndex, xPos, yPos)

	*/
}
func (v *VIC) getCurrentPixelIndex(rasterX uint16, rasterY uint16) int {
	// Only calculate for visible area
	if rasterY < 56 || rasterY > 255 {
		return -1
	}

	// Calculate Y position in pixels (relative to top of visible area)
	pixelY := (rasterY - 56) * 320

	// Convert rasterX cycle to pixel X
	// Visible area starts at cycle 13
	if rasterX < 13 || rasterX >= 53 { // 13 + 40 cycles = 53
		return -1
	}
	pixelX := (rasterX - 13) * 8

	return int(pixelY + pixelX)
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

	if int(pixelIndex) >= len(v.displayBuffer) {
		return
	}
	if pixel == 1 {
		v.displayBuffer[pixelIndex] = colorData
	} else {
		v.displayBuffer[pixelIndex] = v.registers.backgroundColor[0]
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

func (v *VIC) WriteRegister(reg uint8, value uint8) {
	// Registers $D020-$D02E can be written at any time
	if reg >= RegBorderColor && reg <= RegSprite7Color {
		v.registers.colors[reg-RegBorderColor] = value
		return
	}

	// Registers $D000-$D01F can only be written during VBlank or the screen area
	// In real hardware, writes outside these areas are ignored
	rasterX, rasterY := v.GetRasterPosition()
	if (rasterY < 51 || rasterY > 251) || rasterX < 58 {
		switch reg {
		// Sprite positions
		case RegSprite0X, RegSprite1X, RegSprite2X, RegSprite3X,
			RegSprite4X, RegSprite5X, RegSprite6X, RegSprite7X:
			spriteNum := reg >> 1
			v.registers.sprites[spriteNum].xPos = (v.registers.sprites[spriteNum].xPos & 0x100) | uint16(value)

		case RegSprite0Y, RegSprite1Y, RegSprite2Y, RegSprite3Y,
			RegSprite4Y, RegSprite5Y, RegSprite6Y, RegSprite7Y:
			spriteNum := (reg - 1) >> 1
			v.registers.sprites[spriteNum].yPos = value

		case RegSpriteXMSB:
			// Update MSB for all sprite X positions
			for i := uint8(0); i < 8; i++ {
				if value&(1<<i) != 0 {
					v.registers.sprites[i].xPos = v.registers.sprites[i].xPos | 0x100
				} else {
					v.registers.sprites[i].xPos = v.registers.sprites[i].xPos & 0xFF
				}
			}

		case RegScreenControl1:
			// Keep raster MSB in sync
			v.rasterIRQ &= 0xff
			v.rasterIRQ |= (uint16(value) & ScreenControl1Raster8) << 1
			v.registers.sc1 = value
			v.updateDisplayMode()
			v.updateVideoMatrix()

		// A write to the raster register (RegRaster, $D012) sets the raster line at which
		// a raster interrupt should occur. It works in conjunction with bit 7 of the Screen
		// Control Register 1 ($D011) since the raster line value can be from 0-311 (requiring 9 bits).
		case RegRaster:
			v.rasterIRQ = uint16(value) | ((uint16(v.registers.sc1 & ScreenControl1Raster8)) << 1)

		case RegScreenControl2:
			v.registers.sc2 = value
			v.updateDisplayMode()

		case RegMemPointers:
			v.registers.memPtr = value
			// Update screen and character memory pointers
			//v.screenMemPtr = uint16((value&MemPointersScreenMask)>>MemPointersScreenShift) << 10
			//v.charMemPtr = uint16((value&MemPointersCharMask)>>MemPointersCharShift) << 11

		case RegSpriteEnable:
			// Update enabled state for each sprite
			for i := uint8(0); i < 8; i++ {
				v.registers.sprites[i].enabled = (value & (1 << i)) != 0
			}

		case RegSpriteYExpand:
			for i := uint8(0); i < 8; i++ {
				v.registers.sprites[i].expandY = (value & (1 << i)) != 0
			}

		case RegSpritePriority:
			v.spritePriorityRegister = value

		case RegSpriteMulticolor:
			for i := uint8(0); i < 8; i++ {
				v.registers.sprites[i].multicolor = (value & (1 << i)) != 0
			}

		case RegSpriteXExpand:
			for i := uint8(0); i < 8; i++ {
				v.registers.sprites[i].expandX = (value & (1 << i)) != 0
			}

		case RegInterrupt:
			// Writing 1 to a bit clears the interrupt
			v.registers.interrupt &= ^value
			if v.registers.interrupt == 0 {
				// All interrupts cleared, lower IRQ line
				v.irqLine = false
			}

		case RegInterruptEnable:
			v.registers.interruptEnable = value
			// Check if any enabled interrupts are pending
			v.checkInterrupts()

		case RegSpriteCollision, RegSpriteBgCollision:
			// These registers are read-only
			return
		}
	} else {
		fmt.Println("write ignored")
	}
}

func (v *VIC) checkInterrupts() {
	pending := v.registers.interrupt & v.registers.interruptEnable
	if pending != 0 {
		v.irqLine = true
		v.registers.interrupt |= InterruptIRQFlag
	}
}

func (v *VIC) ReadRegister(reg uint8) uint8 {
	switch reg {
	case RegSprite0X, RegSprite1X, RegSprite2X, RegSprite3X,
		RegSprite4X, RegSprite5X, RegSprite6X, RegSprite7X:
		spriteNum := reg >> 1
		return uint8(v.registers.sprites[spriteNum].xPos & 0xFF)

	case RegSprite0Y, RegSprite1Y, RegSprite2Y, RegSprite3Y,
		RegSprite4Y, RegSprite5Y, RegSprite6Y, RegSprite7Y:
		spriteNum := (reg - 1) >> 1
		return v.registers.sprites[spriteNum].yPos

	case RegSpriteXMSB:
		var msb uint8
		for i := uint8(0); i < 8; i++ {
			if v.registers.sprites[i].xPos > 0xFF {
				msb |= 1 << i
			}
		}
		return msb

	case RegScreenControl1:
		// Ensure current raster line MSB is reflected in bit 7
		return (v.registers.sc1 & 0x7F) | uint8((v.rasterCounter&0x100)>>1)

	case RegRaster:
		// Return current raster line (lower 8 bits)
		return uint8(v.rasterCounter & 0xFF)

	// Light pen registers are latched when triggered
	case RegLightPenX:
		return v.registers.penX
	case RegLightPenY:
		return v.registers.penY

	case RegSpriteEnable:
		var enabled uint8
		for i := uint8(0); i < 8; i++ {
			if v.registers.sprites[i].enabled {
				enabled |= 1 << i
			}
		}
		return enabled

	case RegSpriteYExpand:
		var expand uint8
		for i := uint8(0); i < 8; i++ {
			if v.registers.sprites[i].expandY {
				expand |= 1 << i
			}
		}
		return expand

	case RegSpriteMulticolor:
		var multi uint8
		for i := uint8(0); i < 8; i++ {
			if v.registers.sprites[i].multicolor {
				multi |= 1 << i
			}
		}
		return multi

	case RegSpriteXExpand:
		var expand uint8
		for i := uint8(0); i < 8; i++ {
			if v.registers.sprites[i].expandX {
				expand |= 1 << i
			}
		}
		return expand

	case RegSpriteCollision:
		// Reading clears the register after returning its value
		value := v.registers.spriteCollision
		v.registers.spriteCollision = 0
		return value

	case RegSpriteBgCollision:
		// Reading clears the register after returning its value
		value := v.registers.spriteBgCollision
		v.registers.spriteBgCollision = 0
		return value

	case RegInterrupt:
		// Return current interrupt status
		return v.registers.interrupt

	case RegInterruptEnable:
		// Return current interrupt enable mask
		return v.registers.interruptEnable

	case RegBorderColor, RegBgColor0, RegBgColor1, RegBgColor2, RegBgColor3,
		RegSpriteMulti0, RegSpriteMulti1,
		RegSprite0Color, RegSprite1Color, RegSprite2Color, RegSprite3Color,
		RegSprite4Color, RegSprite5Color, RegSprite6Color, RegSprite7Color:
		// Color registers directly return their values
		return v.registers.colors[reg-RegBorderColor]

	default:
		// Handle unused registers ($D03F-$D3FF)
		// They return the last value on the data bus (we'll return 0xFF)
		if reg >= 0x3F {
			return 0xFF
		}
		// All other registers return their current value
		fmt.Println("read ignored", reg)
		//return v.registers[reg]
	}
	return 0
}

func (v *VIC) updateDisplayMode() {
	// Update display mode based on control registers
	ctrl1 := v.registers.sc1
	ctrl2 := v.registers.sc2

	v.displayActive = (ctrl1 & ScreenControl1DEN) != 0

	if ctrl1&ScreenControl1BMM != 0 {
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

//func (v *VIC) updateDisplayMode() {
//	control1 := v.registers.sc1
//	control2 := v.registers[RegScreenControl2]
//
//	v.bitmapMode = (control1 & ScreenControl1BitmapMode) != 0
//	v.extendedBgMode = (control1 & ScreenControl1ExtBgMode) != 0
//	v.displayEnabled = (control1 & ScreenControl1DisplayEnable) != 0
//	v.row25Mode = (control1 & ScreenControl1Row25) != 0
//
//	v.multicolorMode = (control2 & ScreenControl2MultiColor) != 0
//	v.column40Mode = (control2 & ScreenControl2Column40) != 0
//}

func (v *VIC) GetDisplayBuffer() []uint8 {
	return v.displayBuffer
}

func (v *VIC) IsBadLine() bool {
	return v.badLine
}

func (v *VIC) GetRasterPosition() (uint8, uint16) {
	return v.rasterCycle, v.rasterCounter
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
	memControl := v.registers.memPtr

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
	if v.registers.sc1&ScreenControl1BMM != 0 { // Bitmap mode
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

	v.videoMatrix = 0x400

	// Debug output
	v.logMemoryLayout()
}

// Helper function to output memory layout for debugging
func (v *VIC) logMemoryLayout() {
	mode := "text"
	if v.registers.sc1&ScreenControl1BMM != 0 {
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
	if v.registers.sc1&ScreenControl1BMM != 0 { // Bitmap mode
		// In bitmap mode, address is based on pixel position
		return v.bitmapBase + uint16(charCode)*8 + uint16(rowInChar)
	} else {
		// In text mode, address is based on character code
		return v.charGen + uint16(charCode)*8 + uint16(rowInChar)
	}
}
