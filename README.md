# Flappy Pappy

Play at **[https://flappy-pappy.se](https://flappy-pappy.se)**. GitHub Pages only redirects there.

## Run locally (desktop)

```bash
cd src
go run ./cmd
```

Desktop also reads `web/config.js` if `FLAPPY_SCORE_API_URL` is not set.

## Run web locally (with leaderboard)

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

## High scores (self-hosted API)

Every completed game is saved to a shared `highscores` table. The game talks HTTP JSON to `scoreapi` on the same host as the web build (`https://flappy-pappy.se`).

### Desktop

```powershell
$env:FLAPPY_SCORE_API_URL = "https://flappy-pappy.se"
$env:FLAPPY_SCORE_API_KEY = "YOUR_KEY"
cd src
go run ./cmd
```

Or put the same values in `web/config.js` (see `web/config.example.js`).

If neither env vars nor `web/config.js` are set, HIGH SCORES shows a sync hint and scores stay session-only.

### Web (local)

1. Copy `web/config.example.js` to `web/config.js`
2. Set `FLAPPY_SCORE_API_URL` / `FLAPPY_SCORE_API_KEY`
3. Run `.\scripts\serve-web.ps1`

**Note:** The API key is visible in the web client. Use a dedicated key and rotate it if needed.

## Production (Strato / flappy-pappy.se)

DNS A for `flappy-pappy.se` → `31.70.88.32`.

### Install API + Caddy on the container

```bash
sudo bash scripts/install-score-api.sh
```

Default host is `flappy-pappy.se`. That installs Go/build tools if needed, builds `scoreapi`, writes `/etc/flappy/scoreapi.env`, serves static files from `/var/www/flappy`, and puts Caddy in front (Let's Encrypt) for HTTPS.

If the build fails with `signal: killed`, add swap or install a prebuilt binary with `SCOREAPI_BIN=/path/to/scoreapi`.

### Deploy / refresh the game (build off-box)

WASM builds need RAM — prefer building on your PC, then sync:

```bash
# on a machine with enough RAM
bash scripts/build-web.sh

# copy docs/ to the container, then on the container:
sudo WEB_ROOT=/path/to/docs bash scripts/deploy-web.sh
# or from the repo after syncing docs/:
sudo bash scripts/deploy-web.sh
```

`deploy-web.sh` writes `config.js` with `https://flappy-pappy.se` and the key from `/etc/flappy/scoreapi.env`.

You can also pass a build into install: `sudo WEB_ROOT=/path/to/docs bash scripts/install-score-api.sh`.

### GitHub Pages redirect

The Pages workflow only publishes a redirect. Set repository secret:

`FLAPPY_PUBLIC_URL=https://flappy-pappy.se`

(If unset, the workflow defaults to that URL.)

### Import from SQLite Cloud

```bash
pip install sqlitecloud
python3 misc/sqlitecloud_dump.py \
  --connection-string 'sqlitecloud://HOST:8860/flappypappy.sqlite?apikey=OLD_KEY' \
  --dest /var/lib/flappy/flappypappy.sqlite
sudo systemctl restart score-api
```

### Health check

```bash
bash misc/score-api-wakeup.sh https://flappy-pappy.se
```
