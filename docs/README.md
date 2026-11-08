# Documentation

Everything for running Eshu with the assets in this repo. New here? Start
with [Getting started](getting-started/README.md). Looking for what Eshu
itself does? That's in [eshu](https://github.com/eshu-hq/eshu). This repo
only runs it.

## The eight buckets

- [Getting started](getting-started/README.md) — the shortest path to a
  running stack, laptop or shared service.
- [Guides](guides/README.md) — day-to-day operating, upgrading, and
  recovering.
- [Concepts](concepts/README.md) — the deployment topologies in plain
  words.
- [Reference](reference/README.md) — tested versions, settings, and chart
  values to look up.
- [Architecture](architecture/README.md) — what this repo owns, and what
  it never does.
- [Diagrams](diagrams/README.md) — the Compose and Helm shapes, drawn and
  written out.
- [Decisions](decisions/README.md) — decision records for this repo.
  Format and index; none recorded yet.
- [Contributing](contributing/README.md) — local checks and pull requests.
  The full guide is the root [CONTRIBUTING.md](../CONTRIBUTING.md).

## Existing guide pages

These pages predate the buckets and stay where they are. The buckets above
link into them.

- [Run locally](run-locally/index.md) — laptop paths compared, with
  [Docker Compose](run-locally/docker-compose.md),
  [local binaries](run-locally/local-binaries.md), and
  [local MCP](run-locally/mcp-local.md).
- [Deploy to Kubernetes](kubernetes/index.md) — the shared-service home,
  including the [Helm Quickstart](kubernetes/helm-quickstart.md),
  [Prerequisites](kubernetes/prerequisites.md),
  [Storage](kubernetes/storage.md),
  [Helm Values](kubernetes/helm-values.md),
  [Argo CD and GitOps](kubernetes/argocd.md),
  [Upgrades and Rollbacks](kubernetes/upgrades-rollbacks.md), and the
  [Production Checklist](kubernetes/production-checklist.md).
- [Deploy To EKS](eks/index.md) — the EKS variant of the Helm path.
- [Compatibility Matrix](compatibility-matrix.md) — tested version pairs.
- [Rollout and Recovery](rollout-recovery.md) — releases and recovery.
- [Seed Provenance](seed-provenance.md) — where these files came from.
