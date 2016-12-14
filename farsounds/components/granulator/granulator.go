package granulator

import "container/list"

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
	GetTick(timestamp int64) bool
}

// DurationGenerator generates duration
type DurationGenerator interface {
	GetDuration(timestamp int64) float64
}

// ParameterGenerator generates parameter
type ParameterGenerator interface {
	GetParameters(timestamp int64) interface{}
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
	tick := granulator.TickGenerator.GetTick(timestamp)

	// If tick then schedule a new voice
	if tick {
		grainVoice := granulator.getFreeGrainVoice()
		duration := granulator.DurationGenerator.GetDuration(timestamp)
		parameters := granulator.ParameterGenerator.GetParameters(timestamp)
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
