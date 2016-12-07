package components

import (
	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

// Osc uses a phasor to do a lookup
type Osc struct {
	// Lookup table
	*tables.Lookup
	// Phasor for lookup
	*Phasor
	// Amplitude of output
	Amplitude float64
}

// NewOsc creates a new table lookup oscillator
func NewOsc(table []float64, phase float64, inc float64, amp float64) *Osc {
	osc := new(Osc)
	osc.Lookup = tables.NewLookup(table)
	osc.Phasor = NewPhasor(phase, inc)
	osc.Amplitude = amp
	return osc
}

// Process sample please
func (osc *Osc) Process(phaseMod float64) float64 {
	return osc.Look(osc.Phasor.Process(phaseMod)) * osc.Amplitude
}

/*
	Module based oscillator plus Processor interface methods
*/

// OscModule is an oscillator module
type OscModule struct {
	// Inherit from BaseModule
	*farsounds.BaseModule
	// Inherit from Osc
	*Osc
}

// NewOscModule creates a new osc module
func NewOscModule(table tables.WaveTable, phase float64, freq float64, amp float64, buflen int32, sr float64) *OscModule {
	oscModule := new(OscModule)
	oscModule.BaseModule = farsounds.NewBaseModule(3, 1, buflen, sr)
	oscModule.Parent = oscModule
	oscModule.Osc = NewOsc(table, phase, freq/sr, amp)
	return oscModule
}

// OscModuleFactory creates new osc modules
func OscModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	table := tables.SineTable
	phase := 0.0
	freq := 100.0
	amp := 1.0

	module := NewOscModule(table, phase, freq, amp, buflen, sr)

	module.Message(settings)

	return module, nil
}

// DSP fills output buffer for this osc module with samples
func (module *OscModule) DSP(timestamp int64) {
	// First call base module dsp
	module.BaseModule.DSP(timestamp)

	buflen := module.GetBufferLength()
	sr := module.GetSampleRate()

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
			inc := fmodInput[i] / sr
			module.Inc = inc
		}

		if ampInput != nil {
			amp := ampInput[i]
			module.Amplitude = amp
		}

		output[i] = module.Process(pmod)
	}
}

// Message to module
func (module *OscModule) Message(message farsounds.Message) {
	sr := module.GetSampleRate()

	if valueMap, ok := message.(map[string]interface{}); ok {
		if frequency, ok := valueMap["frequency"].(float64); ok {
			module.Inc = frequency / sr
		}

		if phase, ok := valueMap["phase"].(float64); ok {
			module.Phase = phase
		}

		if amplitude, ok := valueMap["amplitude"].(float64); ok {
			module.Amplitude = amplitude
		}

		if tableName, ok := valueMap["table"].(string); ok {
			table, err := farsounds.Registry.GetWaveTable(tableName)
			if err == nil {
				module.Lookup.Table = table
			}
		}
	}
}
