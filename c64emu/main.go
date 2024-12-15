package main

import (
	"github.com/newhook/6502/c64/c64"
	"log"
	"os"
)

func main() {
	computer, err := c64.NewC64()
	if err != nil {
		log.Fatal(err)
	}
	do := func() error {
		mem := computer.Memory
		// Load ROMs
		basicROM, err := os.ReadFile("basic-901226-01.bin")
		if err != nil {
			return err
		}
		kernalROM, err := os.ReadFile("kernal-901227-03.bin")
		if err != nil {
			return err
		}
		charROM, err := os.ReadFile("chargen-901225-01.bin")
		if err != nil {
			return err
		}
		if err := mem.LoadROM(basicROM, "basic"); err != nil {
			return err
		}
		if err := mem.LoadROM(kernalROM, "kernal"); err != nil {
			return err
		}
		if err := mem.LoadROM(charROM, "char"); err != nil {
			return err
		}

		mem.Map()

		// Initialize CPU registers
		// Reset vector
		computer.CPU.PC = uint16(mem.Read(0xFFFC)) | uint16(mem.Read(0xFFFD))<<8

		//p := tea.NewProgram(monitor.NewMonitor(computer, computer.CPU, computer.Memory))
		//if _, err := p.Run(); err != nil {
		//	return err
		//}

		// Main emulation loop
		for computer.IsRunning() {
			computer.Step()

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
