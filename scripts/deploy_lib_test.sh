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

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

export PATH="/usr/bin:/bin"
export ALPHA_PULSE_GO_BIN_DIR="$tmpdir/go-bin"
export ALPHA_PULSE_NODE_BIN_DIR="$tmpdir/node-bin"
export ALPHA_PULSE_NPM_BIN_DIR="$tmpdir/npm-bin"
export ALPHA_PULSE_PM2_BIN_DIR="$tmpdir/pm2-bin"

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

echo "deploy lib tests passed"
