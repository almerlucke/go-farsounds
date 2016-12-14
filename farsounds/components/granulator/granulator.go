package granulator

import (
	"container/list"
	"math/rand"

	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components"
)

/*
	Interfaces
*/

// Grain interface
type Grain interface {
	Initialize(duration float64, sr float64, settings interface{})
	Process() (float64, float64)
}

// TickGenerator generates ticks
type TickGenerator interface {
	GenerateTick(timestamp int64) bool
}

// DurationGenerator generates duration
type DurationGenerator interface {
	GenerateDuration(timestamp int64) float64
}

// ParameterGenerator generates parameter
type ParameterGenerator interface {
	GenerateParameters(timestamp int64) interface{}
}

// GrainFactory generates grains
type GrainFactory interface {
	GetGrain() Grain
	NumChannels() int
}

/*
	Granulator
*/

// grain voice for granulator
type grainVoice struct {
	grain     Grain
	sampsToGo int64
}

// Granulator schedules grains
type Granulator struct {
	TickGenerator      TickGenerator
	DurationGenerator  DurationGenerator
	ParameterGenerator ParameterGenerator
	GrainFactory       GrainFactory
	FreeGrains         *list.List
	UsedGrains         *list.List
}

// NewGranulator creation
func NewGranulator(
	tickGenerator TickGenerator,
	durationGenerator DurationGenerator,
	parameterGenerator ParameterGenerator,
	grainFactory GrainFactory) *Granulator {

	granulator := new(Granulator)

	granulator.TickGenerator = tickGenerator
	granulator.DurationGenerator = durationGenerator
	granulator.ParameterGenerator = parameterGenerator
	granulator.GrainFactory = grainFactory
	granulator.FreeGrains = list.New()
	granulator.UsedGrains = list.New()

	return granulator
}

// NumChannels is an indication to module using granulator how many
// channels are used, can be 1 or 2
func (granulator *Granulator) NumChannels() int {
	return granulator.GrainFactory.NumChannels()
}

func (granulator *Granulator) getFreeGrainVoice() *grainVoice {
	elem := granulator.FreeGrains.Front()

	if elem != nil {
		granulator.FreeGrains.Remove(elem)
		granulator.UsedGrains.PushBack(elem.Value)
		return elem.Value.(*grainVoice)
	}

	grain := granulator.GrainFactory.GetGrain()
	grainVoice := &grainVoice{grain: grain}
	granulator.UsedGrains.PushBack(grainVoice)

	return grainVoice
}

// Process granulator
func (granulator *Granulator) Process(timestamp int64, sr float64) (float64, float64) {
	// First try to get tick
	tick := granulator.TickGenerator.GenerateTick(timestamp)

	// If tick then schedule a new voice
	if tick {
		grainVoice := granulator.getFreeGrainVoice()
		duration := granulator.DurationGenerator.GenerateDuration(timestamp)
		parameters := granulator.ParameterGenerator.GenerateParameters(timestamp)
		grainVoice.grain.Initialize(duration, sr, parameters)
		grainVoice.sampsToGo = int64(duration * sr)
	}

	// Generate left and right sample
	leftOut, rightOut := 0.0, 0.0
	for elem := granulator.UsedGrains.Front(); elem != nil; {
		tmpElem := elem
		elem = elem.Next()
		grainVoice := tmpElem.Value.(*grainVoice)

		if grainVoice.sampsToGo <= 0 {
			granulator.UsedGrains.Remove(tmpElem)
			granulator.FreeGrains.PushBack(grainVoice)
		}

		left, right := grainVoice.grain.Process()
		leftOut += left
		rightOut += right
		grainVoice.sampsToGo--
	}

	return leftOut, rightOut
}

/*
	Granulator module
*/

// GranulatorModule granulator module
type GranulatorModule struct {
	// Base module
	*farsounds.BaseModule

	// Granulator
	granulator *Granulator
}

// NewGranulatorModule creates a new granulator module
func NewGranulatorModule(granulator *Granulator, buflen int32, sr float64) *GranulatorModule {
	granulatorModule := new(GranulatorModule)
	granulatorModule.BaseModule = farsounds.NewBaseModule(0, granulator.NumChannels(), buflen, sr)
	granulatorModule.Parent = granulatorModule
	granulatorModule.granulator = granulator
	return granulatorModule
}

// DSP for granulator module
func (module *GranulatorModule) DSP(timestamp int64) {
	buflen := module.BufferLength

	outBuffer1 := module.Outlets[0].Buffer
	outBuffer2 := outBuffer1

	if len(module.Outlets) == 2 {
		outBuffer2 = module.Outlets[1].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		left, right := module.granulator.Process(timestamp, module.GetSampleRate())

		outBuffer1[i] = left
		outBuffer2[i] = right
	}
}

/*
	Test grain interfaces
*/

type TestGenerator struct {
}

func (generator *TestGenerator) GenerateTick(timestamp int64) bool {
	return rand.Float64() > 0.95
}

func (generator *TestGenerator) GenerateDuration(timestamp int64) float64 {
	return (rand.Float64()*100 + 5) / 1000.0
}

func (generator *TestGenerator) GenerateParameters(timestamp int64) interface{} {
	parameters := make(map[string]interface{})
	parameters["frequency"] = rand.Float64()*2800.0 + 100.0
	parameters["amplitude"] = rand.Float64()*0.8 + 0.2
	return parameters
}

func (generator *TestGenerator) GetGrain() Grain {
	return NewTestGrain()
}

func (generator *TestGenerator) NumChannels() int {
	return 1
}

type TestGrain struct {
	level float64
	slope float64
	curve float64
	osc   *components.Osc
}

func NewTestGrain() *TestGrain {
	testGrain := new(TestGrain)
	testGrain.osc = components.NewOsc(farsounds.SineTable, 0.0, 0.0, 1.0)
	return testGrain
}

func (grain *TestGrain) processEnvelope() float64 {
	level := grain.level
	grain.level += grain.slope
	grain.slope += grain.curve
	return level
}

func (grain *TestGrain) setupEnvelope(duration int64, amp float64) {
	rdur := 1.0 / float64(duration)
	rdur2 := rdur * rdur
	grain.level = 0.0
	grain.slope = 4.0 * amp * (rdur - rdur2)
	grain.curve = -8.0 * amp * rdur2
}

func (grain *TestGrain) Initialize(duration float64, sr float64, settings interface{}) {
	numSamps := int64(duration * sr)

	settingsMap := settings.(map[string]interface{})
	freq := settingsMap["frequency"].(float64)
	amp := settingsMap["amplitude"].(float64)

	grain.setupEnvelope(numSamps, amp)
	grain.osc.Inc = freq / sr
	grain.osc.Phase = 0.0
}

func (grain *TestGrain) Process() (float64, float64) {
	out := grain.osc.Process(0) * grain.processEnvelope()
	return out, out
}
