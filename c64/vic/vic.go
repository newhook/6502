package vic

import (
	"github.com/newhook/6502/c64/memory"
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
)

// VICEvent represents different VIC-II events
type VICEvent struct {
	Type EventType
	Data interface{}
}

type EventType int

const (
	EventBadLine EventType = iota
	EventSpriteDMA
	EventRasterIRQ
	EventDisplayStart
	EventDisplayEnd
)

type VIC struct {
	mem *memory.Manager

	// Raster beam position
	rasterX uint16
	rasterY uint16

	// Registers
	registers [47]uint8

	// Display state
	displayState   bool
	badLine        bool
	badLineEnable  bool
	verticalBorder bool

	// Sprite state
	sprites         [8]Sprite
	spriteDMAActive uint8

	// Memory pointers
	charMemPtr   uint16
	screenMemPtr uint16
	bitmapMemPtr uint16

	// IRQ state
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
		mem:       mem,
		registers: [47]uint8{},
		sprites:   [8]Sprite{},
	}
}

// Update processes one VIC-II cycle and returns any events that occurred
func (v *VIC) Update(cycle uint8) []*VICEvent {
	var events []*VICEvent

	// Update raster position
	v.updateRasterPosition()

	// Check for bad line condition
	if v.isBadLine() {
		events = append(events, &VICEvent{Type: EventBadLine})
	}

	// Check for sprite DMA
	if spriteEvents := v.updateSpriteDMA(); len(spriteEvents) > 0 {
		events = append(events, spriteEvents...)
	}

	// Handle display state
	if displayEvent := v.updateDisplayState(); displayEvent != nil {
		events = append(events, displayEvent)
	}

	// Check for raster interrupt
	if v.checkRasterInterrupt() {
		events = append(events, &VICEvent{Type: EventRasterIRQ})
	}

	return events
}

func (v *VIC) updateRasterPosition() {
	v.rasterX++
	if v.rasterX >= SCREEN_WIDTH {
		v.rasterX = 0
		v.rasterY++
		if v.rasterY > 311 { // PAL timing
			v.rasterY = 0
		}
	}
}

func (v *VIC) isBadLine() bool {
	// Bad lines occur when:
	// 1. Raster Y position is between 0x30-0xF7
	// 2. Lower 3 bits of raster Y match lower 3 bits of scroll register
	// 3. Display enable bit is set
	if v.rasterY >= FIRST_BA_LINE && v.rasterY <= LAST_BA_LINE {
		scrollY := v.registers[0x11] & 0x07
		if uint8(v.rasterY&0x07) == scrollY && (v.registers[0x11]&0x10) != 0 {
			return true
		}
	}
	return false
}

func (v *VIC) updateSpriteDMA() []*VICEvent {
	var events []*VICEvent

	// Calculate sprite DMA cycles
	for i := 0; i < NUM_SPRITES; i++ {
		sprite := &v.sprites[i]
		if !sprite.enabled {
			continue
		}

		// Sprite DMA occurs 2 cycles before the sprite is displayed
		spriteX := sprite.xPos
		if v.rasterX == (spriteX - SPRITE_DMA_CYCLES) {
			events = append(events, &VICEvent{
				Type: EventSpriteDMA,
				Data: i,
			})
			sprite.dmaCount = SPRITE_WIDTH / 8 // Number of bytes to fetch
		}
	}

	return events
}

func (v *VIC) updateDisplayState() *VICEvent {
	oldState := v.displayState

	// Update vertical border flags
	if v.rasterY == FIRST_DISPLAY_LINE {
		v.verticalBorder = false
	} else if v.rasterY == LAST_DISPLAY_LINE {
		v.verticalBorder = true
	}

	// Update display state
	v.displayState = !v.verticalBorder &&
		v.rasterX >= 24 && v.rasterX < (24+VISIBLE_WIDTH)

	// Return event if display state changed
	if v.displayState != oldState {
		eventType := EventDisplayEnd
		if v.displayState {
			eventType = EventDisplayStart
		}
		return &VICEvent{Type: eventType}
	}

	return nil
}

func (v *VIC) checkRasterInterrupt() bool {
	rasterComp := uint16(v.registers[0x12]) |
		(uint16(v.registers[0x11]&0x80) << 1)

	// Check if we've hit the raster compare value
	if v.rasterY == rasterComp {
		if (v.registers[0x1A] & 0x01) != 0 { // Raster interrupt enabled
			v.irqStatus |= 0x01
			return true
		}
	}
	return false
}

// Registers access methods
func (v *VIC) WriteRegister(reg uint8, value uint8) {
	v.registers[reg] = value

	// Handle special register writes
	switch reg {
	case 0x11: // Control register 1
		v.updateMemoryPointers()
	case 0x18: // Memory pointers
		v.updateMemoryPointers()
	}
}

func (v *VIC) ReadRegister(reg uint8) uint8 {
	// Special handling for certain registers
	switch reg {
	case 0x19: // IRQ Status
		value := v.irqStatus
		v.irqStatus = 0 // Reading clears IRQ flags
		return value
	default:
		return v.registers[reg]
	}
}

func (v *VIC) updateMemoryPointers() {
	// Update memory pointers based on registers
	memConfig := v.registers[0x18]
	v.screenMemPtr = uint16(memConfig&0xF0) << 6
	v.charMemPtr = uint16(memConfig&0x0E) << 10
	v.bitmapMemPtr = uint16(memConfig&0x08) << 10
}

// Helper methods for C64 core
func (v *VIC) IsDisplayActive() bool {
	return v.displayState
}

func (v *VIC) IsBadLine() bool {
	return v.badLine
}

func (v *VIC) GetRasterPosition() (uint16, uint16) {
	return v.rasterX, v.rasterY
}

func (v *VIC) GetIRQStatus() uint8 {
	return v.irqStatus
}

func (v *VIC) IsIRQEnabled() uint8 {
	return v.irqEnabled
}
