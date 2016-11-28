package main

import (
	"fmt"
	"math/rand"
	"time"

	"farcoding.me/farsounds/soundwriter"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	outputPath := "/users/almerlucke/Desktop/output"

	writer, err := soundwriter.OpenSoundWriter(outputPath, 2, 44100, true)
	if err != nil {
		fmt.Printf("normalize err: %v\n", err)
		return
	}

	numSamples := 88200 * int64(writer.Channels)
	samples := make([]float64, numSamples)

	for index := range samples {
		newSample := rand.Float64()*20.0 - 10.0
		samples[index] = newSample
	}

	fmt.Printf("num samples %d\n", numSamples)

	err = writer.WriteSamples(samples[:])
	if err != nil {
		writer.Close()
		fmt.Printf("write err: %v\n", err)
		return
	}

	writer.Close()
}
