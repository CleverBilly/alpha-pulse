#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR/backend"
nohup go run ./cmd/server > /tmp/alpha-pulse-backend.log 2>&1 &
BACKEND_PID=$!

echo "backend started: PID=$BACKEND_PID"

echo "starting frontend..."
cd "$ROOT_DIR/frontend"
npm run dev
