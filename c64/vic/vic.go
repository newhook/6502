package vic

import (
	"github.com/newhook/6502/c64/memory"
	"log"
)

const (
// Screen dimensions
)

const (
	// Sprite Constants
	NUM_SPRITES       = 8
	SPRITE_WIDTH      = 24
	SPRITE_DMA_CYCLES = 2

	// Bad Line Constants
	FIRST_BA_LINE = 0x30
	LAST_BA_LINE  = 0xF7

	//
	//	The dimensions of the video display for the different VIC types are as
	// follows:
	//
	//          | Video  | # of    | Visible | Cycles/ |  Visible
	// Type     | system | lines   |  lines  |  line   | pixels/line
	// ---------+--------+---------+---------+---------+------------
	// 6567R56A | NTSC-M |  262    |   234   |   64    |    411
	// 6567R8   | NTSC-M |  263    |   235   |   65    |    418
	// 6569     |  PAL-B |  312    |   284   |   63    |    403
	//
	//           | First   |  Last  |              |   First    |   Last
	//           | vblank  | vblank | First X coo. |  visible   |  visible
	// | Type    |  line  |  line   |  of a line   |   X coo.   |   X coo.
	// ----------+--------+---------+--------------+------------+-----------
	//  6567R56A |   13   |   40    |  412 ($19c)  | 488 ($1e8) | 388 ($184)
	//  6567R8   |   13   |   40    |  412 ($19c)  | 489 ($1e9) | 396 ($18c)
	//  6569     |  300   |   15    |  404 ($194)  | 480 ($1e0) | 380 ($17c)
	//

	// The width of the display window can each be set to two different
	// values with the bits CSEL in the register $d016:
	//
	// CSEL|   Display window width   | First X coo. | Last X coo.
	// ----+--------------------------+--------------+------------
	// 0 | 38 characters/304 pixels   |   31 ($1f)   |  334 ($14e)
	// 1 | 40 characters/320 pixels   |   24 ($18)   |  343 ($157)

	// This vic-ii emulates the 6569 (PAL-B).
	CYCLES_PER_LINE = 63 // CPU cycles per line

	// TOTAL_WIDTH for PAL there are 504 total pixels per line.
	TOTAL_WIDTH = 504
	// SCREEN_HEIGHT represent the visible area of the display.
	// XXX: rename VISIBLE_WIDTH.

	// First visible border starts at cycle 11 (91 pixels from left)
	// Start of visible screen area at cycle 15 (up to cycle 55)
	// Right border begins at cycle 55
	// Horizontal blank starts at cycle 63
	// Each cycle is 8 pixels wide

	//Key timing points:
	// - Line starts at cycle 0
	// - Left border starts: cycle 11
	// - Main screen area: cycles 15-54 (40 columns × 8 pixels)
	// - Right border begins: cycle 55
	// - Horizontal sync: cycles 58-62
	// - Line ends: cycle 63

	// Pal is 403x284, but to make it divisible by 8 we use 408,288.
	PAL_FULL_WIDTH  = 408
	PAL_FULL_HEIGHT = 288

	SCREEN_WIDTH = 403
	// VISIBLE_WIDTH is the portion of the display in between each border.
	LEFT_BORDER_CYCLE_START  = 11
	LEFT_BORDER_CYCLE_END    = 17
	RIGHT_BORDER_CYCLE_START = 57
	RIGHT_BORDER_CYCLE_END   = 61
	VISIBLE_WIDTH            = 320
	LEFT_BORDER_START        = 0
	LEFT_BORDER_END          = 24
	RIGHT_BORDER_START       = LEFT_BORDER_END + VISIBLE_WIDTH

	// For the VIC-II PAL (6569) vertical timing:
	// Total scanlines: 312 (PAL)
	// Breakdown:
	// Top border area:
	// - Starts at line 0
	// - First visible line starts at line 16
	// - Upper border ends at line 51
	// Main display area:
	// - Starts at line 51
	// - 200 visible lines (PAL mode)
	// - Each character row is 8 scanlines high
	// - 25 rows of text in normal text mode
	// Bottom border area:
	// - Starts at line 251
	// - Continues to line 311

	// TOTAL_LINES For PAL there are 312 total lines
	TOTAL_LINES = 312 // Total raster lines
	// SCREEN_HEIGHT represent the visible area of the display.
	SCREEN_HEIGHT = 284
	// VISIBLE_HEIGHT is the visible portion of the display.
	VISIBLE_HEIGHT     = 200
	FIRST_VISIBLE_LINE = 16
	LAST_VISIBLE_LINE  = 298
	// 298-14 = 284 visible lines
	FIRST_DISPLAY_LINE = 51
	LAST_DISPLAY_LINE  = 251

	// The height of the display window can each be set to two different
	// values with the bits RSEL in the register $d011.
	//
	// RSEL|  Display window height   | First line  | Last line
	// ----+--------------------------+-------------+----------
	// 0 | 24 text lines/192 pixels |   55 ($37)  | 246 ($f6)
	// 1 | 25 text lines/200 pixels |   51 ($33)  | 250 ($fa)

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
		mem: mem,
		// Pal is 403x284, but to make it divisible by 8 we use 408,288.
		displayBuffer: make([]uint8, PAL_FULL_WIDTH*PAL_FULL_HEIGHT),
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
	rasterLine  uint16 // y position (0 - 311) pal.
	rasterCycle uint8  // x position (1 - 63).
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
func (v *VIC) Update() *VICEvent {
	v.rasterCycle++

	// Check for bad line condition
	v.updateBadLine()

	// Handle display and sprite generation based on cycle
	if v.rasterLine >= FIRST_VISIBLE_LINE && v.rasterLine < LAST_VISIBLE_LINE {
		// Sprite data fetch cycles.
		switch v.rasterCycle {
		case 58:
			v.fetchSpriteData(0)
		case 60:
			v.fetchSpriteData(1)
		case 62:
			v.fetchSpriteData(2)
		case 1:
			v.fetchSpriteData(3)
		case 3:
			v.fetchSpriteData(4)
		case 5:
			v.fetchSpriteData(5)
		case 7:
			v.fetchSpriteData(6)
		case 9:
			v.fetchSpriteData(7)
		}

		// Character/bitmap fetch and display cycles (cycles 13-53)
		v.generateDisplayData()
	}

	// Handle sprite DMA and collision detection
	//v.updateSprites()

	// Update raster position
	if v.rasterCycle >= CYCLES_PER_LINE {
		v.rasterCycle = 0
		v.rasterLine++

		if v.rasterLine >= TOTAL_LINES {
			v.rasterLine = 0
			v.frameCount++
			return &VICEvent{Type: EventFrameComplete}
		}

		// Check for raster IRQ
		if v.rasterLine == v.rasterIRQ && v.registers[RegInterruptEnable]&0x01 != 0 {
			v.irqStatus |= 0x01
			return &VICEvent{Type: EventRasterIRQ}
		}
	}

	return nil
}

// renderSpriteColumn handles rendering a single column (2 cycles worth) of a sprite
func (v *VIC) renderSpriteColumn(spriteNum uint8) {
	sprite := &v.sprites[spriteNum]

	// Check if sprite is visible on current raster line
	spriteY := int16(sprite.y)
	currentY := int16(v.rasterLine - 50)
	spriteHeight := int16(21)
	if sprite.yExpand {
		spriteHeight *= 2
	}

	// Skip if sprite not visible on this line
	if currentY < spriteY || currentY >= spriteY+spriteHeight {
		return
	}

	x := int(sprite.x) + 24

	// Render 4 pixels worth of sprite data (2 cycles worth)
	// Calculate row in sprite data
	spriteRow := currentY - int16(sprite.y)
	if sprite.yExpand {
		spriteRow /= 2
	}

	// Get sprite data for this row
	spriteDataPtr := uint16(sprite.dataPtr) * 64
	spriteColor := v.registers[RegSprite0Color+spriteNum]

	// Calculate buffer offset
	//bufferOffset := int(currentY-15) * PAL_FULL_WIDTH
	bufferOffset := int(currentY) * PAL_FULL_WIDTH

	for i := uint16(0); i < 3; i++ {
		dataByte := v.mem.Read(spriteDataPtr + uint16(spriteRow)*3 + i)
		// Render 4 pixels
		for j := 0; j < 8; j++ {
			if (dataByte & (1 << j)) == 0 {
				continue
			}
			if sprite.multicolor {
				// Handle multicolor mode
				if i%2 == 0 { // Only process on even pixels
					////colorBits := (bits >> (2 - (i/2)*2)) & 0x03
					//var color uint8
					//switch colorBits {
					//case 1:
					//	color = v.registers[RegSpriteMulti0]
					//case 2:
					//	color = spriteColor
					//case 3:
					//	color = v.registers[RegSpriteMulti1]
					//default:
					//	continue // Transparent
					//}
					//_ = color

					//pixelX := startX + int(i)
					//if sprite.xExpand {
					//	v.plotSpritePixel(bufferOffset, pixelX*2, color, sprite.priority)
					//	v.plotSpritePixel(bufferOffset, pixelX*2+1, color, sprite.priority)
					//} else {
					//	v.plotSpritePixel(bufferOffset, pixelX, color, sprite.priority)
					//}
				}
			} else {
				newX := x + int(i)*8 + 8 - j
				v.displayBuffer[bufferOffset+newX] = spriteColor
			}
		}
	}
}

// renderSpriteColumn handles rendering a single column (2 cycles worth) of a sprite
func (v *VIC) renderSpriteColumnX(spriteNum uint8) {
	sprite := &v.sprites[spriteNum]

	// Check if sprite is visible on current raster line
	spriteY := int16(sprite.y)
	currentY := int16(v.rasterLine - 50)
	spriteHeight := int16(21)
	if sprite.yExpand {
		spriteHeight *= 2
	}

	// Skip if sprite not visible on this line
	if currentY < spriteY || currentY >= spriteY+spriteHeight {
		return
	}

	// Calculate which part of the sprite we're rendering
	spriteColumn := (v.rasterCycle - LEFT_BORDER_CYCLE_END) % (SPRITE_WIDTH / 4) // 6 cycles per sprite (24 pixels/4)
	//startX := int(v.rasterCycle-LEFT_BORDER_CYCLE_END) + int(sprite.x) + int(spriteColumn*4)

	// X=0 corresponds to the start of the visible screen area (cycle 15)
	// sprite.x gives us the offset in pixels from this point
	baseX := int(sprite.x) + 24

	// spriteColumn tells us which part of the sprite we're currently rendering
	// each column is 4 pixels wide
	columnOffset := int(spriteColumn * 4)

	// Final X position is just the base position plus the column offset
	startX := baseX + columnOffset

	// Render 4 pixels worth of sprite data (2 cycles worth)
	// Calculate row in sprite data
	spriteRow := currentY - int16(sprite.y)
	if sprite.yExpand {
		spriteRow /= 2
	}

	// Get sprite data for this row
	spriteDataPtr := uint16(sprite.dataPtr) * 64
	rowOffset := uint16(spriteRow) * 3 // 3 bytes per row

	// Read sprite data and extract relevant bits for this column
	dataByte := v.mem.Read(spriteDataPtr + rowOffset + uint16(spriteColumn/2))
	shift := (1 - (spriteColumn % 2)) * 4 // 0 or 4 depending on which half of byte
	bits := (dataByte >> shift) & 0x0F

	// Get sprite colors
	spriteColor := v.registers[RegSprite0Color+spriteNum]

	// Calculate buffer offset
	bufferOffset := int(currentY-15) * PAL_FULL_WIDTH

	// Render 4 pixels
	for i := uint8(0); i < 4; i++ {
		if sprite.multicolor {
			// Handle multicolor mode
			if i%2 == 0 { // Only process on even pixels
				colorBits := (bits >> (2 - (i/2)*2)) & 0x03
				var color uint8
				switch colorBits {
				case 1:
					color = v.registers[RegSpriteMulti0]
				case 2:
					color = spriteColor
				case 3:
					color = v.registers[RegSpriteMulti1]
				default:
					continue // Transparent
				}

				pixelX := startX + int(i)
				if sprite.xExpand {
					v.plotSpritePixelX(bufferOffset, pixelX*2, color, sprite.priority)
					v.plotSpritePixelX(bufferOffset, pixelX*2+1, color, sprite.priority)
				} else {
					v.plotSpritePixelX(bufferOffset, pixelX, color, sprite.priority)
				}
			}
		} else {
			// Standard mode
			if bits&(0x08>>(i)) != 0 {
				pixelX := startX + int(i)
				if sprite.xExpand {
					v.plotSpritePixelX(bufferOffset, pixelX*2, spriteColor, sprite.priority)
					v.plotSpritePixelX(bufferOffset, pixelX*2+1, spriteColor, sprite.priority)
				} else {
					v.plotSpritePixelX(bufferOffset, pixelX, spriteColor, sprite.priority)
				}
			}
		}
	}
}

// plotSpritePixel plots a single sprite pixel to the display buffer
func (v *VIC) plotSpritePixelX(bufferOffset, x int, color uint8, priority bool) {
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
		v.displayBuffer[pixelIndex] = color
	}
}

