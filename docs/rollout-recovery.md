# Rollout and Recovery

This repository owns rollout and recovery for the deployment assets it
publishes. The release image, the schema-bootstrap and repair
implementation, and durable-state ownership stay core-owned by
`eshu-hq/eshu`; recovery of stored state uses the core-owned repair
commands, never anything in this repository.

## Rollout

1. Cut a chart release through `.github/workflows/chart-release.yml`: lint,
   render, and the Compose/Helm parity verifier run first; packaging and the
   OCI push (`scripts/ci/push-helm-chart.sh`) run only after they pass.
2. Record the published chart version and every image pin in
   `docs/compatibility-matrix.md` as a new tested pair.
3. Deploy through the reviewed GitOps path (see `gitops/` and
   `docs/kubernetes/argocd.md`). The `schema-bootstrap` Job renders the
   pinned core image and must reach `Complete` before long-lived workloads
   are trusted.

## Recovery

Chart versions are immutable: a published version is never overwritten, and
this repository moves no tags. Recovery restores a prior compatible
chart version and its prior values and image pins in configuration, then
redeploys through the same reviewed path. It never rewrites a release.

After recovery, verify the `schema-bootstrap` Job reaches `Complete` and
that API and graph reads agree with the pre-recovery state, then record the
before and after image pins with the verification output in the cutover
evidence.
