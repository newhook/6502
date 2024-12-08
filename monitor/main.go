package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/dis/disassembler"
	"os"
	"strconv"
	"strings"
	"time"
)

// CPUState holds a snapshot of CPU state
type CPUState struct {
	A  uint8
	X  uint8
	Y  uint8
	PC uint16
	SP uint8
	P  uint8
}

// Add tick command for CPU stepping
type stepTick struct{}

func doStep() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return stepTick{}
	})
}

type Memory [65536]uint8

func (c *Memory) Read(address uint16) uint8 {
	return c[address]
}
func (c *Memory) Write(address uint16, value uint8) {
	c[address] = value
}

// Monitor represents the UI state
type Monitor struct {
	mem              *Memory
	cpu              *cpu.CPU
	paused           bool
	width            int
	height           int
	locations        []disassembler.Location
	locationIndex    int
	selectedLocation int

	lastState  CPUState  // Previous CPU state for change detection
	lastMemory [64]uint8 // Only track visible memory (8 rows * 8 bytes)

	memoryAddress uint16 // Start address for memory view
	activePane    string // "disasm", "memory"
	gotoInput     textinput.Model
	showingGoto   bool

	breakpoints map[uint16]bool // Track breakpoint addresses
}

// Define some basic styles
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	changed   = lipgloss.AdaptiveColor{Light: "#FF6B6B", Dark: "#FF6B6B"}

	titleStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1).
			Width(30)

	changedStyle = lipgloss.NewStyle().
			Foreground(changed).
			Bold(true)

	stackStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Padding(1).
			Width(30)

	disasmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1)

	currentLineStyle = lipgloss.NewStyle().
				Background(highlight).
				Foreground(lipgloss.Color("#ffffff"))

	selectedLineStyle = lipgloss.NewStyle().
				Foreground(highlight)

	// Add new style for memory panel
	memoryStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Padding(1).
			Width(50)

	breakpointStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

// Initialize the monitor
func NewMonitor(cpu *cpu.CPU, mem *Memory) *Monitor {
	ti := textinput.New()
	ti.Placeholder = "Enter hex address (e.g. FF00)"
	ti.CharLimit = 4
	ti.Width = 6

	m := &Monitor{
		mem:           mem,
		cpu:           cpu,
		paused:        true,
		locations:     disassembler.DisassembleInstructions(mem[:]),
		memoryAddress: 0,
		activePane:    "disasm",
		gotoInput:     ti,
		breakpoints:   make(map[uint16]bool),
	}
	m.relocate()
	return m
}

// Helper function to capture current memory view state
func (m *Monitor) captureMemoryState() {
	addr := m.memoryAddress
	for i := 0; i < 64; i++ {
		m.lastMemory[i] = m.mem[addr+uint16(i)]
	}
}

// Format memory panel content with change highlighting
func (m Monitor) formatMemory() string {
	var result strings.Builder
	addr := m.memoryAddress

	for row := 0; row < 8; row++ {
		// Add row address
		result.WriteString(fmt.Sprintf("$%04X: ", addr))

		// Add hex bytes
		for col := 0; col < 8; col++ {
			offset := row*8 + col
			value := m.mem[addr+uint16(col)]
			lastValue := m.lastMemory[offset]

			if value != lastValue {
				result.WriteString(changedStyle.Render(fmt.Sprintf("%02X ", value)))
			} else {
				result.WriteString(fmt.Sprintf("%02X ", value))
			}
		}

		// Add ASCII representation
		result.WriteString(" | ")
		for col := 0; col < 8; col++ {
			offset := row*8 + col
			value := m.mem[addr+uint16(col)]
			lastValue := m.lastMemory[offset]

			if value >= 32 && value <= 126 {
				if value != lastValue {
					result.WriteString(changedStyle.Render(string(value)))
				} else {
					result.WriteString(string(value))
				}
			} else {
				if value != lastValue {
					result.WriteString(changedStyle.Render("."))
				} else {
					result.WriteString(".")
				}
			}
		}

		result.WriteString("\n")
		addr += 8
	}

	return result.String()
}

// Implementation of tea.Model interface
func (m Monitor) Init() tea.Cmd {
	return nil
}

func (m *Monitor) relocate() {
	index := 0
	for i, l := range m.locations {
		if l.PC == m.cpu.PC {
			index = i
		}
	}
	m.locationIndex = index
	m.selectedLocation = index
}

