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
| `go/**`, `go.mod`, `go.sum` | `lint`, `go` (gofmt, vet, tidy check, govulncheck, tests) |
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
Pinned downloads (actionlint, kubeconform, kubectl) bootstrap into `.tools/`
on first run. PR lanes use a read-only token and no secrets, so fork PRs run
the same gates.
