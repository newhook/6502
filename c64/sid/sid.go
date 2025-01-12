package sid

type Voice struct {
	frequency   uint16
	pulseWidth  uint16
	waveform    uint8
	attack      uint8
	decay       uint8
	sustain     uint8
	release     uint8
	gateEnabled bool
}

type SID struct {
	voices [3]Voice
	volume uint8

	filterCutoff    uint16
	filterResonance uint8
	filterMode      uint8
	filterEnabled   [3]bool
	Clock           int
}

func NewSID() *SID {
	return &SID{}
}

func (s *SID) Update() {
	// Update audio state for the given number of cycles
}

func (s *SID) AddDelta(i int) {

}
