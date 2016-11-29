package farsounds

import "math"

// SineTable global sine wave table
var SineTable = NewSineTable(8192)

// NewSineTable creates a new sine table
func NewSineTable(length int) []float64 {
	table := make([]float64, length)

	for i := 0; i < length; i++ {
		table[i] = math.Sin((float64(i) / float64(length-1)) * math.Pi * 2.0)
	}

	return table
}