// Handle keyboard input
func (m Monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stepTick:
		// Check if we hit a breakpoint
		if m.paused || m.breakpoints[m.cpu.PC] {
			m.paused = true
			return m, nil
		}

		// Store state before step
		m.lastState = CPUState{
			A:  m.cpu.A,
			X:  m.cpu.X,
			Y:  m.cpu.Y,
			PC: m.cpu.PC,
			SP: m.cpu.SP,
			P:  m.cpu.P,
		}
		m.captureMemoryState()

		// Execute step
		m.cpu.Step()
		m.relocate()

		// Continue stepping
		return m, doStep()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.showingGoto {
			switch msg.Type {
			case tea.KeyEnter:
				if addr, err := strconv.ParseUint(m.gotoInput.Value(), 16, 16); err == nil {
					m.memoryAddress = uint16(addr)
				}
				m.showingGoto = false
				return m, nil
			case tea.KeyEsc:
				m.showingGoto = false
				return m, nil
			}
			var cmd tea.Cmd
			m.gotoInput, cmd = m.gotoInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "g":
			m.showingGoto = true
			m.gotoInput.Focus()
			return m, textinput.Blink
		case "q", "ctrl+c":
			return m, tea.Quit
		case "s":
			// Single step
			if m.paused {
				// Store current state before step
				m.lastState = CPUState{
					A:  m.cpu.A,
					X:  m.cpu.X,
					Y:  m.cpu.Y,
					PC: m.cpu.PC,
					SP: m.cpu.SP,
					P:  m.cpu.P,
				}
				m.captureMemoryState()
				m.cpu.Step()
				m.relocate()
			}
		case "b":
			// Toggle breakpoint at selected address
			addr := m.locations[m.selectedLocation].PC
			if m.breakpoints[addr] {
				delete(m.breakpoints, addr)
			} else {
				m.breakpoints[addr] = true
			}

		case "n":
			if m.paused && len(m.breakpoints) > 0 {
				m.paused = false
				return m, doStep()
			}

		case "p":
			m.paused = !m.paused

		case "tab":
			if m.activePane == "disasm" {
				m.activePane = "memory"
			} else {
				m.activePane = "disasm"
			}

		case "up":
			if m.activePane == "disasm" {
				m.selectedLocation--
				if m.selectedLocation < 0 {
					m.selectedLocation = 0
				}
			} else {
				if m.memoryAddress >= 8 {
					m.memoryAddress -= 8
					m.captureMemoryState() // Capture state for new memory region
				}
			}
		case "down":
			if m.activePane == "disasm" {
				m.selectedLocation++
				if m.selectedLocation > len(m.locations)-20 {
					m.selectedLocation = len(m.locations) - 20
				}
			} else {
				if m.memoryAddress <= 0xFFF8 {
					m.memoryAddress += 8
					m.captureMemoryState() // Capture state for new memory region
				}
			}

		case "pgup":
			if m.activePane == "disasm" {
				m.selectedLocation -= 20
				if m.selectedLocation < 0 {
					m.selectedLocation = 0
				}
			} else if m.activePane == "memory" {
				// Move memory view up by 64 bytes (8 rows)
				if m.memoryAddress >= 64 {
					m.memoryAddress -= 64
				} else {
					m.memoryAddress = 0
				}
				m.captureMemoryState()
			}
		case "pgdown":
			if m.activePane == "disasm" {
				m.selectedLocation += 20
				if m.selectedLocation > len(m.locations)-20 {
					m.selectedLocation = len(m.locations) - 20
				}
			} else if m.activePane == "memory" {
				// Move memory view down by 64 bytes (8 rows)
				if m.memoryAddress <= 0xFFC0 { // Ensure we don't overflow
					m.memoryAddress += 64
				} else {
					m.memoryAddress = 0xFFC0
				}
				m.captureMemoryState()
			}
		}
	case tea.MouseEvent:
		if msg.Type == tea.MouseLeft {
			// TODO: Add click handling for panel selection
		}
	}
	return m, nil
}

// Format register value with highlighting if changed
func (m Monitor) formatReg8(name string, current, last uint8) string {
	value := fmt.Sprintf("%s: $%02X", name, current)
	if current != last {
		return changedStyle.Render(value)
	}
	return value
}

func (m Monitor) formatReg16(name string, current, last uint16) string {
	value := fmt.Sprintf("%s: $%04X", name, current)
	if current != last {
		return changedStyle.Render(value)
	}
	return value
}

