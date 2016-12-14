package granulator

import "container/list"

type Grain interface {
	Initialize(duration float64, sr float64, settings interface{})
	Process() (float64, float64)
}

type TickGenerator interface {
	GetTick(timestamp int64) bool
}

type DurationGenerator interface {
	GetDuration(timestamp int64) float64
}

type ParameterGenerator interface {
	GetParameters(timestamp int64) interface{}
}

type GrainFactory interface {
	GetGrain() Grain
	NumChannels() int
}

type GrainVoice struct {
	Grain     Grain
	SampsToGo int64
}

type Granulator struct {
	TickGenerator      TickGenerator
	DurationGenerator  DurationGenerator
	ParameterGenerator ParameterGenerator
	GrainFactory       GrainFactory
	FreeGrains         *list.List
	UsedGrains         *list.List
}

func (granulator *Granulator) NumChannels() int {
	return granulator.GrainFactory.NumChannels()
}

func (granulator *Granulator) getFreeGrainVoice() *GrainVoice {
	elem := granulator.FreeGrains.Front()

	if elem != nil {
		granulator.FreeGrains.Remove(elem)
		granulator.UsedGrains.PushBack(elem.Value)
		return elem.Value.(*GrainVoice)
	}

	grain := granulator.GrainFactory.GetGrain()
	grainVoice := &GrainVoice{Grain: grain}
	granulator.UsedGrains.PushBack(grainVoice)

	return grainVoice
}

func (granulator *Granulator) Process(timestamp int64, sr float64) (float64, float64) {
	// First try to get tick
	tick := granulator.TickGenerator.GetTick(timestamp)

	// If tick then schedule a new voice
	if tick {
		grainVoice := granulator.getFreeGrainVoice()
		duration := granulator.DurationGenerator.GetDuration(timestamp)
		parameters := granulator.ParameterGenerator.GetParameters(timestamp)
		grainVoice.Grain.Initialize(duration, sr, parameters)
		grainVoice.SampsToGo = int64(duration * sr)
	}

	// Generate left and right sample
	leftOut, rightOut := 0.0, 0.0
	for elem := granulator.UsedGrains.Front(); elem != nil; {
		tmpElem := elem
		elem = elem.Next()
		grainVoice := tmpElem.Value.(*GrainVoice)

		if grainVoice.SampsToGo <= 0 {
			granulator.UsedGrains.Remove(tmpElem)
			granulator.FreeGrains.PushBack(grainVoice)
		}

		left, right := grainVoice.Grain.Process()
		leftOut += left
		rightOut += right
		grainVoice.SampsToGo--
	}

	return leftOut, rightOut
}
