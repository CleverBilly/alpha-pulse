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

wait_for_backend() {
  local attempt

  for attempt in $(seq 1 20); do
    if curl -fsS "http://127.0.0.1:8080/healthz" >/dev/null 2>&1; then
      echo "backend is healthy on http://127.0.0.1:8080"
      return
    fi

    if ! kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
      echo "backend exited during startup. recent log:" >&2
      tail -n 80 /tmp/alpha-pulse-backend.log >&2 || true
      exit 1
    fi

    sleep 1
  done

  echo "backend did not become healthy in time. recent log:" >&2
  tail -n 80 /tmp/alpha-pulse-backend.log >&2 || true
  exit 1
}

cd "$ROOT_DIR/backend"
nohup go run ./cmd/server > /tmp/alpha-pulse-backend.log 2>&1 &
BACKEND_PID=$!

echo "backend started: PID=$BACKEND_PID"
wait_for_backend

echo "starting frontend..."
cd "$ROOT_DIR/frontend"
npm run dev
