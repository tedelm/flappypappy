# Game music

Background music is embedded as MP3 and played through Ebitengine's audio API in [`music.go`](music.go).

| File | Used in |
|------|---------|
| `game_music.mp3` | Gameplay (`StatePlaying`, `StateContinue`) |
| `game_music_menu.mp3` | Menu (`StateReady`, `StateEnterName`, `StateHighScores`) |

`game.go` calls `Manager.SetMode(menu bool)` on state changes to swap tracks.

## Loop glitch fix

### Symptom

Music cut out intermittently and produced clicks or white noise, especially:

- When a track looped (~72 s gameplay, ~2 min menu)
- When switching between menu and gameplay music

### Cause

The original implementation decoded MP3 into a **live streaming** source and passed it directly to `audio.NewInfiniteLoop` at the **exact** decoded length:

```go
stream, _ := mp3.DecodeWithSampleRate(48000, bytes.NewReader(data))
loop := audio.NewInfiniteLoop(stream, stream.Length())
```

Three problems stacked:

1. **Loop seam** — Ebiten documents that looping the full length of a lossy source (MP3) leaves no tail data for crossfading, so each loop restart produces a click or static.
2. **Resample + seek** — MP3s are encoded at 44.1 kHz but the audio context was 48 kHz. `DecodeWithSampleRate` resamples on the fly with a limited cache. When `InfiniteLoop` rewinds to sample 0, the resampler re-seeks and can glitch.
3. **MP3 frame seeks** — The underlying decoder seeks by MP3 frame, not sample-perfect PCM. `Rewind()` and loop restarts both trigger frame-aligned seeks.

### Fix

Decode once to an in-memory PCM buffer, match the native sample rate, and loop with a short blend tail:

```go
stream, _ := mp3.DecodeWithoutResampling(bytes.NewReader(data))
pcm, _ := io.ReadAll(stream)

const channels = 2
bytesPerSample := 2 * channels // go-mp3 always outputs 16-bit stereo
blend := int64(sampleRate) * int64(bytesPerSample) / 10 // 0.1 s tail
loopLen := int64(len(pcm)) - blend

loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), loopLen)
```

Key points:

- **44100 Hz** audio context matches the encoded MP3 sample rate (no resampling).
- **Full PCM buffer** — playback seeks within stable memory, not a streaming decoder.
- **0.1 s blend tail** — loop length ends slightly before EOF so Ebiten can crossfade the tail into the loop start (per [Ebiten `NewInfiniteLoop` docs](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/audio#NewInfiniteLoop)).

### Memory

Decoded PCM is roughly 12 MB (gameplay) + 21 MB (menu) at 44.1 kHz stereo. This is acceptable because the decoder would produce the same data during playback anyway; buffering just moves the cost to init time.

## Compressing assets

Use [`misc/compress-sound.py`](../../../../misc/compress-sound.py) to re-encode embedded MP3s for smaller WASM builds. Always encode from `.orig.mp3` backups to avoid stacking lossy compression:

```powershell
python misc/compress-sound.py --bitrate 48k --mono --from-backup
```

Current working files are 48 kb/s mono at 44.1 kHz. Bitrate and channel count do not affect the loop fix — only how `music.go` decodes and loops the data matters.

To restore originals:

```powershell
python misc/compress-sound.py --restore
```

After re-encoding, rebuild desktop and WASM:

```powershell
cd src; go build ./cmd
cd ..; powershell -ExecutionPolicy Bypass -File scripts\build-web.ps1
```
