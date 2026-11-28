#!/usr/bin/env bash
# Build everything locally and ship it to the production box.
#
#   ./deploy.sh
#
# What ends up in $REMOTE_DIR on the server:
#   docker-compose.yml   services (api, scraper, caddy)
#   caddy/Caddyfile      single origin: static SPA + /api/* -> api:8080
#   www/                 the built frontend (frontend/dist)
#   .env                 DB_URL (synced from ./.env if present; the previous one is kept as .env.bak)
#
# The backend image is tagged with the git commit. To roll back on the server:
#   cd /opt/crossword && sudo docker images crossword-server      # pick a tag
#   sudo env IMAGE_TAG=<tag> docker compose --profile production up -d
#
# Frontend build-time env (e.g. VITE_API_URL) is read from your shell or
# frontend/.env.production.local.
set -euo pipefail

REMOTE_HOST=${REMOTE_HOST:-theodoretliu@136.118.50.156}
REMOTE_DIR=/opt/crossword
PLATFORM=linux/amd64
IMAGE_NAME=crossword-server

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
log() { echo -e "${GREEN}[deploy]${NC} $1"; }
error() {
  echo -e "${RED}[deploy]${NC} $1" >&2
  exit 1
}

cd "$(dirname "$0")"

TAG=$(git rev-parse --short HEAD)
if [[ -n $(git status --porcelain --untracked-files=no) ]]; then
  TAG="$TAG-dirty"
fi
log "Deploying $TAG to $REMOTE_HOST"

# 1. Make sure the server has what it needs (idempotent).
log "Bootstrapping remote..."
ssh "$REMOTE_HOST" bash -s -- "$REMOTE_DIR" <<'EOF'
set -euo pipefail
REMOTE_DIR=$1
if ! command -v docker &>/dev/null; then
  echo "Installing Docker..."
  curl -fsSL https://get.docker.com | sh
fi
if ! command -v rsync &>/dev/null; then
  sudo apt-get update && sudo apt-get install -y rsync
fi
sudo systemctl enable --now docker >/dev/null 2>&1 || true
sudo mkdir -p "$REMOTE_DIR/www" "$REMOTE_DIR/caddy"
sudo chown -R "$(whoami)" "$REMOTE_DIR"
EOF

# 2. Build locally.
log "Building backend image ($PLATFORM)..."
docker build --platform "$PLATFORM" -t "$IMAGE_NAME:$TAG" -t "$IMAGE_NAME:latest" ./backend

log "Building frontend..."
(cd frontend && npm ci --no-audit --no-fund && npm run build)
[[ -f frontend/dist/index.html ]] || error "frontend build produced no dist/index.html"

# 3. Ship it.
log "Uploading backend image..."
docker save "$IMAGE_NAME:$TAG" "$IMAGE_NAME:latest" | gzip | ssh "$REMOTE_HOST" 'gunzip | sudo docker load'

log "Syncing config and frontend..."
rsync -az docker-compose.yml "$REMOTE_HOST:$REMOTE_DIR/"
rsync -az --delete caddy/ "$REMOTE_HOST:$REMOTE_DIR/caddy/"
# Hashed assets are uploaded before index.html (rsync goes in sorted order), and old
# assets are only removed afterwards, so open tabs keep working mid-deploy.
rsync -az --delete-after frontend/dist/ "$REMOTE_HOST:$REMOTE_DIR/www/"
if [[ -f .env ]]; then
  rsync -az --backup --suffix=.bak .env "$REMOTE_HOST:$REMOTE_DIR/.env"
fi

# 4. Restart whatever changed and wait for health checks.
log "Starting services..."
ssh "$REMOTE_HOST" bash -s -- "$REMOTE_DIR" "$TAG" <<'EOF'
set -euo pipefail
cd "$1"
[[ -f .env ]] || { echo ".env missing on server (needs DB_URL)" >&2; exit 1; }
rm -f Caddyfile # moved to caddy/Caddyfile
sudo env IMAGE_TAG="$2" docker compose --profile production up -d --remove-orphans --wait --wait-timeout 120
# The Caddyfile is read at startup; reload in case caddy itself wasn't recreated.
sudo docker compose --profile production exec -T caddy caddy reload --config /etc/caddy/Caddyfile
sudo docker compose --profile production ps
# Keep a week of old images around for rollbacks.
sudo docker image prune -af --filter "until=168h" >/dev/null
EOF

log "Deployed $TAG"
