# Reference

Facts to look up: tested versions, settings, and chart values. For steps,
see [Guides](../guides/README.md). For first install, see
[Getting started](../getting-started/README.md).

## Tested version pairs

- [Compatibility Matrix](../compatibility-matrix.md) — the published
  cross-repository contract. It records the tested deployment artifact
  pairs. Check it before you mix versions.
- [Runtime Parity Matrix](../kubernetes/runtime-parity-matrix.md) — Compose
  and Helm parity detail behind the matrix.

## Helm values

The chart source of truth is
[`helm/eshu/values.yaml`](../../helm/eshu/values.yaml),
`values.schema.json`, and `templates/`. Start from
[Helm Values](../kubernetes/helm-values.md), which routes you to:

- [Routing and Storage Values](../kubernetes/helm-routing-and-storage-values.md)
  and [Storage](../kubernetes/storage.md) — Postgres, graph backend, PVCs,
  exposure, NetworkPolicy, observability.
- [Runtime Values](../kubernetes/helm-runtime-values.md) — schema bootstrap,
  global env, connection tuning, reducer lanes, repo sync.
- [Collector and Webhook Values](../kubernetes/helm-collector-and-webhook-values.md)
  — workflow coordinator, optional collectors, webhook listener.

## Settings you'll actually touch

Set only what your path needs. The full list lives in
[Helm Values](../kubernetes/helm-values.md), the chart defaults in
[`helm/eshu/values.yaml`](../../helm/eshu/values.yaml), and the local examples in
[`.env.example`](../../.env.example). These are the usual suspects:

| Name | When you need it | What it does |
| --- | --- | --- |
| `image.repository` / `image.tag` | Helm | The core Eshu image most workloads run. Defaults to `ghcr.io/eshu-hq/eshu` at `v0.0.3-pre-release-18`. Pin your own build for production. |
| `contentStore.secretName` / `contentStore.dsnKey` | Helm | Names the Secret and key that hold your Postgres connection string. |
| `neo4j.uri` | Helm | The Bolt address of your graph backend. The name is historical. Both graph options use this shape. |
| `ESHU_GRAPH_BACKEND` | Helm, in `env` | Picks your graph backend. Set it with the matching database names from the quickstart. |
| `ESHU_REPO_ROOT` | Compose builds | Points at your `eshu` checkout and its `Dockerfile`. |
| `ESHU_POSTGRES_PASSWORD`, `ESHU_NEO4J_PASSWORD` | Compose | Local database passwords. They ship as `change-me`, so change them before any non-local use. |
| `ESHU_API_KEY` | API auth | An optional explicit key for the API and MCP server. |
| `ESHU_HTTP_PORT` | If you want | Moves the local API port. It defaults to `8080`. |

## Provenance

- [Seed Provenance](../seed-provenance.md) — where the files in this repo
  came from, commit by commit.
