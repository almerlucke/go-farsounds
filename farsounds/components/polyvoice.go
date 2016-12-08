package components

import (
	"container/list"

	"github.com/almerlucke/go-farsounds/farsounds"
)

type PolyVoiceFactory interface {
	NewVoice() farsounds.Voice
}

type PolyVoiceTriggerGenerator interface {
	GenerateTrigger(timestamp int64, sr float64) bool
}

type PolyVoiceDurationGenerator interface {
	GenerateDuration(timestamp int64, sr float64) float64
}

type PolyVoiceSettingsGenerator interface {
	GenerateSettings(timestamp int64, sr float64) interface{}
}

type PolyVoiceModule struct {
	// Inherit from BaseModule
	*farsounds.BaseModule

	// Factory to generate voice modules
	Factory PolyVoiceFactory

	// Settings generator
	SettingsGenerator PolyVoiceSettingsGenerator

	// Duration generator
	DurationGenerator PolyVoiceDurationGenerator

	// Trigger generator
	TriggerGenerator PolyVoiceTriggerGenerator

	// Pool for voice reuse
	VoicePool *list.List
}

// NewPolyVoiceModule creates a new osc module
func NewPolyVoiceModule(
	factory PolyVoiceFactory,
	triggerGenerator PolyVoiceTriggerGenerator,
	durationGenerator PolyVoiceDurationGenerator,
	settingsGenerator PolyVoiceSettingsGenerator,
	buflen int32, sr float64) *PolyVoiceModule {
	// generate new poly voice module
	polyVoiceModule := new(PolyVoiceModule)
	polyVoiceModule.BaseModule = farsounds.NewBaseModule(2, 2, buflen, sr)
	polyVoiceModule.Parent = polyVoiceModule
	polyVoiceModule.VoicePool = list.New()
	polyVoiceModule.Factory = factory
	polyVoiceModule.SettingsGenerator = settingsGenerator
	polyVoiceModule.TriggerGenerator = triggerGenerator
	polyVoiceModule.DurationGenerator = durationGenerator
	return polyVoiceModule
}
