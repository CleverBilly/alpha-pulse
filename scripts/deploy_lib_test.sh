#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HELPER_PATH="$ROOT_DIR/scripts/deploy_lib.sh"

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

assert_contains_path() {
  local needle="$1"
  local message="$2"

  case ":$PATH:" in
    *":$needle:"*)
      ;;
    *)
      echo "assertion failed: $message" >&2
      echo "PATH did not contain: $needle" >&2
      echo "PATH was: $PATH" >&2
      exit 1
      ;;
  esac
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    echo "assertion failed: $message" >&2
    echo "missing substring: $needle" >&2
    echo "full text: $haystack" >&2
    exit 1
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    echo "assertion failed: $message" >&2
    echo "unexpected substring: $needle" >&2
    echo "full text: $haystack" >&2
    exit 1
  fi
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

export PATH="/usr/bin:/bin"
export ALPHA_PULSE_GO_BIN_DIR="$tmpdir/go-bin"
export ALPHA_PULSE_NODE_BIN_DIR="$tmpdir/node-bin"
export ALPHA_PULSE_NPM_BIN_DIR="$tmpdir/npm-bin"
export ALPHA_PULSE_PM2_BIN_DIR="$tmpdir/pm2-bin"
export PM2_CALL_LOG="$tmpdir/pm2-calls.log"
export PM2_MODULE_LIST_OUTPUT_FILE="$tmpdir/pm2-module-list.txt"

mkdir -p \
  "$ALPHA_PULSE_GO_BIN_DIR" \
  "$ALPHA_PULSE_NODE_BIN_DIR" \
  "$ALPHA_PULSE_NPM_BIN_DIR" \
  "$ALPHA_PULSE_PM2_BIN_DIR"

cat >"$ALPHA_PULSE_NODE_BIN_DIR/npm" <<'EOF'
#!/usr/bin/env bash
echo node-bin-npm
EOF

cat >"$ALPHA_PULSE_NPM_BIN_DIR/npm" <<'EOF'
#!/usr/bin/env bash
echo npm-bin-npm
EOF

chmod +x "$ALPHA_PULSE_NODE_BIN_DIR/npm" "$ALPHA_PULSE_NPM_BIN_DIR/npm"

cat >"$ALPHA_PULSE_PM2_BIN_DIR/pm2" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "${PM2_CALL_LOG:?}"

case "${1:-}" in
  module:list)
    if [[ -f "${PM2_MODULE_LIST_OUTPUT_FILE:?}" ]]; then
      cat "${PM2_MODULE_LIST_OUTPUT_FILE:?}"
    fi
    ;;
esac
EOF

chmod +x "$ALPHA_PULSE_PM2_BIN_DIR/pm2"

configure_host_runtime_path

assert_contains_path "$ALPHA_PULSE_GO_BIN_DIR" "go bin dir should be added to PATH"
assert_contains_path "$ALPHA_PULSE_NODE_BIN_DIR" "node bin dir should be added to PATH"
assert_contains_path "$ALPHA_PULSE_NPM_BIN_DIR" "npm bin dir should be added to PATH"
assert_contains_path "$ALPHA_PULSE_PM2_BIN_DIR" "pm2 bin dir should be added to PATH"
assert_eq "$ALPHA_PULSE_NODE_BIN_DIR/npm" "$(command -v npm)" "node bin npm should take precedence over npm module shim"

template_path="$tmpdir/ecosystem.template.cjs"
target_path="$tmpdir/ecosystem.config.cjs"

cat >"$template_path" <<'EOF'
module.exports = { apps: [] };
EOF

ensure_pm2_config "$target_path" "$template_path"
assert_eq "$(cat "$template_path")" "$(cat "$target_path")" "missing PM2 config should be copied from template"

cat >"$target_path" <<'EOF'
module.exports = { apps: ["custom"] };
EOF

ensure_pm2_config "$target_path" "$template_path"
assert_eq 'module.exports = { apps: ["custom"] };' "$(cat "$target_path")" "existing PM2 config should not be overwritten"

missing_template_path="$tmpdir/missing.template.cjs"
ensure_pm2_config "$target_path" "$missing_template_path"
assert_eq 'module.exports = { apps: ["custom"] };' "$(cat "$target_path")" "existing PM2 config should not require a template file"

printf 'default modules\n' >"$PM2_MODULE_LIST_OUTPUT_FILE"
: >"$PM2_CALL_LOG"
ensure_pm2_logrotate "100M" "5"
pm2_calls="$(cat "$PM2_CALL_LOG")"
assert_contains "$pm2_calls" "module:list" "should inspect installed PM2 modules before configuring rotation"
assert_contains "$pm2_calls" "install pm2-logrotate" "should install pm2-logrotate when the module is missing"
assert_contains "$pm2_calls" "set pm2-logrotate:max_size 100M" "should configure the PM2 log max size"
assert_contains "$pm2_calls" "set pm2-logrotate:retain 5" "should configure how many rotated files are kept"
assert_contains "$pm2_calls" "set pm2-logrotate:compress true" "should enable compression for rotated logs"

printf 'pm2-logrotate enabled\n' >"$PM2_MODULE_LIST_OUTPUT_FILE"
: >"$PM2_CALL_LOG"
ensure_pm2_logrotate "100M" "5"
pm2_calls="$(cat "$PM2_CALL_LOG")"
assert_not_contains "$pm2_calls" "install pm2-logrotate" "should not reinstall pm2-logrotate when it already exists"

large_log="$tmpdir/backend.err.log"
printf '0123456789' >"$large_log"
truncate_log_if_oversize "$large_log" 5
assert_eq "0" "$(wc -c < "$large_log" | tr -d ' ')" "oversized log should be truncated"

small_log="$tmpdir/frontend.err.log"
printf '1234' >"$small_log"
truncate_log_if_oversize "$small_log" 10
assert_eq "4" "$(wc -c < "$small_log" | tr -d ' ')" "small logs should be left untouched"

echo "deploy lib tests passed"