func (v *VIC) updateBadLine() {
	// Bad line condition:
	// 1. Current raster line is between 0x30-0xf7
	// 2. Lower 3 bits of raster line match lower 3 bits of scroll register
	// 3. Display enable bit is set
	if v.rasterLine >= 0x30 && v.rasterLine <= 0xf7 {
		if uint8(v.rasterLine&0x07) == (v.registers[RegScreenControl1] & CTRL1_YSCROLL) {
			if v.registers[RegScreenControl1]&CTRL1_DEN != 0 {
				v.badLine = true
				v.badLineEnable = true
				return
			}
		}
	}
	v.badLine = false
}

// First visible border starts at cycle 11 (91 pixels from left)
// Start of visible screen area at cycle 15 (up to cycle 55)
// Right border begins at cycle 55
// Horizontal blank starts at cycle 63
// Each cycle is 8 pixels wide

//For the VIC-II PAL (6569) horizontal timing:
// - Line starts at cycle 0
// - Left border starts: cycle 11
// - Main screen area: cycles 15-54 (40 columns × 8 pixels)
// - Right border begins: cycle 55
// - Horizontal sync: cycles 58-62
// - Line ends: cycle 63

// For the VIC-II PAL (6569) vertical timing:
// Total scanlines: 312 (PAL)
// Breakdown:
// Top border area:
// - Starts at line 0
// - First visible line starts at line 16
// - Upper border ends at line 51
// Main display area:
// - Starts at line 51
// - 200 visible lines (PAL mode)
// - Each character row is 8 scanlines high
// - 25 rows of text in normal text mode
// Bottom border area:
// - Starts at line 251
// - Continues to line 311

