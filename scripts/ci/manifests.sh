#!/usr/bin/env bash
# manifests lane: static validation of plain manifests and GitOps examples.
#
# No cluster. kustomize build proves the kustomizations compose, helm template
# proves the chart renders, and kubeconform -strict proves every core-kind
# resource matches its schema. Schemas come from a commit-pinned
# kubernetes-json-schema checkout (see versions.sh), never floating master.
#
# Skip list (short and explicit): Application (argoproj.io), ExternalSecret
# (external-secrets.io), and PrometheusRule (monitoring.coreos.com) are CRDs
# whose schemas ship with their operators; kubeconform parses these files and
# reports them skipped instead of failing the lane. Non-Kubernetes YAML
# (Prometheus alert rule files, the OTel collector config, Helm values) is
# judged by yamllint in the lint lane.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
# shellcheck source=versions.sh
. scripts/ci/versions.sh

if [ "$(uname -s)" != Linux ] || [ "$(uname -m)" != x86_64 ]; then
  printf 'manifests: pinned tool downloads support linux/amd64 only (got %s/%s)\n' \
    "$(uname -s)" "$(uname -m)" >&2
  exit 1
fi
command -v helm >/dev/null 2>&1 || { printf 'manifests: helm is required (see CONTRIBUTING.md)\n' >&2; exit 1; }

tools_dir="${repo_root}/.tools"
mkdir -p "${tools_dir}"

kubectl_bin="${tools_dir}/kubectl"
if [ ! -x "${kubectl_bin}" ]; then
  printf 'manifests: downloading kubectl %s\n' "${KUBECTL_VERSION}"
  curl -sSL -o "${kubectl_bin}" \
    "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
  curl -sSL -o "${tools_dir}/kubectl.sha256" \
    "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl.sha256"
  (cd "${tools_dir}" && printf '%s  %s\n' "$(cat kubectl.sha256)" kubectl | sha256sum -c -)
  chmod +x "${kubectl_bin}"
fi

kubeconform_bin="${tools_dir}/kubeconform"
if [ ! -x "${kubeconform_bin}" ]; then
  printf 'manifests: downloading kubeconform %s\n' "${KUBECONFORM_VERSION}"
  curl -sSL -o "${tools_dir}/kubeconform.tar.gz" \
    "https://github.com/yannh/kubeconform/releases/download/v${KUBECONFORM_VERSION}/kubeconform-linux-amd64.tar.gz"
  printf '%s  %s\n' "${KUBECONFORM_SHA256}" "${tools_dir}/kubeconform.tar.gz" | sha256sum -c -
  tar -xzf "${tools_dir}/kubeconform.tar.gz" -C "${tools_dir}" kubeconform
  rm -f "${tools_dir}/kubeconform.tar.gz"
fi

out_dir="$(mktemp -d)"
trap 'rm -rf "${out_dir}"' EXIT

printf 'manifests: kustomize build (gitops minimal + argocd base/overlays)\n'
"${kubectl_bin}" kustomize gitops/manifests/minimal > "${out_dir}/minimal.yaml"
"${kubectl_bin}" kustomize gitops/argocd/base > "${out_dir}/argocd-base.yaml"
"${kubectl_bin}" kustomize gitops/argocd/overlays/qa > "${out_dir}/argocd-qa.yaml"
"${kubectl_bin}" kustomize gitops/argocd/overlays/aws > "${out_dir}/argocd-aws.yaml"

printf 'manifests: helm template (default + CI values)\n'
helm template eshu helm/eshu --namespace eshu > "${out_dir}/render-default.yaml"
helm template eshu helm/eshu --namespace eshu \
  -f helm/eshu/ci/governance-two-team-k8s.values.yaml > "${out_dir}/render-ci.yaml"

printf 'manifests: kubeconform -strict (schemas pinned at %s)\n' "${SCHEMA_COMMIT}"
"${kubeconform_bin}" -strict -summary \
  -skip Application,ExternalSecret,PrometheusRule \
  -schema-location "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/${SCHEMA_COMMIT}/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json" \
  "${out_dir}/minimal.yaml" \
  "${out_dir}/argocd-base.yaml" \
  "${out_dir}/argocd-qa.yaml" \
  "${out_dir}/argocd-aws.yaml" \
  "${out_dir}/render-default.yaml" \
  "${out_dir}/render-ci.yaml" \
  observability/id-anchor-census-prometheus-rule.yaml \
  observability/graph-orphan-prometheus-rule.yaml \
  observability/hosted-operations-prometheus-rule.yaml \
  observability/prometheus-rule.yaml

printf 'manifests: pass\n'
