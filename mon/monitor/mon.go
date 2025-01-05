package monitor

import (
	"bytes"
	"fmt"
	"github.com/newhook/6502/c64/t64"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newhook/6502/c64/cia"
	"github.com/newhook/6502/c64/vic"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/dis/disassembler"
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

type continueTick struct{}

func doContinue() tea.Cmd {
	return tea.Tick(1*time.Millisecond, func(t time.Time) tea.Msg {
		return continueTick{}
	})
}

// Monitor represents the UI state
type Monitor struct {
	stepper          Stepper
	mem              cpu.MemoryBus
	cpu              *cpu.CPU
	cia1             *cia.CIA
	cia2             *cia.CIA
	vic              *vic.VIC
	paused           bool
	width            int
	height           int
	region           disassembler.Region
	locationIndex    int
	selectedLocation int

	lastState  CPUState  // Previous CPU state for change detection
	lastMemory [64]uint8 // Only track visible memory (8 rows * 8 bytes)
	lastCIA1   [16]uint8 // Last state of CIA1 registers
	lastCIA2   [16]uint8 // Last state of CIA2 registers

	memoryAddress uint16 // Start address for memory view
	activePane    string // "disasm", "memory"
	gotoInput     textinput.Model
	showingGoto   bool
	gotoDis       bool

	breakpoints map[uint16]bool // Track breakpoint addresses
	nextTo      uint16

	logBuffer       []string // Buffer to store logged lines
	logBufferSize   int      // Maximum number of lines in the buffer
	logScrollIndex  int      // Current scroll position in the buffer
	visibleLogLines int      // Number of lines visible in the output window
}

// Define some basic styles
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	changed   = lipgloss.AdaptiveColor{Light: "#FF6B6B", Dark: "#FF6B6B"}

	leftColumnWidth   = 50 // Fixed width for CPU state and disassembly
	middleColumnWidth = 50 // Fixed width for stack and memory
	rightColumnWidth  = 50 // Fixed width for cia1, cia2, and vic-ii

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			PaddingTop(0).
			PaddingBottom(0).
			PaddingLeft(1).
			PaddingRight(1)

	currentLineStyle = lipgloss.NewStyle().
				Background(highlight).
				Foreground(lipgloss.Color("#ffffff"))

	selectedLineStyle = lipgloss.NewStyle().
				Foreground(highlight)

	changedStyle = lipgloss.NewStyle().
			Foreground(changed).
			Bold(true)

	breakpointStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	cpuStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Width(leftColumnWidth).
			Height(4).
			PaddingLeft(1).
			PaddingRight(1)

	disasmStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Width(leftColumnWidth).
			PaddingLeft(1).
			PaddingRight(1)

	stackStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Height(6).
			Width(middleColumnWidth).
			PaddingLeft(1).
			PaddingRight(1)

	// Add new style for memory panel
	memoryStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Width(middleColumnWidth).
			PaddingLeft(1).
			PaddingRight(1)

	cia1Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Width(rightColumnWidth).
			PaddingLeft(1).
			PaddingRight(1)

	cia2Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Width(rightColumnWidth).
			PaddingLeft(1).
			PaddingRight(1)

	vicStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Padding(1).
			Width(rightColumnWidth)

	outputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Height(5).
			PaddingLeft(1).
			PaddingRight(1)
)

type Stepper interface {
	Step() uint8
}

const (
	nInstructions    = 100
	nInstructionRows = 20
)

