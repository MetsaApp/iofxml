#!/usr/bin/env bash
# Runs locally what .github/workflows/ci.yml ran on pull requests: build,
# vet, race tests and golangci-lint. GitHub Actions CI is off, so this is
# the only check. Usage: scripts/check.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

case "${1:-}" in
  "") ;;
  *) echo "usage: scripts/check.sh (there is no --full: every CI check runs by default)" >&2; exit 2 ;;
esac

step() { printf '\n==> %s\n' "$*"; }
fail() { printf '\ncheck failed: %s\n' "$*" >&2; exit 1; }
command -v go >/dev/null 2>&1 || fail "go is not installed. Install Go 1.26: https://go.dev/doc/install"

# CI's setup-go installs the version in go.mod, and golangci-lint is built
# with it; a newer local Go breaks the linter, so pin it the same way.
v="$(GOTOOLCHAIN=local go list -m -f '{{.GoVersion}}')"
case "$v" in *.*.*) ;; *) v="$v.0" ;; esac
export GOTOOLCHAIN="go$v"

step "go build ($GOTOOLCHAIN)"
go build ./... || fail "go build"

step "go vet"
go vet ./... || fail "go vet"

step "go test -race"
go test -race ./... || fail "go test -race"

step "golangci-lint v2.13.2"
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run --allow-serial-runners ./... || fail "golangci-lint"

printf '\nall checks passed\n'
