package frame

import (
	"math"
	"testing"
)

// Table hits and out-of-range gains must retain upstream math.Pow bits.
func TestRequantizeGain(t *testing.T) {
	for quarter := -2048; quarter <= 1024; quarter++ {
		for _, fraction := range []float64{0, 0.125} {
			exponent := float64(quarter)/4 + fraction
			got, want := requantizeGain(exponent), math.Pow(2, exponent)
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("2^%g: %016x != %016x", exponent, math.Float64bits(got), math.Float64bits(want))
			}
		}
	}
}
