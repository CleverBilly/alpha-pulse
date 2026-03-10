#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "using local services from backend/.env or shell env"
echo "set USE_DOCKER_DEPS=1 if you want to start mysql/redis via docker compose"

if [[ "${USE_DOCKER_DEPS:-0}" == "1" ]]; then
  (
    cd "$ROOT_DIR/docker"
    docker compose up -d mysql redis
  )
fi

cd "$ROOT_DIR/backend"
nohup go run ./cmd/server > /tmp/alpha-pulse-backend.log 2>&1 &
BACKEND_PID=$!

echo "backend started: PID=$BACKEND_PID"

echo "starting frontend..."
cd "$ROOT_DIR/frontend"
npm run dev
