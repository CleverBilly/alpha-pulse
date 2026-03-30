#!/usr/bin/env bash

prepend_path_dir_if_exists() {
  local dir="$1"

  if [[ -z "$dir" || ! -d "$dir" ]]; then
    return 0
  fi

  case ":$PATH:" in
    *":$dir:"*)
      ;;
    *)
      export PATH="$dir:$PATH"
      ;;
  esac
}

configure_host_runtime_path() {
  local go_candidates=(
    "${ALPHA_PULSE_GO_BIN_DIR:-}"
    "/usr/local/btgo/bin"
    "/usr/local/go/bin"
  )
  local node_candidates=(
    "${ALPHA_PULSE_NODE_BIN_DIR:-}"
    "/www/server/nodejs/v24.14.1/bin"
  )
  local npm_candidates=(
    "${ALPHA_PULSE_NPM_BIN_DIR:-}"
    "/www/server/nodejs/v24.14.1/lib/node_modules/npm/bin"
  )
  local pm2_candidates=(
    "${ALPHA_PULSE_PM2_BIN_DIR:-}"
    "/www/server/nodejs/v24.14.1/lib/node_modules/pm2/bin"
  )
  local ordered_dirs=()
  local dir
  local index

  for dir in "${go_candidates[@]}" "${node_candidates[@]}" "${npm_candidates[@]}" "${pm2_candidates[@]}"; do
    if [[ -n "$dir" && -d "$dir" ]]; then
      ordered_dirs+=("$dir")
    fi
  done

  for (( index=${#ordered_dirs[@]}-1; index>=0; index-- )); do
    prepend_path_dir_if_exists "${ordered_dirs[$index]}"
  done
}

print_step() {
  echo
  echo "==> $1"
}

require_command() {
  local command_name="$1"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "missing required command: $command_name" >&2
    return 1
  fi
}

require_file() {
  local file_path="$1"

  if [[ ! -f "$file_path" ]]; then
    echo "missing required file: $file_path" >&2
    return 1
  fi
}

ensure_pm2_config() {
  local target_path="$1"
  local template_path="$2"

  if [[ -f "$target_path" ]]; then
    return 0
  fi

  require_file "$template_path"

  cp "$template_path" "$target_path"
}

ensure_pm2_logrotate() {
  local max_size="$1"
  local retain="$2"
  local compress="${3:-true}"

  if ! pm2 module:list 2>/dev/null | grep -q "pm2-logrotate"; then
    pm2 install pm2-logrotate
  fi

  pm2 set pm2-logrotate:max_size "$max_size"
  pm2 set pm2-logrotate:retain "$retain"
  pm2 set pm2-logrotate:compress "$compress"
}

truncate_log_if_oversize() {
  local file_path="$1"
  local max_bytes="$2"

  if [[ ! -f "$file_path" ]]; then
    return 0
  fi

  local current_size
  current_size="$(wc -c < "$file_path" | tr -d '[:space:]')"

  if (( current_size > max_bytes )); then
    : > "$file_path"
    echo "truncated oversized log: $file_path (${current_size} bytes)"
  fi
}

run_logged_command() {
  local log_path="$1"
  shift

  mkdir -p "$(dirname "$log_path")"

  "$@" 2>&1 | tee "$log_path"
  local command_status="${PIPESTATUS[0]}"
  return "$command_status"
}

fetch_http_status() {
  local url="$1"
  local http_code

  http_code="$(curl -sS -o /dev/null -w "%{http_code}" --max-time 10 "$url" || true)"
  if [[ -z "$http_code" ]]; then
    http_code="000"
  fi

  echo "$http_code"
}

wait_for_http_status() {
  local url="$1"
  local expected_status="$2"
  local label="$3"
  local attempts="${4:-20}"
  local sleep_seconds="${5:-1}"
  local attempt
  local actual_status

  for attempt in $(seq 1 "$attempts"); do
    actual_status="$(fetch_http_status "$url")"
    if [[ "$actual_status" == "$expected_status" ]]; then
      echo "$label: $actual_status"
      return 0
    fi

    sleep "$sleep_seconds"
  done

  echo "$label failed: expected $expected_status, got ${actual_status:-000}" >&2
  return 1
}

pm2_process_exists() {
  local process_name="$1"
  pm2 describe "$process_name" >/dev/null 2>&1
}
