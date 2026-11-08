# Guides

Day-to-day running, once Eshu is up. For first install, start at
[Getting started](../getting-started/README.md). For settings to look up,
see [Reference](../reference/README.md).

## Operate: is it actually working?

Healthy pods don't mean you're done. Health tells you a process is up. It
doesn't tell you indexing finished. If answers come back empty, check the
schema Job, the ingester, and reducer progress in the
[Helm Quickstart](../kubernetes/helm-quickstart.md). Before you call a
shared service ready, work through the
[Production Checklist](../kubernetes/production-checklist.md).

Dashboards and alerts in this repo live in `grafana/` and `observability/`.

## Change values safely

If a render fails after you change values, run `helm template` again and fix
the first error it names. Re-render before every upgrade too. The values
pages are indexed in [Reference](../reference/README.md).

If a Compose build can't find the `Dockerfile`, your `ESHU_REPO_ROOT` is
probably wrong. Point it at your `eshu` checkout and try again.

## Upgrade, roll back, recover

- [Upgrades and Rollbacks](../kubernetes/upgrades-rollbacks.md) — move to a
  new chart or image, and come back if you must.
- [Rollout and Recovery](../rollout-recovery.md) — how releases roll out
  from this repo, and how recovery works. Note the line it draws: recovery
  of stored state uses core-owned repair commands from `eshu`, never
  anything in this repository.
- [Backup and Restore Proof](../kubernetes/backup-restore-proof.md) — prove
  backup and restore before you need them for real.
- [Helm Rollout Proof](../kubernetes/helm-rollout-proof.md) — evidence that
  a rollout did what it claimed.

## More task guides

- [Run locally](../run-locally/index.md) — laptop paths, compared.
- [Deploy to Kubernetes](../kubernetes/index.md) — the shared-service home.
- [Deploy To EKS](../eks/index.md) — the EKS variant.
- [Argo CD and GitOps](../kubernetes/argocd.md) — sync from Git.
- [Run an extracted collector](../kubernetes/run-an-extracted-collector.md) —
  add one collector at a time.
