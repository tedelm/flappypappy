#!/usr/bin/env python3
"""Extract a clip from an MP3 by start/end timestamps in seconds."""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SOUND_DIR = ROOT / "src" / "internal" / "game" / "sound"
DEFAULT_INPUT = "input_file.mp3"
DEFAULT_BITRATE = "128k"
DEFAULT_SAMPLE_RATE = "44100"


def require_tool(name: str) -> str:
    path = shutil.which(name)
    if not path:
        print(f"error: {name} not found on PATH", file=sys.stderr)
        sys.exit(1)
    return path


def human_size(num_bytes: int) -> str:
    if num_bytes < 1024:
        return f"{num_bytes} B"
    if num_bytes < 1024 * 1024:
        return f"{num_bytes / 1024:.1f} KB"
    return f"{num_bytes / (1024 * 1024):.2f} MB"


def format_timestamp(seconds: float) -> str:
    minutes = int(seconds // 60)
    secs = seconds % 60
    return f"{minutes}:{secs:05.2f}"


def probe_duration_seconds(path: Path, ffprobe: str) -> float | None:
    result = subprocess.run(
        [
            ffprobe,
            "-v",
            "error",
            "-show_entries",
            "format=duration",
            "-of",
            "default=noprint_wrappers=1:nokey=1",
            str(path),
        ],
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        return None
    try:
        return float(result.stdout.strip())
    except ValueError:
        return None


def resolve_input(path: Path) -> Path:
    if path.is_absolute():
        return path
    if path.exists():
        return path.resolve()
    in_sound = SOUND_DIR / path
    if in_sound.exists():
        return in_sound
    return path


def resolve_output(path: Path) -> Path:
    if path.is_absolute() or len(path.parts) > 1:
        return path
    return SOUND_DIR / path.name


def default_output_name(input_path: Path, start: float, end: float) -> str:
    return f"{input_path.stem}_{start:.2f}-{end:.2f}.mp3"


def cut_clip(
    input_path: Path,
    output_path: Path,
    *,
    start: float,
    end: float,
    ffmpeg: str,
    ffprobe: str,
    bitrate: str,
    dry_run: bool,
) -> None:
    if not input_path.exists():
        print(f"error: input not found: {input_path}", file=sys.stderr)
        sys.exit(1)

    duration = probe_duration_seconds(input_path, ffprobe)
    if duration is None:
        print(f"error: could not read duration of {input_path}", file=sys.stderr)
        sys.exit(1)

    if start < 0:
        print("error: --start must be >= 0", file=sys.stderr)
        sys.exit(1)
    if end <= start:
        print("error: --end must be greater than --start", file=sys.stderr)
        sys.exit(1)
    if end > duration:
        print(
            f"error: --end {end}s exceeds source duration {duration:.3f}s "
            f"({format_timestamp(duration)})",
            file=sys.stderr,
        )
        sys.exit(1)

    output_path.parent.mkdir(parents=True, exist_ok=True)

    clip_len = end - start
    print(f"input:  {input_path}")
    print(f"source: {human_size(input_path.stat().st_size)} ({format_timestamp(duration)})")
    print(f"clip:   {start:.3f}s - {end:.3f}s ({clip_len:.3f}s)")
    print(f"output: {output_path}")

    cmd = [
        ffmpeg,
        "-y",
        "-i",
        str(input_path),
        "-ss",
        str(start),
        "-to",
        str(end),
        "-vn",
        "-codec:a",
        "libmp3lame",
        "-b:a",
        bitrate,
        "-ar",
        DEFAULT_SAMPLE_RATE,
        str(output_path),
    ]
    print("cmd:", " ".join(cmd))

    if dry_run:
        print("dry-run mode")
        return

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(result.stderr, file=sys.stderr)
        raise SystemExit(f"ffmpeg failed for {input_path.name}")

    after = output_path.stat().st_size
    print(f"wrote:  {human_size(after)}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input",
        "-i",
        default=DEFAULT_INPUT,
        help=f"source MP3 (default: {DEFAULT_INPUT} in sound dir)",
    )
    parser.add_argument(
        "--start",
        type=float,
        required=True,
        help="clip start time in seconds",
    )
    parser.add_argument(
        "--end",
        type=float,
        required=True,
        help="clip end time in seconds",
    )
    parser.add_argument(
        "--output",
        "-o",
        default=None,
        help="output MP3 path (default: {stem}_{start}-{end}.mp3 in sound dir)",
    )
    parser.add_argument(
        "--bitrate",
        default=DEFAULT_BITRATE,
        help=f"output MP3 bitrate (default: {DEFAULT_BITRATE})",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print ffmpeg command without writing output",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    ffmpeg = require_tool("ffmpeg")
    ffprobe = require_tool("ffprobe")

    input_path = resolve_input(Path(args.input))
    if args.output:
        output_path = resolve_output(Path(args.output))
    else:
        output_path = SOUND_DIR / default_output_name(input_path, args.start, args.end)

    cut_clip(
        input_path,
        output_path,
        start=args.start,
        end=args.end,
        ffmpeg=ffmpeg,
        ffprobe=ffprobe,
        bitrate=args.bitrate,
        dry_run=args.dry_run,
    )


if __name__ == "__main__":
    main()
