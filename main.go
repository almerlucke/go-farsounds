package main

import (
	"fmt"

	"github.com/almerlucke/go-farsounds/examples"
	"github.com/almerlucke/go-farsounds/farsounds"

	// make sure standard components are loaded
	_ "github.com/almerlucke/go-farsounds/farsounds/components"
)

func setup() {
	farsounds.Registry.RegisterPolyVoiceFactory("sinvoice", examples.NewSinVoiceModule, 2)
}

func main() {
	setup()

	inputJSONFile := "examples/exampleScripts/patchvoice/mainPatch.json"
	outputSoundFile := "/users/almerlucke/Desktop/sinvoice"

	fmt.Printf("Generate soundfile from patch %v\n", inputJSONFile)

	err := farsounds.SoundFileFromScript(inputJSONFile, outputSoundFile, 20.0)
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
}
