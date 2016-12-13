package components

import "github.com/almerlucke/go-farsounds/farsounds"

// Allpass structure
type Allpass struct {
	*Delay
	Feedback float64
}

// NewAllpass create a new allpass
func NewAllpass(length int, feedback float64) *Allpass {
	return &Allpass{
		Delay:    NewDelay(length),
		Feedback: feedback,
	}
}

// Process allpass
func (allpass *Allpass) Process(xn float64, location float64) float64 {
	vm := allpass.Delay.Read(location)
	vn := xn - allpass.Feedback*vm
	yn := vn*allpass.Feedback + vm

	allpass.Delay.Write(vn)

	return yn
}

/*
   Allpass module
*/

// AllpassModule module version of allpass
type AllpassModule struct {
	// Base module
	*farsounds.BaseModule

	// Allpass
	Allpass *Allpass

	// Read location
	ReadLocation float64
}

// AllpassModuleFactory creates new allpass modules
func AllpassModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	maxDelay := 1.0
	delay := 1.0
	feedback := 0.4

	if valueMap, ok := settings.(map[string]interface{}); ok {
		if _maxDelay, ok := valueMap["maxDelay"].(float64); ok {
			maxDelay = _maxDelay
		}
		if _delay, ok := valueMap["delay"].(float64); ok {
			delay = _delay
		}
		if _feedback, ok := valueMap["feedback"].(float64); ok {
			feedback = _feedback
		}
	}

	module := NewAllpassModule(maxDelay, delay, feedback, buflen, sr)

	return module, nil
}

// NewAllpassModule new allpass module
func NewAllpassModule(maxDelay float64, delay float64, feedback float64, buflen int32, sr float64) *AllpassModule {
	allpassModule := new(AllpassModule)
	allpassModule.BaseModule = farsounds.NewBaseModule(3, 1, buflen, sr)
	allpassModule.Parent = allpassModule

	lengthInSamples := int(maxDelay * sr)

	allpassModule.Allpass = NewAllpass(lengthInSamples, feedback)

	allpassModule.ReadLocation = delay * sr
	if allpassModule.ReadLocation > float64(lengthInSamples) {
		allpassModule.ReadLocation = float64(lengthInSamples)
	}

	return allpassModule
}

// DSP for allpass module
func (module *AllpassModule) DSP(timestamp int64) {
	buflen := module.GetBufferLength()
	sr := module.GetSampleRate()

	var sampleInput []float64
	var readInput []float64
	var feedbackInput []float64

	output := module.Outlets[0].Buffer

	// Check if inlet is connected for input
	if module.Inlets[0].Connections.Len() > 0 {
		sampleInput = module.Inlets[0].Buffer
	}

	if module.Inlets[1].Connections.Len() > 0 {
		readInput = module.Inlets[1].Buffer
	}

	if module.Inlets[2].Connections.Len() > 0 {
		feedbackInput = module.Inlets[2].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		inSample := 0.0
		location := module.ReadLocation

		if sampleInput != nil {
			inSample = sampleInput[i]
		}

		if readInput != nil {
			location = readInput[i] * sr
		}

		if feedbackInput != nil {
			module.Allpass.Feedback = feedbackInput[i]
		}

		output[i] = module.Allpass.Process(inSample, location)
	}
}

// Message to module
func (module *AllpassModule) Message(message farsounds.Message) {
	if valueMap, ok := message.(map[string]interface{}); ok {
		if feedback, ok := valueMap["feedback"].(float64); ok {
			module.Allpass.Feedback = feedback
		}
	}
}
