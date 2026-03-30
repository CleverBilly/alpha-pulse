#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$ROOT_DIR/deploy/.tmp"
CURRENT_STEP="initialization"
LAST_LOG=""
LOG_MAX_SIZE_MB="${ALPHA_PULSE_PM2_LOG_MAX_SIZE_MB:-100}"
LOG_RETAIN="${ALPHA_PULSE_PM2_LOG_RETAIN:-5}"
LOG_MAX_SIZE_HUMAN="${LOG_MAX_SIZE_MB}M"
LOG_MAX_SIZE_BYTES="$((LOG_MAX_SIZE_MB * 1024 * 1024))"

# shellcheck source=/dev/null
source "$ROOT_DIR/scripts/deploy_lib.sh"

handle_error() {
  local exit_code="$?"

  echo >&2
  echo "deploy failed during: $CURRENT_STEP" >&2
  if [[ -n "$LAST_LOG" ]]; then
    echo "check log: $LAST_LOG" >&2
  fi

  exit "$exit_code"
}

trap handle_error ERR

mkdir -p "$ROOT_DIR/logs" "$TMP_DIR"

CURRENT_STEP="check environment"
print_step "检查环境"
configure_host_runtime_path
require_command go
require_command npm
require_command pm2
require_command curl
require_file "$ROOT_DIR/backend/.env"
require_file "$ROOT_DIR/frontend/.env.production"
ensure_pm2_config "$ROOT_DIR/ecosystem.config.cjs" "$ROOT_DIR/deploy/ecosystem.host.example.cjs"
LAST_LOG="$TMP_DIR/pm2-logrotate.log"
run_logged_command "$LAST_LOG" ensure_pm2_logrotate "$LOG_MAX_SIZE_HUMAN" "$LOG_RETAIN" "true"

CURRENT_STEP="build backend"
print_step "编译后端"
LAST_LOG="$TMP_DIR/backend-go-mod-download.log"
run_logged_command "$LAST_LOG" bash -lc "cd '$ROOT_DIR/backend' && go mod download"
LAST_LOG="$TMP_DIR/backend-build.log"
run_logged_command "$LAST_LOG" bash -lc "cd '$ROOT_DIR/backend' && mkdir -p bin && go build -o ./bin/alpha-pulse ./cmd/server"

CURRENT_STEP="build frontend"
print_step "构建前端"
LAST_LOG="$TMP_DIR/frontend-npm-ci.log"
run_logged_command "$LAST_LOG" bash -lc "cd '$ROOT_DIR/frontend' && npm ci"
LAST_LOG="$TMP_DIR/frontend-build.log"
run_logged_command "$LAST_LOG" bash -lc "cd '$ROOT_DIR/frontend' && npm run build"

CURRENT_STEP="trim oversized logs"
print_step "治理日志"
LAST_LOG="$TMP_DIR/log-maintenance.log"
run_logged_command "$LAST_LOG" truncate_log_if_oversize "$ROOT_DIR/logs/backend.err.log" "$LOG_MAX_SIZE_BYTES"
run_logged_command "$LAST_LOG" truncate_log_if_oversize "$ROOT_DIR/logs/backend.out.log" "$LOG_MAX_SIZE_BYTES"
run_logged_command "$LAST_LOG" truncate_log_if_oversize "$ROOT_DIR/logs/frontend.err.log" "$LOG_MAX_SIZE_BYTES"
run_logged_command "$LAST_LOG" truncate_log_if_oversize "$ROOT_DIR/logs/frontend.out.log" "$LOG_MAX_SIZE_BYTES"

CURRENT_STEP="restart processes"
print_step "重启进程"
LAST_LOG="$TMP_DIR/pm2-restart.log"
if pm2_process_exists "alpha-pulse-backend" && pm2_process_exists "alpha-pulse-frontend"; then
  run_logged_command "$LAST_LOG" pm2 restart alpha-pulse-backend
  run_logged_command "$LAST_LOG" pm2 restart alpha-pulse-frontend
else
  run_logged_command "$LAST_LOG" bash -lc "cd '$ROOT_DIR' && pm2 start ecosystem.config.cjs"
fi
run_logged_command "$LAST_LOG" pm2 save

CURRENT_STEP="health checks"
print_step "健康检查"
LAST_LOG=""
wait_for_http_status "http://127.0.0.1:8080/healthz" "200" "backend /healthz"
wait_for_http_status "http://127.0.0.1:3000/login" "200" "frontend /login"
wait_for_http_status "http://127.0.0.1:3000/api/trade-settings" "401" "frontend /api/trade-settings"

echo
echo "deploy finished"
echo "- backend /healthz: 200"
echo "- frontend /login: 200"
echo "- frontend /api/trade-settings: 401"
