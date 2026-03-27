#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HELPER_PATH="$ROOT_DIR/scripts/dev_env.sh"

# shellcheck source=/dev/null
source "$HELPER_PATH"

assert_eq() {
  local expected="$1"
  local actual="$2"
  local message="$3"

  if [[ "$expected" != "$actual" ]]; then
    echo "assertion failed: $message" >&2
    echo "expected: $expected" >&2
    echo "actual:   $actual" >&2
    exit 1
  fi
}

assert_command_fails() {
  local message="$1"
  shift

  if "$@" >/dev/null 2>&1; then
    echo "assertion failed: $message" >&2
    exit 1
  fi
}

create_env_file() {
  local path="$1"
  shift

  cat >"$path" <<EOF
$*
EOF
}

reset_auth_env() {
  unset NEXT_PUBLIC_AUTH_ENABLED
  unset AUTH_COOKIE_NAME
  unset AUTH_SESSION_SECRET
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

backend_env="$tmpdir/backend.env"
frontend_env="$tmpdir/frontend.env"

create_env_file "$backend_env" "ENABLE_SINGLE_USER_AUTH=true
AUTH_COOKIE_NAME=backend_cookie
AUTH_SESSION_SECRET=backend-secret"

reset_auth_env
resolve_frontend_auth_env "$backend_env" "$tmpdir/missing.env"
assert_eq "true" "${NEXT_PUBLIC_AUTH_ENABLED:-}" "backend auth flag should enable frontend auth"
assert_eq "backend_cookie" "${AUTH_COOKIE_NAME:-}" "backend cookie name should flow to frontend"
assert_eq "backend-secret" "${AUTH_SESSION_SECRET:-}" "backend session secret should flow to frontend"

create_env_file "$frontend_env" "NEXT_PUBLIC_AUTH_ENABLED=false
AUTH_COOKIE_NAME=frontend_cookie
AUTH_SESSION_SECRET=frontend-secret"

reset_auth_env
resolve_frontend_auth_env "$backend_env" "$frontend_env"
assert_eq "true" "${NEXT_PUBLIC_AUTH_ENABLED:-}" "backend auth flag should remain authoritative during local dev"
assert_eq "backend_cookie" "${AUTH_COOKIE_NAME:-}" "backend cookie name should remain authoritative during local dev"
assert_eq "backend-secret" "${AUTH_SESSION_SECRET:-}" "backend session secret should remain authoritative during local dev"

reset_auth_env
resolve_frontend_auth_env "$tmpdir/missing.env" "$frontend_env"
assert_eq "false" "${NEXT_PUBLIC_AUTH_ENABLED:-}" "frontend auth flag should be used when backend auth config is absent"
assert_eq "frontend_cookie" "${AUTH_COOKIE_NAME:-}" "frontend cookie name should be used when backend auth config is absent"
assert_eq "frontend-secret" "${AUTH_SESSION_SECRET:-}" "frontend secret should be used when backend auth config is absent"

reset_auth_env
export NEXT_PUBLIC_AUTH_ENABLED=false
export AUTH_COOKIE_NAME=shell_cookie
export AUTH_SESSION_SECRET=shell-secret
resolve_frontend_auth_env "$backend_env" "$frontend_env"
assert_eq "false" "${NEXT_PUBLIC_AUTH_ENABLED:-}" "shell auth flag should override file values"
assert_eq "shell_cookie" "${AUTH_COOKIE_NAME:-}" "shell cookie name should override file values"
assert_eq "shell-secret" "${AUTH_SESSION_SECRET:-}" "shell secret should override file values"

reset_auth_env
create_env_file "$frontend_env" "NEXT_PUBLIC_AUTH_ENABLED=true"
resolve_frontend_auth_env "$tmpdir/missing.env" "$frontend_env"
assert_command_fails \
  "validation should fail when frontend auth is enabled without a session secret" \
  validate_frontend_auth_env

get_listening_port_entry() {
  case "$1" in
    8080)
      echo "server 1234 billy 7u IPv6 TCP *:8080 (LISTEN)"
      ;;
    *)
      echo ""
      ;;
  esac
}

assert_command_fails \
  "port availability check should fail when a listener already exists" \
  ensure_port_available 8080

get_listening_port_entry() {
  echo ""
}

ensure_port_available 8080

echo "dev auth env tests passed"
