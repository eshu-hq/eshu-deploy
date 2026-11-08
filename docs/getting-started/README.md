# Getting started

The shortest path to a running Eshu. Pick the row that matches where Eshu
will live, and follow the linked guide. The detail lives there, not here.

## On your laptop

You need Docker and a checkout of `eshu`, because Compose builds from its
`Dockerfile`:

```bash
export ESHU_REPO_ROOT="/path/to/eshu"
docker compose -f compose/docker-compose.yaml up --build
curl -sS http://localhost:8080/healthz
```

You're up when the API on port 8080 and the MCP server on port 8081 answer.
Then follow the [Docker Compose guide](../run-locally/docker-compose.md). It
lists the other Compose files and explains how to tell whether indexing
finished. The [Run locally](../run-locally/index.md) page compares the local
paths, including local binaries and local MCP.

## As a shared team service

Read [Prerequisites](../kubernetes/prerequisites.md) and
[Storage](../kubernetes/storage.md) first. You'll need a Postgres database
and a graph backend ready, because the chart installs neither for you. Then
follow the [Helm Quickstart](../kubernetes/helm-quickstart.md). Render before
you install:

```bash
helm template eshu ./helm/eshu -f values.eshu.yaml
helm upgrade --install eshu ./helm/eshu --namespace eshu -f values.eshu.yaml
```

You're up when the render is clean, the schema Job completes, and the API
rollout succeeds. On EKS, use the same install with the steps in
[Deploy To EKS](../eks/index.md). With Argo CD, start from
`gitops/argocd/base/` and read
[Argo CD and GitOps](../kubernetes/argocd.md).

## Before you start

- Copy what you need from [`.env.example`](../../.env.example), and change
  every `change-me` password before you expose the stack.
- New to Eshu itself? Learn what it does in
  [eshu](https://github.com/eshu-hq/eshu) first. This repo only runs it.
- Want the shapes first? [Diagrams](../diagrams/README.md) draws both paths.
