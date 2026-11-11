package bits

import (
	"bytes"
	"testing"
)

// Reusing the reservoir must keep its tail and reset a partially consumed cursor.
func TestReuseReservoir(t *testing.T) {
	b := New([]byte{1, 2, 3, 4})
	b.SetPos(13)
	copy(b.Reuse(2, 3), []byte{5, 6, 7})
	if !bytes.Equal(b.vec, []byte{3, 4, 5, 6, 7}) || b.BitPos() != 0 {
		t.Fatalf("reuse: %v at %d", b.vec, b.BitPos())
	}
	b.SetPos(31)
	copy(b.Reuse(1, 1), []byte{8})
	if !bytes.Equal(b.vec, []byte{7, 8}) || b.BitPos() != 0 {
		t.Fatalf("shrink: %v at %d", b.vec, b.BitPos())
	}
	b.SetPos(7)
	copy(b.Extend(2), []byte{9, 10})
	if !bytes.Equal(b.vec, []byte{7, 8, 9, 10}) || b.BitPos() != 0 {
		t.Fatalf("extend: %v at %d", b.vec, b.BitPos())
	}
}
