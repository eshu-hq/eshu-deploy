#!/usr/bin/env bash
# go lane: gofmt, go vet, and the Go contract test suite.
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
export GOFLAGS=-mod=mod

printf 'go: %s\n' "$(go version)"
printf 'go: gofmt (no output below means clean)\n'
test -z "$(gofmt -l go/)"
printf 'go: go vet\n'
go vet ./go/...
printf 'go: go test\n'
go test ./go/... -count=1
printf 'go: pass\n'
