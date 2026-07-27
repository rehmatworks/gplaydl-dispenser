# gplaydl dispenser

**Live instance: [dispenser.gplaydl.com](https://dispenser.gplaydl.com)**

A small community-run login pool for open-source app store clients. Contributors share
spare Google accounts (publicly or privately), and the service hands out anonymous
session tokens to compatible clients like [gplaydl](https://github.com/rehmatworks/gplaydl).

## Features

- **App-managed account pools** — Android contributors add Google accounts and choose
  per account whether to share it with the community (`public`) or keep it private.
- **Atomic LRU rotation in Postgres** — accounts are claimed with
  `FOR UPDATE SKIP LOCKED`, so concurrent requests rotate through distinct accounts
  without contention. Rotation state survives restarts and works across replicas.
- **Encrypted at rest** — AAS tokens are sealed with AES-256-GCM before storage.
- **Self-healing pool** — accounts are auto-flagged after 5 consecutive failures and
  drop out of rotation; a successful mint reactivates them.
- **Concurrent minting** — bounded parallel handshakes, per-mint timeouts, and
  automatic failover to the next account.
- **Built-in web app** — a passwordless, phone-paired dashboard embedded in the
  single Go binary for account health and sharing controls.
- **Rate limiting** — anonymous dispenses are limited per IP; API-key users are exempt.

## API

| Endpoint | Description |
|---|---|
| `GET /api/auth` | Anonymous auth bundle `{email, auth}` from the public pool |
| `POST /api/auth` | Full `AuthBundle` minted with the device config supplied in the body |
| `GET /api/health` | Liveness probe |
| `/api/v1/*` | App enrolment, phone pairing, account management, and stats |

Query params for `/api/auth`:

- `locale` — locale for the bundle (default `en`)
- `device` — device profile name from `resources/` (GET only, default `arm64_xxhdpi`)
- `pool=private` — restrict to your own accounts (requires API key)

Authenticated dispensing: pass your API key as `X-Api-Key` header (or `api_key` query
param). Authenticated requests draw from your private accounts *and* the public pool,
and skip anonymous rate limits.

## Quick start

Requirements: Go 1.24+, Node 20+, Docker (or any PostgreSQL 14+).

```bash
# 1. Start Postgres
docker compose up -d

# 2. Configure
cp .env.example .env
# edit .env and set DISPENSER_ENCRYPTION_KEY (openssl rand -hex 32)

# 3. Build the frontend
cd web && npm install && npm run build && cd ..

# 4. Build & run the server (frontend is embedded into the binary)
go build -o dispenser ./cmd/dispenser
set -a && source .env && set +a && ./dispenser
```

Open http://localhost:8080. Add accounts in the Android app, then use its
one-time pairing code to open the web dashboard.

### Getting an AAS token

Use the [Authenticator app](https://github.com/whyorean/Authenticator/releases) to
generate an AAS token for a Google account. AAS tokens don't expire unless the account
password changes.

## Development

```bash
# backend (auto-restarts not included; just re-run)
set -a && source .env && set +a && go run ./cmd/dispenser

# frontend dev server with API proxy to :8080
cd web && npm run dev
```

Regenerate protobuf code after editing `proto/google_play.proto`:

```bash
protoc --proto_path=proto --go_out=internal/pb --go_opt=paths=source_relative proto/google_play.proto
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `DISPENSER_ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | local docker | Postgres connection string |
| `DISPENSER_ENCRYPTION_KEY` | — (required) | 64 hex chars; AES-256 key for AAS tokens |
| `DISPENSER_DEV` | off | `1` allows session cookies over plain HTTP |
| `MINT_CONCURRENCY` | `64` | Max simultaneous Google handshakes |
| `MINT_TIMEOUT_SECONDS` | `90` | Per-mint deadline |
| `RESOURCES_DIR` | `resources` | Device `.properties` profiles |
| `DEFAULT_DEVICE` | `arm64_xxhdpi` | Default device profile |
| `SESSION_TTL_HOURS` | `336` | Web session lifetime |
| `PUBLIC_URL` | `https://dispenser.gplaydl.com` | Base URL embedded in minted bundles and pairing links |

## Architecture

```
cmd/dispenser        entrypoint
internal/config      env configuration
internal/crypto      AES-GCM box and token hashing
internal/gplay       Google Play protocol (checkin → deviceConfig → auth → toc)
internal/pb          generated protobuf (from proto/google_play.proto)
internal/store       pgx pool, embedded migrations, rotation queries
internal/api         chi router, sessions, rate limiting, handlers
web/                 React + Tailwind v4 + shadcn/ui frontend (embedded via go:embed)
```

Deploy behind nginx/Caddy with TLS; the binary serves both the API and the web app.

## Deployment

Pushes to `main` build a static Linux binary (frontend embedded) and deploy it to a
bare Ubuntu server over SSH — no Docker. See [`deploy/README.md`](deploy/README.md)
for the one-time server setup (systemd unit, Caddy config, Postgres role, and the
GitHub secrets the workflow expects).

## License

GPL-3.0-only — derived from Aurora Dispenser by Aurora OSS.
