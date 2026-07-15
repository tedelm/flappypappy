from math import sin, pi
import wave
import struct
from pathlib import Path

sample_rate = 44100
duration = 1.2
n_samples = int(sample_rate * duration)

# Build a simple "glass clink" from several short resonant tones.
samples = [0.0] * n_samples

events = [
    (0.00, [1200, 1850, 2600]),
    (0.18, [1100, 1700, 2450]),
    (0.42, [1300, 1900, 2750]),
]

for start, freqs in events:
    start_idx = int(start * sample_rate)
    event_len = int(0.25 * sample_rate)
    for i in range(event_len):
        t = i / sample_rate
        # Exponential decay envelope.
        env = (2.71828 ** (-18 * t))
        value = sum(sin(2 * pi * f * t) for f in freqs) / len(freqs)
        idx = start_idx + i
        if idx < n_samples:
            samples[idx] += 0.35 * env * value

# Normalize.
peak = max(abs(x) for x in samples) or 1.0
samples = [x / peak * 0.8 for x in samples]

out_path = Path("/mnt/data/klirrande_glas.wav")
with wave.open(str(out_path), "w") as wf:
    wf.setnchannels(1)
    wf.setsampwidth(2)
    wf.setframerate(sample_rate)
    for s in samples:
        wf.writeframes(struct.pack("<h", int(s * 32767)))

print(out_path)
