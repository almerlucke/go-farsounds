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

	oscModule1 := farsounds.NewOscModule(farsounds.SineTable, 0.0, 1000.0/samplerate, 1.0, buflen)
	oscModule2 := farsounds.NewOscModule(farsounds.SineTable, 0.0, 4.0/samplerate, 100.0/samplerate, buflen)

	oscModule2.Connect(0, oscModule1, 0)

	// osc1 := farsounds.NewOsc(farsounds.SineTable, 0, 100.0/samplerate, 1)
	// osc2 := farsounds.NewOsc(farsounds.SineTable, 0, 201.0/samplerate, 0.6)
	// osc3 := farsounds.NewOsc(farsounds.SineTable, 0, 430.0/samplerate, 0.4)
	// osc4 := farsounds.NewOsc(farsounds.SineTable, 0, 510.0/samplerate, 0.2)

	outputPath := "/users/almerlucke/Desktop/output"

	writer, err := farsounds.OpenSoundWriter(outputPath, 1, int32(samplerate), true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	timestamp := int64(0)

	for i := 0; i < 200; i++ {
		oscModule1.Processed = false
		oscModule2.Processed = false
		oscModule1.DSP(buflen, timestamp, int32(samplerate))
		err = writer.WriteSamples(oscModule1.Outlets[0].Buffer)
		if err != nil {
			writer.Close()
			fmt.Printf("write err: %v\n", err)
			return
		}
		timestamp += int64(buflen)

		if i == 100 {
			oscModule2.Disconnect(0, oscModule1, 0)
		}
	}

	writer.Close()
	oscModule1.Cleanup()
	oscModule2.Cleanup()
}
