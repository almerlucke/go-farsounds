package components

import (
	"math"

	"github.com/almerlucke/go-farsounds/farsounds"
)

// Lookup holds table data
type Lookup struct {
	Table farsounds.WaveTable
}

// NewLookup creates a new lookup
func NewLookup(table farsounds.WaveTable) *Lookup {
	lookup := new(Lookup)
	lookup.Table = table
	return lookup
}

// Look at table with phase (must be between 0 up to but not including 1)
func (lookup *Lookup) Look(phase float64) float64 {
	firstIndex, fraction := math.Modf(phase * float64(len(lookup.Table)-1))
	secondIndex := firstIndex + 1
	table := lookup.Table
	v1 := table[int(firstIndex)]
	v2 := table[int(secondIndex)]
	return v1 + (v2-v1)*fraction
}
