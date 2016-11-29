package farsounds

import "math"

// Lookup holds table data
type Lookup struct {
	table []float64
}

// NewLookup creates a new lookup
func NewLookup(table []float64) *Lookup {
	lookup := new(Lookup)
	lookup.table = table
	return lookup
}

// Look at table with phase (must be between 0 up to but not including 1)
func (lookup *Lookup) Look(phase float64) float64 {
	firstIndex, fraction := math.Modf(phase * float64(len(lookup.table)-1))
	secondIndex := firstIndex + 1
	table := lookup.table
	v1 := table[int(firstIndex)]
	v2 := table[int(secondIndex)]

	return v1 + (v2-v1)*fraction
}
