# go-mp3 — bit-exact performance fork

A maintained performance fork of [hajimehoshi/go-mp3](https://github.com/hajimehoshi/go-mp3), an MP3 decoder based on [PDMP3](https://github.com/technosaurus/PDMP3).

The original project is no longer maintained. This fork prioritizes lower allocation and decode cost **without changing the original decoder's PCM results**.

- No cgo, external decoder DLL, FFmpeg dependency, or audio-device dependency in the decoder.
- Decoder-owned reusable frame, PCM, main-data and reservoir storage.
- Huffman lookup decoding, per-band gain caching, circular synthesis history.
- Ordered SSE dot products and windowing on amd64; scalar implementation on other architectures or with `-tags purego`.
- SSE processes independent outputs, preserves each sum's order, and does not use FMA, reduced precision or fast-math transformations.
- No new gapless trimming, volume adjustment or output-format change.

## Scope

The upstream public API and module path are retained. Output is signed 16-bit little-endian stereo PCM. MPEG-1/2 Layer III support and upstream format limitations remain; this is not a new codec implementation or a claim of perfect ISO conformance. In particular, preserving an upstream seek result is not proof that the upstream seek algorithm is sample-accurate.

Use a pinned `replace github.com/hajimehoshi/go-mp3 => github.com/YHQY-Dev/go-mp3 <published-version>` after a version is published. Do not point production builds at a mutable branch. During development a local replacement can be used in a disposable test project. The music application's dependency has **not** been changed by this work.

Go 1.20 or later is required. The Oto dependency in `go.mod` is used only by the upstream `example` player, not by the decoder package.

## Validation

```sh
go test ./...
go vet ./...
go test -tags purego . ./internal/...
go test ./internal/... -count=5
go test ./internal/huffman -run '^$' -fuzz FuzzLookupMatchesTree -fuzztime=10s
```

Five bundled files cover stereo/mono, MPEG-1/2, transient signals and variable bitrate. The regression suite compares full PCM hashes, byte counts, short reads, truncated input, seek output and concurrent independent decoders with **unmodified v0.3.4**. Set `MP3_CORPUS` to the six-file experiment corpus to include the longer pink-noise recording.

These tests establish exact agreement for the tested corpus and scalar kernels, not an exhaustive proof for every possible MP3 or machine. Assembly paths are checked against independent scalar references. The `purego` path is tested separately.

```sh
# Complete decode throughput, not realtime playback CPU usage:
go test -run '^$' -bench BenchmarkDecode -benchmem -count=5
# Optional: MP3_FILE=/path/to/file.mp3
```

## Results and provenance

See [OPTIMIZATIONS.md](OPTIMIZATIONS.md) for borrowed ideas, exact source revisions, local benchmark results and limitations. Retain all original Apache-2.0 notices and [LICENSE](LICENSE) when redistributing.
