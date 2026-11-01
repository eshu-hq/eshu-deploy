#!/usr/bin/env bash
# charts lane: helm lint plus a helm template render check per values set.
#
# Rendered output is schema-validated in the manifests lane; this lane proves
# the chart lints and renders, and that a changed chart carries a version bump.
# No cluster, no push.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
# shellcheck source=versions.sh
. scripts/ci/versions.sh

command -v helm >/dev/null 2>&1 || { printf 'charts: helm is required (see CONTRIBUTING.md)\n' >&2; exit 1; }
printf 'charts: %s\n' "$(helm version --short)"

chart=helm/eshu
helm lint "${chart}"
helm template eshu "${chart}" --namespace eshu > /dev/null
helm template eshu "${chart}" --namespace eshu \
  -f "${chart}/ci/governance-two-team-k8s.values.yaml" > /dev/null

# Published chart versions are immutable: a content change needs a version bump
# (eshu-hq/eshu-deploy#27). The test is hermetic; the check reads the published
# chart anonymously.
bash scripts/test-verify-chart-version-bumped.sh
bash scripts/ci/verify-chart-version-bumped.sh

printf 'charts: pass\n'

# fork-pr lane proof, do not merge
