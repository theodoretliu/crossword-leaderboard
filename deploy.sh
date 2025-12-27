#!/usr/bin/env bash
set -euo pipefail

# Configuration
REMOTE_HOST="theodoretliu@136.118.50.156"
REMOTE_DIR="/opt/crossword"
PLATFORM="linux/amd64"
IMAGE_NAME="crossword-server"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[deploy]${NC} $1"; }
warn() { echo -e "${YELLOW}[deploy]${NC} $1"; }
error() { echo -e "${RED}[deploy]${NC} $1" >&2; exit 1; }

# Validate configuration
[[ ! -f ".env" ]] && error ".env file is required (must contain DB_URL)"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Phase 1: Bootstrap remote
log "Bootstrapping remote server..."
ssh "$REMOTE_HOST" bash <<'EOF'
set -euo pipefail
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
fi
if ! command -v rsync &> /dev/null; then
    echo "Installing rsync..."
    sudo apt-get update && sudo apt-get install -y rsync
fi
EOF

ssh "$REMOTE_HOST" "sudo mkdir -p $REMOTE_DIR && sudo chown \$(whoami) $REMOTE_DIR"

# Phase 2: Build locally
log "Building Docker image for $PLATFORM..."
docker build --platform "$PLATFORM" -t "$IMAGE_NAME" ./backend

log "Saving image to tarball..."
docker save "$IMAGE_NAME" | gzip > /tmp/${IMAGE_NAME}.tar.gz

# Phase 3: Transfer files
log "Transferring files to remote..."
rsync -avz --progress /tmp/${IMAGE_NAME}.tar.gz "$REMOTE_HOST:$REMOTE_DIR/"
rsync -avz docker-compose.yml Caddyfile .env "$REMOTE_HOST:$REMOTE_DIR/"

# Phase 4: Deploy on remote
log "Deploying on remote..."
ssh "$REMOTE_HOST" bash <<EOF
set -euo pipefail
cd $REMOTE_DIR
echo "Loading Docker image..."
gunzip -c ${IMAGE_NAME}.tar.gz | sudo docker load
echo "Starting services..."
sudo docker compose --profile production up -d
echo "Cleaning up tarball..."
rm -f ${IMAGE_NAME}.tar.gz
sudo docker image prune -f
EOF

# Cleanup local tarball
rm -f /tmp/${IMAGE_NAME}.tar.gz

log "Deployment complete!"
