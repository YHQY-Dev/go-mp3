package sideinfo

import (
	"github.com/hajimehoshi/go-mp3/internal/frameheader"
	"testing"
)

type zeroReader struct {
	first   *byte
	changed bool
}

func (r *zeroReader) ReadFull(b []byte) (int, error) {
	if r.first == nil {
		r.first = &b[0]
	} else if r.first != &b[0] {
		r.changed = true
	}
	for i := range b {
		b[i] = 0
	}
	return len(b), nil
}

// Repeated frames reuse their input bytes without retaining parsed channel state.
func TestReadReusesBytesAndClearsFields(t *testing.T) {
	r := &zeroReader{}
	var si *SideInfo
	for _, header := range []frameheader.FrameHeader{0xfffb9000, 0xfff380c0, 0xfffb9000} {
		previous := si
		var err error
		si, err = Read(r, header, si)
		if err != nil {
			t.Fatal(err)
		}
		if previous != nil && si != previous {
			t.Fatal("side-info object replaced")
		}
		if si.Scfsi[1][3] != 0 || si.Count1[1][1] != 0 {
			t.Fatal("stale parsed fields")
		}
		si.Scfsi[1][3] = 1
		si.Count1[1][1] = 42
	}
	if r.changed {
		t.Fatal("allocated new side-info bytes")
	}
}
