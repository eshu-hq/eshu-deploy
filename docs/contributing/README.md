# Contributing

The full guide is the root
[CONTRIBUTING.md](../../CONTRIBUTING.md). This page is the short version:
what to run, and where things go.

## Before you push

Run `make ci`. It runs the same checks as CI. You'll need `python3` with
`yamllint`, plus `shellcheck`, `helm`, and `docker`. Pinned tools bootstrap
into `.tools/` on first run. The Go version is pinned in `go.mod`.

CI is path-aware: only the lanes your change can affect will run, and
`ci-complete` is the one required check. Docs-only changes merge without
lane runs. The lane table lives in
[CONTRIBUTING.md](../../CONTRIBUTING.md).

## What goes where

- Chart changes go in `helm/eshu/`. Compose stacks go in `compose/`.
  GitOps examples go in `gitops/`.
- Docs go in `docs/`, following the buckets in the
  [docs index](../README.md). Existing guide pages under
  `docs/kubernetes/`, `docs/run-locally/`, and `docs/eks/` stay where
  they are.
- When you cut a release, record the new tested pair in the
  [Compatibility Matrix](../compatibility-matrix.md). See
  [Rollout and Recovery](../rollout-recovery.md).
- Hard-to-reverse deployment choices get a record in
  [Decisions](../decisions/README.md).

## Ground rules

- No secrets in Git. Examples hold names and placeholders only.
- Fixes to the Eshu programs belong in
  [eshu](https://github.com/eshu-hq/eshu), not here. They reach this repo
  as new images.
- Found a vulnerability? Use the [Security policy](../../SECURITY.md),
  not a public issue.
