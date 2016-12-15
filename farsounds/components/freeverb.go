package components

import (
	"math"

	"github.com/almerlucke/go-farsounds/farsounds"
)

/*
   Constants
*/

const (
	minPositiveNormalFloat = 2.225073858507201e-308
	numcombs               = 8
	numallpasses           = 4
	muted                  = 0
	fixedgain              = 0.015
	scalewet               = 3
	scaledry               = 2
	scaledamp              = 0.4
	scaleroom              = 0.28
	offsetroom             = 0.7
	initialroom            = 0.5
	initialdamp            = 0.5
	initialwet             = 0.4 / scalewet
	initialdry             = 0.2
	initialwidth           = 1.0
	initialmode            = 0
	initialfeedback        = 0.5
	freezemode             = 0.5
	stereospread           = 23
	combtuningL1           = 1116
	combtuningR1           = combtuningL1 + stereospread
	combtuningL2           = 1188
	combtuningR2           = combtuningL2 + stereospread
	combtuningL3           = 1277
	combtuningR3           = combtuningL3 + stereospread
	combtuningL4           = 1356
	combtuningR4           = combtuningL4 + stereospread
	combtuningL5           = 1422
	combtuningR5           = combtuningL5 + stereospread
	combtuningL6           = 1491
	combtuningR6           = combtuningL6 + stereospread
	combtuningL7           = 1557
	combtuningR7           = combtuningL7 + stereospread
	combtuningL8           = 1617
	combtuningR8           = combtuningL8 + stereospread
	allpasstuningL1        = 556
	allpasstuningR1        = allpasstuningL1 + stereospread
	allpasstuningL2        = 441
	allpasstuningR2        = allpasstuningL2 + stereospread
	allpasstuningL3        = 341
	allpasstuningR3        = allpasstuningL3 + stereospread
	allpasstuningL4        = 225
	allpasstuningR4        = allpasstuningL4 + stereospread
)

/*
   undenormalise
*/

func undenormalise(val float64) float64 {
	aval := math.Abs(val)
	if aval < minPositiveNormalFloat {
		return 0.0
	}

	return val
}

/*
   Allpass
*/
type freeVerbAllpass struct {
	buffer   []float64
	bufidx   int
	feedback float64
}

func newFreeVerbAllpass(buflen int, feedback float64) *freeVerbAllpass {
	allpass := new(freeVerbAllpass)
	allpass.buffer = make([]float64, buflen)
	allpass.feedback = feedback
	return allpass
}

func (allpass *freeVerbAllpass) mute() {
	buffer := allpass.buffer
	for i := 0; i < len(buffer); i++ {
		buffer[i] = 0.0
	}
}

func (allpass *freeVerbAllpass) process(input float64) float64 {
	buffer := allpass.buffer
	bufidx := allpass.bufidx
	bufout := undenormalise(buffer[bufidx])
	output := -input + bufout
	buffer[bufidx] = input + (bufout * allpass.feedback)
	bufidx++
	if bufidx >= len(buffer) {
		bufidx = 0
	}
	allpass.bufidx = bufidx
	return output
}

/*
   Comb
*/

type freeVerbComb struct {
	feedback    float64
	filterstore float64
	damp1       float64
	damp2       float64
	buffer      []float64
	bufidx      int
}

func newFreeVerbComb(buflen int, feedback float64) *freeVerbComb {
	comb := new(freeVerbComb)
	comb.buffer = make([]float64, buflen)
	comb.feedback = feedback
	comb.setDamp(initialdamp)
	return comb
}

func (comb *freeVerbComb) setDamp(val float64) {
	comb.damp1 = val
	comb.damp2 = 1.0 - val
}

func (comb *freeVerbComb) mute() {
	buffer := comb.buffer
	for i := 0; i < len(buffer); i++ {
		buffer[i] = 0.0
	}
}

func (comb *freeVerbComb) process(input float64) float64 {
	buffer := comb.buffer
	bufidx := comb.bufidx
	filterstore := comb.filterstore
	output := undenormalise(buffer[bufidx])
	filterstore = undenormalise(output*comb.damp2 + filterstore*comb.damp1)
	buffer[bufidx] = input + filterstore*comb.feedback
	bufidx++
	if bufidx >= len(buffer) {
		bufidx = 0
	}
	comb.bufidx = bufidx
	comb.filterstore = filterstore
	return output
}

