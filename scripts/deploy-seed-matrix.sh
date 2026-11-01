#!/usr/bin/env bash
# Seeded RED/GREEN matrix for the eshu-deploy CI lanes.
#
# Local runner, not a CI gate: for each seed it plants a defect, runs the real
# lane body from scripts/ci/, asserts the lane exits non-zero AND its output
# names the seed, removes the seed, and asserts the lane exits zero again. It
# never compares the tree against a committed count or list.
#
# 19 checks (RED + GREEN each, plus one GREEN baseline):
#   yaml     a YAML syntax error (new file)        -> scripts/ci/lint.sh
#   chart    a template that cannot render (new)   -> scripts/ci/charts.sh
#   compose  an unknown top-level Compose key      -> scripts/ci/compose.sh
#   manifest an unknown field in a Service         -> scripts/ci/manifests.sh
#   version  a values.yaml edit, no version bump   -> scripts/ci/verify-chart-version-bumped.sh
#   lint     an unused function (new .go file)     -> scripts/ci/go.sh (golangci-lint run)
#   format   a gofumpt-only violation (new .go)    -> scripts/ci/go.sh (golangci-lint fmt --diff)
#   cap      a 501-line .go file (new)             -> scripts/ci/go.sh (500-line cap)
#   race     a racy test (new _test.go file)       -> scripts/ci/go.sh (go test -race)
# The compose and manifest seeds edit a file the lane already validates (the
# lanes name their inputs, so a new file would never be read); the version
# seed needs the published chart, read anonymously from ghcr.io. The four Go
# seeds are new files in go/internal/runtime (package runtime) and run the
# whole go lane, so they need the lane's network on a cold cache (Go
# toolchain tarball, `go install` of golangci-lint and govulncheck, the
# vulnerability database) and a C compiler for -race. The lane downloads the
# linux/amd64 toolchain, so a non-amd64 host needs the pinned toolchain
# already unpacked in .tools/ (or the checks report SKIPPED).
#
# A check whose tool or network is unavailable prints SKIPPED with the reason
# and is counted separately; it is never reported as PASS.
#
# Tooling: yamllint must be importable by python3 (the lint lane runs
# `python3 -m yamllint`), so put a venv with yamllint on PATH first. The
# manifests lane downloads linux/amd64 tools and refuses other hosts; on such
# a host this runner reports SKIPPED rather than bypassing the lane's guard.
#
# Usage: bash scripts/deploy-seed-matrix.sh
# Leaves the tree exactly as found: every seed is removed (new files deleted,
# edited files restored with `git checkout --`) on every exit path including
# INT and TERM, and the final `git status --porcelain` is compared with the
# start snapshot.
set -euo pipefail

cd "$(dirname "$0")/.."

TMP="$(mktemp -d)"
LOG="$TMP/last.log"
pass=0
fail=0
skipped=0

# Files this runner creates (new seeds) and edits (restored from HEAD).
go_seed_dir=go/internal/runtime
new_seeds=(
	zz_seed_yaml.yaml
	helm/eshu/templates/zz_seed_unrenderable.yaml
	"$go_seed_dir/zz_seed_lint.go"
	"$go_seed_dir/zz_seed_format.go"
	"$go_seed_dir/zz_seed_cap.go"
	"$go_seed_dir/zz_seed_race_test.go"
)
edited=(compose/docker-compose.yaml gitops/manifests/minimal/service.yaml helm/eshu/values.yaml)

