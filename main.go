package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components"
	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

func setup() {
	rand.Seed(time.Now().UTC().UnixNano())

	farsounds.Registry.RegisterWaveTable("sine", tables.SineTable)

	farsounds.Registry.RegisterModuleFactory("patch", farsounds.PatchFactory)
	farsounds.Registry.RegisterModuleFactory("osc", components.OscModuleFactory)
	farsounds.Registry.RegisterModuleFactory("square", components.SquareModuleFactory)
	farsounds.Registry.RegisterModuleFactory("adsr", components.ADSRModuleFactory)
	farsounds.Registry.RegisterModuleFactory("delay", components.DelayModuleFactory)
	farsounds.Registry.RegisterModuleFactory("allpass", components.AllpassModuleFactory)
}

func main() {
	setup()

	inputJSONFile := "exampleScripts/stereoPatch.json"
	outputSoundFile := "/users/almerlucke/Desktop/output"

	err := farsounds.SoundFileFromScript(inputJSONFile, outputSoundFile, 20.0)
	if err != nil {
		fmt.Printf("err %v\n", err)
	}
}
