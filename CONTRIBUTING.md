# Contributing to eshu-deploy

## CI lanes

Every PR runs path-aware gates (`.github/workflows/ci.yml`). A `changes` job
maps touched paths to lanes; only lanes a PR can affect run, and `ci-complete`
is the one required context (it passes when needed lanes passed and the rest
were legitimately skipped, so docs-only PRs merge without lane runs).

| PR touches | Lanes |
|---|---|
| Docs only (`**/*.md` except `docs/run-locally/docker-compose.md`, `LICENSE`) | none; `ci-complete` succeeds |
| `charts/**`, `helm/**` | `lint`, `charts` |
| `compose/**`, `docker-compose*.y*ml` | `lint`, `compose` |
| `manifests/**`, `gitops/**`, `examples/**` | `lint`, `manifests` |
| `**/*.sh` | `lint` (shellcheck) |
| `go/**`, `go.mod`, `go.sum`, `.golangci.yml` | `lint`, `go` (gofumpt, golangci-lint, 500-line cap, vet, tidy check, govulncheck, tests with `-race`) |
| `scripts/ci/versions.sh` | `lint`, `go`, `charts`, `manifests` (sourced by every lane body) |
| `scripts/ci/go.sh` | `lint`, `go` (a lane-body change must run its lane) |
| `helm/**`, `compose/**` | owning lane(s) plus `go` (contract suite pins these trees) |
| `docs/run-locally/docker-compose.md`, `.env.remote-e2e.example` | owning lane(s) plus `go` (read by the contract suite) |
| `.github/workflows/**` | all lanes (proves a lane still passes) |
| Push to `main` / weekly schedule | all lanes |

Lane bodies live in `scripts/ci/*.sh`, called by both CI and `make ci`.
There are no static gates: no lane compares the tree against a committed
count, list, or parity file. Validation runs on the whole tree
(no new-code-only mode); findings are fixed, and any exclusion states its
reason plus a linked open issue (see `.yamllint.yml` and `manifests.sh`).

## Local checks

Run `make ci` before pushing. It needs: `python3` with
`pip install "yamllint==1.37.1"`, `shellcheck`, `helm`, and `docker`.
Pinned downloads (actionlint, kubeconform, kubectl, the Go toolchain) and
`go install`ed tools (golangci-lint, govulncheck) bootstrap into `.tools/` on
first run. The go lane also needs a C compiler (`gcc`) for `go test -race`,
and downloads a linux/amd64 Go toolchain: on another host, unpack the pinned
`GO_VERSION` toolchain for that host into `.tools/go<GO_VERSION>` first. PR lanes use a read-only token and no secrets, so fork PRs run
the same gates.

## Proving the lanes catch defects

`bash scripts/deploy-seed-matrix.sh` is a local runner, not a CI gate. For
each of nine seeds (a YAML syntax error, an unrenderable chart template, an
invalid Compose file, a bad Kubernetes manifest, a `helm/eshu/values.yaml`
edit with no chart version bump, and four Go seeds: an unused function, a
gofumpt-only format violation, a 501-line `.go` file, and a data race) it runs the real lane body from
`scripts/ci/`, requires it to fail and name the seed, removes the seed, and
requires the lane to pass again. A check whose tool or network is missing
prints `SKIPPED` with the reason instead of passing. The runner refuses to
start over leftover seeds or local edits to the files it seeds, and restores
the tree on every exit path.

## Go lane

`scripts/ci/go.sh` (job `go`, `make ci-go`) gates, in order: `golangci-lint
fmt --diff` (gofumpt), `golangci-lint run` (`.golangci.yml`: govet,
staticcheck, ineffassign, unused, errcheck; no exclusions), a 500-line cap on
every `.go` file, `go vet`, `go mod tidy -diff`, `govulncheck`, and `go test
-race`. Tool versions are pinned in `scripts/ci/versions.sh`.

Not applicable, with the reason:

- **Coverage floor.** The module has zero production statements (`go test
  -cover` prints `coverage: [no statements]`; the tree is `doc.go` plus
  contract tests). There is no measured number to floor, and none is invented.
  Add a floor measured from the real figure when production code lands.
- **gosec.** It scans non-test code by default, which today is one 8-line
  `doc.go` (0 issues); it would pass without gating anything. With `-tests` it
  reports 24 findings (G306, G101, G304, G204, G301) in test fixtures that
  write executable stub scripts and read repo files by design, and passing
  that would need exclusions. Adopt it when production Go lands.
