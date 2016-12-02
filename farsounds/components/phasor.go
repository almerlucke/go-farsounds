package components

// Phasor phases between 0 - 1 with inc steps
type Phasor struct {
	Phase float64
	Inc   float64
}

// NewPhasor creates a new phasor
func NewPhasor(phase float64, inc float64) *Phasor {
	phasor := new(Phasor)
	phasor.Inc = inc
	phasor.Phase = phase
	return phasor
}

// Next sample please
func (phasor *Phasor) Next(phaseMod float64) float64 {
	out := phasor.Phase + phaseMod

	// Clip out between 0.0 and 1.0
	for ; out >= 1.0; out -= 1.0 {
	}

	for ; out < 0.0; out += 1.0 {
	}

	phase := out + phasor.Inc

	// Clip phase between 0.0 and 1.0
	for ; phase >= 1.0; phase -= 1.0 {
	}

	for ; phase < 0.0; phase += 1.0 {
	}

	phasor.Phase = phase

	return out
}
