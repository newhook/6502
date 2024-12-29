package flow

import (
	"fmt"
	"github.com/newhook/6502/cpu"
	"github.com/newhook/6502/dis/disassembler"
	"strings"
)

type AddressType int

const (
	Unknown AddressType = iota
	Code
	Data
	IndirectTarget
	JumpTable
	JumpTableEntry
)

const MaxJumpTableSize = 256

type AddressInfo struct {
	Type       AddressType
	Referenced bool
	// For tracking cross-references
	ReferencedBy        []uint16
	JumpsTo             []uint16
	IndirectJumpTargets []uint16 // Stores potential indirect jump targets
	JumpTableSize       int      // If this is a jump table, how many entries
	JumpTableBase       uint16   // Start address of the jump table if this is an entry
	IndirectJumpSite    bool     // Track locations of indirect jumps for second pass
	Label               string   // Store generated label if this address needs one
}

type FlowTracer struct {
	memory      cpu.MemoryBus
	addressMap  map[uint16]*AddressInfo
	entryPoints []uint16
}

func NewFlowTracer(memory cpu.MemoryBus) *FlowTracer {
	return &FlowTracer{
		memory:      memory,
		addressMap:  make(map[uint16]*AddressInfo),
		entryPoints: make([]uint16, 0),
	}
}

func (ft *FlowTracer) AddEntryPoint(addr uint16) {
	ft.entryPoints = append(ft.entryPoints, addr)
}

func (ft *FlowTracer) TraceFlow() []disassembler.Location {
	// Start with known entry points
	workList := make([]uint16, len(ft.entryPoints))
	copy(workList, ft.entryPoints)

	var indirectJumps []uint16

	for len(workList) > 0 {
		// Pop next address to analyze
		currentAddr := workList[0]
		workList = workList[1:]

		// Skip if we've already processed this address
		if info, exists := ft.addressMap[currentAddr]; exists && info.Referenced {
			continue
		}

		// Analyze the instruction at this address
		loc := disassembler.DisassembleLocation(ft.memory, int(currentAddr))
		info := ft.markAsCode(currentAddr)
		info.Referenced = true

		if loc.Inst == nil {
			continue
		}

		// Handle different types of control flow
		switch {
		case loc.Inst.Name == "JMP":
			if loc.Inst.Mode == disassembler.Absolute {
				target := uint16(loc.OperandBytes[0]) | uint16(loc.OperandBytes[1])<<8
				info.JumpsTo = append(info.JumpsTo, target)
				ft.addReference(currentAddr, target)
				workList = append(workList, target)
			} else if loc.Inst.Mode == disassembler.Indirect {
				info.IndirectJumpSite = true
				indirectJumps = append(indirectJumps, currentAddr)

				ptrAddr := uint16(loc.OperandBytes[0]) | uint16(loc.OperandBytes[1])<<8

				// Mark the indirect target location as data
				ft.markAddressType(ptrAddr, IndirectTarget)
				ft.markAddressType(ptrAddr+1, IndirectTarget)

				// Try to follow all potential targets in first pass
				targets := ft.scanPotentialJumpTable(ptrAddr)
				for _, target := range targets {
					// Record the indirect jump target
					info.IndirectJumpTargets = append(info.IndirectJumpTargets, target)
					info.JumpsTo = append(info.JumpsTo, target)
					ft.addReference(currentAddr, target)
					workList = append(workList, target)
				}
			}
		case loc.Inst.Name == "JSR":
			// Handle subroutine calls
			target := uint16(loc.OperandBytes[0]) | uint16(loc.OperandBytes[1])<<8
			info.JumpsTo = append(info.JumpsTo, target)
			ft.addReference(currentAddr, target)
			workList = append(workList, target)
			// Continue analysis after the JSR
			workList = append(workList, currentAddr+uint16(loc.Size()))
		case strings.HasPrefix(loc.Inst.Name, "B"):
			// Handle conditional branches
			if loc.Inst.Mode == disassembler.Relative {
				offset := int8(loc.OperandBytes[0])
				target := currentAddr + uint16(loc.Size()) + uint16(offset)
				info.JumpsTo = append(info.JumpsTo, target)
				ft.addReference(currentAddr, target)
				workList = append(workList, target)
				// Also follow the non-taken branch
				workList = append(workList, currentAddr+uint16(loc.Size()))
			}
		default:
			// For normal instructions, continue to next instruction
			if !strings.HasPrefix(loc.Inst.Name, "RTS") && !strings.HasPrefix(loc.Inst.Name, "RTI") {
				workList = append(workList, currentAddr+uint16(loc.Size()))
			}
		}
	}

	// Second pass - analyze potential jump tables
	for _, addr := range indirectJumps {
		loc := disassembler.DisassembleLocation(ft.memory, int(addr))
		ptrAddr := uint16(loc.OperandBytes[0]) | uint16(loc.OperandBytes[1])<<8

		targets := ft.scanPotentialJumpTable(ptrAddr)
		validTargets := 0

		// Validate targets are actually code
		for _, target := range targets {
			if info, exists := ft.addressMap[target]; exists && info.Type == Code {
				validTargets++
			}
		}

		// If we found enough valid targets, mark it as a jump table
		if validTargets >= 2 {
			ft.markAddressType(ptrAddr, JumpTable)
			info := ft.addressMap[ptrAddr]
			info.JumpTableSize = validTargets

			// Validate targets are actually code
			for i, target := range targets {
				if i == 0 {
					continue
				}
				if info, exists := ft.addressMap[target]; exists && info.Type == Code {
					entryAddr := ptrAddr + uint16(i*2)
					ft.markAddressType(entryAddr, JumpTableEntry)
				}
			}
		}
	}

	// Generate labels before creating final output
	ft.generateLabels()

	// Generate final disassembly with flow information
	var result []disassembler.Location
	for addr := uint16(0); addr < disassembler.MaxMemory; addr++ {
		if info, exists := ft.addressMap[addr]; exists && info.Type == Code {
			loc := disassembler.DisassembleLocation(ft.memory, int(addr))
			if loc.Inst != nil {
				// Handle relative addressing specially to use labels
				if loc.Inst.Mode == disassembler.Relative {
					offset := int8(loc.OperandBytes[0])
					target := loc.Address + 2 + uint16(offset)

					// Get target label if it exists
					if info := ft.addressMap[target]; info != nil && info.Label != "" {
						loc.Label = info.Label
					}
				}

				// Handle absolute jumps/calls
				if loc.Inst.Mode == disassembler.Absolute && (strings.HasPrefix(loc.Inst.Name, "JMP") || strings.HasPrefix(loc.Inst.Name, "JSR")) {
					target := uint16(loc.OperandBytes[0]) | uint16(loc.OperandBytes[1])<<8
					if info := ft.addressMap[target]; info != nil && info.Label != "" {
						loc.Label = info.Label
					}
				}
			}
			result = append(result, loc)
			addr += uint16(loc.Size() - 1)
		}
	}

	return result
}

