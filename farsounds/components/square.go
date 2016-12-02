package components

import "github.com/almerlucke/go-farsounds/farsounds"

// Square wave generator
type Square struct {
	*Phasor
	// Amplitude of output
	Amplitude float64
}

// NewSquare new square wave generator
func NewSquare(phase float64, inc float64, amp float64) *Square {
	square := new(Square)
	square.Phasor = NewPhasor(phase, inc)
	square.Amplitude = amp
	return square
}

// Process square wave generator
func (square *Square) Process(phaseMod float64) float64 {
	val := square.Phasor.Process(phaseMod)

	if val >= 0.5 {
		val = -1.0
	} else {
		val = 1.0
	}

	return val * square.Amplitude
}

/*
	Square wave module plus Processor interface methods
*/

// SquareModule is an square wave module
type SquareModule struct {
	// Inherit from BaseModule
	*farsounds.BaseModule
	// Inherit from Osc
	*Square
}

// NewSquareModule creates a new square module
func NewSquareModule(phase float64, inc float64, amp float64, buflen int32) *SquareModule {
	squareModule := new(SquareModule)
	squareModule.BaseModule = farsounds.NewBaseModule(3, 1, buflen)
	squareModule.Parent = squareModule
	squareModule.Square = NewSquare(phase, inc, amp)
	return squareModule
}

// DSP fills output buffer for this square module with samples
func (module *SquareModule) DSP(buflen int32, timestamp int64, samplerate int32) {
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

		output[i] = module.Process(pmod)
	}
}
