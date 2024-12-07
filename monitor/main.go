package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/dis/disassembler"
	"os"
	"strconv"
	"strings"
)

// Monitor represents the UI state
type Monitor struct {
	cpu    *cpu.CPU
	paused bool
	width  int
	height int
}

// Define some basic styles
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	titleStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1).
			Width(30)

	stackStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Padding(1).
			Width(30)

	disasmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1)
)

// Initialize the monitor
func NewMonitor(cpu *cpu.CPU) *Monitor {
	return &Monitor{
		cpu:    cpu,
		paused: true,
	}
}

// Implementation of tea.Model interface
func (m Monitor) Init() tea.Cmd {
	return nil
}

// Handle keyboard input
func (m Monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "s":
			// Single step
			if m.paused {
				m.cpu.Step()
			}
		case "n":
			// Run until next instruction
			if m.paused {
				// TODO: Implement running until next instruction
			}
		case "p":
			m.paused = !m.paused
		}
	}
	return m, nil
}

// Format CPU flags
func (m Monitor) formatFlags() string {
	flags := []struct {
		name string
		flag uint8
	}{
		{"N", cpu.FlagN},
		{"V", cpu.FlagV},
		{"B", cpu.FlagB},
		{"D", cpu.FlagD},
		{"I", cpu.FlagI},
		{"Z", cpu.FlagZ},
		{"C", cpu.FlagC},
	}

	var result strings.Builder
	for _, f := range flags {
		if m.cpu.P&f.flag != 0 {
			result.WriteString(fmt.Sprintf("%s ", f.name))
		} else {
			result.WriteString("- ")
		}
	}
	return result.String()
}

// Disassemble memory around PC
func (m Monitor) disassemble() string {
	return disassembler.DisassembleMemory(m.cpu.Memory[:], int(m.cpu.PC), 100)
	//var result strings.Builder
	//start := m.cpu.PC - 6
	//end := m.cpu.PC + 6
	//
	//for addr := start; addr <= end; addr += 2 {
	//	instruction := m.cpu.Memory[addr]
	//	if addr == m.cpu.PC {
	//		result.WriteString("→ ")
	//	} else {
	//		result.WriteString("  ")
	//	}
	//	result.WriteString(fmt.Sprintf("$%04X: %02X\n", addr, instruction))
	//}
	//return result.String()
}

// Show stack contents
func (m Monitor) formatStack() string {
	var result strings.Builder
	for i := uint16(0xFF); i >= uint16(m.cpu.SP); i-- {
		result.WriteString(fmt.Sprintf("$%02X: %02X\n", i, m.cpu.Memory[0x100+i]))
	}
	return result.String()
}

func (m Monitor) View() string {

	// Calculate column widths
	rightColumnWidth := 32
	leftColumnWidth := 40 // Fixed width for disassembly

	// Update style widths
	infoStyle = infoStyle.Width(rightColumnWidth)
	stackStyle = stackStyle.Width(rightColumnWidth)
	disasmStyle = disasmStyle.Width(leftColumnWidth)

	// Left column: Disassembly
	disasm := disasmStyle.Render(fmt.Sprintf(
		"Disassembly\n\n%s",
		m.disassemble(),
	))

	// Right column: CPU State and Stack
	cpuState := infoStyle.Render(fmt.Sprintf(
		"CPU State\n\n"+
			"A:  $%02X    X: $%02X    Y: $%02X\n"+
			"PC: $%04X  SP: $%02X\n\n"+
			"Flags: %s\n",
		m.cpu.A, m.cpu.X, m.cpu.Y,
		m.cpu.PC, m.cpu.SP,
		m.formatFlags(),
	))

	stack := stackStyle.Render(fmt.Sprintf(
		"Stack\n\n%s",
		m.formatStack(),
	))

	// Combine right column elements
	right := lipgloss.JoinVertical(
		lipgloss.Left,
		cpuState,
		stack,
	)

	// Help section at the bottom
	help := titleStyle.Render(
		"s: step • n: next • p: pause/resume • q: quit",
	)

	// Join columns horizontally with spacing
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		disasm,
		lipgloss.PlaceHorizontal(3, lipgloss.Left, right),
	)

	// Join everything vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func LoadAndSetupBinary(c *cpu.CPU, filename string, startAddr int) (int, error) {
	// Read the binary file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read binary file: %v", err)
	}

	// Check if the binary will fit in memory
	if int(startAddr)+len(data) > len(c.Memory) {
		return 0, fmt.Errorf("binary file too large for available memory")
	}

	// Copy binary data into CPU memory starting at 0xF000
	for i, b := range data {
		c.Memory[uint16(startAddr)+uint16(i)] = b
	}

	// Set up reset vector at 0xFFFC-0xFFFD to point to 0xF000
	c.Memory[0xFFFC] = 0x00 // Low byte
	c.Memory[0xFFFD] = 0xF0 // High byte

	// Set up IRQ vector at 0xFFFE-0xFFFF to point to 0xF5A4
	c.Memory[0xFFFE] = 0xA4 // Low byte
	c.Memory[0xFFFF] = 0xF5 // High byte

	// Set the Program Counter to the reset vector location
	c.PC = uint16(startAddr)

	return len(data), nil
}

func main() {
	// Command line flags
	inputFile := flag.String("i", "", "Input binary file")
	startAddr := flag.String("a", "", "Start address")
	flag.Parse()

	addrStr := *startAddr
	if strings.HasPrefix(addrStr, "$") {
		addrStr = "0x" + addrStr[1:]
	}
	startAddrInt, err := strconv.ParseUint(addrStr, 0, 16)
	if err != nil {
		fmt.Printf("Error parsing start address: %v\n", err)
		return
	}

	// Create and initialize CPU
	c := cpu.NewCPU()
	_, err = LoadAndSetupBinary(c, *inputFile, int(startAddrInt))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	p := tea.NewProgram(NewMonitor(c))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
