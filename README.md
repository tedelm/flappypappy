# Flappy Pappy

## Run locally (desktop)

```bash
cd src
go run ./cmd
```

Desktop also reads `web/config.js` if `FLAPPY_SQLITECLOUD_URL` is not set.

## Run web (with leaderboard)

The IDE **WASM Run** button builds a temp `main.wasm` without `config.js`, so the leaderboard will not sync. Use the web dev server instead:

```powershell
.\scripts\serve-web.ps1
```

Then open http://localhost:8080

On Linux/macOS:

```bash
bash scripts/serve-web.sh
```

Or use the VS Code / Cursor launch config **Flappy Web (with leaderboard)**.

## High scores (SQLite Cloud)

Every completed game is saved to a shared `highscores` table so players can compete globally.

### Desktop

Set the connection string before launching, **or** put it in `web/config.js` (desktop reads that file as a fallback):

```powershell
$env:FLAPPY_SQLITECLOUD_URL = "sqlitecloud://cepetvllvk.g5.sqlite.cloud:8860/flappypappy.sqlite?apikey=YOUR_KEY"
cd src
go run ./cmd
```

If neither env var nor `web/config.js` is set, HIGH SCORES shows a sync hint and scores stay session-only. If the URL is set but the database cannot be reached, HIGH SCORES shows "Could not reach leaderboard".

### Web

1. Copy `web/config.example.js` to `web/config.js`
2. Paste your SQLite Cloud connection string into `FLAPPY_SQLITECLOUD_URL`
3. Run `.\scripts\serve-web.ps1` (or build with `build-web.ps1` for GitHub Pages)

For GitHub Pages, add a repository secret named `FLAPPY_SQLITECLOUD_URL`; the deploy workflow writes `web/config.js` at build time.

**Note:** The API key is visible in the web client. Use a dedicated key and rotate it if needed.

Without a configured URL, the game runs normally but high scores are session-only (in memory).