/*
   Module
*/

// FreeVerbModule module
type FreeVerbModule struct {
	// Base module
	*farsounds.BaseModule

	// Freeverb
	combL     []*freeVerbComb
	combR     []*freeVerbComb
	allpassL  []*freeVerbAllpass
	allpassR  []*freeVerbAllpass
	gain      float64
	roomsize  float64
	roomsize1 float64
	damp      float64
	damp1     float64
	wet       float64
	wet1      float64
	wet2      float64
	dry       float64
	width     float64
	mode      float64
}

func (freeverb *FreeVerbModule) setWet(wet float64) {
	freeverb.wet = wet * scalewet
	freeverb.update()
}

func (freeverb *FreeVerbModule) setRoomSize(roomsize float64) {
	freeverb.roomsize = (roomsize * scaleroom) + offsetroom
	freeverb.update()
}

func (freeverb *FreeVerbModule) setDry(dry float64) {
	freeverb.dry = dry * scaledry
}

func (freeverb *FreeVerbModule) setDamp(damp float64) {
	freeverb.damp = damp * scaledamp
	freeverb.update()
}

func (freeverb *FreeVerbModule) setWidth(width float64) {
	freeverb.width = width
	freeverb.update()
}

func (freeverb *FreeVerbModule) setMode(mode float64) {
	freeverb.mode = mode
	freeverb.update()
}

func (freeverb *FreeVerbModule) update() {
	freeverb.wet1 = freeverb.wet * (freeverb.width/2.0 + 0.5)
	freeverb.wet2 = freeverb.wet * ((1.0 - freeverb.width) / 2.0)

	if freeverb.mode >= freezemode {
		freeverb.roomsize1 = 1
		freeverb.damp1 = 0
		freeverb.gain = muted
	} else {
		freeverb.roomsize1 = freeverb.roomsize
		freeverb.damp1 = freeverb.damp
		freeverb.gain = fixedgain
	}

	for i := 0; i < numcombs; i++ {
		freeverb.combL[i].feedback = freeverb.roomsize1
		freeverb.combR[i].feedback = freeverb.roomsize1
		freeverb.combL[i].setDamp(freeverb.damp1)
		freeverb.combR[i].setDamp(freeverb.damp1)
	}
}

// FreeVerbModuleFactory factory
func FreeVerbModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	module := NewFreeVerbModule(buflen, sr)
	module.Message(settings)
	return module, nil
}

