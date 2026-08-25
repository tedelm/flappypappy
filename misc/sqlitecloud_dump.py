#!/usr/bin/env python3
"""Copy highscores rows from SQLite Cloud into a local SQLite database."""

import argparse
import sqlite3

import sqlitecloud


CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS highscores (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	player_name  TEXT    NOT NULL,
	score        INTEGER NOT NULL,
	level        INTEGER NOT NULL DEFAULT 1,
	difficulty   TEXT    NOT NULL,
	created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);
"""


def main() -> None:
    parser = argparse.ArgumentParser(description="Dump SQLite Cloud highscores into a local .sqlite file.")
    parser.add_argument("--connection-string", required=True, help="sqlitecloud:// connection string")
    parser.add_argument("--dest", required=True, help="Destination SQLite file path")
    parser.add_argument("--table", default="highscores")
    args = parser.parse_args()

    src = sqlitecloud.connect(args.connection_string)
    try:
        rows = src.execute(
            f"SELECT player_name, score, level, difficulty FROM {args.table}"
        ).fetchall()
    finally:
        src.close()

    dest = sqlite3.connect(args.dest)
    try:
        dest.execute(CREATE_TABLE_SQL)
        dest.execute("CREATE INDEX IF NOT EXISTS idx_highscores_score ON highscores(score DESC)")
        normalized = []
        for row in rows:
            if isinstance(row, dict):
                name = row["player_name"]
                score = int(row["score"])
                level = int(row["level"] or 1)
                difficulty = row["difficulty"]
            else:
                name, score, level, difficulty = row[0], row[1], row[2], row[3]
                score = int(score)
                level = int(level or 1)
            normalized.append((name, score, level, difficulty))
        dest.executemany(
            "INSERT INTO highscores (player_name, score, level, difficulty) VALUES (?, ?, ?, ?)",
            normalized,
        )
        dest.commit()
    finally:
        dest.close()

    print(f"imported {len(rows)} row(s) into {args.dest}")


if __name__ == "__main__":
    main()
