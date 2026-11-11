package mp3_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	gomp3 "github.com/hajimehoshi/go-mp3"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type optReaderOnly struct {
	r     *bytes.Reader
	short bool
}

func (r *optReaderOnly) Read(b []byte) (int, error) {
	if r.short && len(b) > 17 {
		b = b[:17]
	}
	return r.r.Read(b)
}

type optDigest struct {
	File, Mode, Hash, End string
	Rate                  int
	Length, Bytes         int64
	Seeks                 []string
}

func optDecode(data []byte, file, mode string) optDigest {
	r := bytes.NewReader(data)
	var source io.Reader = r
	if mode != "seekable" {
		source = &optReaderOnly{r: r, short: mode == "short"}
	}
	out := optDigest{File: file, Mode: mode}
	dec, err := gomp3.NewDecoder(source)
	if err != nil {
		out.End = fmt.Sprintf("open:%T:%v", err, err)
		return out
	}
	out.Rate, out.Length = dec.SampleRate(), dec.Length()
	h := sha256.New()
	buf := make([]byte, 777)
	for {
		n, e := dec.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
			out.Bytes += int64(n)
		}
		if e != nil {
			out.End = fmt.Sprintf("%T:%v", e, e)
			break
		}
		if n == 0 {
			panic("decoder stopped without EOF")
		}
	}
	out.Hash = hex.EncodeToString(h.Sum(nil))
	if mode == "seekable" && out.Length > 40000 {
		for _, at := range []int64{out.Length / 3 / 4 * 4, 0, 10000, out.Length / 2 / 4 * 4, 0} {
			p, e := dec.Seek(at, io.SeekStart)
			if e != nil {
				out.Seeks = append(out.Seeks, fmt.Sprintf("seek:%d:%v", p, e))
				continue
			}
			b := make([]byte, 12000)
			n, e := io.ReadFull(dec, b)
			sum := sha256.Sum256(b[:n])
			out.Seeks = append(out.Seeks, fmt.Sprintf("%d:%d:%x:%v", p, n, sum, e))
		}
	}
	return out
}

// Optimizations must retain upstream PCM, short-read/EOF behavior and seek output.
// Goldens were captured from unmodified v0.3.4, not from this implementation.
func TestPCMRegression(t *testing.T) {
	raw, e := os.ReadFile("testdata/upstream_pcm.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []optDigest
	if e = json.Unmarshal(raw, &cases); e != nil {
		t.Fatal(e)
	}
	root := os.Getenv("MP3_CORPUS")
	for _, expected := range cases {
		expected := expected
		name := expected.File
		cut := -1
		if at := strings.LastIndex(name, "-cut-"); at >= 0 {
			cut, e = strconv.Atoi(name[at+5:])
			if e != nil {
				t.Fatal(e)
			}
			name = name[:at]
		}
		if root == "" && name == "pink90.mp3" {
			continue
		} // Optional external 90-second corpus; five bundled files always run.
		file := filepath.Join("testdata", name)
		if name == "classic.mp3" || name == "mpeg2.mp3" {
			file = filepath.Join("example", name)
		}
		if root != "" {
			file = filepath.Join(root, name)
		}
		t.Run(expected.File+"/"+expected.Mode, func(t *testing.T) {
			if expected.Mode == "parallel" {
				t.Parallel()
			}
			data, e := os.ReadFile(file)
			if e != nil {
				t.Fatal(e)
			}
			if cut >= 0 {
				data = data[:cut]
			}
			got := optDecode(data, expected.File, expected.Mode)
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("PCM regression: got %+v; want %+v", got, expected)
			}
		})
	}
}

// Benchmark complete decoding without audio-device waits or retaining PCM.
func BenchmarkDecode(b *testing.B) {
	file := os.Getenv("MP3_FILE")
	if file == "" {
		file = "example/classic.mp3"
	}
	data, e := os.ReadFile(file)
	if e != nil {
		b.Fatal(e)
	}
	buf := make([]byte, 32768)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d, e := gomp3.NewDecoder(&optReaderOnly{r: bytes.NewReader(data)})
		if e != nil {
			b.Fatal(e)
		}
		if _, e = io.CopyBuffer(io.Discard, d, buf); e != nil {
			b.Fatal(e)
		}
	}
}
