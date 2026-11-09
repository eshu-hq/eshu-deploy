# Architecture

How this deployment repo is put together, and where its edges are. For the
ideas in plain words, read [Concepts](../concepts/README.md). For the
pictures, see [Diagrams](../diagrams/README.md).

## The design in one paragraph

You supply settings and secrets. These assets start the Eshu services. A
setup job prepares the databases, the ingester stores facts in Postgres, and
the reducer links them into your graph backend. You read it back through the
API and MCP server. Postgres holds facts and queues. The graph backend holds
the linked truth. Collectors are optional watchers on the side, off until
you turn them on.

## What you bring, and what this repo never does

You bring the Eshu release images, a cluster or a Docker host, a Postgres
database, a graph backend, and your secrets. This repo talks to your secret
store by reference only. Examples here hold names and placeholders, never real
credentials, so don't commit yours either.

This repo never touches Eshu source code. Fixes to the programs land in `eshu`
and reach you as new images. It also never creates your cloud account,
cluster, or database for you. And if you leave a collector disabled, nothing
here will go near that system.

Two exposure choices are yours on Kubernetes. The API Service is `ClusterIP`
by default, and Ingress stays off. You decide how to expose it. Before you
do, work through the
[Production Checklist](../kubernetes/production-checklist.md).

## Where things live in this repo

The chart lives in `helm/eshu/`. Compose stacks live in `compose/`, and GitOps
examples live in `gitops/`. You'll find dashboards and alerts in `grafana/`
and `observability/`, guides in `docs/`, and check helpers in `scripts/`. The
chart release workflow, `.github/workflows/chart-release.yml`, pushes the chart
to `oci://ghcr.io/<owner>/charts`.

One path warning. Some older guide pages still show `deploy/helm/eshu`,
`deploy/argocd`, or `deploy/manifests/minimal`. In this repo, use `helm/eshu`,
`gitops/argocd`, and `gitops/manifests/minimal` instead.

## How the pieces connect

- Laptop shape: [Docker Compose guide](../run-locally/docker-compose.md).
- Shared-service shape: [Deploy to Kubernetes](../kubernetes/index.md) and the
  [Helm Quickstart](../kubernetes/helm-quickstart.md).
- From Git: start from `gitops/argocd/base/` and read
  [Argo CD and GitOps](../kubernetes/argocd.md).
- Where these files came from: [Seed Provenance](../seed-provenance.md).
- Which versions were tested together:
  [Compatibility Matrix](../compatibility-matrix.md).