// NewFreeVerbModule generate new freeverb module
func NewFreeVerbModule(buflen int32, sr float64) *FreeVerbModule {
	scale := sr / 44100.0
	freeverb := new(FreeVerbModule)
	freeverb.BaseModule = farsounds.NewBaseModule(2, 2, buflen, sr)
	freeverb.Parent = freeverb

	freeverb.combL = make([]*freeVerbComb, numcombs)
	freeverb.combR = make([]*freeVerbComb, numcombs)
	freeverb.combL[0] = newFreeVerbComb(int(combtuningL1*scale), initialfeedback)
	freeverb.combR[0] = newFreeVerbComb(int(combtuningR1*scale), initialfeedback)
	freeverb.combL[1] = newFreeVerbComb(int(combtuningL2*scale), initialfeedback)
	freeverb.combR[1] = newFreeVerbComb(int(combtuningR2*scale), initialfeedback)
	freeverb.combL[2] = newFreeVerbComb(int(combtuningL3*scale), initialfeedback)
	freeverb.combR[2] = newFreeVerbComb(int(combtuningR3*scale), initialfeedback)
	freeverb.combL[3] = newFreeVerbComb(int(combtuningL4*scale), initialfeedback)
	freeverb.combR[3] = newFreeVerbComb(int(combtuningR4*scale), initialfeedback)
	freeverb.combL[4] = newFreeVerbComb(int(combtuningL5*scale), initialfeedback)
	freeverb.combR[4] = newFreeVerbComb(int(combtuningR5*scale), initialfeedback)
	freeverb.combL[5] = newFreeVerbComb(int(combtuningL6*scale), initialfeedback)
	freeverb.combR[5] = newFreeVerbComb(int(combtuningR6*scale), initialfeedback)
	freeverb.combL[6] = newFreeVerbComb(int(combtuningL7*scale), initialfeedback)
	freeverb.combR[6] = newFreeVerbComb(int(combtuningR7*scale), initialfeedback)
	freeverb.combL[7] = newFreeVerbComb(int(combtuningL8*scale), initialfeedback)
	freeverb.combR[7] = newFreeVerbComb(int(combtuningR8*scale), initialfeedback)

	freeverb.allpassL = make([]*freeVerbAllpass, numallpasses)
	freeverb.allpassR = make([]*freeVerbAllpass, numallpasses)
	freeverb.allpassL[0] = newFreeVerbAllpass(int(allpasstuningL1*scale), initialfeedback)
	freeverb.allpassR[0] = newFreeVerbAllpass(int(allpasstuningR1*scale), initialfeedback)
	freeverb.allpassL[1] = newFreeVerbAllpass(int(allpasstuningL2*scale), initialfeedback)
	freeverb.allpassR[1] = newFreeVerbAllpass(int(allpasstuningR2*scale), initialfeedback)
	freeverb.allpassL[2] = newFreeVerbAllpass(int(allpasstuningL3*scale), initialfeedback)
	freeverb.allpassR[2] = newFreeVerbAllpass(int(allpasstuningR3*scale), initialfeedback)
	freeverb.allpassL[3] = newFreeVerbAllpass(int(allpasstuningL4*scale), initialfeedback)
	freeverb.allpassR[3] = newFreeVerbAllpass(int(allpasstuningR4*scale), initialfeedback)

	freeverb.setWet(initialwet)
	freeverb.setRoomSize(initialroom)
	freeverb.setDry(initialdry)
	freeverb.setDamp(initialdamp)
	freeverb.setWidth(initialwidth)
	freeverb.setMode(initialmode)

	return freeverb
}

// DSP for free verb
func (freeverb *FreeVerbModule) DSP(timestamp int64) {
	buflen := freeverb.BufferLength
	outBuffer1 := freeverb.Outlets[0].Buffer
	outBuffer2 := freeverb.Outlets[1].Buffer

	var inBuffer1 []float64
	var inBuffer2 []float64

	if freeverb.Inlets[0].Connections.Len() > 0 {
		inBuffer1 = freeverb.Inlets[0].Buffer
	}

	if freeverb.Inlets[1].Connections.Len() > 0 {
		inBuffer2 = freeverb.Inlets[1].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		outL, outR, inputL, inputR, input := 0.0, 0.0, 0.0, 0.0, 0.0

		if inBuffer1 != nil {
			inputL = inBuffer1[i]
		}

		if inBuffer2 != nil {
			inputR = inBuffer2[i]
		} else {
			inputR = inputL
		}

		input = (inputL + inputR) * freeverb.gain

		for j := 0; j < numcombs; j++ {
			outL += freeverb.combL[j].process(input)
			outR += freeverb.combR[j].process(input)
		}

		for j := 0; j < numallpasses; j++ {
			outL = freeverb.allpassL[j].process(outL)
			outR = freeverb.allpassR[j].process(outR)
		}

		outBuffer1[i] = outL*freeverb.wet1 + outR*freeverb.wet2 + inputL*freeverb.dry
		outBuffer2[i] = outR*freeverb.wet1 + outL*freeverb.wet2 + inputR*freeverb.dry
	}
}

// Message received
func (freeverb *FreeVerbModule) Message(message farsounds.Message) {
	if valueMap, ok := message.(map[string]interface{}); ok {
		if wet, ok := valueMap["wet"].(float64); ok {
			freeverb.setWet(wet)
		}
		if roomSize, ok := valueMap["roomSize"].(float64); ok {
			freeverb.setRoomSize(roomSize)
		}
		if dry, ok := valueMap["dry"].(float64); ok {
			freeverb.setDry(dry)
		}
		if damp, ok := valueMap["damp"].(float64); ok {
			freeverb.setDamp(damp)
		}
		if width, ok := valueMap["width"].(float64); ok {
			freeverb.setWidth(width)
		}
		if mode, ok := valueMap["mode"].(float64); ok {
			freeverb.setMode(mode)
		}
	}
}
