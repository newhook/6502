package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newhook/6502/cpu"
	"log"
	"os"
	"time"
	"unicode"
)

var runMonitor = false

type Interface interface {
	ReadRegister(reg uint8) uint8
	WriteRegister(reg uint8, value uint8)
}

type Memory struct {
	// 4K RAM from 0x0000 to 0x0FFF
	ram [4096]uint8
	// 4K RAM from 0xE000 to 0xEFFF
	e000 [4096]uint8

	// 256 bytes rom from 0xFF00 to 0xFFFF
	rom [256]uint8

	// PIA registers (0xD010-0xD013)
	PIA Interface
}

func (m *Memory) Read(addr uint16) uint8 {
	switch {
	case addr <= 0x0FFF:
		return m.ram[addr]
	case addr >= 0xE000 && addr <= 0xEFFF:
		return m.e000[addr-0xE000]
	case addr >= 0xD010 && addr <= 0xD013:
		return m.PIA.ReadRegister(uint8(addr - 0xD010))
	case addr >= 0xFF00:
		return m.rom[addr-0xFF00]
	default:
		// Unmapped memory reads typically return 0
		return 0
	}
}

func (m *Memory) Write(addr uint16, val uint8) {
	switch {
	case addr <= 0x0FFF:
		m.ram[addr] = val
	case addr >= 0xE000 && addr <= 0xEFFF:
		m.e000[addr-0xE000] = val
	case addr >= 0xD010 && addr <= 0xD013:
		m.PIA.WriteRegister(uint8(addr-0xD010), val)
	}
}

// LoadROM loads ROM data into the specified ROM area
func (m *Memory) LoadROM(data []uint8, romType string) error {
	switch romType {
	case "basic":
		if len(data) != 4096 {
			return fmt.Errorf("BASIC ROM must be 4K, got %d bytes", len(data))
		}
		copy(m.e000[:], data)
	case "monitor":
		if len(data) != 256 {
			return fmt.Errorf("montior ROM must be 256 bytes, got %d bytes", len(data))
		}
		copy(m.rom[:], data)
	default:
		return fmt.Errorf("unknown ROM type: %s", romType)
	}
	return nil
}

type PIA struct {
	kbd   uint8 // Keyboard Data Register (0xD010)
	kbdcr uint8 // Keyboard Control Register (0xD011)
	dsp   uint8 // Display Register (0xD012)
	dspcr uint8 // Display Control Register (0xD013)

	// Callback for when character should be displayed
	OnDisplay func(char byte)
}

func NewPIA(displayCallback func(char byte)) *PIA {
	return &PIA{
		dspcr:     0x80, // Display initially ready
		OnDisplay: displayCallback,
	}
}

func (p *PIA) ReadRegister(addr uint8) uint8 {
	switch addr {
	case 0x0:
		return p.kbd
	case 0x1:
		return p.kbdcr
	case 0x2:
		return p.dsp
	case 0x3:
		return p.dspcr
	default:
		return 0
	}
}

func (p *PIA) WriteRegister(addr uint8, val uint8) {
	switch addr {
	case 0x1:
		// Clear keyboard strobe
		p.kbd &= 0x7F
	case 0x2:
		// Write to display
		if p.dspcr&0x80 != 0 { // Check if display is ready
			p.dsp = val
			if val&0x80 != 0 { // Only display if bit 7 is set
				p.OnDisplay(val & 0x7F) // Strip bit 7 for ASCII
			}
		}
	}
}

// Called when a key is pressed on the emulated keyboard
func (p *PIA) KeyPress(key byte) {
	p.kbd = key | 0x80 // Set ASCII code and strobe bit
}

type A1 struct {
	CPU    *cpu.CPU
	Memory *Memory
	PIA    *PIA
}

func NewA1() (*A1, error) {
	pia := NewPIA(func(char byte) {
		fmt.Println("pia output")
		fmt.Print(string(char))
	})
	mem := &Memory{
		PIA: pia,
	}

	c := cpu.NewCPU(mem)
	a1 := &A1{
		CPU:    c,
		Memory: mem,
		PIA:    pia,
	}

	return a1, nil
}

