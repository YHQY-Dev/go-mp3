import array
import math
import random
from pathlib import Path

import av

# Synthetic test audio, Apache-2.0. Requires PyAV 19.0.2 and its FFmpeg build.
root = Path(__file__).parent
for name, rate, channels, bitrate, vbr in [
    ("transient-stereo-44100", 44100, 2, 128000, False),
    ("transient-mono-22050", 22050, 1, 64000, False),
    ("transient-vbr-48000", 48000, 2, 0, True),
]:
    rng = random.Random(123)
    count = rate * 8
    pcm = array.array("h")
    for at in range(count):
        for ch in range(channels):
            t = at / rate
            tone = 0.2 * math.sin(2 * math.pi * (440 if ch == 0 else 997) * t)
            burst = 0.4 * rng.uniform(-1, 1) if at % (rate // 4) < rate // 100 else 0.0
            value = (tone + burst) * (1.0 if int(t) % 2 else 0.05)
            pcm.append(int(max(-1.0, min(1.0, value)) * 32767))
    path = root / f"{name}.mp3"
    layout = "stereo" if channels == 2 else "mono"
    with av.open(str(path), "w") as container:
        stream = container.add_stream("libmp3lame", rate=rate)
        stream.layout = layout
        if vbr:
            stream.codec_context.qscale = True
            stream.codec_context.global_quality = 4 * 118
        else:
            stream.bit_rate = bitrate
        for at in range(0, count, 4096):
            n = min(4096, count - at)
            frame = av.AudioFrame(format="s16", layout=layout, samples=n)
            frame.sample_rate = rate
            frame.pts = at
            frame.planes[0].update(pcm[at * channels : (at + n) * channels].tobytes())
            for packet in stream.encode(frame):
                container.mux(packet)
        for packet in stream.encode(None):
            container.mux(packet)
    sizes = set()
    with av.open(str(path)) as container:
        for packet in container.demux(audio=0):
            if packet.size:
                sizes.add(packet.size)
    print(name, rate, channels, "VBR" if vbr else "CBR", path.stat().st_size, "packet sizes", sorted(sizes))
    if vbr and len(sizes) < 4:
        raise RuntimeError("VBR fixture did not vary enough")
