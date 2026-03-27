#!/usr/bin/env bash

read_env_value() {
  local file_path="$1"
  local key="$2"

  if [[ ! -f "$file_path" ]]; then
    return 0
  fi

  awk -v key="$key" '
    /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
    {
      line = $0
      separator = index(line, "=")
      if (separator == 0) {
        next
      }

      current_key = substr(line, 1, separator - 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", current_key)
      if (current_key != key) {
        next
      }

      value = substr(line, separator + 1)
      sub(/\r$/, "", value)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)

      if ((value ~ /^".*"$/) || (value ~ /^'\''.*'\''$/)) {
        value = substr(value, 2, length(value) - 2)
      }

      print value
      exit
    }
  ' "$file_path"
}

normalize_bool() {
  local value="${1:-}"
  value="${value,,}"

  case "$value" in
    1|true|yes|on)
      echo "true"
      ;;
    0|false|no|off)
      echo "false"
      ;;
    *)
      echo ""
      ;;
  esac
}

resolve_frontend_auth_env() {
  local backend_env_path="$1"
  local frontend_env_path="$2"
  local resolved_auth_enabled="${NEXT_PUBLIC_AUTH_ENABLED:-}"
  local resolved_cookie_name="${AUTH_COOKIE_NAME:-}"
  local resolved_session_secret="${AUTH_SESSION_SECRET:-}"
  local backend_auth_enabled

  if [[ -z "$resolved_auth_enabled" ]]; then
    backend_auth_enabled="$(normalize_bool "$(read_env_value "$backend_env_path" "ENABLE_SINGLE_USER_AUTH")")"
    resolved_auth_enabled="$backend_auth_enabled"
  fi

  if [[ -z "$resolved_auth_enabled" ]]; then
    resolved_auth_enabled="$(normalize_bool "$(read_env_value "$frontend_env_path" "NEXT_PUBLIC_AUTH_ENABLED")")"
  fi

  if [[ -z "$resolved_auth_enabled" ]]; then
    resolved_auth_enabled="false"
  fi

  if [[ -z "$resolved_cookie_name" ]]; then
    resolved_cookie_name="$(read_env_value "$backend_env_path" "AUTH_COOKIE_NAME")"
  fi

  if [[ -z "$resolved_cookie_name" ]]; then
    resolved_cookie_name="$(read_env_value "$frontend_env_path" "AUTH_COOKIE_NAME")"
  fi

  if [[ -z "$resolved_cookie_name" ]]; then
    resolved_cookie_name="alpha_pulse_session"
  fi

  if [[ -z "$resolved_session_secret" ]]; then
    resolved_session_secret="$(read_env_value "$backend_env_path" "AUTH_SESSION_SECRET")"
  fi

  if [[ -z "$resolved_session_secret" ]]; then
    resolved_session_secret="$(read_env_value "$frontend_env_path" "AUTH_SESSION_SECRET")"
  fi

  export NEXT_PUBLIC_AUTH_ENABLED="$resolved_auth_enabled"
  export AUTH_COOKIE_NAME="$resolved_cookie_name"
  export AUTH_SESSION_SECRET="$resolved_session_secret"
}

validate_frontend_auth_env() {
  local auth_enabled

  auth_enabled="$(normalize_bool "${NEXT_PUBLIC_AUTH_ENABLED:-false}")"

  if [[ "$auth_enabled" == "true" && -z "${AUTH_SESSION_SECRET:-}" ]]; then
    echo "frontend auth is enabled but AUTH_SESSION_SECRET is missing." >&2
    echo "set it in shell env, backend/.env, or frontend/.env.local before running ./scripts/dev.sh" >&2
    return 1
  fi
}

get_listening_port_entry() {
  local port="$1"

  lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | awk 'NR == 2 { print; exit }'
}

ensure_port_available() {
  local port="$1"
  local entry

  entry="$(get_listening_port_entry "$port")"
  if [[ -z "$entry" ]]; then
    return 0
  fi

  echo "port $port is already in use." >&2
  echo "stop the existing listener before running ./scripts/dev.sh again." >&2
  echo "listener: $entry" >&2
  return 1
}