func (v *VIC) generateDisplayData() {
	rasterLine := v.rasterLine
	rasterCycle := v.rasterCycle

	// Only render during visible area
	if rasterLine < FIRST_VISIBLE_LINE || rasterLine >= LAST_VISIBLE_LINE ||
		rasterCycle < LEFT_BORDER_CYCLE_START || rasterCycle >= RIGHT_BORDER_CYCLE_END {
		return
	}

	// Calculate where in display buffer to put the pixels
	bufferIndex := v.getCurrentPixelIndex(uint16(rasterCycle), rasterLine)

	// border.
	if rasterLine < FIRST_DISPLAY_LINE || rasterLine >= LAST_DISPLAY_LINE ||
		rasterCycle < LEFT_BORDER_CYCLE_END || rasterCycle >= RIGHT_BORDER_CYCLE_START {
		// Render all 8 pixels for this character line
		for bit := uint8(0); bit < 8; bit++ {
			v.displayBuffer[bufferIndex+int(bit)] = v.registers[RegBorderColor] // Background color
		}
		return
	}

	// Calculate which character row and column we're rendering
	charRow := (rasterLine - FIRST_DISPLAY_LINE) / 8 // Which row of characters
	charCol := rasterCycle - LEFT_BORDER_CYCLE_END   // Which column in the current row
	charIndex := (charRow * 40) + uint16(charCol)    // Character position in screen RAM

	// Calculate which line of the character we're drawing (0-7)
	charLine := (rasterLine - FIRST_DISPLAY_LINE) % 8

	// Get character from screen RAM (screen matrix)
	// Screen RAM location is determined by memory pointers register
	screenAddr := v.videoMatrix + charIndex
	char := v.mem.Read(screenAddr)

	// Get character color from color RAM ($D800-$DBFF)
	colorAddr := 0xD800 + uint16(charIndex)
	charColor := v.mem.Read(colorAddr)

	// Get character data from character ROM/RAM
	// Character memory location determined by memory pointers register
	charData := v.mem.ReadChar(uint16(char)*8 + charLine)

	// Render all 8 pixels for this character line
	for bit := uint8(0); bit < 8; bit++ {
		pixel := (charData >> (7 - bit)) & 1
		if pixel == 1 {
			v.displayBuffer[bufferIndex+int(bit)] = charColor
		} else {
			v.displayBuffer[bufferIndex+int(bit)] = v.registers[RegBgColor0] // Background color
		}
	}

	for i := 0; i < 8; i++ {
		spriteNum := uint8(i)
		if v.sprites[spriteNum].enabled {
			v.renderSpriteColumn(spriteNum)
		}
	}
}

