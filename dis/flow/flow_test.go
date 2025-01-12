package flow

import (
	"github.com/newhook/6502/as/assembler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

// MockMemory implements cpu.MemoryBus for testing
type MockMemory struct {
	data map[uint16]byte
}

func NewMockMemory() *MockMemory {
	return &MockMemory{
		data: make(map[uint16]byte),
	}
}

func (m *MockMemory) Read(addr uint16) byte {
	return m.data[addr]
}

func (m *MockMemory) Write(addr uint16, val byte) {
	m.data[addr] = val
}

// Helper function to assemble code into memory
func assembleToMemory(t *testing.T, source string) *MockMemory {
	as := assembler.NewAssembler()
	err := as.Assemble(source)
	require.NoError(t, err, "Assembly should succeed")

	mem := NewMockMemory()
	code := as.GetOutput()
	org := uint16(0x600)
	for addr, val := range code {
		mem.Write(org+uint16(addr), val)
	}

	//s := disassembler.DisassembleMemory(mem, int(org), len(code))
	//fmt.Print(s)
	//
	return mem
}

func TestBasicFlowTracing(t *testing.T) {
	source := `
        .org $0600
        LDA #$01
        BEQ done
        STA $02
        JMP other
    done:
        NOP
        RTS
    other:
        NOP
        RTS
    `
	mem := assembleToMemory(t, source)

	tracer := NewFlowTracer(mem)
	tracer.AddEntryPoint(0x0600)
	tracer.TraceFlow()

	// Verify basic flow tracing
	refs, _ := tracer.GetReferences(0x0600)
	assert.Empty(t, refs, "Entry point should have no references")

	// Check branch paths are followed
	info := tracer.addressMap[0x0602]
	require.NotNil(t, info)
	assert.Contains(t, info.JumpsTo, uint16(0x0609), "BEQ target should be tracked")

	// Check jump targets
	info = tracer.addressMap[0x0606]
	require.NotNil(t, info)
	assert.Contains(t, info.JumpsTo, uint16(0x060B), "JMP target should be tracked")

	// Verify all instructions were found
	expectedAddrs := []uint16{0x0600, 0x0602, 0x0604, 0x0606, 0x0609, 0x060A, 0x060B, 0x060C}
	for _, addr := range expectedAddrs {
		info := tracer.addressMap[addr]
		assert.NotNil(t, info, "Address %04X should be marked as code", addr)
		assert.Equal(t, Code, info.Type)
	}
}

func TestIndirectJump(t *testing.T) {
	source := `
        .org $0600
        JMP (vector)
        BRK             ; Should not reach
        NOP            ; Target of indirect jump
        RTS
        
        .org $0610
    vector:
        .word target   ; Jump target address
        
        .org $0604
    target:
        NOP
        RTS
    `
	mem := assembleToMemory(t, source)

	tracer := NewFlowTracer(mem)
	tracer.AddEntryPoint(0x0600)
	tracer.TraceFlow()

	// Check indirect jump handling
	info := tracer.addressMap[0x0600]
	require.NotNil(t, info)
	assert.Contains(t, info.IndirectJumpTargets, uint16(0x0604), "Indirect jump target should be tracked")

	// Verify indirect target storage is marked
	info = tracer.addressMap[0x0610]
	require.NotNil(t, info)
	assert.Equal(t, IndirectTarget, info.Type, "Indirect jump vector should be marked as data")

	// Verify target is traced
	info = tracer.addressMap[0x0604]
	require.NotNil(t, info)
	assert.Equal(t, Code, info.Type, "Indirect jump target should be marked as code")
}

func TestJumpTableDetection(t *testing.T) {
	source := `
        .org $0600
        JMP (jumptable)
        BRK
        
    routine1:
        NOP            ; Target 1
        RTS
    routine2:
        LDA #$00       ; Target 2
        RTS
    routine3:
        STA $02        ; Target 3
        RTS
        
        .org $0610
    jumptable:
        .word routine1  ; Jump table entry 1
        .word routine2  ; Jump table entry 2
        .word routine3  ; Jump table entry 3
    `
	mem := assembleToMemory(t, source)

	tracer := NewFlowTracer(mem)
	tracer.AddEntryPoint(0x0600)
	tracer.TraceFlow()

	// Check jump table detection
	isTable, size, entries := tracer.GetJumpTableInfo(0x0610)
	assert.True(t, isTable, "Should detect jump table")
	assert.Equal(t, 3, size, "Jump table should have 3 entries")
	assert.Equal(t, []uint16{0x0604, 0x0606, 0x0609}, entries, "Jump table entries should match")

	// Verify table entries are marked correctly
	info := tracer.addressMap[0x0610]
	require.NotNil(t, info)
	assert.Equal(t, JumpTable, info.Type)
	assert.Equal(t, 3, info.JumpTableSize)

	// Check that all targets are traced
	for _, target := range []uint16{0x0604, 0x0606, 0x0609} {
		info := tracer.addressMap[target]
		require.NotNil(t, info)
		assert.Equal(t, Code, info.Type, "Jump table target should be marked as code")
	}
}

func TestJumpTablePageBoundary(t *testing.T) {
	source := `
        .org $0600
        JMP (pagetable)
        
    target1:
        NOP
        RTS
    target2:
        NOP
        RTS
        
        .org $06FF
    pagetable:
        .word target1   ; Will cross page boundary
        .word target2
    `
	mem := assembleToMemory(t, source)

	tracer := NewFlowTracer(mem)
	tracer.AddEntryPoint(0x0600)
	tracer.TraceFlow()

	// Verify page boundary handling
	isTable, size, _ := tracer.GetJumpTableInfo(0x06FF)
	assert.True(t, isTable, "Should detect jump table across page boundary")
	assert.Equal(t, 2, size, "Jump table should have 2 entries")

	// Check that high byte is read from correct address
	info := tracer.addressMap[0x0600]
	require.NotNil(t, info)
	// target1 is $0603, target2 is $0605
	assert.Contains(t, info.IndirectJumpTargets, uint16(0x0605),
		"Should correctly read target across page boundary")
	assert.Contains(t, info.IndirectJumpTargets, uint16(0x6C03),
		"Should correctly read target across page boundary")
}

func TestLabelGeneration(t *testing.T) {
	source := `
        .org $0600
        start:
            LDA #$01
            BEQ done    ; Should generate label
            STA $02
            JMP other   ; Should generate label
        done:
            NOP
            RTS
        other:
            NOP
            RTS
    `
	mem := assembleToMemory(t, source)
	tracer := NewFlowTracer(mem)
	tracer.AddEntryPoint(0x0600)
	instructions := tracer.TraceFlow()

	// Verify labels were generated
	found := make(map[string]bool)
	for _, inst := range instructions {
		if strings.Contains(inst.String(), "loc_") {
			found[strings.Split(inst.String(), ":")[0]] = true
		}
	}

	// Should have labels for 'done' and 'other'
	require.Equal(t, 2, len(found), "Should have generated 2 labels")

	// Verify instructions use labels
	var hasLabelledBranch, hasLabelledJump bool
	for _, inst := range instructions {
		if strings.Contains(inst.String(), "BEQ loc_") {
			hasLabelledBranch = true
		}
		if strings.Contains(inst.String(), "JMP loc_") {
			hasLabelledJump = true
		}
	}

	assert.True(t, hasLabelledBranch, "Branch should use label")
	assert.True(t, hasLabelledJump, "Jump should use label")
}
