package t64

import (
	"encoding/binary"
	"fmt"
	"os"
)

// T64Header represents the T64 tape container header
type T64Header struct {
	Signature   [32]byte // "C64 tape image file" signature
	Version     uint16   // Version number (usually $0100)
	MaxEntries  uint16   // Maximum number of entries
	UsedEntries uint16   // Currently used entries
	Name        [24]byte // Tape container name
	Reserved    [4]byte  // Reserved bytes
}

// T64Entry represents a directory entry in the T64 container
type T64Entry struct {
	EntryType  byte     // $00=free, $01=normal, $02=memory snapshot
	FileType   byte     // $00=free, $01=normal tape file, $02=tape file with header
	StartAddr  uint16   // Start address in C64 memory
	EndAddr    uint16   // End address in C64 memory
	Reserved   uint16   // Reserved
	DataOffset uint32   // Offset to data in T64 file
	Reserved2  [4]byte  // More reserved bytes
	Filename   [16]byte // C64 filename, PETSCII
}

func loadT64(filename string) (*T64Header, []T64Entry, []byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read T64 file: %v", err)
	}

	if len(data) < 64 { // Minimum size for header
		return nil, nil, nil, fmt.Errorf("invalid T64 file: too short")
	}

	// Parse header
	header := &T64Header{}
	copy(header.Signature[:], data[0:32])
	header.Version = binary.LittleEndian.Uint16(data[32:34])
	header.MaxEntries = binary.LittleEndian.Uint16(data[34:36])
	header.UsedEntries = binary.LittleEndian.Uint16(data[36:38])
	copy(header.Name[:], data[40:64])

	// Read directory entries
	entries := make([]T64Entry, header.UsedEntries)
	offset := 64 // Directory starts after header

	for i := uint16(0); i < header.UsedEntries; i++ {
		entry := &entries[i]
		entryData := data[offset : offset+32]

		entry.EntryType = entryData[0]
		entry.FileType = entryData[1]
		entry.StartAddr = binary.LittleEndian.Uint16(entryData[2:4])
		entry.EndAddr = binary.LittleEndian.Uint16(entryData[4:6])
		entry.DataOffset = binary.LittleEndian.Uint32(entryData[8:12])
		copy(entry.Filename[:], entryData[16:32])

		offset += 32
	}

	return header, entries, data, nil
}

// LoadProgramFromT64 loads a specific program from the T64 image
func LoadProgramFromT64(filename string, programIndex uint16) (uint16, []byte, error) {
	header, entries, data, err := loadT64(filename)
	if err != nil {
		return 0, nil, err
	}

	if programIndex >= header.UsedEntries {
		return 0, nil, fmt.Errorf("program index out of range")
	}

	entry := entries[programIndex]
	if entry.EntryType == 0 {
		return 0, nil, fmt.Errorf("empty entry")
	}

	// Calculate program size
	size := entry.EndAddr - entry.StartAddr

	// Extract program data
	programData := make([]byte, size)
	copy(programData, data[entry.DataOffset:entry.DataOffset+uint32(size)])

	return entry.StartAddr, programData, nil
}

// Helper function to convert PETSCII filename to string
func petsciiToString(petscii []byte) string {
	result := make([]byte, len(petscii))
	for i, c := range petscii {
		if c == 0 {
			return string(result[:i])
		}
		// Convert PETSCII to ASCII (simplified)
		if c >= 0x41 && c <= 0x5A {
			c += 0x20
		} else if c >= 0x61 && c <= 0x7A {
			c -= 0x20
		}
		result[i] = c
	}
	return string(result)
}

// ListT64Contents prints the contents of a T64 file
func ListT64Contents(filename string) error {
	header, entries, _, err := loadT64(filename)
	if err != nil {
		return err
	}

	fmt.Printf("T64 Name: %s\n", petsciiToString(header.Name[:]))
	fmt.Printf("Entries: %d/%d\n\n", header.UsedEntries, header.MaxEntries)

	for i, entry := range entries {
		if entry.EntryType != 0 {
			fmt.Printf("%d: %s (Start: $%04X, End: $%04X)\n",
				i,
				petsciiToString(entry.Filename[:]),
				entry.StartAddr,
				entry.EndAddr)
		}
	}
	return nil
}
