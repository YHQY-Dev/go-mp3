//go:build amd64 && !purego
// +build amd64,!purego

package dsp

// Dot4 accumulates four independent dot products in their original order.
// weights contains n consecutive groups of four coefficients; input contains
// n values spaced stride float32 elements apart. out has four elements.
// SSE2 is part of the amd64 baseline; no AVX, FMA, cgo or runtime probing is used.
//
//go:noescape
func Dot4(input *float32, stride int, weights *float32, n int, out *float32)

// Window consumes circular synthesis history in the original 16-term order.
// pos is a multiple of 64 in [0,1024); output lanes never cross a ring edge.
//
//go:noescape
func Window(input *[1024]float32, pos int, weights *[512]float32, out *[32]float32)