// Format CPU flags with highlighting for changes
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
		current := m.cpu.P&f.flag != 0
		last := m.lastState.P&f.flag != 0

		if current {
			if current != last {
				result.WriteString(changedStyle.Render(f.name + " "))
			} else {
				result.WriteString(f.name + " ")
			}
		} else {
			result.WriteString("- ")
		}
	}
	return result.String()
}

// Disassemble memory around PC
func (m Monitor) disassemble() string {
	var result strings.Builder

	for i := 0; i < 20; i++ {
		offset := m.selectedLocation + i
		l := m.locations[offset]
		line := l.String()
		// Style the line based on whether it's the PC or selected line
		if m.breakpoints[l.PC] {
			if l.PC == m.cpu.PC {
				line = currentLineStyle.Render("● " + line) // Show both current line and breakpoint
			} else {
				line = breakpointStyle.Render("● " + line)
			}
		} else if l.PC == m.cpu.PC {
			line = currentLineStyle.Render(line)
		} else if offset == m.selectedLocation {
			line = selectedLineStyle.Render(line)
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// Show stack contents
func (m Monitor) formatStack() string {
	var result strings.Builder
	for i := uint16(0xFF); i >= uint16(m.cpu.SP); i-- {
		result.WriteString(fmt.Sprintf("$%02X: %02X\n", i, m.mem[0x100+i]))
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

	// Right column: CPU State with change highlighting
	cpuState := infoStyle.Render(fmt.Sprintf(
		"CPU State\n\n%s    %s    %s\n%s  %s\n\nFlags: %s\n",
		m.formatReg8("A", m.cpu.A, m.lastState.A),
		m.formatReg8("X", m.cpu.X, m.lastState.X),
		m.formatReg8("Y", m.cpu.Y, m.lastState.Y),
		m.formatReg16("PC", m.cpu.PC, m.lastState.PC),
		m.formatReg8("SP", m.cpu.SP, m.lastState.SP),
		m.formatFlags(),
	))

	stack := stackStyle.Render(fmt.Sprintf(
		"Stack\n\n%s",
		m.formatStack(),
	))

	memory := memoryStyle.Render(fmt.Sprintf(
		"Memory (↑↓ to scroll)\n\n%s",
		m.formatMemory(),
	))

	// Combine right column elements
	right := lipgloss.JoinVertical(
		lipgloss.Left,
		cpuState,
		stack,
		memory,
	)

	// Help section at the bottom
	var help string
	if !m.paused {
		help = titleStyle.Render(
			"p: pause • q: quit",
		)
	} else {
		help = titleStyle.Render(
			"s: step • n: run to break • p: pause/resume • b: toggle break • " +
				"↑↓: scroll • pgup/pgdn: page • tab: switch pane • g: goto • q: quit",
		)
	}

	// Join columns horizontally with spacing
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		disasm,
		lipgloss.PlaceHorizontal(3, lipgloss.Left, right),
	)

	// Add goto dialog if active
	if m.showingGoto {
		dialog := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Width(30).
			Render(
				"Go to address:\n\n" +
					m.gotoInput.View(),
			)

		return lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			help,
			dialog,
		)
	}

	// Join everything vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func LoadAndSetupBinary(c *cpu.CPU, mem *Memory, filename string, startAddr int) (int, error) {
	// Read the binary file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read binary file: %v", err)
	}

	// Check if the binary will fit in memory
	if int(startAddr)+len(data) > len(mem) {
		return 0, fmt.Errorf("binary file too large for available memory")
	}

	// Copy binary data into CPU memory starting at 0xF000
	for i, b := range data {
		mem[uint16(startAddr)+uint16(i)] = b
	}

	// Set up reset vector at 0xFFFC-0xFFFD to point to 0xF000
	mem[0xFFFC] = 0x00 // Low byte
	mem[0xFFFD] = 0xF0 // High byte

	// Set up IRQ vector at 0xFFFE-0xFFFF to point to 0xF5A4
	mem[0xFFFE] = 0xA4 // Low byte
	mem[0xFFFF] = 0xF5 // High byte

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
	memory := &Memory{}
	c := cpu.NewCPU(memory)
	_, err = LoadAndSetupBinary(c, memory, *inputFile, int(startAddrInt))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	p := tea.NewProgram(NewMonitor(c, memory))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
