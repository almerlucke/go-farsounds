package main

import (
	"fmt"

	"github.com/almerlucke/go-farsounds/examples"
	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/go-farsounds/farsounds/components"
	"github.com/almerlucke/go-farsounds/farsounds/components/granulator"
)

func setup() {
	farsounds.Registry.RegisterPolyVoiceFactory("sinvoice", examples.NewSinVoiceModule, 2)
}

func main() {
	setup()

	portaudio.Initialize()
	defer portaudio.Terminate()

	// fmt.Printf("Start rendering...\n\n")
	// startTime := time.Now()
	//
	// inputJSONFile := "examples/exampleScripts/player/player.json"
	// outputSoundFile := "/users/almerlucke/Desktop/player"
	//
	// fmt.Printf("Generate soundfile from patch %v\n", inputJSONFile)
	//
	// err := farsounds.RenderScript(inputJSONFile, outputSoundFile, 10.0)
	// if err != nil {
	// 	fmt.Printf("err %v\n", err)
	// }
	//
	// fmt.Printf("Soundfile rendered in %f sec\n\n", time.Now().Sub(startTime).Seconds())

	testGenerator := &granulator.TestGenerator{}
	gr := granulator.NewGranulator(testGenerator, testGenerator, testGenerator, testGenerator)
	grModule := granulator.NewGranulatorModule(gr, 512, 44100.0)
	freeverb := components.NewFreeVerbModule(512, 44100.0)
	patch := farsounds.NewPatch(0, 2, 512, 44100.0)
	patch.Modules.PushBack(grModule)
	patch.Modules.PushBack(freeverb)
	grModule.Connect(0, freeverb, 0)
	freeverb.Connect(0, patch.OutletModules[0], 0)
	freeverb.Connect(1, patch.OutletModules[1], 0)
	// farsounds.SoundFileFromPatch(patch, "/Users/almerlucke/Desktop/grains", 20.0)

	stream, err := farsounds.NewModuleStream(patch)
	if err != nil {
		return
	}

	defer stream.Close()
	stream.Start()

	var inputString string
	fmt.Scanln(&inputString)

	stream.Stop()
}
