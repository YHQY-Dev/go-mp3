// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imdct

import (
	"math"

	"github.com/hajimehoshi/go-mp3/internal/dsp"
)

var imdctWinData = [4][36]float32{}

func init() {
	for i := 0; i < 36; i++ {
		imdctWinData[0][i] = float32(math.Sin(math.Pi / 36 * (float64(i) + 0.5)))
	}
	for i := 0; i < 18; i++ {
		imdctWinData[1][i] = float32(math.Sin(math.Pi / 36 * (float64(i) + 0.5)))
	}
	for i := 18; i < 24; i++ {
		imdctWinData[1][i] = 1.0
	}
	for i := 24; i < 30; i++ {
		imdctWinData[1][i] = float32(math.Sin(math.Pi / 12 * (float64(i) + 0.5 - 18.0)))
	}
	for i := 30; i < 36; i++ {
		imdctWinData[1][i] = 0.0
	}
	for i := 0; i < 12; i++ {
		imdctWinData[2][i] = float32(math.Sin(math.Pi / 12 * (float64(i) + 0.5)))
	}
	for i := 12; i < 36; i++ {
		imdctWinData[2][i] = 0.0
	}
	for i := 0; i < 6; i++ {
		imdctWinData[3][i] = 0.0
	}
	for i := 6; i < 12; i++ {
		imdctWinData[3][i] = float32(math.Sin(math.Pi / 12 * (float64(i) + 0.5 - 6.0)))
	}
	for i := 12; i < 18; i++ {
		imdctWinData[3][i] = 1.0
	}
	for i := 18; i < 36; i++ {
		imdctWinData[3][i] = float32(math.Sin(math.Pi / 36 * (float64(i) + 0.5)))
	}
}

var cosN12 [3][6][4]float32

func init() {
	const N = 12
	for i := 0; i < 6; i++ {
		for j := 0; j < 12; j++ {
			cosN12[j/4][i][j%4] = float32(math.Cos(math.Pi / (2 * N) * (2*float64(j) + 1 + N/2) * (2*float64(i) + 1)))
		}
	}
}

var cosN36 [9][18][4]float32

func init() {
	const N = 36
	for i := 0; i < 18; i++ {
		for j := 0; j < 36; j++ {
			cosN36[j/4][i][j%4] = float32(math.Cos(math.Pi / (2 * N) * (2*float64(j) + 1 + N/2) * (2*float64(i) + 1)))
		}
	}
}

func Win(in []float32, blockType int, out *[36]float32) {
	for i := range out {
		out[i] = 0
	}
	if blockType == 2 {
		iwd := imdctWinData[blockType]
		const N = 12
		for i := 0; i < 3; i++ {
			for p := 0; p < N; p += 4 {
				var a, b, c, d float32
				for m := 0; m < N/2; m++ {
					x, w := in[i+3*m], &cosN12[p/4][m]
					a += float32(x * w[0])
					b += float32(x * w[1])
					c += float32(x * w[2])
					d += float32(x * w[3])
				}
				o := 6*i + p + 6
				out[o] += float32(a * iwd[p])
				out[o+1] += float32(b * iwd[p+1])
				out[o+2] += float32(c * iwd[p+2])
				out[o+3] += float32(d * iwd[p+3])
			}
		}
		return
	}
	const N = 36
	iwd := imdctWinData[blockType]
	for p := 0; p < N; p += 4 {
		// SIMD lanes are independent outputs, never partial sums of one output.
		dsp.Dot4(&in[0], 1, &cosN36[p/4][0][0], N/2, &out[p])
		out[p] *= iwd[p]
		out[p+1] *= iwd[p+1]
		out[p+2] *= iwd[p+2]
		out[p+3] *= iwd[p+3]
	}
}
