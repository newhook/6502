package d64

import (
	"encoding/binary"
	"fmt"
	"os"
)

// D64 constants
const (
	SectorSize    = 256
	TracksPerDisk = 35
	// Track 1-17: 21 sectors, 18-24: 19 sectors, 25-30: 18 sectors, 31-35: 17 sectors

	// File types
	FileTypeSCRATCHED = 0
	FileTypeSEQ       = 1
	FileTypePRG       = 2
	FileTypeUSR       = 3
	FileTypeREL       = 4
	FileTypeCBM       = 5
	FileTypeDIR       = 6

	// File flags
	FileFlagLocked = 0x40
	FileFlagClosed = 0x80
)

var (
	SectorsPerTrack = [TracksPerDisk]int{
		21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
		19, 19, 19, 19, 19, 19, 19,
		18, 18, 18, 18, 18, 18,
		17, 17, 17, 17, 17,
	}
)

// D64File represents a D64 disk image
type D64File struct {
	data     []byte
	filename string
}

// LoadD64 reads a D64 file and returns a D64File struct
func LoadD64(filename string) (*D64File, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read D64 file: %w", err)
	}

	// Validate file size (174848 bytes for standard D64)
	if len(data) != 174848 {
		return nil, fmt.Errorf("invalid D64 file size: %d bytes", len(data))
	}

	return &D64File{
		data:     data,
		filename: filename,
	}, nil
}

// ReadPRG reads a PRG file from the disk given its directory entry
func (d *D64File) ReadPRG(entry DirEntry) ([]byte, error) {
	if entry.FileType != 2 { // PRG files have type 2
		return nil, fmt.Errorf("not a PRG file")
	}

	var data []byte
	track, sector := entry.Track, entry.Sector

	// Follow the chain of sectors
	for {
		sectorData, err := d.ReadSector(track, sector)
		if err != nil {
			return nil, fmt.Errorf("error reading track %d sector %d: %w", track, sector, err)
		}

		// First two bytes are next track/sector
		nextTrack := int(sectorData[0])
		nextSector := int(sectorData[1])

		// Append the actual file data (skip track/sector link)
		if nextTrack == 0 {
			// Last sector - only append valid bytes
			bytesToCopy := nextSector - 1
			if bytesToCopy > 0 {
				data = append(data, sectorData[2:2+bytesToCopy]...)
			}
			break
		} else {
			// Regular sector - append all bytes after track/sector link
			data = append(data, sectorData[2:]...)
		}

		track, sector = nextTrack, nextSector
	}

	return data, nil
}

// ReadPRGWithAddress reads a PRG file and returns both the load address and data
func (d *D64File) ReadPRGWithAddress(entry DirEntry) (uint16, []byte, error) {
	data, err := d.ReadPRG(entry)
	if err != nil {
		return 0, nil, err
	}

	if len(data) < 2 {
		return 0, nil, fmt.Errorf("PRG file too short")
	}

	loadAddress := uint16(data[0]) | uint16(data[1])<<8
	return loadAddress, data[2:], nil
}

// ReadSector reads a specific track/sector combination
func (d *D64File) ReadSector(track, sector int) ([]byte, error) {
	if track < 1 || track > TracksPerDisk {
		return nil, fmt.Errorf("invalid track number: %d", track)
	}

	if sector < 0 || sector >= SectorsPerTrack[track-1] {
		return nil, fmt.Errorf("invalid sector number: %d for track %d", sector, track)
	}

	// Calculate offset in the D64 file
	offset := 0
	for t := 1; t < track; t++ {
		offset += SectorsPerTrack[t-1] * SectorSize
	}
	offset += sector * SectorSize

	return d.data[offset : offset+SectorSize], nil
}

// ReadDirectory reads the directory track (18) and returns file entries
func (d *D64File) ReadDirectory() ([]DirEntry, error) {
	sector := 1 // Directory starts at track 18, sector 1
	var entries []DirEntry

	for {
		data, err := d.ReadSector(18, sector)
		if err != nil {
			return nil, err
		}

		// Process 8 directory entries per sector
		for i := 0; i < 8; i++ {
			entry := data[i*32 : (i+1)*32]

			// Check if this is a valid file entry
			fileType := entry[2] & 0x07
			if fileType == 0 {
				continue // Skip empty entries
			}

			// Extract filename (padded with $A0)
			filename := make([]byte, 16)
			copy(filename, entry[5:21])

			// Create directory entry
			dirEntry := DirEntry{
				FileType: int(fileType),
				Track:    int(entry[3]),
				Sector:   int(entry[4]),
				Blocks:   int(binary.LittleEndian.Uint16(entry[30:32])),
				Name:     trimPETSCII(filename),
			}
			entries = append(entries, dirEntry)
		}

		// Check if there are more directory sectors to read
		if data[0] == 0 {
			break // No more sectors
		}
		sector = int(data[1])
	}

	return entries, nil
}

func trimPETSCII(name []byte) string {
	end := len(name)
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] != 0xA0 {
			break
		}
		end = i
	}
	return string(name[:end])
}

// DirEntry represents a directory entry in the D64 image
type DirEntry struct {
	FileType int
	Track    int
	Sector   int
	Blocks   int
	Name     string
}
