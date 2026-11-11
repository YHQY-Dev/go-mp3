package frame

import (
	"bytes"
	"github.com/hajimehoshi/go-mp3/internal/frameheader"
	"github.com/hajimehoshi/go-mp3/internal/maindata"
	"math"
	"math/rand"
	"testing"
)

var scalarNWin = func() (a [64][32]float32) {
	for i := range a {
		for j := range a[i] {
			a[i][j] = float32(math.Cos(float64((16+i)*(2*j+1)) * (math.Pi / 64.0)))
		}
	}
	return
}()

// Circular synthesis must preserve every PCM sample through many ring wraps.
func TestSynthesisMatchesShiftedHistory(t *testing.T) {
	r := rand.New(rand.NewSource(23))
	for _, h := range []frameheader.FrameHeader{0xfffb9000, 0xfffb90c0} {
		a, b := Frame{header: h, mainData: &maindata.MainData{}}, Frame{header: h, mainData: &maindata.MainData{}}
		for step := 0; step < 80; step++ {
			for gr := 0; gr < 2; gr++ {
				for ch := 0; ch < h.NumberOfChannels(); ch++ {
					for i := range a.mainData.Is[gr][ch] {
						v := float32(r.NormFloat64() * math.Pow(2, float64(step%20-15)))
						a.mainData.Is[gr][ch][i] = v
						b.mainData.Is[gr][ch][i] = v
					}
					got, want := make([]byte, 2304), make([]byte, 2304)
					a.subbandSynthesis(gr, ch, got)
					b.scalarSynthesis(gr, ch, want)
					if !bytes.Equal(got, want) {
						t.Fatalf("header %x step %d granule %d channel %d", h, step, gr, ch)
					}
					for i := range b.v_vec[ch] {
						if math.Float32bits(a.v_vec[ch][(a.vPos[ch]+i)&1023]) != math.Float32bits(b.v_vec[ch][i]) {
							t.Fatalf("history differs step %d at %d", step, i)
						}
					}
				}
			}
		}
	}
}

// Reference: hajimehoshi/go-mp3 v0.3.4, Apache-2.0.
func (f *Frame) scalarSynthesis(gr int, ch int, out []byte) {
	u_vec := make([]float32, 512)
	s_vec := make([]float32, 32)

	nch := f.header.NumberOfChannels()
	// Setup the n_win windowing vector and the v_vec intermediate vector
	for ss := 0; ss < 18; ss++ { // Loop through 18 samples in 32 subbands
		copy(f.v_vec[ch][64:1024], f.v_vec[ch][0:1024-64])
		d := f.mainData.Is[gr][ch]
		for i := 0; i < 32; i++ { // Copy next 32 time samples to a temp vector
			s_vec[i] = d[i*18+ss]
		}
		for i := 0; i < 64; i++ { // Matrix multiply input with n_win[][] matrix
			sum := float32(0)
			for j := 0; j < 32; j++ {
				sum += scalarNWin[i][j] * s_vec[j]
			}
			f.v_vec[ch][i] = sum
		}
		v := f.v_vec[ch]
		for i := 0; i < 512; i += 64 { // Build the U vector
			copy(u_vec[i:i+32], v[(i<<1):(i<<1)+32])
			copy(u_vec[i+32:i+64], v[(i<<1)+96:(i<<1)+128])
		}
		for i := 0; i < 512; i++ { // Window by u_vec[i] with synthDtbl[i]
			u_vec[i] *= synthDtbl[i]
		}
		for i := 0; i < 32; i++ { // Calc 32 samples,store in outdata vector
			sum := float32(0)
			for j := 0; j < 512; j += 32 {
				sum += u_vec[j+i]
			}
			// sum now contains time sample 32*ss+i. Convert to 16-bit signed int
			samp := int(sum * 32767)
			if samp > 32767 {
				samp = 32767
			} else if samp < -32767 {
				samp = -32767
			}
			s := int16(samp)
			idx := 4 * (32*ss + i)
			if nch == 1 {
				// We always run in stereo mode and duplicate channels here for mono.
				out[idx] = byte(s)
				out[idx+1] = byte(s >> 8)
				out[idx+2] = byte(s)
				out[idx+3] = byte(s >> 8)
				continue
			}
			if ch == 0 {
				out[idx] = byte(s)
				out[idx+1] = byte(s >> 8)
			} else {
				out[idx+2] = byte(s)
				out[idx+3] = byte(s >> 8)
			}
		}
	}
}
