package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eiannone/keyboard"
	"github.com/newhook/6502/cpu"
	"log"
	"log/slog"
	"os"
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
		if p.kbd != 0 {
			p.kbdcr &= 0x7F // Clear keyboard strobe
		}
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
		p.kbdcr &= 0x7F
	case 0x2:
		// Write to display
		if p.dspcr&0x80 != 0 { // Check if display is ready
			p.dsp = val
			if val&0x80 != 0 { // Only display if bit 7 is set
				p.OnDisplay(val & 0x7F) // Strip bit 7 for ASCII
				p.dsp &= 0x7F
			}
		}
	}
}

// Called when a key is pressed on the emulated keyboard
func (p *PIA) KeyPress(key byte) {
	p.kbd = key | 0x80 // Set ASCII code and strobe bit
	p.kbdcr |= 0x80
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

		if true {
			mon := NewMonitor(computer, computer.CPU, computer.Memory, computer.PIA)
			p := tea.NewProgram(mon)

			logger := slog.New(slog.NewTextHandler(mon, nil))
			slog.SetDefault(logger)

			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running program: %v", err)
			}
			return nil
		}

		if err := keyboard.Open(); err != nil {
			return err
		}
		defer keyboard.Close()

		computer.PIA.OnDisplay = func(char byte) {
			if char == '\r' {
				fmt.Print("\n")
			} else {
				fmt.Print(string(char))
			}
		}

		keysEvents, err := keyboard.GetKeys(10)
		if err != nil {
			return err
		}

		// Main emulation loop
	loop:
		for computer.IsRunning() {
			select {
			case k := <-keysEvents:
				char := k.Rune
				// Convert lowercase to uppercase
				if char >= 'a' && char <= 'z' {
					char = char - 32
				}

				switch k.Key {
				case keyboard.KeyF10:
					break loop
				case keyboard.KeyEsc:
					computer.PIA.KeyPress(0x1B) // Send ESC to PIA
				case keyboard.KeySpace:
					computer.PIA.KeyPress(' ')
				case keyboard.KeyEnter:
					computer.PIA.KeyPress(0x0D) // CR
				case keyboard.KeyBackspace, keyboard.KeyBackspace2:
					computer.PIA.KeyPress(0x08) // BS
				default:
					if char >= 0x20 && char <= 0x7F {
						computer.PIA.KeyPress(byte(char))
					}
				}
			default:
			}
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
