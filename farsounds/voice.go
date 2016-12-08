package farsounds

// Voice interface must be implemented by components to be used
// as voice by other components such as polyvoice
type Voice interface {
	// Check if voice is finished releasing
	IsFinished() bool

	// Note on for voice
	NoteOn(settings interface{}, sr float64)

	// Call note off on voice, voice can still have
	// a release period after this
	NoteOff()

	// Process always generates left and right sample
	Process() (float64, float64)
}
