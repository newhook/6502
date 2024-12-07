package main

import (
	"flag"
	"fmt"
	"github.com/newhook/6502/as/assembler"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Command line flags
	inputFile := flag.String("i", "", "Input assembly file")
	outputFile := flag.String("o", "", "Output binary file")
	listFile := flag.String("l", "", "Generate listing file")
	flag.Parse()
	*inputFile = "/Users/matthew/6502/6502/AllSuiteA.asm"

	if *inputFile == "" {
		fmt.Println("Error: Input file is required")
		flag.Usage()
		os.Exit(1)
	}

	// If no output file specified, use input filename with .bin extension
	if *outputFile == "" {
		*outputFile = strings.TrimSuffix(*inputFile, filepath.Ext(*inputFile)) + ".bin"
	}

	// Read source file
	source, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Create and run assembler
	as := assembler.NewAssembler()
	err = as.Assemble(string(source))
	if err != nil {
		fmt.Printf("Assembly error: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	err = os.WriteFile(*outputFile, as.GetOutput(), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	// Generate listing file if requested
	if *listFile != "" {
		listing := generateListing(string(source), as)
		err = os.WriteFile(*listFile, []byte(listing), 0644)
		if err != nil {
			fmt.Printf("Error writing listing file: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Successfully assembled %s to %s\n", *inputFile, *outputFile)
	fmt.Printf("Output size: %d bytes\n", len(as.GetOutput()))
}

func generateListing(source string, as *assembler.Assembler) string {
	var listing strings.Builder
	lines := strings.Split(source, "\n")
	addr := uint16(0)

	for _, line := range lines {
		listing.WriteString(fmt.Sprintf("%04X  %-32s %s\n", addr, line, ""))
		// Note: This is a simplified listing - a full implementation would
		// need to track addresses and bytes more accurately
		addr += 1
	}

	return listing.String()
}
