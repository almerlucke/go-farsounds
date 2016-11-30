package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/almerlucke/go-farsounds/farsounds"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	samplerate := 44100.0 * 4
	buflen := int32(1024)

	patch := farsounds.NewPatch(1, 1, buflen)
	oscModule1 := farsounds.NewOscModule(farsounds.SineTable, 0.0, 1000.0/samplerate, 1.0, buflen)
	patch.Modules.PushBack(oscModule1)
	patch.InletModules[0].Connect(0, oscModule1, 0)
	oscModule1.Connect(0, patch.OutletModules[0], 0)

	oscModule2 := farsounds.NewOscModule(farsounds.SineTable, 0.0, 4.0/samplerate, 100.0/samplerate, buflen)
	oscModule2.Connect(0, patch, 0)

	outputPath := "/users/almerlucke/Desktop/output"

	writer, err := farsounds.OpenSoundWriter(outputPath, 1, int32(samplerate), true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	timestamp := int64(0)

	for i := 0; i < 200; i++ {
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

		if i == 100 {
			oscModule2.Disconnect(0, patch, 0)
		}
	}

	writer.Close()
	patch.Cleanup()
	oscModule2.Cleanup()
}
