#!/usr/bin/env python3
"""Re-encode embedded game MP3s to smaller files for WASM embed."""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SOUND_DIR = ROOT / "src" / "internal" / "game" / "sound"
TRACKS = ("game_music.mp3", "game_music_menu.mp3")
DEFAULT_BITRATE = "96k"
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


def probe_duration(path: Path, ffprobe: str) -> str:
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
        return "unknown"
    try:
        seconds = float(result.stdout.strip())
    except ValueError:
        return "unknown"
    minutes = int(seconds // 60)
    secs = seconds % 60
    return f"{minutes}:{secs:05.2f}"


def backup_path(path: Path) -> Path:
    return path.with_name(path.stem + ".orig.mp3")


def ensure_backup(path: Path, dry_run: bool) -> Path:
    dest = backup_path(path)
    if dest.exists():
        print(f"  backup exists: {dest.name}")
        return dest
    print(f"  backup -> {dest.name}")
    if not dry_run:
        shutil.copy2(path, dest)
    return dest


def restore_track(path: Path, dry_run: bool) -> bool:
    src = backup_path(path)
    if not src.exists():
        print(f"  skip {path.name}: no backup {src.name}")
        return False
    print(f"  restore {path.name} <- {src.name}")
    if not dry_run:
        shutil.copy2(src, path)
    return True


def compress_track(
    path: Path,
    *,
    ffmpeg: str,
    ffprobe: str,
    bitrate: str,
    sample_rate: str,
    channels: int,
    dry_run: bool,
    from_backup: bool,
) -> None:
    if not path.exists() and not (from_backup and backup_path(path).exists()):
        print(f"skip missing file: {path}")
        return

    src = backup_path(path) if from_backup else path
    if from_backup:
        if not src.exists():
            print(f"skip {path.name}: no backup {src.name}")
            return
        print(f"{path.name} (from {src.name})")
    else:
        print(f"{path.name}")

    before = src.stat().st_size
    duration = probe_duration(src, ffprobe)
    print(f"  before: {human_size(before)} ({duration})")

    if from_backup:
        print(f"  source: {src.name}")
    else:
        ensure_backup(path, dry_run)

    temp = path.with_suffix(".tmp.mp3")
    cmd = [
        ffmpeg,
        "-y",
        "-i",
        str(src),
        "-vn",
        "-codec:a",
        "libmp3lame",
        "-b:a",
        bitrate,
        "-ar",
        sample_rate,
        "-ac",
        str(channels),
        str(temp),
    ]
    print("  cmd:", " ".join(cmd))
    if dry_run:
        return

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        if temp.exists():
            temp.unlink()
        print(result.stderr, file=sys.stderr)
        raise SystemExit(f"ffmpeg failed for {path.name}")

    temp.replace(path)
    after = path.stat().st_size
    saved = 100 * (1 - after / before) if before else 0
    print(f"  after:  {human_size(after)} ({saved:.1f}% smaller)")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--bitrate",
        default=DEFAULT_BITRATE,
        help=f"target MP3 bitrate (default: {DEFAULT_BITRATE})",
    )
    parser.add_argument(
        "--mono",
        action="store_true",
        help="encode mono (-ac 1) for extra savings",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print actions without writing files",
    )
    parser.add_argument(
        "--restore",
        action="store_true",
        help="restore working files from .orig.mp3 backups",
    )
    parser.add_argument(
        "--from-backup",
        action="store_true",
        help="encode from .orig.mp3 backup (avoids double lossy compression)",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    ffmpeg = require_tool("ffmpeg")
    ffprobe = require_tool("ffprobe")

    if not SOUND_DIR.is_dir():
        print(f"error: sound dir not found: {SOUND_DIR}", file=sys.stderr)
        sys.exit(1)

    print(f"sound dir: {SOUND_DIR}")
    if args.restore:
        print("restore mode")
        restored = 0
        for name in TRACKS:
            if restore_track(SOUND_DIR / name, args.dry_run):
                restored += 1
        if restored == 0:
            print("nothing restored")
        return

    channels = 1 if args.mono else 2
    print(f"bitrate={args.bitrate} sample_rate={DEFAULT_SAMPLE_RATE} channels={channels}")
    if args.from_backup:
        print("from-backup mode")
    if args.dry_run:
        print("dry-run mode")

    for name in TRACKS:
        compress_track(
            SOUND_DIR / name,
            ffmpeg=ffmpeg,
            ffprobe=ffprobe,
            bitrate=args.bitrate,
            sample_rate=DEFAULT_SAMPLE_RATE,
            channels=channels,
            dry_run=args.dry_run,
            from_backup=args.from_backup,
        )


if __name__ == "__main__":
    main()
