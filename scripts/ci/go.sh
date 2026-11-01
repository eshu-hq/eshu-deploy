#!/usr/bin/env bash
# go lane: gofumpt format check (golangci-lint fmt), golangci-lint, the
# 500-line cap on .go files, go vet, go mod tidy drift check, govulncheck,
# and the Go contract test suite under the race detector. Requires helm on
# PATH: the suite skips helm-dependent tests when it is absent, so CI
# provisions it pinned. The race detector needs cgo, so a C compiler (gcc)
# must be on PATH; ubuntu-24.04 runners ship one.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
# shellcheck source=versions.sh
. scripts/ci/versions.sh

tools_dir="${repo_root}/.tools"
mkdir -p "${tools_dir}"
go_root="${tools_dir}/go${GO_VERSION}"
if [ ! -x "${go_root}/bin/go" ]; then
  printf 'go: downloading Go %s\n' "${GO_VERSION}"
  curl -sSL -o "${tools_dir}/go.tar.gz" \
    "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
  printf '%s  %s\n' "${GO_SHA256}" "${tools_dir}/go.tar.gz" | sha256sum -c -
  rm -rf "${go_root}"
  mkdir -p "${go_root}"
  tar -xzf "${tools_dir}/go.tar.gz" -C "${go_root}" --strip-components=1
  rm -f "${tools_dir}/go.tar.gz"
fi
export PATH="${go_root}/bin:${PATH}"
export GOTOOLCHAIN=local
# An ambient GOROOT (e.g. mise shims) mixes its stdlib with the pinned
# toolchain and breaks the build; the lane is hermetic.
unset GOROOT
# No -mod=mod: the Go default (-mod=readonly) must hold so this lane fails
# when go.mod/go.sum drift from the imports instead of silently repairing
# them in-worktree. (Locally a warm module cache masks the difference;
# CI's cold cache is where readonly refuses.) Verified: plain
# `go build ./...` passes at the module root.

printf 'go: %s\n' "$(go version)"

# Keyed by version: a pin bump must install the new binary instead of
# reusing a stale cached one. Installed with `go install` (module-proxy and
# checksum-DB verified) like govulncheck.
lint_bin="${tools_dir}/golangci-lint-${GOLANGCI_LINT_VERSION}"
if [ ! -x "${lint_bin}" ]; then
  printf 'go: installing golangci-lint %s\n' "${GOLANGCI_LINT_VERSION}"
  GOBIN="${tools_dir}" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
  mv "${tools_dir}/golangci-lint" "${lint_bin}"
fi
printf 'go: golangci-lint fmt --diff (gofumpt; fails on drift, rewrites nothing)\n'
"${lint_bin}" fmt --diff ./go/...
printf 'go: golangci-lint run\n'
"${lint_bin}" run ./go/...
# 500-line cap on every .go file, tracked or not yet added, minus gitignored
# paths (.tools holds the Go toolchain sources). No allowlist.
printf 'go: 500-line cap on .go files\n'
over=""
while IFS= read -r f; do
  [ -f "${f}" ] || continue
  lines="$(wc -l <"${f}")"
  if [ "${lines}" -gt 500 ]; then
    over="${over}${f}: ${lines} lines\n"
  fi
done < <(git ls-files --cached --others --exclude-standard -- '*.go')
if [ -n "${over}" ]; then
  printf '%b' "${over}"
  printf 'go: FAIL: .go files over 500 lines\n'
  exit 1
fi
printf 'go: go vet\n'
go vet ./go/...
printf 'go: go mod tidy -diff (fails on drift, repairs nothing)\n'
go mod tidy -diff
# Keyed by version: a pin bump must install the new binary instead of
# reusing a stale cached one.
vuln_bin="${tools_dir}/govulncheck-${GOVULNCHECK_VERSION}"
if [ ! -x "${vuln_bin}" ]; then
  printf 'go: installing govulncheck %s\n' "${GOVULNCHECK_VERSION}"
  GOBIN="${tools_dir}" go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
  mv "${tools_dir}/govulncheck" "${vuln_bin}"
fi
printf 'go: govulncheck\n'
"${vuln_bin}" ./go/...
# -race subsumes a plain run. Coverage is not gated: the module has no
# production statements (`go test -cover` prints "[no statements]"), so
# there is no measured number to floor.
printf 'go: go test -race\n'
go test ./go/... -race -count=1
printf 'go: pass\n'
