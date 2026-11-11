//go:build !amd64 || purego
// +build !amd64 purego

package dsp

import "unsafe"

// Window preserves each scalar window product and ordered accumulation.
func Window(input *[1024]float32, pos int, weights *[512]float32, out *[32]float32) {
	for i := 0; i < 32; i++ {
		var sum float32
		for j := 0; j < 512; j += 64 {
			sum += float32(input[(pos+2*j+i)&1023] * weights[j+i])
			sum += float32(input[(pos+2*j+96+i)&1023] * weights[j+32+i])
		}
		out[i] = sum
	}
}

// Dot4 is the scalar implementation of the four-output ordered dot product.
func Dot4(input *float32, stride int, weights *float32, n int, out *float32) {
	var sums [4]float32
	for i := 0; i < n; i++ {
		x := *(*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(input)) + uintptr(i*stride)*4))
		for j := 0; j < 4; j++ {
			w := *(*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(weights)) + uintptr(i*4+j)*4))
			sums[j] += float32(w * x)
		}
	}
	*(*[4]float32)(unsafe.Pointer(out)) = sums
}
