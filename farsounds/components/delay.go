package components

import (
	"math"

	"github.com/almerlucke/go-farsounds/farsounds"
)

// Delay structure
type Delay struct {
	Buffer    []float64
	WriteHead int
}

// NewDelay create a new delay
func NewDelay(length int) *Delay {
	return &Delay{
		Buffer: make([]float64, length),
	}
}

// Write to delay
func (delay *Delay) Write(sample float64) {
	delay.Buffer[delay.WriteHead] = sample
	delay.WriteHead++
	if delay.WriteHead >= len(delay.Buffer) {
		delay.WriteHead = 0
	}
}

// Read from delay
func (delay *Delay) Read(location float64) float64 {
	buflen := float64(len(delay.Buffer))
	sampleLocation := float64(delay.WriteHead) - location

	for sampleLocation < 0.0 {
		sampleLocation += buflen
	}

	firstIndex, fraction := math.Modf(sampleLocation)
	secondIndex := firstIndex + 1

	if firstIndex >= buflen {
		firstIndex = buflen - 1.0
	}

	if secondIndex >= buflen {
		secondIndex = buflen - 1.0
	}

	buffer := delay.Buffer
	v1 := buffer[int(firstIndex)]
	v2 := buffer[int(secondIndex)]

	return v1 + (v2-v1)*fraction
}

/*
   Delay module
*/

// DelayModule module version of delay
type DelayModule struct {
	// Base module
	*farsounds.BaseModule

	// Delay
	Delay *Delay

	// Read location
	ReadLocation float64
}

// DelayModuleFactory creates new delay modules
func DelayModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	maxDelay := 1.0
	delay := 1.0

	if valueMap, ok := settings.(map[string]interface{}); ok {
		if _maxDelay, ok := valueMap["maxDelay"].(float64); ok {
			maxDelay = _maxDelay
		}
		if _delay, ok := valueMap["delay"].(float64); ok {
			delay = _delay
		}
	}

	module := NewDelayModule(maxDelay, delay, buflen, sr)

	return module, nil
}

// NewDelayModule new delay module
func NewDelayModule(lengthInSeconds float64, readLocationInSeconds float64, buflen int32, sr float64) *DelayModule {
	delayModule := new(DelayModule)
	delayModule.BaseModule = farsounds.NewBaseModule(2, 1, buflen, sr)
	delayModule.Parent = delayModule

	lengthInSamples := int(lengthInSeconds * sr)

	delayModule.Delay = NewDelay(lengthInSamples)

	delayModule.ReadLocation = readLocationInSeconds * sr
	if delayModule.ReadLocation > float64(lengthInSamples) {
		delayModule.ReadLocation = float64(lengthInSamples)
	}

	return delayModule
}

// DSP for delay module
func (module *DelayModule) DSP(timestamp int64) {
	// First call base module dsp
	module.BaseModule.DSP(timestamp)

	buflen := module.GetBufferLength()
	sr := module.GetSampleRate()

	var sampleInput []float64
	var readInput []float64

	output := module.Outlets[0].Buffer

	// Check if inlet is connected for phase modulation
	if module.Inlets[0].Connections.Len() > 0 {
		sampleInput = module.Inlets[0].Buffer
	}

	if module.Inlets[1].Connections.Len() > 0 {
		readInput = module.Inlets[1].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		inSample := 0.0
		readLocation := module.ReadLocation

		if sampleInput != nil {
			inSample = sampleInput[i]
		}

		if readInput != nil {
			readLocation = readInput[i] * sr
		}

		module.Delay.Write(inSample)

		output[i] = module.Delay.Read(readLocation)
	}
}

// Message to module
func (module *DelayModule) Message(message farsounds.Message) {
	sr := module.GetSampleRate()

	if valueMap, ok := message.(map[string]interface{}); ok {
		if delay, ok := valueMap["delay"].(float64); ok {
			module.ReadLocation = delay * sr
			if module.ReadLocation >= float64(len(module.Delay.Buffer)) {
				module.ReadLocation = float64(len(module.Delay.Buffer) - 1)
			}
		}
	}
}
