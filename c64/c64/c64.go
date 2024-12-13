package c64

import (
	"fmt"
	"github.com/newhook/6502/c64/cia"
	"github.com/newhook/6502/c64/memory"
	"github.com/newhook/6502/c64/sid"
	"github.com/newhook/6502/c64/vic"
	"github.com/newhook/6502/cpu"
	"github.com/veandco/go-sdl2/sdl"
	"time"
	"unsafe"
)

const (
	// Clock frequencies
	PAL_CLOCK_HZ  = 985248  // PAL C64 clock frequency
	NTSC_CLOCK_HZ = 1022727 // NTSC C64 clock frequency

	// Video timing constants (PAL)
	CYCLES_PER_LINE    = 63
	LINES_PER_FRAME    = 312
	CYCLES_PER_FRAME   = CYCLES_PER_LINE * LINES_PER_FRAME
	RASTER_FIRST_LINE  = 0
	RASTER_LAST_LINE   = 311
	RASTER_FIRST_CYCLE = 0
	RASTER_LAST_CYCLE  = 62

	// VIC-II visible area
	FIRST_VISIBLE_LINE = 16
	LAST_VISIBLE_LINE  = 287
	VISIBLE_LINES      = LAST_VISIBLE_LINE - FIRST_VISIBLE_LINE + 1
)

type TimingConfig struct {
	clockFrequency int
	cyclesPerLine  int
	linesPerFrame  int
}

// Timing represents the cycle-accurate timing system
type Timing struct {
	config TimingConfig

	// Current timing state
	currentCycle   uint64
	cyclesThisLine int
	currentLine    int
	frameCount     uint64

	// Timing control
	targetCycles uint64
	lastUpdate   time.Time

	// Component cycle counts
	cpuCycles  uint64
	vicCycles  uint64
	sidCycles  uint64
	cia1Cycles uint64
	cia2Cycles uint64
}

func NewTiming(isPAL bool) *Timing {
	config := TimingConfig{
		clockFrequency: PAL_CLOCK_HZ,
		cyclesPerLine:  CYCLES_PER_LINE,
		linesPerFrame:  LINES_PER_FRAME,
	}

	if !isPAL {
		config.clockFrequency = NTSC_CLOCK_HZ
		config.linesPerFrame = 263 // NTSC has fewer lines
	}

	return &Timing{
		config:     config,
		lastUpdate: time.Now(),
	}
}

// Step advances the system by one CPU cycle
func (t *Timing) Step() {
	t.currentCycle++
	t.cyclesThisLine++

	// Check for end of line
	if t.cyclesThisLine >= t.config.cyclesPerLine {
		t.cyclesThisLine = 0
		t.currentLine++

		// Check for end of frame
		if t.currentLine >= t.config.linesPerFrame {
			t.currentLine = 0
			t.frameCount++
		}
	}
}

// IsVisible returns true if we're in the visible screen area
func (t *Timing) IsVisible() bool {
	return t.currentLine >= FIRST_VISIBLE_LINE &&
		t.currentLine <= LAST_VISIBLE_LINE
}

type C64 struct {
	CPU    *cpu.CPU
	Memory *memory.Manager
	VIC    *vic.VIC
	SID    *sid.SID
	CIA1   *cia.CIA
	CIA2   *cia.CIA

	Timing *Timing

	// Interrupt handling
	irqLine bool
	nmiLine bool

	// Rendering.
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	pixels   []byte
	running  bool
}

func NewC64() (*C64, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	window, err := sdl.CreateWindow("C64 Emulator",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		640, 400, // Double the original resolution for better visibility
		sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		window.Destroy()
		return nil, err
	}

	// Create texture that matches C64's native resolution
	texture, err := renderer.CreateTexture(
		uint32(sdl.PIXELFORMAT_ABGR8888),
		sdl.TEXTUREACCESS_STREAMING,
		320, 200)
	if err != nil {
		renderer.Destroy()
		window.Destroy()
		return nil, err
	}

	mem := memory.NewManager()

	c := cpu.NewCPU(mem)

	// Initialize CPU registers
	// Reset vector
	c.PC = uint16(mem.Read(0xFFFC)) | uint16(mem.Read(0xFFFD))<<8
	// Initialize stack pointer
	c.SP = 0xFF
	// Set interrupt disable flag
	c.P = cpu.FlagI

	return &C64{
		CPU:    c,
		Memory: mem,
		VIC:    vic.NewVIC(mem),
		SID:    sid.NewSID(),
		CIA1:   cia.NewCIA(),
		CIA2:   cia.NewCIA(),
		Timing: NewTiming(false),

		window:   window,
		renderer: renderer,
		texture:  texture,
		pixels:   make([]byte, 320*200*4),
		running:  false,
	}, nil
}

