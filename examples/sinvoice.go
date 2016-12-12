package examples

import (
	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components"
	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

// SinVoiceModule simple example voice
type SinVoiceModule struct {
	// Inherit from BaseModule
	*farsounds.BaseModule

	// Voice with ADSR and oscillator
	adsr *components.ADSR
	osc  *components.Osc
}

// NewSinVoiceModule new voice module
func NewSinVoiceModule(buflen int32, sr float64) farsounds.VoiceModule {
	// generate new sin voice module
	sinVoiceModule := new(SinVoiceModule)
	sinVoiceModule.BaseModule = farsounds.NewBaseModule(0, 1, buflen, sr)
	sinVoiceModule.Parent = sinVoiceModule

	sinVoiceModule.adsr = components.NewADSR()
	sinVoiceModule.osc = components.NewOsc(tables.SineTable, 0, 100.0/sr, 1.0)

	return sinVoiceModule
}

// DSP for module
func (module *SinVoiceModule) DSP(timestamp int64) {
	module.BaseModule.DSP(timestamp)

	buflen := module.GetBufferLength()
	output := module.Outlets[0].Buffer

	for i := int32(0); i < buflen; i++ {
		output[i] = module.adsr.Process() * module.osc.Process(0)
	}
}

// IsFinished check
func (module *SinVoiceModule) IsFinished() bool {
	return module.adsr.Idle()
}

// NoteOff action
func (module *SinVoiceModule) NoteOff() {
	module.adsr.Gate(0.0)
}

// NoteOn action
func (module *SinVoiceModule) NoteOn(duration float64, sr float64, settings interface{}) {
	settingsMap := settings.(map[string]interface{})
	module.adsr.SetAttackRate(0.1 * sr)
	module.adsr.SetDecayRate(0.1 * sr)
	module.adsr.SetReleaseRate(1.4 * sr)
	module.adsr.SetSustainLevel(0.3)
	module.adsr.Gate(1.0)

	if frequency, ok := settingsMap["frequency"].(float64); ok {
		module.osc.Inc = frequency / sr
	}

	if amplitude, ok := settingsMap["amplitude"].(float64); ok {
		module.osc.Amplitude = amplitude
	}

	module.osc.Phase = 0.0
}
