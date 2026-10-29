#!/usr/bin/env bash
# go lane: gofmt, go vet, go mod tidy drift check, govulncheck, and the
# Go contract test suite. Requires helm on PATH: the suite skips
# helm-dependent tests when it is absent, so CI provisions it pinned.
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
printf 'go: gofmt (no output below means clean)\n'
test -z "$(gofmt -l go/)"
printf 'go: go vet\n'
go vet ./go/...
printf 'go: go mod tidy -diff (fails on drift, repairs nothing)\n'
go mod tidy -diff
vuln_bin="${tools_dir}/govulncheck"
if [ ! -x "${vuln_bin}" ]; then
  printf 'go: installing govulncheck %s\n' "${GOVULNCHECK_VERSION}"
  GOBIN="${tools_dir}" go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
fi
printf 'go: govulncheck\n'
"${vuln_bin}" ./go/...
printf 'go: go test\n'
go test ./go/... -count=1
printf 'go: pass\n'
