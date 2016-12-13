package farsounds

import (
	"container/list"
	"errors"
	"fmt"
)

/*
	Interfaces
*/

// Voice interface must be implemented by components to be used
// as voice by other components such as polyvoice
type Voice interface {
	// Check if voice is finished releasing
	IsFinished() bool

	// Note on for voice
	NoteOn(float64, float64, interface{})

	// Call note off on voice, voice can still have
	// a release period after this
	NoteOff()
}

// VoiceModule implements voice and module interface
type VoiceModule interface {
	Module
	Voice
}

/*
	Poly factory and module
*/

// PolyVoiceFactory factory for voice modules
type PolyVoiceFactory func(buflen int32, sr float64) VoiceModule

// PolyVoiceModule poly voice module. Play multiple voice modules at the same time,
// no limit to amount of voices, voice can be any module including patches, as
// long as they implement VoiceModule interface. Voices are triggered via Message
// function
type PolyVoiceModule struct {
	// Inherit from BaseModule
	*BaseModule

	// Factory to generate voice modules
	Factory PolyVoiceFactory

	// Free voice pool
	FreeVoicePool *list.List

	// Used voice pool
	UsedVoicePool *list.List
}

type polyVoiceInstance struct {
	voice            VoiceModule
	sampsTillNoteOff int64
	noteOffSend      bool
}

/*
	Poly methods
*/

// NewPolyVoiceModule creates a new osc module
func NewPolyVoiceModule(factory PolyVoiceFactory, numOutlets int, buflen int32, sr float64) *PolyVoiceModule {
	// generate new poly voice module
	polyVoiceModule := new(PolyVoiceModule)
	polyVoiceModule.BaseModule = NewBaseModule(0, numOutlets, buflen, sr)
	polyVoiceModule.Parent = polyVoiceModule
	polyVoiceModule.FreeVoicePool = list.New()
	polyVoiceModule.UsedVoicePool = list.New()
	polyVoiceModule.Factory = factory
	return polyVoiceModule
}

// PolyVoiceModuleFactory creates poly voice modules
func PolyVoiceModuleFactory(settings interface{}, buflen int32, sr float64) (Module, error) {
	factorySettings, ok := settings.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Poly voice settings error %v", settings)
	}

	factoryName, ok := factorySettings["factory"].(string)
	if !ok {
		return nil, errors.New("Poly voice expected a factory name")
	}

	entry := Registry.GetPolyVoiceFactoryEntry(factoryName)
	if entry == nil {
		return nil, fmt.Errorf("Unknown voice factory %v for poly voice", factoryName)
	}

	module := NewPolyVoiceModule(entry.Factory, entry.NumOutlets, buflen, sr)

	return module, nil
}

func (module *PolyVoiceModule) getFreeVoice() *polyVoiceInstance {
	var instance *polyVoiceInstance

	e := module.FreeVoicePool.Front()

	if e == nil {
		// No free module, get a new voice from the factory
		voiceModule := module.Factory(module.GetBufferLength(), module.GetSampleRate())
		instance = new(polyVoiceInstance)
		instance.voice = voiceModule

		// Add the new voice instance to the used voice pool
		module.UsedVoicePool.PushBack(instance)
	} else {
		// Get instance from free list and add to used list
		instance = module.FreeVoicePool.Remove(e).(*polyVoiceInstance)
		module.UsedVoicePool.PushFront(instance)
	}

	return instance
}

// DSP do some dsp
func (module *PolyVoiceModule) DSP(timestamp int64) {
	buflen := module.GetBufferLength()

	// First zero out buffers
	for _, outlet := range module.Outlets {
		buffer := outlet.Buffer

		for i := int32(0); i < buflen; i++ {
			buffer[i] = 0.0
		}
	}

	// Loop through voice modules
	for elem := module.UsedVoicePool.Front(); elem != nil; {
		instance := elem.Value.(*polyVoiceInstance)
		voice := instance.voice

		// Set temp elem, so we can savely remove elem from list
		tmpElem := elem
		elem = elem.Next()

		if voice.IsFinished() {
			module.UsedVoicePool.Remove(tmpElem)
			module.FreeVoicePool.PushBack(instance)
		} else {
			voice.PrepareDSP()
			voice.RequestDSP(timestamp)

			for outletIndex, voiceOutlet := range voice.GetOutlets() {
				voiceBuffer := voiceOutlet.Buffer
				polyBuffer := module.Outlets[outletIndex].Buffer

				for i := int32(0); i < buflen; i++ {
					polyBuffer[i] += voiceBuffer[i]
				}
			}

			if !instance.noteOffSend {
				instance.sampsTillNoteOff -= int64(buflen)
				if instance.sampsTillNoteOff <= 0 {
					voice.NoteOff()
					instance.noteOffSend = true
				}
			}
		}
	}
}

// Message to module
func (module *PolyVoiceModule) Message(message Message) {
	sr := module.GetSampleRate()

	valueMap, ok := message.(map[string]interface{})
	if !ok {
		return
	}

	duration, ok := valueMap["duration"].(float64)
	if !ok {
		return
	}

	settings, ok := valueMap["settings"]
	if !ok {
		return
	}

	sampsTillNoteOff := int64(sr * duration)

	instance := module.getFreeVoice()
	instance.voice.NoteOn(duration, sr, settings)
	instance.noteOffSend = false
	instance.sampsTillNoteOff = sampsTillNoteOff
}