// Helper function to get references to/from an address
func (ft *FlowTracer) GetReferences(addr uint16) (from []uint16, to []uint16) {
	if info, exists := ft.addressMap[addr]; exists {
		return info.ReferencedBy, info.JumpsTo
	}
	return nil, nil
}

// Add helper method to get jump table information
func (ft *FlowTracer) GetJumpTableInfo(addr uint16) (isTable bool, size int, entries []uint16) {
	if info, exists := ft.addressMap[addr]; exists && info.Type == JumpTable {
		entries = make([]uint16, info.JumpTableSize)
		for i := 0; i < info.JumpTableSize; i++ {
			entryAddr := addr + uint16(i*2)
			low := ft.memory.Read(entryAddr)
			high := ft.memory.Read(entryAddr + 1)
			entries[i] = uint16(low) | uint16(high)<<8
		}
		return true, info.JumpTableSize, entries
	}
	return false, 0, nil
}

// Add label generation helper
func (ft *FlowTracer) generateLabels() {
	labelCount := 1

	// First find all addresses that need labels (branch/jump targets)
	for addr, info := range ft.addressMap {
		if info.Type != Code {
			continue
		}

		// If this address is jumped to, it needs a label
		if len(info.ReferencedBy) > 0 {
			if info.Label == "" {
				info.Label = fmt.Sprintf("loc_%04x", addr)
				labelCount++
			}
		}
	}
}

// Add helper to record references
func (ft *FlowTracer) addReference(from, to uint16) {
	info, exists := ft.addressMap[to]
	if !exists {
		info = &AddressInfo{Type: Unknown}
		ft.addressMap[to] = info
	}
	info.ReferencedBy = append(info.ReferencedBy, from)
}

func (ft *FlowTracer) markAsCode(addr uint16) *AddressInfo {
	//fmt.Printf("code: %x\n", addr)
	info, exists := ft.addressMap[addr]
	if !exists {
		info = &AddressInfo{Type: Code}
		ft.addressMap[addr] = info
	}
	info.Type = Code
	return info
}

// Add method to try detecting and scanning a jump table early
func (ft *FlowTracer) scanPotentialJumpTable(addr uint16) []uint16 {
	var targets []uint16

	// Quick scan for what looks like valid addresses
	for i := 0; i < MaxJumpTableSize; i++ {
		currentAddr := addr + uint16(i*2)
		if currentAddr > 0xFFFE {
			break
		}

		target := ft.readIndirectTarget(currentAddr)
		// Basic sanity check on target address
		if target < 0x0200 || target > 0xC000 {
			break
		}

		// Quick check if target contains valid opcode
		opcode := ft.memory.Read(target)
		if _, exists := disassembler.Decode(opcode); !exists {
			break
		}

		targets = append(targets, target)
	}

	return targets
}

// Helper for reading indirect targets (handles page boundary bug)
func (ft *FlowTracer) readIndirectTarget(addr uint16) uint16 {
	// It's a bug in the 6502 that wraps around the LSB without incrementing
	// the MSB. So instead of reading address from 0x02FF-0x0300 you should be
	// looking at 0x02FF-0x0200.
	lowByte := ft.memory.Read(addr)
	var highByte byte
	if addr&0xFF == 0xFF {
		highByte = ft.memory.Read(addr & 0xFF00)
	} else {
		highByte = ft.memory.Read(addr + 1)
	}
	return uint16(lowByte) | uint16(highByte)<<8
}

// Add helper method for marking address types
func (ft *FlowTracer) markAddressType(addr uint16, addrType AddressType) {
	info, exists := ft.addressMap[addr]
	if !exists {
		info = &AddressInfo{Type: addrType}
		ft.addressMap[addr] = info
	}
	info.Type = addrType
}
