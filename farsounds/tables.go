package farsounds

import "math"

// WaveTable is an alias for a float64 slice
type WaveTable []float64

// SineTable global sine wave table
var SineTable = NewSineTable(8192)

// NewSineTable creates a new sine table
func NewSineTable(length int) WaveTable {
	table := make(WaveTable, length)

	for i := 0; i < length; i++ {
		table[i] = math.Sin((float64(i) / float64(length-1)) * math.Pi * 2.0)
	}

	return table
}
