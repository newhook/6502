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
	CTRL1_Raster8 = 0x80 // Bit 7: Bit 8 of raster compare register
	CTRL1_ECM     = 0x40 // Bit 6: Extended Color Mode
	CTRL1_BMM     = 0x20 // Bit 5: Bitmap Mode
	CTRL1_DEN     = 0x10 // Bit 4: Display Enable
	CTRL1_RSEL    = 0x08 // Bit 3: Row Select (24/25 rows)
	CTRL1_YSCROLL = 0x07 // Bits 2-0: Vertical Scroll
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

func NewVIC(mem *memory.Manager) *VIC {
	vic := &VIC{
		mem:           mem,
		displayBuffer: make([]uint8, 320*200),
		colorBuffer:   make([]uint8, VISIBLE_WIDTH),
	}
	vic.registers[RegBgColor0] = 0x0E       // Background color 0 (light blue)
	vic.registers[RegBorderColor] = 0x0E    // Border color (light blue)
	vic.registers[RegScreenControl1] = 0x1B // Default: Screen on, 25 rows, Y scroll = 3
	vic.registers[RegScreenControl2] = 0x08 // Default: No multicolor, 40 columns, X scroll = 0
	vic.registers[RegMemPointers] = 0x17    // Default memory layout
	return vic
}

const NUM_REGISTERS = 0x2F

// Sprite data structure to track sprite state
type Sprite struct {
	x          uint16 // X position (including MSB)
	y          uint8  // Y position
	enabled    bool   // Sprite enabled status
	multicolor bool   // Multicolor mode
	xExpand    bool   // X expansion
	yExpand    bool   // Y expansion
	priority   bool   // Sprite-to-background priority
	dataPtr    uint8  // Pointer to sprite data
}

type VIC struct {
	mem *memory.Manager

	registers [NUM_REGISTERS]uint8

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

	sprites                   [NUM_SPRITES]Sprite
	spriteSpriteCollision     uint8
	spriteBackgroundCollision uint8
}

