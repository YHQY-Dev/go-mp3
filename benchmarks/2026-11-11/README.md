# Local comparison, 2026-11-11

See ../../OPTIMIZATIONS.md for methodology, attribution and limitations.

- `results.json`: all 55 formal throughput samples; `summary.json`: median/min/max and allocations.
- `*.mp3-<round>-<variant>.log`: raw Go benchmark output.
- `*-pcm.log`: complete PCM byte counts and SHA-256 for six files. Gapless disabled for llehouerou. All byte counts match. This fork/asvedr match the original hashes; llehouerou hashes differ. Hash differences alone are not evidence of audible degradation.
- `playback-smoke.log`: three runs in one process, full MyGo output and FFT. The historical harness label is `all`, but the executable was rebuilt with the new local fork. These are correctness smoke checks, not same-cohort CPU measurements.
- `fixture-sha256.txt`: hashes of the five bundled audio inputs.
- `executables-sha256.txt`: measured executables retained outside this repository, not distributed here.
- `source-sha256.txt`: final decoder source fingerprint; documentation/tests and keyed error-literal cleanup are not a change to measured decoding behavior.

Variants: baseline = unmodified hajimehoshi/go-mp3 v0.3.4; previous = earlier five-change local combination; ours = this fork with ordered SSE; competitor SHAs are in ../../OPTIMIZATIONS.md.

The source tree was uncommitted when these experiments ran. Results are specific to the recorded machine and fixture corpus; do not market them as a cross-platform or all-input guarantee.
