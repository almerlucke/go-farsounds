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

	osc1 := farsounds.NewOsc(farsounds.SineTable, 0, 100.0/samplerate, 1)
	osc2 := farsounds.NewOsc(farsounds.SineTable, 0, 201.0/samplerate, 0.6)
	osc3 := farsounds.NewOsc(farsounds.SineTable, 0, 430.0/samplerate, 0.4)
	osc4 := farsounds.NewOsc(farsounds.SineTable, 0, 510.0/samplerate, 0.2)

	outputPath := "/users/almerlucke/Desktop/output"

	writer, err := soundwriter.OpenSoundWriter(outputPath, 1, int32(samplerate), true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	numSamples := int64(int64(samplerate) * 2 * int64(writer.Channels))
	samples := make([]float64, numSamples)

	for index := range samples {
		newSample := osc1.Next(0) + osc2.Next(0) + osc3.Next(0) + osc4.Next(0)
		samples[index] = newSample
	}

	err = writer.WriteSamples(samples[:])
	if err != nil {
		writer.Close()
		fmt.Printf("write err: %v\n", err)
		return
	}

	writer.Close()
}