func (c *C64) Step() {
	// Execute one CPU instruction
	cpuCycles := c.CPU.Step()

	// Update each component for the number of cycles
	for i := uint8(0); i < cpuCycles; i++ {
		// Update global timing
		c.Timing.Step()

		// Update VIC-II
		if event := c.VIC.Update(1); event != nil {
			switch event.Type {
			case vic.EventRasterIRQ:
				//c.CPU.TriggerIRQ()
			case vic.EventFrameComplete:
				if err := c.RenderFrame(c.VIC.GetDisplayBuffer()); err != nil {
					fmt.Println(err)
				}
			}
		}

		// If it's a bad line, stall the CPU
		if c.VIC.IsBadLine() {
			//c.CPU.Stall(43)
		}

		// Update SID (runs at a different clock rate)
		c.SID.AddDelta(1)
		if c.SID.Clock%(c.Timing.config.clockFrequency/44100) == 0 {
			c.SID.Update()
		}

		// Update CIAs
		//if cia1Event := c.CIA1.Update(1); cia1Event != nil {
		//	c.handleCIA1Event(cia1Event)
		//}
		//if cia2Event := c.CIA2.Update(1); cia2Event != nil {
		//	c.handleCIA2Event(cia2Event)
		//}

		// Check interrupts
		c.updateInterrupts()
	}
}

func (c *C64) updateInterrupts() {
	// Check and handle IRQ conditions
	//newIRQ := c.VIC.IsIRQEnabled() || c.CIA1.IRQActive() || c.CIA2.IRQActive()
	//if newIRQ != c.irqLine {
	//	c.irqLine = newIRQ
	//	if newIRQ {
	//		c.CPU.TriggerIRQ()
	//	}
	//}
	//
	//// Check and handle NMI conditions
	//newNMI := c.CIA2.NMIActive()
	//if newNMI != c.nmiLine {
	//	c.nmiLine = newNMI
	//	if newNMI {
	//		c.CPU.TriggerNMI()
	//	}
	//}
}

func (c *C64) handleSpriteDMA(spriteNum int) {
	// Perform sprite DMA during appropriate cycles
	//baseAddr := c.VIC.GetSpritePointer(spriteNum) * 64
	//c.Memory.DMA(baseAddr, 63) // Copy sprite data
}

// C64Colors represents the standard C64 palette
var C64Colors = []uint32{
	0x000000, // Black
	0xFFFFFF, // White
	0x880000, // Red
	0xAAFFEE, // Cyan
	0xCC44CC, // Purple
	0x00CC55, // Green
	0x0000AA, // Blue
	0xEEEE77, // Yellow
	0xDD8855, // Orange
	0x664400, // Brown
	0xFF7777, // Light red
	0x333333, // Dark grey
	0x777777, // Medium grey
	0xAAFF66, // Light green
	0x0088FF, // Light blue
	0xBBBBBB, // Light grey
}

func (c *C64) RenderFrame(buffer []uint8) error {
	//for y := 0; y < 200; y++ {
	//	for x := 0; x < 320; x++ {
	//		fmt.Printf("%02x", buffer[y*320+x])
	//	}
	//	fmt.Printf("\n")
	//}
	// Handle SDL events
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			c.running = false
			return nil
		}
	}
	// Convert the VIC output buffer to RGBA pixels
	for i := 0; i < len(buffer); i++ {
		colorIndex := buffer[i] & 0x0F // Get color index (0-15)
		color := C64Colors[colorIndex]

		// Convert 32-bit color to RGBA components
		pixelOffset := i * 4
		c.pixels[pixelOffset+0] = byte((color >> 24) & 0xFF) // R
		c.pixels[pixelOffset+1] = byte((color >> 16) & 0xFF) // G
		c.pixels[pixelOffset+2] = byte((color >> 8) & 0xFF)  // B
		c.pixels[pixelOffset+3] = 0xFF                       // A (full opacity)
	}

	// Update texture with new pixel data
	if err := c.texture.Update(nil, unsafe.Pointer(&c.pixels[0]), 320*4); err != nil {
		return err
	}

	// Clear the renderer
	if err := c.renderer.Clear(); err != nil {
		return err
	}

	// Copy texture to renderer, scaling it to window size
	if err := c.renderer.Copy(c.texture, nil, nil); err != nil {
		return err
	}

	// Present the renderer
	c.renderer.Present()

	return nil
}

func (c *C64) Cleanup() {
	if c.texture != nil {
		c.texture.Destroy()
	}
	if c.renderer != nil {
		c.renderer.Destroy()
	}
	if c.window != nil {
		c.window.Destroy()
	}
	sdl.Quit()
}
