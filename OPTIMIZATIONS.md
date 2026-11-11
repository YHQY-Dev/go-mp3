# Performance work — 2026-11-11

## Design constraint

Preserve upstream v0.3.4's PCM arithmetic, not merely similar perceptual quality. No resampling, precision reduction, output gain changes, gapless trimming or transform factorization was introduced. A change that fails the PCM or scalar-reference comparisons is not accepted to improve a benchmark.

This is stronger than a listening comparison for the tested inputs, but does not imply that the old decoder is perfect, that all inputs have been tested, or that every platform has been measured.

## Changes and attribution

Original files retain Hajime Hoshi's Apache-2.0 notices. Work in this fork is also under Apache-2.0.

- Earlier local work: reusable Frame, PCM queue, MainData and caller-owned IMDCT output; buffered input with relative-seek position correction.
- [asvedr/go-mp3 at 1e32faeeb5823ac63129ff0fd225042754352ea1](https://github.com/asvedr/go-mp3/tree/1e32faeeb5823ac63129ff0fd225042754352ea1), Apache-2.0: `internal/bits/bits.go` reservoir reuse/lookahead and `internal/huffman/huffman.go` 9-bit lookup implementation were adapted directly. Its per-band gains, independent accumulation chains, SideInfo reuse and fused circular windowing informed this fork's implementations. Original license headers are retained. These changes are not presented as wholly original work.
- [llehouerou/go-mp3 at 4cc857bbd553a63fa9b7d228afeb51ef00985b5e](https://github.com/llehouerou/go-mp3/tree/4cc857bbd553a63fa9b7d228afeb51ef00985b5e), Apache-2.0: gain lookup, circular synthesis and transformed coefficient layouts were reviewed. The FFT/DCT factorization and default gapless behavior were **not** adopted because they do not preserve the selected upstream PCM baseline.
- Additional implementation here: immutable quarter-step gains evaluated with upstream `math.Pow`; four-output transposed coefficient layouts; amd64 SSE dot products and fused circular windowing. SIMD lanes correspond to different outputs, never separately accumulated parts of one sum. MULPS and ADDPS stay separate, without FMA. SSE is in the amd64 baseline; no AVX or CPU-feature package is needed. Portable scalar code remains available on other architectures and with `-tags purego`.

Decoder scratch buffers are per instance. Shared tables become read-only after initialization. A decoder is not newly made safe for simultaneous Read/Seek calls; independently owned decoders can run concurrently.

## Same-cohort comparison

Windows 11, Intel Core i5-14600K, Go 1.27.3 windows/amd64. Decode complete in-memory MP3s to discarded PCM, using the same small Go benchmark harness. No sound device, UI, network, FFT or renderer. Input wrapper is not seekable; output buffer is 32 KiB. `llehouerou` uses `Gapless:false` so byte counts are comparable. Competitor source revisions are pinned above.

Each round rotates implementation order. Pink noise: five rounds, five decodes per benchmark iteration group. Music/MPEG-2: three rounds, three decodes per group. There are 55 formal samples. All samples are retained. Values below are median milliseconds for **one complete file**, not percentages of realtime CPU:

| Implementation | 90-second pink noise | classic.mp3 | mpeg2.mp3 |
|---|---:|---:|---:|
| Original v0.3.4 | 696.49 | 3970.90 | 156.00 |
| Earlier local five-change combination | 835.25 | 4292.46 | 166.85 |
| asvedr | 272.31 | 1386.94 | 59.68 |
| llehouerou | 200.35 | 961.02 | 43.79 |
| This fork, SSE | **185.96** | **830.01** | **40.60** |

This fork vs llehouerou: approximately **7.2%, 13.6%, 7.3% less decode time** in this cohort. It does not establish a universal win on all music, CPUs, platforms, power states or application workloads.

Current fork ranges: 184.72–187.47 / 829.91–940.66 / 40.58–40.81 ms. llehouerou ranges: 199.11–201.17 / 938.33–972.65 / 43.25–43.97 ms. The music-file ranges overlap. The CPU is a hybrid-core desktop and scheduling/frequency were not fixed.

The earlier local combination did not beat the original in this cohort, despite its large allocation reduction. This differs from the earlier application-host benchmark; host, machine state and timing differ, so do not combine those measurements into a single speedup claim or omit the unfavorable comparison.

For pink noise, cumulative Go allocation per complete decode:

| Implementation | Bytes/op | Allocs/op |
|---|---:|---:|
| Original | 197795672 | 475843 |
| Earlier combination | 11616193 | 20707 |
| This fork | 217480 | 6928 |
| asvedr | 179707 | 6925 |
| llehouerou | 116849 | 3466 |

This fork wins measured decode time, **not every metric**. The competitors still allocate fewer bytes. These are cumulative allocation counters, not peak/resident memory.

## Correctness and integration checks

- Six-file, 38-case v0.3.4 PCM/seek/EOF parity passed; full regression repeated five times. Five files are bundled; the long pink-noise file remains an optional external fixture.
- IMDCT reference covers all four block types and randomized inputs. Synthesis reference checks PCM and internal history over repeated circular-buffer wraps, stereo/mono and both granules. SSE dot-product tests check stride, length and scalar float32 bits.
- Huffman lookup checked against the original scalar tree on 139264 randomized prefixes across all 34 tables. A 10-second fuzz run performed over two million comparisons without a failure; this is a bounded check, not exhaustive verification.
- Default and `purego` suites pass; `go test ./...` and `go vet ./...` pass. Linux/arm64 and Darwin/arm64 decoder cross-builds pass; they were **not executed** locally.
- A disposable copy of the music application passed its player tests with a local replacement. Three 15-second real WASAPI playback smoke checks passed nonzero PCM, spectrum, advancement, pause and seek checks. They share a process and are correctness checks, not a formal comparison of application CPU usage.
- No production music-application dependency, installed application, Go module cache or global PATH was modified.

## Artifacts

`benchmarks/2026-11-11/` contains formal raw benchmark output, JSON summaries and hashes. The original full temporary experiment environment and its predecessor cohorts remain outside the source repository. `benchmarks/compare.cjs` can construct isolated test hosts from local checkouts; it does not fetch or alter dependencies in the music application.

Test fixtures are documented in `testdata/README.md`. Upstream example audio retains its existing `example/license.md` terms. This fork has not yet fixed all pre-existing format/seek limitations; correctness fixes should be reviewed separately from these performance changes.
