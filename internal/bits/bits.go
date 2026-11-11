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

// Reservoir reuse and bit lookahead adapted from asvedr/go-mp3 (Apache-2.0).
// See OPTIMIZATIONS.md for provenance.
package bits

type Bits struct {
	vec     []byte
	bitPos  int
	bytePos int
}

func New(vec []byte) *Bits {
	return &Bits{
		vec: vec,
	}
}

// Extend grows the vector by size bytes, resets the read position and returns
// the newly added region to be filled by the caller.
func (b *Bits) Extend(size int) []byte {
	n := len(b.vec)
	if cap(b.vec)-n < size {
		vec := make([]byte, n+size)
		copy(vec, b.vec)
		b.vec = vec
	} else {
		b.vec = b.vec[:n+size]
	}
	b.bytePos = 0
	b.bitPos = 0
	return b.vec[n:]
}

// Reuse keeps only the last tail bytes of the vector (moving them to the
// front), extends the vector to tail+size bytes and resets the read position.
// It returns the region of size bytes to be filled with new data by the
// caller.
func (b *Bits) Reuse(tail, size int) []byte {
	if cap(b.vec) < tail+size {
		vec := make([]byte, tail+size)
		copy(vec, b.vec[len(b.vec)-tail:])
		b.vec = vec
	} else {
		copy(b.vec, b.vec[len(b.vec)-tail:])
		b.vec = b.vec[:tail+size]
	}
	b.bytePos = 0
	b.bitPos = 0
	return b.vec[tail:]
}

func (b *Bits) Bit() int {
	if len(b.vec) <= b.bytePos {
		// TODO: Should this return error?
		return 0
	}
	tmp := uint(b.vec[b.bytePos]) >> (7 - uint(b.bitPos))
	tmp &= 0x01
	b.bytePos += (b.bitPos + 1) >> 3
	b.bitPos = (b.bitPos + 1) & 0x07
	return int(tmp)
}

func (b *Bits) Bits(num int) int {
	if num == 0 {
		return 0
	}
	if len(b.vec) <= b.bytePos {
		// TODO: Should this return error?
		return 0
	}
	bb := make([]byte, 4)
	copy(bb, b.vec[b.bytePos:])
	tmp := (uint32(bb[0]) << 24) | (uint32(bb[1]) << 16) | (uint32(bb[2]) << 8) | (uint32(bb[3]))
	tmp <<= uint(b.bitPos)
	tmp >>= (32 - uint(num))
	b.bytePos += (b.bitPos + num) >> 3
	b.bitPos = (b.bitPos + num) & 0x07
	return int(tmp)
}

func (b *Bits) BitPos() int {
	return b.bytePos<<3 + b.bitPos
}

// Peek32 returns the next 32 bits at the current position (in the high-to-low
// order they would be returned by Bit) without advancing. Bits past the end
// of the data are zero, matching Bit's behavior there.
func (b *Bits) Peek32() uint32 {
	var w uint64
	n := len(b.vec) - b.bytePos
	if n > 5 {
		n = 5
	}
	for i := 0; i < n; i++ {
		w |= uint64(b.vec[b.bytePos+i]) << uint(32-8*i)
	}
	return uint32(w >> uint(8-b.bitPos))
}

// AvailBits returns the number of bits left before the end of the data.
func (b *Bits) AvailBits() int {
	n := len(b.vec)<<3 - b.BitPos()
	if n < 0 {
		return 0
	}
	return n
}

func (b *Bits) SetPos(pos int) {
	b.bytePos = pos >> 3
	b.bitPos = pos & 0x7
}

func (b *Bits) LenInBytes() int {
	return len(b.vec)
}