// Update processes one VIC-II cycle
func (v *VIC) Update(cycle uint8) *VICEvent {
	v.rasterCycle += cycle

	// Check for bad line condition
	v.updateBadLine()

	// Handle display generation
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
		if v.rasterCounter == v.rasterIRQ && v.registers[RegInterruptEnable]&0x01 != 0 {
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
		if uint8(v.rasterCounter&0x07) == (v.registers[RegScreenControl1] & CTRL1_YSCROLL) {
			if v.registers[RegScreenControl1]&CTRL1_DEN != 0 {
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
	charData := v.mem.ReadChar(uint16(char)*8 + charLine)

	// Calculate where in display buffer to put the pixels
	bufferIndex := v.getCurrentPixelIndex(uint16(rasterCycle), rasterCounter)

	// Render all 8 pixels for this character line
	for bit := uint8(0); bit < 8; bit++ {
		pixel := (charData >> (7 - bit)) & 1
		if pixel == 1 {
			v.displayBuffer[bufferIndex+int(bit)] = charColor
		} else {
			v.displayBuffer[bufferIndex+int(bit)] = v.registers[RegBgColor0] // Background color
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
	v.renderSprites()
}

// Add these methods to your VIC struct implementation

// renderSprites renders all enabled sprites for the current raster line
func (v *VIC) renderSprites() {
	// Only render during visible area
	if v.rasterCounter < FIRST_VISIBLE_LINE || v.rasterCounter >= LAST_VISIBLE_LINE {
		return
	}

	// Process sprites in reverse order (sprite 7 first, as it has lowest priority)
	for i := NUM_SPRITES - 1; i >= 0; i-- {
		if v.sprites[i].enabled {
			v.renderSprite(uint8(i))
		}
	}
}

// renderSprite renders a single sprite for the current raster line
func (v *VIC) renderSprite(spriteNum uint8) {
	sprite := &v.sprites[spriteNum]

	// Check if sprite is visible on current raster line
	spriteY := int16(sprite.y)
	currentY := int16(v.rasterCounter - FIRST_VISIBLE_LINE)
	spriteHeight := int16(21)
	if sprite.yExpand {
		spriteHeight *= 2
	}

	// Skip if sprite not visible on this line
	if currentY < int16(spriteY) || currentY >= int16(spriteY)+spriteHeight {
		return
	}

	// Calculate which row of the sprite we're rendering
	spriteRow := currentY - int16(spriteY)
	if sprite.yExpand {
		spriteRow /= 2
	}

	// Get sprite data pointer
	spriteDataPtr := uint16(sprite.dataPtr) * 64
	rowOffset := uint16(spriteRow) * 3 // 3 bytes per row

	// Read sprite data for current row (3 bytes = 24 bits)
	data1 := v.mem.Read(spriteDataPtr + rowOffset)
	data2 := v.mem.Read(spriteDataPtr + rowOffset + 1)
	data3 := v.mem.Read(spriteDataPtr + rowOffset + 2)

	// Calculate x position in screen space
	screenX := int16(sprite.x) - 24 // Adjust for sprite border offset

	// Get sprite colors
	spriteColor := v.registers[RegSprite0Color+spriteNum]
	multicolor0 := v.registers[RegSpriteMulti0]
	multicolor1 := v.registers[RegSpriteMulti1]

	// Render the sprite pixels
	if sprite.multicolor {
		v.renderMulticolorSprite(screenX, currentY, data1, data2, data3,
			spriteColor, multicolor0, multicolor1, sprite.xExpand, sprite.priority)
	} else {
		v.renderStandardSprite(screenX, currentY, data1, data2, data3,
			spriteColor, sprite.xExpand, sprite.priority)
	}
}

// renderStandardSprite renders a standard (high-resolution) sprite
func (v *VIC) renderStandardSprite(x int16, y int16, data1, data2, data3 uint8,
	color uint8, xExpand bool, priority bool) {

	// Convert the three data bytes into 24 bits
	bits := uint32(data1)<<16 | uint32(data2)<<8 | uint32(data3)

	// Calculate pixel index in display buffer
	bufferOffset := int(y) * VISIBLE_WIDTH

	// Render all 24 bits
	for bit := uint8(0); bit < 24; bit++ {
		if bits&(1<<(23-bit)) != 0 {
			pixelX := x + int16(bit)
			if xExpand {
				// In x-expanded mode, each bit is rendered twice
				pixelX *= 2
				v.plotSpritePixel(bufferOffset, int(pixelX), color, priority)
				v.plotSpritePixel(bufferOffset, int(pixelX+1), color, priority)
			} else {
				v.plotSpritePixel(bufferOffset, int(pixelX), color, priority)
			}
		}
	}
}

// renderMulticolorSprite renders a multicolor sprite
func (v *VIC) renderMulticolorSprite(x int16, y int16, data1, data2, data3 uint8,
	spriteColor, multicolor0, multicolor1 uint8, xExpand bool, priority bool) {

	// Convert the three data bytes into 24 bits
	bits := uint32(data1)<<16 | uint32(data2)<<8 | uint32(data3)

	// Calculate pixel index in display buffer
	bufferOffset := int(y) * VISIBLE_WIDTH

	// In multicolor mode, bits are processed in pairs
	for bitPair := uint8(0); bitPair < 12; bitPair++ {
		// Extract 2 bits
		pixelBits := (bits >> (22 - (bitPair * 2))) & 0x3

		// Determine color based on bit pair
		var color uint8
		switch pixelBits {
		case 0:
			continue // Transparent
		case 1:
			color = multicolor0
		case 2:
			color = spriteColor
		case 3:
			color = multicolor1
		}

		// Calculate x position (each pixel is twice as wide in multicolor mode)
		pixelX := x + int16(bitPair*2)

		if xExpand {
			// In x-expanded mode, each multicolor pixel is 4 pixels wide
			pixelX *= 2
			for i := 0; i < 4; i++ {
				v.plotSpritePixel(bufferOffset, int(pixelX)+i, color, priority)
			}
		} else {
			// Normal multicolor pixel is 2 pixels wide
			v.plotSpritePixel(bufferOffset, int(pixelX), color, priority)
			v.plotSpritePixel(bufferOffset, int(pixelX+1), color, priority)
		}
	}
}

// plotSpritePixel plots a single sprite pixel to the display buffer
func (v *VIC) plotSpritePixel(bufferOffset, x int, color uint8, priority bool) {
	// Check if pixel is within visible screen area
	if x < 0 || x >= VISIBLE_WIDTH {
		return
	}

	pixelIndex := bufferOffset + x

	// Check array bounds
	if pixelIndex < 0 || pixelIndex >= len(v.displayBuffer) {
		return
	}

	// Handle sprite-background priority
	if priority {
		// Sprite appears behind background
		if v.displayBuffer[pixelIndex] == v.registers[RegBgColor0] {
			v.displayBuffer[pixelIndex] = color
		}
	} else {
		// Sprite appears in front of background
		v.displayBuffer[pixelIndex] = color
	}
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

//func (v *VIC) generateTextMode(pixelIndex uint16, charIndex uint16, xPos uint16, yPos uint16) {
//	// Get character from video matrix
//	charPtr := v.videoMatrix + charIndex
//	char := v.mem.Read(charPtr)
//
//	// Get character data from character generator
//	charDataPtr := v.charGen + uint16(char)*8 + (yPos % 8)
//	charData := v.mem.Read(charDataPtr)
//
//	// Get color data
//	colorData := v.mem.Read(COLOR_RAM_BASE + charIndex)
//
//	// Calculate pixel
//	bitPos := 7 - (xPos % 8)
//	pixel := (charData >> bitPos) & 0x01
//
//	if int(pixelIndex) >= len(v.displayBuffer) {
//		return
//	}
//	if pixel == 1 {
//		v.displayBuffer[pixelIndex] = colorData
//	} else {
//		v.displayBuffer[pixelIndex] = v.registers.backgroundColor[0]
//	}
//}

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
	// Process sprite DMA cycles
	if v.rasterCycle >= 15 && v.rasterCycle <= 54 {
		// Sprite data fetch happens during visible screen area
		v.fetchSpriteData()
	}

	// Update sprite positions and check collisions
	v.updateSpritePositions()
	v.checkSpriteCollisions()
}

func (v *VIC) fetchSpriteData() {
	// Each sprite needs 2 cycles for DMA
	spriteIndex := (v.rasterCycle - 15) / SPRITE_DMA_CYCLES

	if spriteIndex < NUM_SPRITES && v.sprites[spriteIndex].enabled {
		// Get sprite data pointer from $07F8-$07FF
		basePtr := SPRITE_POINTER_BASE + uint16(spriteIndex)
		v.sprites[spriteIndex].dataPtr = v.mem.Read(basePtr)
	}
}

func (v *VIC) updateSpritePositions() {
	for i := uint8(0); i < NUM_SPRITES; i++ {
		// Update X position (including MSB from $D010)
		xLow := v.registers[RegSprite0X+(i*2)]
		xMsb := (v.registers[RegSpriteXMSB] >> i) & 1
		v.sprites[i].x = uint16(xLow) | (uint16(xMsb) << 8)

		// Update Y position
		v.sprites[i].y = v.registers[RegSprite0Y+(i*2)]

		// Update sprite attributes
		v.sprites[i].enabled = (v.registers[RegSpriteEnable]>>i)&1 == 1
		v.sprites[i].multicolor = (v.registers[RegSpriteMulticolor]>>i)&1 == 1
		v.sprites[i].xExpand = (v.registers[RegSpriteXExpand]>>i)&1 == 1
		v.sprites[i].yExpand = (v.registers[RegSpriteYExpand]>>i)&1 == 1
		v.sprites[i].priority = (v.registers[RegSpritePriority]>>i)&1 == 1
	}
}

func (v *VIC) checkSpriteCollisions() {
	// Reset collision registers if they were just read
	if v.registers[RegSpriteCollision] == 0 {
		v.spriteSpriteCollision = 0
	}
	if v.registers[RegSpriteBgCollision] == 0 {
		v.spriteBackgroundCollision = 0
	}

	// Check sprite-sprite collisions
	for i := uint8(0); i < NUM_SPRITES-1; i++ {
		if !v.sprites[i].enabled {
			continue
		}

		for j := i + 1; j < NUM_SPRITES; j++ {
			if !v.sprites[j].enabled {
				continue
			}

			if v.spriteOverlaps(i, j) {
				// Set collision bits for both sprites
				v.spriteSpriteCollision |= (1 << i) | (1 << j)
				// Set interrupt if enabled
				if v.registers[RegInterruptEnable]&InterruptSpriteSprite != 0 {
					v.registers[RegInterrupt] |= InterruptSpriteSprite
					v.irqLine = true
				}
			}
		}
	}

	// Check sprite-background collisions
	for i := uint8(0); i < NUM_SPRITES; i++ {
		if !v.sprites[i].enabled {
			continue
		}

		if v.spriteIntersectsBackground(i) {
			v.spriteBackgroundCollision |= (1 << i)
			// Set interrupt if enabled
			if v.registers[RegInterruptEnable]&InterruptSpriteBg != 0 {
				v.registers[RegInterrupt] |= InterruptSpriteBg
				v.irqLine = true
			}
		}
	}
}

func (v *VIC) spriteOverlaps(s1, s2 uint8) bool {
	// Get sprite dimensions (account for expansion)
	s1Width := SPRITE_WIDTH
	if v.sprites[s1].xExpand {
		s1Width *= 2
	}
	s1Height := uint8(21)
	if v.sprites[s1].yExpand {
		s1Height *= 2
	}

	s2Width := SPRITE_WIDTH
	if v.sprites[s2].xExpand {
		s2Width *= 2
	}
	s2Height := uint8(21)
	if v.sprites[s2].yExpand {
		s2Height *= 2
	}

	// Check for overlap
	if v.sprites[s1].x >= v.sprites[s2].x+uint16(s2Width) ||
		v.sprites[s2].x >= v.sprites[s1].x+uint16(s1Width) ||
		v.sprites[s1].y >= v.sprites[s2].y+s2Height ||
		v.sprites[s2].y >= v.sprites[s1].y+s1Height {
		return false
	}

	return true
}

func (v *VIC) spriteIntersectsBackground(spriteNum uint8) bool {
	// Only check if sprite is in visible area
	if v.sprites[spriteNum].y < 30 ||
		v.sprites[spriteNum].y > 249 ||
		v.sprites[spriteNum].x < uint16(LEFT_BORDER_END) ||
		v.sprites[spriteNum].x > uint16(RIGHT_BORDER_START) {
		return false
	}

	// Get sprite data
	spriteDataPtr := uint16(v.sprites[spriteNum].dataPtr) * 64

	// Check each row of the sprite
	height := uint8(21)
	if v.sprites[spriteNum].yExpand {
		height *= 2
	}

	for row := uint8(0); row < height; row++ {
		data := v.mem.Read(spriteDataPtr + uint16(row))

		// If any sprite pixel overlaps with non-background pixel
		if data != 0 {
			// This is a simplified check - in reality you'd need to check
			// pixel by pixel against the actual background content
			return true
		}
	}

	return false
}

func (v *VIC) WriteRegister(reg uint8, value uint8) {
	// Registers $D020-$D02E can be written at any time
	if reg >= RegBorderColor && reg <= RegSprite7Color {
		v.registers[reg] = value
		return
	}

	// Registers $D000-$D01F can only be written during VBlank or the screen area
	rasterX, rasterY := v.GetRasterPosition()
	if (rasterY < 51 || rasterY > 251) || rasterX < 58 {
		switch reg {
		case RegSprite0X, RegSprite1X, RegSprite2X, RegSprite3X,
			RegSprite4X, RegSprite5X, RegSprite6X, RegSprite7X:
			v.registers[reg] = value
			// Update sprite X position
			spriteNum := reg >> 1
			xpos := uint16(value)
			if v.registers[RegSpriteXMSB]&(1<<spriteNum) != 0 {
				xpos |= 0x100
			}
			v.sprites[spriteNum].x = xpos

		case RegSprite0Y, RegSprite1Y, RegSprite2Y, RegSprite3Y,
			RegSprite4Y, RegSprite5Y, RegSprite6Y, RegSprite7Y:
			v.registers[reg] = value
			spriteNum := (reg - 1) >> 1
			v.sprites[spriteNum].y = value

		case RegSpriteXMSB:
			v.registers[reg] = value
			// Update all sprite X positions to account for MSB changes
			for i := uint8(0); i < NUM_SPRITES; i++ {
				xpos := uint16(v.registers[RegSprite0X+i*2])
				if value&(1<<i) != 0 {
					xpos |= 0x100
				}
				v.sprites[i].x = xpos
			}

		case RegSpriteEnable:
			v.registers[reg] = value
			for i := uint8(0); i < NUM_SPRITES; i++ {
				v.sprites[i].enabled = (value & (1 << i)) != 0
			}

			// Sprite Y-expansion ($D017)
		case RegSpriteYExpand:
			v.registers[reg] = value
			for i := uint8(0); i < NUM_SPRITES; i++ {
				v.sprites[i].yExpand = (value & (1 << i)) != 0
			}

			// Sprite priority ($D01B)
		case RegSpritePriority:
			v.registers[reg] = value
			for i := uint8(0); i < NUM_SPRITES; i++ {
				v.sprites[i].priority = (value & (1 << i)) != 0
			}

			// Sprite multicolor ($D01C)
		case RegSpriteMulticolor:
			v.registers[reg] = value
			for i := uint8(0); i < NUM_SPRITES; i++ {
				v.sprites[i].multicolor = (value & (1 << i)) != 0
			}

		// Sprite X-expansion ($D01D)
		case RegSpriteXExpand:
			v.registers[reg] = value
			for i := uint8(0); i < NUM_SPRITES; i++ {
				v.sprites[i].xExpand = (value & (1 << i)) != 0
			}

		// Read-only collision registers
		case RegSpriteCollision, RegSpriteBgCollision:
			return

		case RegScreenControl1:
			// Keep raster MSB in sync
			v.rasterIRQ &= 0xff
			v.rasterIRQ |= (uint16(value) & CTRL1_Raster8) << 1
			v.registers[reg] = value
			v.updateDisplayMode()
			v.updateVideoMatrix()

		case RegRaster:
			v.registers[reg] = value
			v.rasterIRQ = uint16(value) | ((uint16(v.registers[RegScreenControl1] & CTRL1_Raster8)) << 1)

		case RegInterrupt:
			// Writing 1 to a bit clears the interrupt
			v.registers[reg] &= ^value
			if v.registers[reg] == 0 {
				v.irqLine = false
			}

		case RegInterruptEnable:
			v.registers[reg] = value
			v.checkInterrupts()

		case RegMemPointers:
			v.registers[reg] = value
			v.updateVideoMatrix()

		case RegScreenControl2:
			v.registers[reg] = value
			v.updateDisplayMode()
		}
	}
}

func (v *VIC) ReadRegister(reg uint8) uint8 {
	switch reg {
	case RegScreenControl1:
		// Ensure current raster line MSB is reflected in bit 7
		return (v.registers[reg] & 0x7F) | uint8((v.rasterCounter&0x100)>>1)

	case RegRaster:
		// Return current raster line (lower 8 bits)
		return uint8(v.rasterCounter & 0xFF)

	case RegSpriteCollision, RegSpriteBgCollision:
		// Reading clears the register after returning its value
		value := v.registers[reg]
		v.registers[reg] = 0
		return value

	default:
		if reg >= NUM_REGISTERS {
			return 0xFF // Unused registers return last value on data bus
		}
		return v.registers[reg]
	}
}

func (v *VIC) updateDisplayMode() {
	// Update display mode based on control registers
	ctrl1 := v.registers[RegScreenControl1]
	ctrl2 := v.registers[RegScreenControl2]

	v.displayActive = (ctrl1 & CTRL1_DEN) != 0

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

func (v *VIC) checkInterrupts() {
	pending := v.registers[RegInterrupt] & v.registers[RegInterruptEnable]
	if pending != 0 {
		v.irqLine = true
		v.registers[RegInterrupt] |= InterruptIRQFlag
	}
}

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
	memControl := v.registers[RegMemPointers]

	// Get bank selection from CIA2 Port A (top 2 bits)
	// Bank 0: $0000-$3FFF
	// Bank 1: $4000-$7FFF
	// Bank 2: $8000-$BFFF
	// Bank 3: $C000-$FFFF
	bankSelect := v.mem.Read(0xDD00) & 0x03
	bankBase := uint16(^bankSelect&0x03) << 14 // Convert to actual base address

	// Video Matrix Base Address (VM13-VM10)
	// Bits 4-7 of $D018 specify video matrix base address within selected bank
	// bits xxxx----
	videoBase := uint16(memControl&0xF0) << 6
	v.videoMatrix = bankBase | videoBase

	// Character Generator/Bitmap Base
	// Bits 1-2 select character generator base in text modes
	if v.registers[RegScreenControl1]&CTRL1_BMM != 0 { // Bitmap mode
		v.bitmapBase = bankBase
		// bit  ----x---
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
			// bits ----xxx-
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
	if v.registers[RegScreenControl1]&CTRL1_BMM != 0 {
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
	if v.registers[RegScreenControl1]&CTRL1_BMM != 0 { // Bitmap mode
		// In bitmap mode, address is based on pixel position
		return v.bitmapBase + uint16(charCode)*8 + uint16(rowInChar)
	} else {
		// In text mode, address is based on character code
		return v.charGen + uint16(charCode)*8 + uint16(rowInChar)
	}
}
