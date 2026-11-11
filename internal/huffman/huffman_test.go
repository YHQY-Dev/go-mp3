package huffman

import (
	"fmt"
	"github.com/hajimehoshi/go-mp3/internal/bits"
	"math/rand"
	"testing"
)

// Lookup decoding must preserve symbols, signs, linbits, EOF and bit position.
func TestLookupMatchesTree(t *testing.T) {
	r := rand.New(rand.NewSource(17))
	for table := range huffmanMain {
		for trial := 0; trial < 4096; trial++ {
			data := make([]byte, trial%12)
			r.Read(data)
			a, b := bits.New(data), bits.New(data)
			a.SetPos(trial % 8)
			b.SetPos(trial % 8)
			x, y, v, w, e := Decode(a, table)
			xx, yy, vv, ww, ee := scalarDecode(b, table)
			if x != xx || y != yy || v != vv || w != ww || fmt.Sprint(e) != fmt.Sprint(ee) || a.BitPos() != b.BitPos() {
				t.Fatalf("table %d trial %d: (%d %d %d %d %v @%d) != (%d %d %d %d %v @%d)", table, trial, x, y, v, w, e, a.BitPos(), xx, yy, vv, ww, ee, b.BitPos())
			}
		}
	}
}

// Arbitrary truncated prefixes must match the original scalar tree decoder.
func FuzzLookupMatchesTree(f *testing.F) {
	f.Add([]byte{0xFF, 0x00, 0xC1, 0x77}, uint8(16), uint8(3))
	f.Add([]byte{}, uint8(33), uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, table, offset uint8) {
		if len(data) > 128 {
			return
		}
		tn := int(table) % len(huffmanMain)
		a, b := bits.New(data), bits.New(data)
		a.SetPos(int(offset) % 8)
		b.SetPos(int(offset) % 8)
		x, y, v, w, e := Decode(a, tn)
		xx, yy, vv, ww, ee := scalarDecode(b, tn)
		if x != xx || y != yy || v != vv || w != ww || fmt.Sprint(e) != fmt.Sprint(ee) || a.BitPos() != b.BitPos() {
			t.Fatalf("table %d position %d: lookup differs from tree", tn, offset)
		}
	})
}

// Reference: hajimehoshi/go-mp3 v0.3.4, Apache-2.0; scalar tree walk.
func scalarDecode(m *bits.Bits, table_num int) (x, y, v, w int, err error) {
	point := 0
	error := 1
	bitsleft := 32
	treelen := huffmanMain[table_num].treelen
	linbits := huffmanMain[table_num].linbits
	if treelen == 0 { // Check for empty tables
		return 0, 0, 0, 0, nil
	}
	htptr := huffmanMain[table_num].hufftable
	for { // Start reading the Huffman code word,bit by bit
		// Check if we've matched a code word
		if (htptr[point] & 0xff00) == 0 {
			error = 0
			x = int((htptr[point] >> 4) & 0xf)
			y = int(htptr[point] & 0xf)
			break
		}
		if m.Bit() != 0 { // Go right in tree
			for (htptr[point] & 0xff) >= 250 {
				point += int(htptr[point]) & 0xff
			}
			point += int(htptr[point]) & 0xff
		} else { // Go left in tree
			for (htptr[point] >> 8) >= 250 {
				point += int(htptr[point]) >> 8
			}
			point += int(htptr[point]) >> 8
		}
		bitsleft--
		if bitsleft <= 0 || point >= treelen {
			break
		}
	}
	if error != 0 { // Check for error.
		err := fmt.Errorf("mp3: illegal Huff code in data. bleft = %d, point = %d. tab = %d.",
			bitsleft, point, table_num)
		return 0, 0, 0, 0, err
	}
	if table_num > 31 { // Process sign encodings for quadruples tables.
		v = (y >> 3) & 1
		w = (y >> 2) & 1
		x = (y >> 1) & 1
		y = y & 1
		if (v != 0) && (m.Bit() == 1) {
			v = -v
		}
		if (w != 0) && (m.Bit() == 1) {
			w = -w
		}
		if (x != 0) && (m.Bit() == 1) {
			x = -x
		}
		if (y != 0) && (m.Bit() == 1) {
			y = -y
		}
	} else {
		if (linbits != 0) && (x == 15) {
			x += m.Bits(linbits) // Get linbits
		}
		if (x != 0) && (m.Bit() == 1) {
			x = -x // Get sign bit
		}
		if (linbits != 0) && (y == 15) {
			y += m.Bits(linbits) // Get linbits
		}
		if (y != 0) && (m.Bit() == 1) {
			y = -y // Get sign bit
		}
	}
	return x, y, v, w, nil
}
