package farsounds

// Osc uses a phasor to do a lookup
type Osc struct {
	*Lookup
	*Phasor
	Amplitude float64
}

// NewOsc creates a new table lookup oscillator
func NewOsc(table []float64, phase float64, inc float64, amp float64) *Osc {
	osc := new(Osc)
	osc.Lookup = NewLookup(table)
	osc.Phasor = NewPhasor(phase, inc)
	osc.Amplitude = amp
	return osc
}

// Next sample please
func (osc *Osc) Next(phaseMod float64) float64 {
	return osc.Look(osc.Phasor.Next(phaseMod)) * osc.Amplitude
}
