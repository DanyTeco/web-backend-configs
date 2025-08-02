#!/bin/bash

set -e

# === CONFIG ===
GITHUB_USER="your_github_username"
GITHUB_TOKEN="your_github_token"

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: $0 <project_name> <clone_url>"
  exit 1
fi

PROJECT_NAME=$(echo "$1" | tr '[:upper:]' '[:lower:]')
CLONE_URL="$2"

PROJECTS_DIR="/path/to/your/projects"  # Change this to your projects directory
PROJECT_PATH="$PROJECTS_DIR/$PROJECT_NAME"
LOG_FILE="/var/log/webhook-deploy.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting deployment for project: $PROJECT_NAME"

# === Build authenticated clone URL if HTTPS ===
if [[ "$CLONE_URL" == https://* ]]; then
    # Remove https:// from original
    CLEAN_URL="${CLONE_URL#https://}"
    AUTHED_CLONE_URL="https://$GITHUB_USER:$GITHUB_TOKEN@$CLEAN_URL"
else
    # Assume it's SSH or already authenticated
    AUTHED_CLONE_URL="$CLONE_URL"
fi

# === Clone if project folder doesn't exist ===
if [ ! -d "$PROJECT_PATH" ]; then
    log "Project directory $PROJECT_PATH does not exist. Cloning from $CLONE_URL..."
    git clone "$AUTHED_CLONE_URL" "$PROJECT_PATH" || { log "git clone failed"; exit 1; }
fi

cd "$PROJECT_PATH"

# === Stop & remove all containers ===
log "Stopping and removing containers with docker compose..."
docker compose down || log "docker compose down failed"

# === GET LATEST CODE VERSION ---
git pull || { log "Git pull error"; exit 1; }

# === Build image(s) ===
log "Building Docker images with docker compose..."
docker compose build || { log "docker compose build failed"; exit 1; }

# === Start containers ===
log "Starting containers with docker compose..."
docker compose up -d || { log "docker compose up failed"; exit 1; }


log "$PROJECT_NAME run"
log "--------------------------------------------"
