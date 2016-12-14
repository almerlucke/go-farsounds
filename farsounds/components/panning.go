package components

import "math"

const (
	sinusoidalPanningParam = math.Pi / 2.0
)

// SinusoidalPanning function
func SinusoidalPanning(value float64, pan float64) (float64, float64) {
	a := pan * sinusoidalPanningParam

	return value * math.Sin(a), value * math.Cos(a)
}
