package memory

import "fmt"

const (
	// Memory regions
	BASIC_ROM_START  = 0xA000
	BASIC_ROM_END    = 0xBFFF
	IO_START         = 0xD000
	IO_END           = 0xDFFF
	KERNAL_ROM_START = 0xE000
	KERNAL_ROM_END   = 0xFFFF

	// Memory configuration register
	PROCESSOR_PORT = 0x0001
	PLA_PORT       = 0x0000
)

// MemoryConfig represents different memory configurations based on control lines
type MemoryConfig struct {
	LORAM  bool // BASIC ROM visible
	HIRAM  bool // KERNAL ROM visible
	CHAREN bool // I/O area visible (true) or Character ROM visible (false)
}

type Manager struct {
	ram    [65536]uint8
	basic  [8192]uint8 // 8K BASIC ROM
	kernal [8192]uint8 // 8K KERNAL ROM
	char   [4096]uint8 // 4K Character ROM
	io     [4096]uint8 // 4K I/O area

	// Control registers
	processorPort uint8 // Controls ROM banking
	plaPort       uint8 // Additional banking control
	config        MemoryConfig
}

func NewManager() *Manager {
	m := &Manager{
		// Initialize with default memory configuration (all ROMs visible)
		processorPort: 0x37, // Default value after reset
		plaPort:       0x00,
	}

	// Set initial memory configuration
	m.updateMemoryConfig()
	return m
}

// LoadROM loads ROM data into the specified ROM area
func (m *Manager) LoadROM(data []uint8, romType string) error {
	switch romType {
	case "basic":
		if len(data) != 8192 {
			return fmt.Errorf("BASIC ROM must be 8K, got %d bytes", len(data))
		}
		copy(m.basic[:], data)
	case "kernal":
		if len(data) != 8192 {
			return fmt.Errorf("KERNAL ROM must be 8K, got %d bytes", len(data))
		}
		copy(m.kernal[:], data)
	case "char":
		if len(data) != 4096 {
			return fmt.Errorf("Character ROM must be 4K, got %d bytes", len(data))
		}
		copy(m.char[:], data)
	default:
		return fmt.Errorf("unknown ROM type: %s", romType)
	}
	return nil
}

// Read handles memory reads with banking
func (m *Manager) Read(address uint16) uint8 {
	switch {
	case address == PROCESSOR_PORT:
		return m.processorPort
	case address == PLA_PORT:
		return m.plaPort
	case address >= BASIC_ROM_START && address <= BASIC_ROM_END:
		if m.config.LORAM {
			return m.basic[address-BASIC_ROM_START]
		}
		return m.ram[address]
	case address >= IO_START && address <= IO_END:
		if m.config.CHAREN {
			return m.io[address-IO_START]
		}
		return m.char[address-IO_START]
	case address >= KERNAL_ROM_START && address <= KERNAL_ROM_END:
		if m.config.HIRAM {
			return m.kernal[address-KERNAL_ROM_START]
		}
		return m.ram[address]
	default:
		return m.ram[address]
	}
}

// Write handles memory writes with banking
func (m *Manager) Write(address uint16, value uint8) {
	switch {
	case address == PROCESSOR_PORT:
		m.processorPort = value
		m.updateMemoryConfig()
	case address == PLA_PORT:
		m.plaPort = value
		m.updateMemoryConfig()
	case address >= BASIC_ROM_START && address <= BASIC_ROM_END:
		// Can always write to RAM under ROM
		m.ram[address] = value
	case address >= IO_START && address <= IO_END:
		if m.config.CHAREN {
			m.io[address-IO_START] = value
		} else {
			// Can write to RAM under Character ROM
			m.ram[address] = value
		}
	case address >= KERNAL_ROM_START && address <= KERNAL_ROM_END:
		// Can always write to RAM under ROM
		m.ram[address] = value
	default:
		m.ram[address] = value
	}
}

// updateMemoryConfig updates the memory configuration based on control registers
func (m *Manager) updateMemoryConfig() {
	// Bits 0-2 of processor port control memory configuration
	// Bit 0: LORAM (BASIC ROM control)
	// Bit 1: HIRAM (KERNAL ROM control)
	// Bit 2: CHAREN (I/O vs Character ROM)

	m.config = MemoryConfig{
		LORAM:  (m.processorPort & 0x01) != 0,
		HIRAM:  (m.processorPort & 0x02) != 0,
		CHAREN: (m.processorPort & 0x04) != 0,
	}
}

// WriteIO allows other components (VIC, SID, CIA) to write directly to I/O space
func (m *Manager) WriteIO(offset uint16, value uint8) {
	if offset < 4096 {
		m.io[offset] = value
	}
}

// ReadIO allows other components to read directly from I/O space
func (m *Manager) ReadIO(offset uint16) uint8 {
	if offset < 4096 {
		return m.io[offset]
	}
	return 0
}

// DumpMemory dumps a region of memory for debugging
func (m *Manager) DumpMemory(start uint16, length uint16) []uint8 {
	dump := make([]uint8, length)
	for i := uint16(0); i < length; i++ {
		dump[i] = m.Read(start + i)
	}
	return dump
}

func (m *Manager) DMA(address uint16, data []uint8) {
	for i, value := range data {
		m.Write(address+uint16(i), value)
	}
}
