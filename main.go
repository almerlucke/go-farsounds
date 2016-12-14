package main

import (
	"github.com/almerlucke/go-farsounds/examples"
	"github.com/almerlucke/go-farsounds/farsounds"

	// make sure standard components are loaded
	_ "github.com/almerlucke/go-farsounds/farsounds/components"

	"github.com/almerlucke/go-farsounds/farsounds/components/granulator"
)

func setup() {
	farsounds.Registry.RegisterPolyVoiceFactory("sinvoice", examples.NewSinVoiceModule, 2)
}

func main() {
	setup()

	/*
		inputJSONFile := "examples/exampleScripts/patchvoice/mainPatch.json"
		outputSoundFile := "/users/almerlucke/Desktop/sinvoice"

		fmt.Printf("Generate soundfile from patch %v\n", inputJSONFile)

		err := farsounds.SoundFileFromScript(inputJSONFile, outputSoundFile, 20.0)
		if err != nil {
			fmt.Printf("err %v\n", err)
		}
	*/

	testGenerator := &granulator.TestGenerator{}
	gr := granulator.NewGranulator(testGenerator, testGenerator, testGenerator, testGenerator)
	grModule := granulator.NewGranulatorModule(gr, 512, 44100.0)
	patch := farsounds.NewPatch(0, 1, 512, 44100.0)
	patch.Modules.PushBack(grModule)
	grModule.Connect(0, patch.OutletModules[0], 0)

	farsounds.SoundFileFromPatch(patch, "/Users/almerlucke/Desktop/grains", 10.0)
}
