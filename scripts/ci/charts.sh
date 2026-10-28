#!/usr/bin/env bash
# charts lane: helm lint plus a helm template render check per values set.
#
# Rendered output is schema-validated in the manifests lane; this lane proves
# the chart lints and renders. No cluster, no push.
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

printf 'charts: pass\n'
