package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components"
	"github.com/almerlucke/go-farsounds/farsounds/io"
	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

func setup() {
	rand.Seed(time.Now().UTC().UnixNano())

	farsounds.Registry.RegisterWaveTable("sine", tables.SineTable)

	farsounds.Registry.RegisterModuleFactory("osc", components.OscModuleFactory)
	farsounds.Registry.RegisterModuleFactory("square", components.SquareModuleFactory)
	farsounds.Registry.RegisterModuleFactory("adsr", components.ADSRModuleFactory)
	farsounds.Registry.RegisterModuleFactory("patch", farsounds.PatchFactory)
}

/*
patch := farsounds.NewPatch(1, 1, buflen, samplerate)
oscModule1, _ := farsounds.Registry.NewModule("osc", "osc1", map[string]interface{}{
	"phase":     0.0,
	"frequency": 1000.0,
	"amplitude": 1.0,
	"table":     "sine",
}, buflen, samplerate)

patch.InletModules[0].Connect(0, oscModule1, 0)
oscModule1.Connect(0, patch.OutletModules[0], 0)
patch.Modules.PushBack(oscModule1)

adsrModule := components.NewADSRModule(buflen, samplerate)
adsrModule.SetAttackRate(0.1 * samplerate)
adsrModule.SetDecayRate(0.2 * samplerate)
adsrModule.SetReleaseRate(0.6 * samplerate)
adsrModule.SetSustainLevel(0.1)
adsrModule.Connect(0, oscModule1, 2)
patch.Modules.PushBack(adsrModule)

gateModule, _ := farsounds.Registry.NewModule("square", "square1", map[string]interface{}{
	"phase":     0.0,
	"frequency": 0.7,
	"amplitude": 1.0,
}, buflen, samplerate)

gateModule.Connect(0, adsrModule, 0)
patch.Modules.PushBack(gateModule)

oscModule2, _ := farsounds.Registry.NewModule("osc", "osc2", map[string]interface{}{
	"phase":     0.0,
	"frequency": 57.0,
	"amplitude": 100.0 / samplerate,
	"table":     "sine",
}, buflen, samplerate)

oscModule2.Connect(0, patch, 0)
*/

func main() {
	setup()

	patch, err := farsounds.LoadMainScript("patcher.json")
	if err != nil {
		fmt.Printf("Error opening main script: %v\n", err)
		return
	}

	sr := patch.GetSampleRate()
	buflen := patch.GetBufferLength()
	numChannels := len(patch.GetOutlets())

	writer, err := io.OpenSoundWriter("/users/almerlucke/Desktop/output", int32(numChannels), int32(sr), true)
	if err != nil {
		fmt.Printf("Open writer error: %v\n", err)
		return
	}

	numSeconds := 4.0
	timestamp := int64(0)
	numCycles := int64((numSeconds * sr) / float64(buflen))

	for i := int64(0); i < numCycles; i++ {
		patch.PrepareDSP()
		patch.DSP(timestamp)
		err = writer.WriteSamples(patch.Outlets[0].Buffer)
		if err != nil {
			writer.Close()
			fmt.Printf("write err: %v\n", err)
			return
		}
		timestamp += int64(buflen)
	}

	writer.Close()
	patch.Cleanup()
}