func (c *A1) IsRunning() bool {
	return true
}

func (c *A1) Step() uint8 {
	// Execute one CPU instruction
	cpuCycles := c.CPU.Step()
	return cpuCycles
}

func (c *A1) HandleKey(r rune) {
	// Apple 1 only accepted uppercase ASCII
	// Convert to uppercase and ensure it's in ASCII range
	ascii := byte(unicode.ToUpper(r))

	// Special key translations
	switch r {
	case '\n', '\r':
		ascii = 0x0D // Return key -> Carriage Return
	case '\b', 127:
		ascii = 0x5F // Backspace/Delete -> Underscore (Apple 1's delete char)
	case '\t':
		ascii = 0x09 // Tab
	case 0x1B:
		ascii = 0x1B // Escape
	}

	// Only process printable ASCII chars and control chars
	if ascii <= 0x7F {
		c.PIA.KeyPress(ascii)
	}
}

type Model struct {
	emu      *A1
	display  string
	quitting bool
	paused   bool
}

func NewModel(emu *A1) Model {
	m := Model{
		emu:     emu,
		display: "",
	}

	emu.PIA.OnDisplay = func(char byte) {
		// Append the new character to display
		//char := byte(char)
		if char == 0x0D { // Carriage return
			m.display += "\n"
		} else {
			m.display += string(char)
		}
	}
	return m

}

// Define our messages
func (m Model) Init() tea.Cmd {
	return tick()
}

// Add a tick command to drive the emulation
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

var cyclesPerTick = 1

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		default:
			// Convert key press to Apple 1 format and send to PIA
			m.emu.HandleKey(rune(msg.String()[0]))
			return m, nil
		}

	case tickMsg:
		if !m.paused {
			// Run CPU for cyclesPerTick cycles
			for i := 0; i < cyclesPerTick; i++ {
				m.emu.Step()
			}
		}
		return m, tick()
	}

	return m, nil
}

func (m Model) View() string {
	status := "RUNNING"
	if m.paused {
		status = "PAUSED"
	}

	return fmt.Sprintf(`
╭────── Apple 1 Emulator ──────╮
Status: %s
PC: %04X  A: %02X  X: %02X  Y: %02X
%s
╰─────────────────────────────╯
Commands:
 Space: Pause/Resume
 S: Single step (when paused)
 Q: Quit
`, status, m.emu.CPU.PC, m.emu.CPU.A, m.emu.CPU.X, m.emu.CPU.Y, m.display)
}

func main() {
	computer, err := NewA1()
	if err != nil {
		log.Fatal(err)
	}
	do := func() error {
		mem := computer.Memory
		// Load ROMs
		basicROM, err := os.ReadFile("basic.rom")
		if err != nil {
			return err
		}
		monitorROM, err := os.ReadFile("monitor.rom")
		if err != nil {
			return err
		}

		if err := mem.LoadROM(basicROM, "basic"); err != nil {
			return err
		}
		if err := mem.LoadROM(monitorROM, "monitor"); err != nil {
			return err
		}

		// Initialize CPU registers
		// Reset vector
		computer.CPU.PC = uint16(mem.Read(0xFFFC)) | uint16(mem.Read(0xFFFD))<<8
		computer.CPU.PC = 0xFF1F

		fmt.Printf("%x\n", computer.CPU.PC)

		p := tea.NewProgram(NewModel(computer))
		if err := p.Start(); err != nil {
			fmt.Printf("Error running program: %v", err)
		}
		return nil

		// Main emulation loop
		for computer.IsRunning() {
			computer.Step()
			if err := computer.CPU.Error(); err != nil {
				panic(err)
			}

			// Optional: Add delay to match real C64 speed
			//if computer.Timing.ShouldDelay() {
			//	time.Sleep(c64.Timing.GetDelay())
			//}
		}
		return nil
	}
	if err := do(); err != nil {
		log.Fatal("error", err)
	}
}
