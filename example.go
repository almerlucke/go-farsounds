package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/soundwriter"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	samplerate := 44100.0 * 4
	buflen := int32(2048)
	oscModule := farsounds.NewOscModule(farsounds.SineTable, 0.0, 1000.0/samplerate, 1.0, buflen)

	// osc1 := farsounds.NewOsc(farsounds.SineTable, 0, 100.0/samplerate, 1)
	// osc2 := farsounds.NewOsc(farsounds.SineTable, 0, 201.0/samplerate, 0.6)
	// osc3 := farsounds.NewOsc(farsounds.SineTable, 0, 430.0/samplerate, 0.4)
	// osc4 := farsounds.NewOsc(farsounds.SineTable, 0, 510.0/samplerate, 0.2)

	outputPath := "/users/almerlucke/Desktop/output"

	writer, err := soundwriter.OpenSoundWriter(outputPath, 1, int32(samplerate), true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	timestamp := int64(0)

	for i := 0; i < 100; i++ {
		oscModule.Processed = false
		oscModule.DSP(buflen, timestamp, int32(samplerate))
		err = writer.WriteSamples(oscModule.Outlets[0].Buffer)
		if err != nil {
			writer.Close()
			fmt.Printf("write err: %v\n", err)
			return
		}
		timestamp += int64(buflen)
	}

	writer.Close()
}
