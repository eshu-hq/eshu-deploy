# Seed Provenance

Seed of `eshu-deploy` from the contracted in-tree deployment assets for
`eshu-hq/eshu#8871` (cutover step 1: population plus independent
build and release proof).

- Source: `eshu-hq/eshu` `origin/main` at
  `e4c385f493494ed9d6c7f192b04ce0255361f621` (contains the merged deploy
  contract and census).
- Destination base: `eshu-hq/eshu-deploy` `main` at `9f18dca`
  (`Initial commit`, README only).
- Chart templates: 102 (`helm/eshu/templates`), matching the census.
- Compose lanes: 20 root files in `compose/` (18 in the census plus
  `docker-compose.tier2-tfstate.yaml` and
  `docker-compose.tier2-tfstate-v25.yaml`, added in-tree after the census
  base) plus 4 fragments in `compose/fragments/`.

## Byte-identity checks

Every verbatim family was copied unchanged from the source SHA above:

```bash
diff -r <eshu>/deploy/helm/eshu helm/eshu
diff -r <eshu>/deploy/argocd gitops/argocd
diff -r <eshu>/deploy/manifests gitops/manifests
diff -r <eshu>/deploy/observability observability
diff -r <eshu>/deploy/grafana grafana
diff -r <eshu>/docs/public/deploy docs  # plus docs/run-locally, new here
cmp <eshu>/scripts/ci/push-helm-chart.sh scripts/ci/push-helm-chart.sh
```

All return empty output. New files in this seed (not copies): the rewired
`scripts/verify-compose-helm-runtime-parity.sh` (identical checks,
parameterized paths), `.github/workflows/chart-release.yml`,
`SECURITY.md`, `docs/compatibility-matrix.md`,
`docs/rollout-recovery.md`, this file, and the rewritten `README.md`.

## Independent build and release proof

From this repository root, with no eshu checkout on the path:

```bash
helm lint helm/eshu
helm template eshu helm/eshu --namespace eshu
bash scripts/verify-compose-helm-runtime-parity.sh
```

Release runs `.github/workflows/chart-release.yml`: lint, render, parity
verify, `helm package`, then `scripts/ci/push-helm-chart.sh` to
`oci://ghcr.io/<owner>/charts`. Chart versions are immutable and this
repository moves no tags; see `docs/rollout-recovery.md`.
