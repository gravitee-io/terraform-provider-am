#!/usr/bin/env bash
# Fail if Speakeasy removes a hand-maintained region during generation. The logic lives in hand-written
# files; a region only calls it, so a lost region is re-added by copying the call back (see CLAUDE.md).
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

check_count() {
  local file="$1"
  local text="$2"
  local expected="$3"
  local actual

  actual=$(grep -F -c "$text" "$file" || true)
  if [[ "$actual" -ne "$expected" ]]; then
    echo "$file: expected $expected occurrence(s) of '$text', found $actual" >&2
    return 1
  fi
}

check_count internal/provider/application_resource.go '// BEGIN GRAVITEE DERIVED OAUTH SETTINGS' 1
check_count internal/provider/application_resource.go 'omitUndeclaredDerivedOAuthSettings(ctx, req, data)' 1
check_count internal/provider/application_resource.go '// END GRAVITEE DERIVED OAUTH SETTINGS' 1
