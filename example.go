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
	farsounds.RegisterWaveTable("sine", tables.SineTable)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	samplerate := 44100.0 * 4
	buflen := int32(1024)

	patch := farsounds.NewPatch(1, 1, buflen)
	oscModule1 := components.NewOscModule(tables.SineTable, 0.0, 1000.0/samplerate, 1.0, buflen)
	patch.InletModules[0].Connect(0, oscModule1, 0)
	oscModule1.Connect(0, patch.OutletModules[0], 0)
	patch.Modules.PushBack(oscModule1)

	adsrModule := components.NewADSRModule(buflen)
	adsrModule.SetAttackRate(0.1 * samplerate)
	adsrModule.SetDecayRate(0.2 * samplerate)
	adsrModule.SetReleaseRate(0.6 * samplerate)
	adsrModule.SetSustainLevel(0.1)
	adsrModule.Connect(0, oscModule1, 2)
	patch.Modules.PushBack(adsrModule)

	gateModule := components.NewSquareModule(0, 0.7/samplerate, 1, buflen)
	gateModule.Connect(0, adsrModule, 0)
	patch.Modules.PushBack(gateModule)

	oscModule2 := components.NewOscModule(tables.SineTable, 0.0, 57.0/samplerate, 100.0/samplerate, buflen)
	oscModule2.Connect(0, patch, 0)

	writer, err := io.OpenSoundWriter("/users/almerlucke/Desktop/output", 1, int32(samplerate), true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	numSeconds := 4.0
	timestamp := int64(0)
	numCycles := int64((numSeconds * samplerate) / float64(buflen))

	for i := int64(0); i < numCycles; i++ {
		patch.PrepareDSP()
		oscModule2.PrepareDSP()
		patch.DSP(buflen, timestamp, int32(samplerate))
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
	oscModule2.Cleanup()
}