cleanup() {
	local f
	for f in "${new_seeds[@]}"; do
		rm -f "$f"
	done
	git checkout -q -- "${edited[@]}" 2>/dev/null || true
	rm -rf "$TMP"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Refuse to start over leftovers or local edits: the trap would destroy them.
status_now="$(git status --porcelain --untracked-files=all)"
if [[ "$status_now" == *zz_seed_* ]]; then
	echo "FAIL: leftover seed files present; refusing to seed"
	printf '%s\n' "$status_now"
	exit 1
fi
if ! git diff --quiet -- "${edited[@]}"; then
	echo "FAIL: files this runner edits already have local changes; refusing to seed"
	git diff --stat -- "${edited[@]}"
	exit 1
fi
SNAPSHOT="$TMP/status-before.txt"
printf '%s\n' "$status_now" >"$SNAPSHOT"

ok() { # $1: check name
	echo "PASS: $1"
	pass=$((pass + 1))
}
bad() { # $1: check name
	echo "FAIL: $1"
	tail -n 20 "$LOG" 2>/dev/null || true
	fail=$((fail + 1))
}
skip() { # $1: check name, $2: reason
	echo "SKIPPED: $1 ($2)"
	skipped=$((skipped + 1))
}

# red_check NAME PATTERN CMD...: CMD must exit non-zero and its output must
# contain PATTERN (a finding that names the seed).
red_check() {
	local name=$1 pattern=$2
	shift 2
	if "$@" >"$LOG" 2>&1; then
		bad "$name (lane stayed green)"
	elif [[ "$(<"$LOG")" == *"$pattern"* ]]; then
		ok "$name"
	else
		bad "$name (red, but no finding names the seed)"
	fi
}

# green_check NAME CMD...: CMD must exit zero.
green_check() {
	local name=$1
	shift
	if "$@" >"$LOG" 2>&1; then
		ok "$name"
	else
		bad "$name"
	fi
}

# edit_applied FILE: the seed edit must have changed the file, or the RED
# result would prove nothing.
edit_applied() {
	if git diff --quiet -- "$1"; then
		echo "FAIL: seed edit did not change $1"
		fail=$((fail + 1))
		return 1
	fi
}

# The lane commands: the real lane bodies CI and `make ci` run.
lane_lint() { bash scripts/ci/lint.sh; }
lane_charts() { bash scripts/ci/charts.sh; }
lane_compose() { bash scripts/ci/compose.sh; }
lane_manifests() { bash scripts/ci/manifests.sh; }
lane_version() { bash scripts/ci/verify-chart-version-bumped.sh; }
lane_go() { bash scripts/ci/go.sh; }

# --- prerequisites: SKIPPED with a reason, never a silent pass ---
lint_skip=""
if ! command -v python3 >/dev/null 2>&1 || ! python3 -m yamllint --version >/dev/null 2>&1; then
	lint_skip="yamllint is not importable by python3; put a venv with it on PATH"
elif ! command -v shellcheck >/dev/null 2>&1; then
	lint_skip="shellcheck is not on PATH"
fi
helm_skip=""
command -v helm >/dev/null 2>&1 || helm_skip="helm is not on PATH"
compose_skip=""
if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
	compose_skip="docker compose is not available"
fi
manifests_skip="$helm_skip"
if [ "$(uname -s)/$(uname -m)" != "Linux/x86_64" ]; then
	manifests_skip="the manifests lane downloads linux/amd64 tools only (host $(uname -s)/$(uname -m))"
fi

# The go lane downloads the linux/amd64 toolchain unless the pinned one is
# already unpacked, and -race needs cgo (a C compiler).
go_skip=""
go_version="$(. scripts/ci/versions.sh && printf '%s' "$GO_VERSION")"
if [ "$(uname -s)/$(uname -m)" != "Linux/x86_64" ] && [ ! -x ".tools/go${go_version}/bin/go" ]; then
	go_skip="the go lane downloads a linux/amd64 toolchain (host $(uname -s)/$(uname -m)); unpack go${go_version} for this host into .tools/go${go_version}"
elif ! command -v gcc >/dev/null 2>&1 && ! command -v cc >/dev/null 2>&1; then
	go_skip="no C compiler on PATH (go test -race needs cgo)"
fi

# The version guard reads the published chart: prove it is readable first.
version_skip="$helm_skip"
if [ -z "$version_skip" ]; then
	registry="${CHART_REGISTRY:-oci://ghcr.io/${GITHUB_REPOSITORY_OWNER:-eshu-hq}/charts}"
	chart_version="$(awk '/^version:/ {print $2; exit}' helm/eshu/Chart.yaml)"
	mkdir -p "$TMP/precheck"
	if ! helm pull "$registry/eshu" --version "$chart_version" --destination "$TMP/precheck" >"$TMP/precheck.out" 2>&1; then
		version_skip="chart $chart_version not readable from $registry: $(tr '\n' ' ' <"$TMP/precheck.out")"
	fi
fi

# --- yaml: a syntax error turns the lint lane red (yamllint names the file) ---
if [ -n "$lint_skip" ]; then
	skip "yaml RED on seeded syntax error" "$lint_skip"
	skip "yaml GREEN after seed removal" "$lint_skip"
else
	printf 'zz_seed_key: [unclosed\n' >zz_seed_yaml.yaml
	red_check "yaml RED on seeded syntax error" "zz_seed_yaml.yaml" lane_lint
	rm -f zz_seed_yaml.yaml
	green_check "yaml GREEN after seed removal" lane_lint
fi

# --- chart: a template that cannot render turns the charts lane red ---
if [ -n "$helm_skip" ]; then
	skip "chart RED on seeded unrenderable template" "$helm_skip"
	skip "chart GREEN after seed removal" "$helm_skip"
else
	printf '{{ fail "zz_seed_unrenderable" }}\n' >helm/eshu/templates/zz_seed_unrenderable.yaml
	red_check "chart RED on seeded unrenderable template" "zz_seed_unrenderable" lane_charts
	rm -f helm/eshu/templates/zz_seed_unrenderable.yaml
	green_check "chart GREEN after seed removal" lane_charts
fi

# --- compose: an unknown top-level key fails `docker compose config` ---
if [ -n "$compose_skip" ]; then
	skip "compose RED on seeded invalid Compose file" "$compose_skip"
	skip "compose GREEN after seed removal" "$compose_skip"
else
	printf 'zz_seed_invalid_key: true\n' >>compose/docker-compose.yaml
	if edit_applied compose/docker-compose.yaml; then
		red_check "compose RED on seeded invalid Compose file" "zz_seed_invalid_key" lane_compose
	fi
	git checkout -q -- compose/docker-compose.yaml
	green_check "compose GREEN after seed removal" lane_compose
fi

# --- manifest: an unknown Service field fails kubeconform -strict ---
if [ -n "$manifests_skip" ]; then
	skip "manifest RED on seeded bad manifest" "$manifests_skip"
	skip "manifest GREEN after seed removal" "$manifests_skip"
else
	sed -i 's/^  type: ClusterIP$/&\n  zzSeedBadField: true/' gitops/manifests/minimal/service.yaml
	if edit_applied gitops/manifests/minimal/service.yaml; then
		red_check "manifest RED on seeded bad manifest" "zzSeedBadField" lane_manifests
	fi
	git checkout -q -- gitops/manifests/minimal/service.yaml
	green_check "manifest GREEN after seed removal" lane_manifests
fi

# --- version: a chart content edit with no version bump fails the guard ---
if [ -n "$version_skip" ]; then
	skip "version RED on values.yaml edit without bump" "$version_skip"
	skip "version GREEN after restore" "$version_skip"
else
	printf '# zz_seed_values_edit\n' >>helm/eshu/values.yaml
	if edit_applied helm/eshu/values.yaml; then
		red_check "version RED on values.yaml edit without bump" "values.yaml" lane_version
	fi
	git checkout -q -- helm/eshu/values.yaml
	green_check "version GREEN after restore" lane_version
fi

# --- Go seeds: each is a new file in package runtime, removed after its check ---
# go_seed NAME FILE PATTERN: RED must exit non-zero and name PATTERN; after
# the seed is removed the lane must pass again.
go_seed() {
	local name=$1 file=$2 pattern=$3
	red_check "$name RED on seeded defect" "$pattern" lane_go
	rm -f "$file"
	green_check "$name GREEN after seed removal" lane_go
}

if [ -n "$go_skip" ]; then
	skip "go lane GREEN on clean tree (baseline)" "$go_skip"
	for n in lint format cap race; do
		skip "go $n RED on seeded defect" "$go_skip"
		skip "go $n GREEN after seed removal" "$go_skip"
	done
else
	# Warm the tool cache and prove the clean tree passes before seeding, so a
	# RED below is the seed's doing and not a cold-cache or broken-lane failure.
	green_check "go lane GREEN on clean tree (baseline)" lane_go

	# lint: an unused function; fmt and vet accept it, golangci-lint does not.
	printf 'package runtime\n\nfunc zzSeedUnused() {}\n' >"$go_seed_dir/zz_seed_lint.go"
	go_seed "go lint" "$go_seed_dir/zz_seed_lint.go" "zzSeedUnused"

	# format: a blank first line in a function body is gofmt-clean but a
	# gofumpt violation, so only `golangci-lint fmt --diff` catches it.
	printf 'package runtime\n\nfunc zzSeedFormat() {\n\n\t_ = 1\n}\n' >"$go_seed_dir/zz_seed_format.go"
	go_seed "go format" "$go_seed_dir/zz_seed_format.go" "zz_seed_format.go"

	# cap: 501 lines of comments, otherwise clean (gofumpt would regroup
	# repeated single-line var declarations and trip the format gate first).
	{
		echo "package runtime"
		echo ""
		for i in $(seq 1 499); do echo "// zz seed filler line $i"; done
	} >"$go_seed_dir/zz_seed_cap.go"
	if [ "$(wc -l <"$go_seed_dir/zz_seed_cap.go")" -ne 501 ]; then
		echo "FAIL: cap seed is not 501 lines"
		fail=$((fail + 1))
	fi
	go_seed "go cap" "$go_seed_dir/zz_seed_cap.go" "zz_seed_cap.go: 501 lines"

	# race: two goroutines increment one int with no synchronisation.
	cat >"$go_seed_dir/zz_seed_race_test.go" <<'GOSEED'
package runtime

import (
	"sync"
	"testing"
)

func TestZzSeedRace(t *testing.T) {
	var wg sync.WaitGroup
	counter := 0
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}
	wg.Wait()
	if counter < 0 {
		t.Fatal("unreachable")
	}
}
GOSEED
	go_seed "go race" "$go_seed_dir/zz_seed_race_test.go" "DATA RACE"
fi

echo "seed matrix: pass=$pass fail=$fail skipped=$skipped"
if ! diff -u "$SNAPSHOT" <(git status --porcelain --untracked-files=all) >"$TMP/status-diff.txt" 2>&1; then
	echo "FAIL: tree changed by the matrix run:"
	cat "$TMP/status-diff.txt"
	exit 1
fi
[ "$fail" -eq 0 ]
