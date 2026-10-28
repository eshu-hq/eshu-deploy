# eshu-deploy

Eshu deployment assets (Helm, Compose, GitOps examples). Seeded from
`eshu-hq/eshu` for `eshu-hq/eshu#8871`; holds no cutover authority until its
reviewed cutover.

## Layout

- `helm/eshu/` — the supported Kubernetes chart (chart, values, schema, CI
  overlay, all templates). Core-owned boundary: the `schema-bootstrap` Job
  template renders the core image and execs the core-owned bootstrap binary;
  schema and repair implementation stay in `eshu-hq/eshu`.
- `compose/` — every root Compose lane plus `compose/fragments/` from the
  run-locally set.
- `gitops/` — ArgoCD base and overlays plus the minimal raw manifests.
  Secret-reference examples only; no credential material.
- `observability/`, `grafana/` — collector config, alert rules, dashboards.
- `docs/` — deployment guides, the run-locally guides, the compatibility
  matrix that supersedes `docs/kubernetes/runtime-parity-matrix.md`, and the
  rollout and recovery procedures.
- `scripts/` — the chart push helper and the Compose/Helm runtime parity
  verifier, rewired to this layout.
- `.github/workflows/chart-release.yml` — the independent chart release
  lane: lint, render, parity verify, package, push to
  `oci://ghcr.io/<owner>/charts`.

## Proof

See `docs/seed-provenance.md` for the source SHA, the file inventory, and
the byte-identical and render-parity checks that prove this seed matches
the contracted in-tree source.