// Initialize the monitor
func NewMonitor(stepper Stepper, cpu *cpu.CPU, mem cpu.MemoryBus, cia1 *cia.CIA, cia2 *cia.CIA, vic *vic.VIC) *Monitor {
	ti := textinput.New()
	ti.Placeholder = "Enter hex address (e.g. FF00)"
	ti.CharLimit = 4
	ti.Width = 6

	m := &Monitor{
		stepper:         stepper,
		mem:             mem,
		cpu:             cpu,
		cia1:            cia1,
		cia2:            cia2,
		vic:             vic,
		paused:          true,
		region:          disassembler.DisassembleRegion(mem, 0, nInstructions),
		memoryAddress:   0,
		activePane:      "disasm",
		gotoInput:       ti,
		breakpoints:     make(map[uint16]bool),
		logBuffer:       make([]string, 0),
		logBufferSize:   100, // Maximum of 100 log lines
		logScrollIndex:  0,
		visibleLogLines: 5, // Assume 5 lines fit in the output window; adjust as needed
	}
	m.relocate(0)
	return m
}

func (m *Monitor) captureState() {
	m.captureMemoryState()
	m.captureCIAState()
}

// Helper function to capture current memory view state
func (m *Monitor) captureMemoryState() {
	addr := m.memoryAddress
	for i := 0; i < 64; i++ {
		m.lastMemory[i] = m.mem.Read(addr + uint16(i))
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
			value := m.mem.Read(addr + uint16(col))
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
			value := m.mem.Read(addr + uint16(col))
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
	// Initialize with full-screen mode
	return tea.Batch(tea.EnterAltScreen, tea.ClearScreen)
}

func (m *Monitor) relocate(pc uint16) {
	// Check to see whether the current region of memory has changed.
	region := disassembler.DisassembleRegion(m.mem, m.region.StartAddr, len(m.region.Instructions))
	if !bytes.Equal(m.region.Bytes, region.Bytes) {
		// The memory region has changed.
		// Load instructions at the current pc.
		m.region = disassembler.DisassembleRegion(m.mem, int(m.cpu.PC), nInstructions)
		m.locationIndex = 0
		m.selectedLocation = 0
		return
	}

	// Find the current location of the PC.
	index := -1
	for i, l := range m.region.Instructions {
		if l.Address == pc {
			index = i
		}
	}
	if index == -1 {
		// PC cannot be found.
		m.region = disassembler.DisassembleRegion(m.mem, int(pc), nInstructions)
		m.locationIndex = 0
		m.selectedLocation = 0
	} else {
		m.locationIndex = index
		m.selectedLocation = index
	}
}

func (m *Monitor) Write(p []byte) (n int, err error) {
	line := strings.TrimSpace(string(p))
	// Append the new line to the buffer
	m.logBuffer = append(m.logBuffer, line)

	// Trim the buffer if it exceeds the maximum size
	if len(m.logBuffer) > m.logBufferSize {
		m.logBuffer = m.logBuffer[1:]
	}

	// Scroll to the bottom automatically when adding new lines
	m.logScrollIndex = len(m.logBuffer) - m.visibleLogLines
	if m.logScrollIndex < 0 {
		m.logScrollIndex = 0
	}
	return len(p), nil
}

func (m Monitor) formatLogBuffer() string {
	var result strings.Builder

	// Determine the range of lines to display
	start := m.logScrollIndex
	end := start + m.visibleLogLines
	if end > len(m.logBuffer) {
		end = len(m.logBuffer)
	}

	// Add visible lines to the output
	for _, line := range m.logBuffer[start:end] {
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// Handle keyboard input
func (m *Monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case continueTick:
		if m.paused || m.cpu.PC == m.nextTo || m.breakpoints[m.cpu.PC] {
			m.nextTo = 0
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
		m.captureState()

		// Execute step until we hit a breakpoint
		for cycles := 0; cycles < 10_000; {
			cycles += int(m.stepper.Step())
			if err := m.cpu.Error(); err != nil {
				m.Write([]byte(fmt.Sprintf("Error: %v", err)))
				m.paused = true
				break
			}
			if m.nextTo == m.cpu.PC || m.breakpoints[m.cpu.PC] {
				m.nextTo = 0
				m.paused = true
				break
			}
		}
		m.relocate(m.cpu.PC)
		return m, doContinue()

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
		m.captureState()

		// Execute step
		m.stepper.Step()
		if err := m.cpu.Error(); err != nil {
			m.Write([]byte(fmt.Sprintf("Error: %v", err)))
			m.paused = true
			break
		}
		m.relocate(m.cpu.PC)

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
					if m.gotoDis {
						m.relocate(uint16(addr))
					} else {
						m.memoryAddress = uint16(addr)
					}
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
		case "l":
			//loadAddr, programData, err := LoadProgramFromT64("/Users/matthew/6502/6502/c64emu/ch/CHOPLI-E.T64", 0)
			loadAddr, programData, err := t64.LoadProgramFromT64("roms/CHOPLI-D.T64", 0)
			if err != nil {
				panic(err)
				//return err
			}
			for i, b := range programData {
				m.mem.Write(loadAddr+uint16(i), b)
			}
			m.Write([]byte(fmt.Sprintf("loaded at %d bytes at address %x", len(programData), loadAddr)))

		case "G":
			m.showingGoto = true
			m.gotoDis = true
			m.gotoInput.Focus()
			return m, textinput.Blink

		case "c":
			m.paused = false
			return m, doContinue()

		case "g":
			m.showingGoto = true
			m.gotoInput.Focus()
			return m, textinput.Blink

		case "r":
			// Refresh the screen by clearing and re-rendering
			return m, tea.Batch(tea.ClearScreen)

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
				m.captureState()
				m.stepper.Step()
				if err := m.cpu.Error(); err != nil {
					m.Write([]byte(fmt.Sprintf("Error: %v", err)))
					m.paused = true
					break
				}
				m.relocate(m.cpu.PC)
			}

		case "b":
			// Toggle breakpoint at selected address
			addr := m.region.Instructions[m.selectedLocation].Address
			if m.breakpoints[addr] {
				delete(m.breakpoints, addr)
			} else {
				m.breakpoints[addr] = true
			}

		case "n":
			m.nextTo = m.region.Instructions[m.selectedLocation+1].Address
			m.paused = false
			return m, doContinue()

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
				if m.selectedLocation < len(m.region.Instructions)-1 {
					m.selectedLocation++
				}
			} else {
				if m.memoryAddress <= 0xFFF8 {
					m.memoryAddress += 8
					m.captureMemoryState() // Capture state for new memory region
				}
			}

		case "pgup":
			if m.activePane == "disasm" {
				m.selectedLocation -= nInstructionRows
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
				m.selectedLocation += nInstructionRows
				if m.selectedLocation > len(m.region.Instructions) {
					m.selectedLocation = len(m.region.Instructions) - 1
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

func (m Monitor) formatCPU() string {
	return fmt.Sprintf(
		"%s    %s    %s\n%s  %s\n\nFlags: %s",
		m.formatReg8("A", m.cpu.A, m.lastState.A),
		m.formatReg8("X", m.cpu.X, m.lastState.X),
		m.formatReg8("Y", m.cpu.Y, m.lastState.Y),
		m.formatReg16("PC", m.cpu.PC, m.lastState.PC),
		m.formatReg8("SP", m.cpu.SP, m.lastState.SP),
		m.formatFlags(),
	)
}

// Disassemble memory around PC
func (m Monitor) disassemble() string {
	var result strings.Builder

	top := m.selectedLocation - nInstructionRows
	if top < 0 {
		top = 0
	}

	for i := 0; i < nInstructionRows; i++ {
		offset := i + top
		if offset >= len(m.region.Instructions)-1 {
			break
		}
		l := m.region.Instructions[offset]
		line := l.String()
		// Style the line based on whether it's the PC or selected line
		if m.breakpoints[l.Address] {
			if l.Address == m.cpu.PC {
				line = currentLineStyle.Render("● " + line) // Show both current line and breakpoint
			} else {
				line = breakpointStyle.Render("● " + line)
			}
		} else if l.Address == m.cpu.PC {
			line = currentLineStyle.Render(line)
		} else if offset == m.selectedLocation {
			line = selectedLineStyle.Render(line)
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

func (m Monitor) formatStack() string {
	// Calculate how many stack values we have
	stackSize := uint16(0xFF) - uint16(m.cpu.SP)
	if stackSize > 18 {
		stackSize = 18 // Limit to maximum 18 values
	}

	stackTop := uint16(m.cpu.SP) + stackSize

	var values []string
	for i := stackTop; i > uint16(m.cpu.SP); i-- {
		values = append(values, fmt.Sprintf("$%02X: %02X", i, m.mem.Read(0x100+i)))
	}

	valuesPerCol := 6

	var result strings.Builder
	// Build rows
	for row := 0; row < valuesPerCol; row++ {
		// First column
		if row < len(values) {
			result.WriteString(values[row])
		}

		// Add padding
		result.WriteString("    ")

		// Second column
		if row+valuesPerCol < len(values) {
			result.WriteString(values[row+valuesPerCol])
		}

		// Add padding
		result.WriteString("    ")

		// Third column
		if row+2*valuesPerCol < len(values) {
			result.WriteString(values[row+2*valuesPerCol])
		}

		result.WriteString("\n")
	}

	return result.String()
}

/*
func (m Monitor) formatStack() string {
	var result strings.Builder
	for i := uint16(0xFF); i >= uint16(m.cpu.SP); i-- {
		result.WriteString(fmt.Sprintf("$%02X: %02X\n", i, m.mem.Read(0x100+i)))
	}
	return result.String()
}

*/

func formatBitfield(value uint8, bitNames map[uint8]string) string {
	var result strings.Builder
	for bit, name := range bitNames {
		if value&bit != 0 {
			result.WriteString(fmt.Sprintf("%s: 1\n", name))
		} else {
			result.WriteString(fmt.Sprintf("%s: 0\n", name))
		}
	}
	return result.String()
}

func (m Monitor) formatCIA(c *cia.CIA, lastState [16]uint8) string {
	// Right column: CIA1, CIA2, and VIC-II
	registerNames := map[int]string{
		cia.PRA:       "PRA",
		cia.PRB:       "PRB",
		cia.DDRA:      "DDRA",
		cia.DDRB:      "DDRB",
		cia.TA_LO:     "TA_LO",
		cia.TA_HI:     "TA_HI",
		cia.TB_LO:     "TB_LO",
		cia.TB_HI:     "TB_HI",
		cia.TOD_10THS: "TOD_10",
		cia.TOD_SEC:   "TOD_SEC",
		cia.TOD_MIN:   "TOD_MIN",
		cia.TOD_HR:    "TOD_HR",
		cia.SDR:       "SDR",
		cia.ICR:       "ICR",
		cia.CRA:       "CRA",
		cia.CRB:       "CRB",
	}

	var ciaDetails strings.Builder
	for i := 0; i < len(c.Registers); i += 3 {
		// Get the first register
		name1, reg1 := registerNames[i], c.Registers[i]
		line := fmt.Sprintf("%-8s: $%02X", name1, reg1)
		if reg1 != lastState[i] {
			line = changedStyle.Render(line)
		}

		// Check if there's a second register
		if i+1 < len(c.Registers) {
			name2, reg2 := registerNames[i+1], c.Registers[i+1]
			part := fmt.Sprintf("    %-8s: $%02X", name2, reg2)
			if reg2 != lastState[i+1] {
				part = changedStyle.Render(part)
			}
			line += part
		}

		if i+2 < len(c.Registers) {
			name3, reg3 := registerNames[i+2], c.Registers[i+2]
			part := fmt.Sprintf("    %-8s: $%02X", name3, reg3)
			if reg3 != lastState[i+2] {
				part = changedStyle.Render(part)
			}
			line += part
		}

		ciaDetails.WriteString(line + "\n")
	}

	// Add timer details
	ciaDetails.WriteString(fmt.Sprintf("\nTimer A: %04X Latch: %04X\n", c.TimerA, c.TimerALatch))
	ciaDetails.WriteString(fmt.Sprintf("Timer B: %04X Latch: %04X\n", c.TimerB, c.TimerBLatch))

	// Add bitfield details for ICR, CRA, and CRB
	if false {
		ciaDetails.WriteString("\nICR:\n")
		ciaDetails.WriteString(formatBitfield(c.Registers[cia.ICR], map[uint8]string{
			cia.ICR_TA:   "Timer A Interrupt",
			cia.ICR_TB:   "Timer B Interrupt",
			cia.ICR_TOD:  "TOD Alarm Interrupt",
			cia.ICR_SDR:  "Serial Port Interrupt",
			cia.ICR_FLAG: "FLAG Line Interrupt",
			cia.ICR_SET:  "Set/Clear Flag",
		}))

		ciaDetails.WriteString("\nCRA:\n")
		ciaDetails.WriteString(formatBitfield(c.Registers[cia.CRA], map[uint8]string{
			cia.CRA_START:   "Start Timer A",
			cia.CRA_PBON:    "Timer A Output on PB6",
			cia.CRA_OUTMODE: "Timer A Output Mode",
			cia.CRA_RUNMODE: "Timer A Run Mode",
			cia.CRA_FORCE:   "Force Timer A Load",
			cia.CRA_INMODE:  "Timer A Input Mode",
			cia.CRA_SPMODE:  "Serial Port Mode",
			cia.CRA_TODIN:   "TOD Frequency",
		}))

		ciaDetails.WriteString("\nCRB:\n")
		ciaDetails.WriteString(formatBitfield(c.Registers[cia.CRB], map[uint8]string{
			cia.CRB_START:   "Start Timer B",
			cia.CRB_PBON:    "Timer B Output on PB7",
			cia.CRB_OUTMODE: "Timer B Output Mode",
			cia.CRB_RUNMODE: "Timer B Run Mode",
			cia.CRB_FORCE:   "Force Timer B Load",
			cia.CRB_INMODE:  "Timer B Input Mode",
			cia.CRB_ALARM:   "TOD Alarm",
		}))
	}
	return ciaDetails.String()
}

func (m Monitor) View() string {
	// Left column: CPU state and disassembly
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("CPU"),
		cpuStyle.Render(m.formatCPU()),
		labelStyle.Render("Disassembly"),
		disasmStyle.Render(m.disassemble()),
	)

	// Middle column: Stack and memory
	middleColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("Stack"),
		stackStyle.Render(m.formatStack()),
		labelStyle.Render("Memory (↑↓ to scroll)"),
		memoryStyle.Render(m.formatMemory()),
	)

	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		labelStyle.Render("CIA1"),
		cia1Style.Render(m.formatCIA(m.cia1, m.lastCIA1)),
		labelStyle.Render("CIA2"),
		cia2Style.Render(m.formatCIA(m.cia2, m.lastCIA2)),
		labelStyle.Render("VIC-II"),
		vicStyle.Render("VIC-II\n\n<vic-ii details here>"),
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

	// Combine the columns
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		lipgloss.PlaceHorizontal(3, lipgloss.Left, middleColumn),
		lipgloss.PlaceHorizontal(3, lipgloss.Left, rightColumn),
	)

	// Output window at the bottom
	outputWindow := outputStyle.Width(m.width).Render(m.formatLogBuffer())

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
			lipgloss.Left,
			content,
			labelStyle.Render("Output"),
			outputWindow,
			help,
			dialog,
		)
	}

	// Join everything vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		labelStyle.Render("Output"),
		outputWindow,
		help,
	)
}

func (m *Monitor) captureCIAState() {
	copy(m.lastCIA1[:], m.cia1.Registers[:])
	copy(m.lastCIA2[:], m.cia2.Registers[:])
}
