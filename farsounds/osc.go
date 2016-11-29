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

/*
	Module based oscillator plus Processor interface methods
*/

// NewOscModule creates a new osc module
func NewOscModule(table WaveTable, phase float64, inc float64, amp float64, buflen int32) *Module {
	return NewModule(3, 1, buflen, NewOsc(table, phase, inc, amp))
}

// DSP fills output buffer for this osc module with samples
func (osc *Osc) DSP(module *Module, buflen int32, timestamp int64, samplerate int32) {
	var pmodInput []float64
	var fmodInput []float64
	var ampInput []float64

	output := module.Outlets[0].Buffer

	// Check if inlet is connected for phase modulation
	if module.Inlets[0].Connections.Len() > 0 {
		pmodInput = module.Inlets[0].Buffer
	}

	if module.Inlets[1].Connections.Len() > 0 {
		fmodInput = module.Inlets[1].Buffer
	}

	if module.Inlets[2].Connections.Len() > 0 {
		ampInput = module.Inlets[2].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		pmod := 0.0

		if pmodInput != nil {
			pmod = pmodInput[i]
		}

		if fmodInput != nil {
			inc := fmodInput[i] / float64(samplerate)
			osc.Inc = inc
		}

		if ampInput != nil {
			amp := ampInput[i]
			osc.Amplitude = amp
		}

		output[i] = osc.Next(pmod)
	}
}

// Cleanup is mandatory, does nothing for osc
func (osc *Osc) Cleanup(module *Module) {}
