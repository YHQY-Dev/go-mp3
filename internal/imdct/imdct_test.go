package imdct

import (
	"math"
	"math/rand"
	"testing"
)

// Interleaving outputs must not change rounding, including short-block overlap.
func TestWinMatchesScalar(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for block := 0; block < 4; block++ {
		for trial := 0; trial < 1000; trial++ {
			var in [18]float32
			for i := range in {
				in[i] = float32(r.NormFloat64() * math.Pow(2, float64(trial%31-15)))
			}
			var got, want [36]float32
			for i := range got {
				got[i] = 123
			} // Caller storage may contain a preceding frame.
			Win(in[:], block, &got)
			if block == 2 {
				for i := 0; i < 3; i++ {
					for p := 0; p < 12; p++ {
						var sum float32
						for m := 0; m < 6; m++ {
							w := float32(math.Cos(math.Pi / 24 * (2*float64(p) + 7) * (2*float64(m) + 1)))
							sum += float32(in[i+3*m] * w)
						}
						want[6*i+p+6] += float32(sum * imdctWinData[block][p])
					}
				}
			} else {
				for p := 0; p < 36; p++ {
					var sum float32
					for m := 0; m < 18; m++ {
						w := float32(math.Cos(math.Pi / 72 * (2*float64(p) + 19) * (2*float64(m) + 1)))
						sum += float32(in[m] * w)
					}
					want[p] = sum * imdctWinData[block][p]
				}
			}
			for i := range want {
				if math.Float32bits(got[i]) != math.Float32bits(want[i]) {
					t.Fatalf("block %d trial %d sample %d: %08x != %08x", block, trial, i, math.Float32bits(got[i]), math.Float32bits(want[i]))
				}
			}
		}
	}
}
