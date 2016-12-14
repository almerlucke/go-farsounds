package examples

import (
	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components"
)

// SinVoiceModule simple example voice
type SinVoiceModule struct {
	// Inherit from BaseModule
	*farsounds.BaseModule

	// Voice with ADSR and oscillator
	adsr *components.ADSR
	osc  *components.Osc
	pan  float64
}

// NewSinVoiceModule new voice module
func NewSinVoiceModule(buflen int32, sr float64) farsounds.VoiceModule {
	// generate new sin voice module
	sinVoiceModule := new(SinVoiceModule)
	sinVoiceModule.BaseModule = farsounds.NewBaseModule(0, 2, buflen, sr)
	sinVoiceModule.Parent = sinVoiceModule

	sinVoiceModule.adsr = components.NewADSR()
	sinVoiceModule.osc = components.NewOsc(farsounds.SineTable, 0, 100.0/sr, 1.0)
	sinVoiceModule.pan = 0.5

	return sinVoiceModule
}

// DSP for module
func (module *SinVoiceModule) DSP(timestamp int64) {
	buflen := module.GetBufferLength()
	leftOutput := module.Outlets[0].Buffer
	rightOutput := module.Outlets[1].Buffer

	for i := int32(0); i < buflen; i++ {
		value := module.adsr.Process() * module.osc.Process(0)
		left, right := components.SinusoidalPanning(value, module.pan)
		leftOutput[i] = left
		rightOutput[i] = right
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
	settingsMap, ok := settings.(map[string]interface{})
	if !ok {
		return
	}

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

	if pan, ok := settingsMap["pan"].(float64); ok {
		module.pan = pan
	}

	module.osc.Phase = 0.0
}
