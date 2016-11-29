package farsounds

import "github.com/almerlucke/go-farsounds/farsounds/module"

// Osc uses a phasor to do a lookup
type Osc struct {
	*Lookup
	*Phasor
	Amplitude float64
}

// OscModule wraps an osc in a module
type OscModule struct {
	*module.Module
	*Osc
}

// NewOscModule creates a new osc module
func NewOscModule(table WaveTable, phase float64, inc float64, amp float64, buflen int32) *OscModule {
	oscModule := new(OscModule)
	oscModule.Module = module.NewModule(3, 1, buflen, oscDspFunction, nil)
	oscModule.Osc = NewOsc(table, phase, inc, amp)
	return oscModule
}

func oscDspFunction(module interface{}, buflen int32, timestamp int64, samplerate int32) {
	oscModule := module.(*OscModule)

	var pmodInput []float64
	var fmodInput []float64
	var ampInput []float64

	output := oscModule.Outlets[0].Buffer

	if oscModule.Inlets[0].Connections.Len() > 0 {
		pmodInput = oscModule.Inlets[0].Buffer
	}

	if oscModule.Inlets[1].Connections.Len() > 0 {
		fmodInput = oscModule.Inlets[1].Buffer
	}

	if oscModule.Inlets[2].Connections.Len() > 0 {
		ampInput = oscModule.Inlets[2].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		pmod := 0.0

		if pmodInput != nil {
			pmod = pmodInput[i]
		}

		if fmodInput != nil {
			inc := fmodInput[i] / float64(samplerate)
			oscModule.Osc.Inc = inc
		}

		if ampInput != nil {
			amp := ampInput[i]
			oscModule.Osc.Amplitude = amp
		}

		output[i] = oscModule.Next(pmod)
	}
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
