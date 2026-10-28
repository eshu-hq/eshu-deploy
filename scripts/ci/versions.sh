#!/usr/bin/env bash
# shellcheck disable=SC2034
# Pinned validator versions for eshu-deploy CI lanes.
# (SC2034: this file only assigns; the lane scripts consume the variables.)
#
# CI (.github/workflows/ci.yml) and local runs (make ci) source this file, so
# both judge with the same tool versions. Action pins (checkout, paths-filter,
# setup-python, setup-helm, action-shellcheck) live in the workflow with a
# version comment next to each SHA; tool pins live here. Re-verify every pin
# with git ls-remote (actions) or the publisher checksum file (tarballs) when
# auditing; Dependabot proposes action bumps, weekly grouped.
#
# Note on shellcheck: no publisher checksum file exists for koalaman
# tarballs, so the lint lane uses the pinned ludeeus/action-shellcheck action
# in CI and a PATH shellcheck locally (CONTRIBUTING.md lists the install).
# Both must pass.
set -euo pipefail

# yamllint via pip (CI installs it with setup-python; local: pip install).
YAMLLINT_VERSION=1.37.1
# actionlint tarball, sha256-checked on download (rhysd/actionlint).
ACTIONLINT_VERSION=1.7.12
ACTIONLINT_SHA256=8aca8db96f1b94770f1b0d72b6dddcb1ebb8123cb3712530b08cc387b349a3d8
# kubeconform tarball, sha256-checked on download (yannh/kubeconform).
KUBECONFORM_VERSION=0.8.0
KUBECONFORM_SHA256=9bc2bffbf71f261128533edaf912153948b7ff238f9a531ae6d34466ec287883
# kubectl binary, sha256-checked on download (dl.k8s.io publishes .sha256).
KUBECTL_VERSION=v1.34.2
KUBECTL_SHA256=9591f3d75e1581f3f7392e6ad119aab2f28ae7d6c6e083dc5d22469667f27253
# Kubernetes JSON schemas, commit-pinned (yannh/kubernetes-json-schema).
# The manifests lane replaces only the repo revision in kubeconform's default
# template with this commit; the strict standalone layout is unchanged.
SCHEMA_COMMIT=89a35a5e9a1412a692a2603d79c3786f602238a4
# Helm tool version (matches the chart-release lane).
HELM_VERSION=v4.2.3
