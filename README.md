# Flappy Pappy

## Run locally

```bash
cd src
go run ./cmd
```

## Run web (WASM dev server)

```bash
cd src
go run github.com/hajimehoshi/wasmserve@latest ./cmd
```

## High scores (SQLite Cloud)

Every completed game is saved to a shared `highscores` table so players can compete globally.

### Desktop

Set the connection string before launching:

```powershell
$env:FLAPPY_SQLITECLOUD_URL = "sqlitecloud://host.g5.sqlite.cloud:8860/database?apikey=YOUR_KEY"
go run ./cmd
```

### Web

1. Copy `web/config.example.js` to `web/config.js`
2. Paste your SQLite Cloud connection string into `FLAPPY_SQLITECLOUD_URL`
3. Build or serve the web bundle (`build-web.ps1` / `wasmserve`)

For GitHub Pages, add a repository secret named `FLAPPY_SQLITECLOUD_URL`; the deploy workflow writes `web/config.js` at build time.

**Note:** The API key is visible in the web client. Use a dedicated key and rotate it if needed.

Without a configured URL, the game runs normally but high scores are session-only (in memory).
