import argparse
import sqlitecloud


def quote_ident(ident: str) -> str:
    """
    Quote SQLite identifiers defensively.

    Note: this is for identifiers (table/column names), not SQL values.
    """
    if not ident:
        raise ValueError("Identifier cannot be empty")
    # Keep it strict: only allow common identifier chars.
    # This avoids accidentally injecting SQL via identifiers.
    allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
    if any(ch not in allowed for ch in ident):
        raise ValueError(f"Invalid identifier: {ident!r}")
    return f'"{ident}"'


def main() -> None:
    parser = argparse.ArgumentParser(description="Insert one row then delete it (SQLite Cloud).")
    parser.add_argument("--connection-string", required=True, help="SQLite Cloud connection string")
    parser.add_argument("--table", default="highscores", help="Target table name (default: highscores)")
    parser.add_argument("--id-col", default="id", help="ID column name (default: id)")
    parser.add_argument("--player-name", required=True, help="Player name (player_name)")
    parser.add_argument("--score", required=True, type=int, help="Score (score)")
    parser.add_argument("--difficulty", required=True, help="Difficulty (difficulty)")
    parser.add_argument("--level", required=False, type=int, default=1, help="Level (level), default: 1")
    args = parser.parse_args()

    table_q = quote_ident(args.table)
    id_col_q = quote_ident(args.id_col)
    player_name_col_q = quote_ident("player_name")
    score_col_q = quote_ident("score")
    difficulty_col_q = quote_ident("difficulty")
    level_col_q = quote_ident("level")

    conn = sqlitecloud.connect(args.connection_string)
    try:
        insert_sql = (
            f"INSERT INTO {table_q} ({player_name_col_q}, {score_col_q}, {difficulty_col_q}, {level_col_q}) "
            f"VALUES (?, ?, ?, ?)"
        )
        select_sql = f"SELECT * FROM {table_q} WHERE {id_col_q} = ?"
        delete_sql = f"DELETE FROM {table_q} WHERE {id_col_q} = ?"

        # Insert row (using parameter binding for safety).
        conn.execute(insert_sql, (args.player_name, args.score, args.difficulty, args.level))

        # AUTOINCREMENT id is available via last_insert_rowid().
        insert_id = conn.execute("SELECT last_insert_rowid()").fetchone()[0]

        # Fetch the inserted row (for visibility).
        row = conn.execute(select_sql, (insert_id,)).fetchone()
        print("Inserted row:", row)

        # Delete it right away.
        cur = conn.execute(delete_sql, (insert_id,))
        # sqlitecloud's execute cursor typically supports rowcount; keep it best-effort.
        deleted = getattr(cur, "rowcount", None)
        print("Deleted rowcount:", deleted if deleted is not None else "unknown")
    finally:
        conn.close()


if __name__ == "__main__":
    main()