func (v *VIC) getCurrentPixelIndex(rasterX uint16, rasterY uint16) int {
	pixelY := (int(rasterY) - FIRST_VISIBLE_LINE) * PAL_FULL_WIDTH
	pixelX := (int(rasterX) - LEFT_BORDER_CYCLE_START) * 8

	return pixelY + pixelX
}

func (v *VIC) updateSprites() {
	//// Process sprite DMA cycles
	//if v.rasterCycle >= 15 && v.rasterCycle <= 54 {
	//	// Sprite data fetch happens during visible screen area
	//	v.fetchSpriteData()
	//}

	// Update sprite positions and check collisions
	v.updateSpritePositions()
	v.checkSpriteCollisions()
}

func (v *VIC) fetchSpriteData(spriteIndex uint8) {
	if v.sprites[spriteIndex].enabled {
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
	// XXX: not sure what 58 means here.
	if (rasterY < FIRST_DISPLAY_LINE || rasterY > LAST_DISPLAY_LINE) || rasterX < 58 {
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
		return (v.registers[reg] & 0x7F) | uint8((v.rasterLine&0x100)>>1)

	case RegRaster:
		// Return current raster line (lower 8 bits)
		return uint8(v.rasterLine & 0xFF)

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
	return v.rasterCycle, v.rasterLine
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
