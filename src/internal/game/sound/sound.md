# Game audio

Sound assets live in this directory. Playback code is in [`music.go`](music.go). Helper scripts are in [`misc/`](../../../../misc/).

| Script | Purpose |
|--------|---------|
| [`cut-sound.py`](../../../../misc/cut-sound.py) | Extract a clip from an MP3 by start/end (seconds) |
| [`compress-sound.py`](../../../../misc/compress-sound.py) | Re-encode embedded music MP3s for smaller WASM builds |

Requires `ffmpeg` and `ffprobe` on PATH.

## Background music

Music is embedded as MP3 and played through Ebitengine's audio API.

| File | Used in |
|------|---------|
| `game_music.mp3` | Gameplay (`StatePlaying`, `StateContinue`) |
| `game_music_menu.mp3` | Menu (`StateReady`, `StateEnterName`, `StateHighScores`) |

`game.go` calls `Manager.SetMode(menu bool)` on state changes to swap tracks.

## Sound effects

| File | Used in |
|------|---------|
| `beer_glass_hit.mp3` | Life lost (`loseLife()` — pipe/ground collision) |
| `jump.mp3` | Flap / jump (`flap()` — tap, space, start, continue) |
| `laugh.mp3` | Game over (`loseLife()` when lives reach 0; plays with `beer_glass_hit.mp3`) |

Extracted from `beer_crash.mp3` via [`cut-sound.py`](../../../../misc/cut-sound.py). `Manager.PlayLifeLost()` rewinds and plays the one-shot clip; music keeps playing underneath. `Manager.PlayGameOver()` plays the laugh on the final death.

## Loading

All embedded MP3s are decoded to PCM in `sound.NewManager()`, which runs asynchronously while the game is in `StateLoading`. The menu (`StateReady`) is not shown until loading completes.

- **Desktop:** in-game loading screen ("Pouring beer..." / "Loading audio...")
- **Web:** HTML `#loading` overlay stays visible through WASM startup and audio decode; Go calls `HideLoadingScreen()` when entering `StateReady`

Bridge: [`loading_js.go`](../loading_js.go) / [`loading_stub.go`](../loading_stub.go) (desktop no-ops).

### Loop glitch fix

#### Symptom

Music cut out intermittently and produced clicks or white noise, especially:

- When a track looped (~72 s gameplay, ~2 min menu)
- When switching between menu and gameplay music

#### Cause

The original implementation decoded MP3 into a **live streaming** source and passed it directly to `audio.NewInfiniteLoop` at the **exact** decoded length:

```go
stream, _ := mp3.DecodeWithSampleRate(48000, bytes.NewReader(data))
loop := audio.NewInfiniteLoop(stream, stream.Length())
```

Three problems stacked:

1. **Loop seam** — Ebiten documents that looping the full length of a lossy source (MP3) leaves no tail data for crossfading, so each loop restart produces a click or static.
2. **Resample + seek** — MP3s are encoded at 44.1 kHz but the audio context was 48 kHz. `DecodeWithSampleRate` resamples on the fly with a limited cache. When `InfiniteLoop` rewinds to sample 0, the resampler re-seeks and can glitch.
3. **MP3 frame seeks** — The underlying decoder seeks by MP3 frame, not sample-perfect PCM. `Rewind()` and loop restarts both trigger frame-aligned seeks.

#### Fix

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

#### Memory

Decoded PCM is roughly 12 MB (gameplay) + 21 MB (menu) at 44.1 kHz stereo. This is acceptable because the decoder would produce the same data during playback anyway; buffering just moves the cost to init time.

## Extracting SFX clips

Use [`cut-sound.py`](../../../../misc/cut-sound.py) to cut one segment from a source MP3. Default input is `beer_crash.mp3` in this directory.

```powershell
python misc/cut-sound.py --start 0.5 --end 1.8 --output beer_tap.mp3
python misc/cut-sound.py --start 1.2 --end 3.4 --dry-run
```

### CLI flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--input`, `-i` | `beer_crash.mp3` in this dir | Source MP3 |
| `--start` | required | Clip start time in seconds (float) |
| `--end` | required | Clip end time in seconds (float) |
| `--output`, `-o` | `{stem}_{start}-{end}.mp3` here | Output filename or path |
| `--bitrate` | `128k` | Output MP3 bitrate |
| `--dry-run` | off | Print ffmpeg command without writing |

Validation: `start >= 0`, `end > start`, and `end` must not exceed source duration (checked via `ffprobe`). The script prints source size, duration, clip range, and output path before cutting.

Omit `--output` to auto-name the file, e.g. `beer_crash_1.20-3.40.mp3`.

Output files are written here when given as a bare filename. Relative paths with directories or absolute paths are used as-is.

Cuts are sample-accurate (`-ss` after `-i`) and re-encoded as 44.1 kHz MP3. Channel count (mono/stereo) is preserved from the source.

### Example workflow

```powershell
# Preview a cut before writing
python misc/cut-sound.py --start 2.1 --end 3.0 --dry-run

# Extract a named SFX clip
python misc/cut-sound.py --start 2.1 --end 3.0 --output beer_glass_hit.mp3

# Cut from a different source file
python misc/cut-sound.py -i path/to/other.mp3 --start 10 --end 12.5 -o clip.mp3
```

After extracting, embed the clip in Go (`//go:embed`) and wire playback in game code as needed.

## Compressing music assets

Use [`compress-sound.py`](../../../../misc/compress-sound.py) to re-encode embedded music MP3s for smaller WASM builds. Always encode from `.orig.mp3` backups to avoid stacking lossy compression:

```powershell
python misc/compress-sound.py --bitrate 48k --mono --from-backup
```

Current working music files are 48 kb/s mono at 44.1 kHz. Bitrate and channel count do not affect the loop fix — only how `music.go` decodes and loops the data matters.

| Flag | Purpose |
|------|---------|
| `--bitrate` | Target MP3 bitrate (default `96k`) |
| `--mono` | Encode mono (`-ac 1`) |
| `--from-backup` | Source `.orig.mp3` instead of the working file |
| `--restore` | Restore working files from `.orig.mp3` backups |
| `--dry-run` | Print actions without writing |

To restore originals:

```powershell
python misc/compress-sound.py --restore
```

## Rebuild after asset changes

After changing embedded MP3s, rebuild desktop and WASM:

```powershell
cd src; go build ./cmd
cd ..; powershell -ExecutionPolicy Bypass -File scripts\build-web.ps1
```
