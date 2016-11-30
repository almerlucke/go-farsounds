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

// OscModule is an oscillator module
type OscModule struct {
	*BaseModule
	*Osc
}

// NewOscModule creates a new osc module
func NewOscModule(table WaveTable, phase float64, inc float64, amp float64, buflen int32) *OscModule {
	return &OscModule{
		BaseModule: NewBaseModule(3, 1, buflen),
		Osc:        NewOsc(table, phase, inc, amp),
	}
}

// DSP fills output buffer for this osc module with samples
func (module *OscModule) DSP(buflen int32, timestamp int64, samplerate int32) {
	// First call base module dsp
	module.BaseModule.DSP(buflen, timestamp, samplerate)

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
			module.Inc = inc
		}

		if ampInput != nil {
			amp := ampInput[i]
			module.Amplitude = amp
		}

		output[i] = module.Next(pmod)
	}
}
