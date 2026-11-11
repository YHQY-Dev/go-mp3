package dsp

import (
	"math"
	"math/rand"
	"testing"
)

// SIMD must preserve scalar float32 rounding and stride, not reassociate sums.
func TestDot4MatchesScalar(t *testing.T) {
	r := rand.New(rand.NewSource(41))
	for _, stride := range []int{1, 3} {
		for _, n := range []int{0, 1, 6, 18, 32} {
			for trial := 0; trial < 300; trial++ {
				var in [96]float32
				var weights [128]float32
				var got, want [4]float32
				for i := range in {
					in[i] = float32(r.NormFloat64() * math.Pow(2, float64(trial%81-40)))
				}
				for i := range weights {
					weights[i] = float32(r.NormFloat64())
				}
				Dot4(&in[0], stride, &weights[0], n, &got[0])
				for i := 0; i < n; i++ {
					for j := 0; j < 4; j++ {
						want[j] += float32(weights[4*i+j] * in[i*stride])
					}
				}
				for j := range want {
					if math.Float32bits(got[j]) != math.Float32bits(want[j]) {
						t.Fatalf("stride %d n %d trial %d lane %d: %08x != %08x", stride, n, trial, j, math.Float32bits(got[j]), math.Float32bits(want[j]))
					}
				}
			}
		}
	}
}
