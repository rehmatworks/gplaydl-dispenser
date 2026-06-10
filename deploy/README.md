# Production deployment (Ubuntu + Caddy + PostgreSQL, no Docker)

One-time server setup. After this, every push to `main` deploys automatically
via GitHub Actions.

## 1. Create the service user and directories

```bash
sudo useradd --system --create-home --shell /bin/bash dispenser
sudo mkdir -p /opt/gplaydl-dispenser
sudo chown dispenser:dispenser /opt/gplaydl-dispenser
```

## 2. PostgreSQL

```bash
sudo -u postgres psql <<'SQL'
CREATE ROLE dispenser WITH LOGIN PASSWORD 'CHANGE-ME';
CREATE DATABASE dispenser OWNER dispenser;
SQL
```

The app creates its schema (and the `citext` extension) on first start. The
extension requires superuser or appropriate grants; the simplest is to
pre-create it once:

```bash
sudo -u postgres psql -d dispenser -c 'CREATE EXTENSION IF NOT EXISTS citext;'
```

## 3. App configuration

```bash
sudo mkdir -p /etc/gplaydl-dispenser
sudo tee /etc/gplaydl-dispenser/env > /dev/null <<EOF
DISPENSER_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://dispenser:CHANGE-ME@localhost:5432/dispenser?sslmode=disable
DISPENSER_ENCRYPTION_KEY=$(openssl rand -hex 32)
MINT_CONCURRENCY=64
RESOURCES_DIR=/opt/gplaydl-dispenser/resources
DEFAULT_DEVICE=arm64_xxhdpi
PUBLIC_URL=https://dispenser.gplaydl.com
BREVO_API_KEY=CHANGE-ME
MAIL_FROM=no-reply@gplaydl.com
MAIL_FROM_NAME=gplaydl dispenser
EOF
sudo chmod 600 /etc/gplaydl-dispenser/env
```

Back up `DISPENSER_ENCRYPTION_KEY` — losing it makes all stored AAS tokens
unrecoverable.

## 4. systemd service

```bash
sudo cp gplaydl-dispenser.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable gplaydl-dispenser
```

Allow the deploy user to restart the service without a password:

```bash
echo 'dispenser ALL=(root) NOPASSWD: /usr/bin/systemctl restart gplaydl-dispenser' \
  | sudo tee /etc/sudoers.d/gplaydl-dispenser
sudo chmod 440 /etc/sudoers.d/gplaydl-dispenser
```

## 5. Caddy

Merge `Caddyfile.example` into `/etc/caddy/Caddyfile`, then:

```bash
sudo systemctl reload caddy
```

Caddy terminates TLS automatically; the app only listens on localhost.

## 6. Deploy key for GitHub Actions

```bash
sudo -u dispenser ssh-keygen -t ed25519 -f /home/dispenser/.ssh/deploy -N ''
sudo -u dispenser sh -c 'cat /home/dispenser/.ssh/deploy.pub >> /home/dispenser/.ssh/authorized_keys'
sudo -u dispenser cat /home/dispenser/.ssh/deploy   # private key -> DEPLOY_SSH_KEY secret
ssh-keyscan -H 157.173.207.54                       # output -> DEPLOY_KNOWN_HOSTS secret
```

## 7. GitHub repository secrets

Create a `production` environment in the repo settings and add:

| Secret | Value |
|---|---|
| `DEPLOY_HOST` | `157.173.207.54` |
| `DEPLOY_USER` | `dispenser` |
| `DEPLOY_PORT` | SSH port (optional, defaults to 22) |
| `DEPLOY_SSH_KEY` | Private key from step 6 |
| `DEPLOY_KNOWN_HOSTS` | `ssh-keyscan` output from step 6 |

Push to `main` — the workflow builds the frontend, embeds it into a static
Linux binary, uploads it, swaps it atomically, restarts the service, and rolls
back automatically if the health check fails.
