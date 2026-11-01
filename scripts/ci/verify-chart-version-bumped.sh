#!/usr/bin/env bash
# Fails when the chart contents changed but Chart.yaml still carries a version
# that is already published (eshu-hq/eshu-deploy#27, from #8).
#
# Published chart versions are immutable (see scripts/ci/push-helm-chart.sh and
# docs/rollout-recovery.md). Without this check the drift only surfaced as a red
# post-merge release run that is not a required check, and the push step is
# skipped on pull_request events. The published chart is readable anonymously,
# so this runs on PRs with no package access.
#
# Contents are compared, not archive bytes: helm package embeds file times, so
# two builds of identical content need not be byte-identical.
#
# A version that cannot be pulled (not published yet, or the registry is
# unreachable) passes with a note; the release lane still refuses to overwrite.
#
# Env: CHART_PATH (default helm/eshu), CHART_REGISTRY (default
# oci://ghcr.io/${GITHUB_REPOSITORY_OWNER:-eshu-hq}/charts).
set -euo pipefail

chart_path="${CHART_PATH:-helm/eshu}"
registry="${CHART_REGISTRY:-oci://ghcr.io/${GITHUB_REPOSITORY_OWNER:-eshu-hq}/charts}"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

version="$(awk '/^version:/ {print $2; exit}' "${chart_path}/Chart.yaml")"
[ -n "${version}" ] || { echo "verify-chart-version-bumped: no version in ${chart_path}/Chart.yaml" >&2; exit 1; }

helm package "${chart_path}" --destination "${work}/local" >/dev/null
# helm pull needs the destination directory to exist.
mkdir -p "${work}/remote"
if ! helm pull "${registry}/eshu" --version "${version}" --destination "${work}/remote" >"${work}/pull.out" 2>&1; then
	if grep -Eqi 'not found|manifest unknown|404' "${work}/pull.out"; then
		echo "verify-chart-version-bumped: ${version} is not published yet; nothing to compare"
	else
		# Anything else (network, auth, a changed error text) must stay visible
		# rather than read as "not published", but must not block a PR.
		echo "::warning::verify-chart-version-bumped: could not read ${version} from ${registry}: $(tr '\n' ' ' <"${work}/pull.out")"
		echo "verify-chart-version-bumped: could not read the published chart; skipped"
	fi
	echo "verify-chart-version-bumped: pass"
	exit 0
fi

mkdir -p "${work}/local-x" "${work}/remote-x"
tar -mxzf "${work}"/local/eshu-*.tgz -C "${work}/local-x"
tar -mxzf "${work}"/remote/eshu-*.tgz -C "${work}/remote-x"
if diff -rq "${work}/remote-x" "${work}/local-x" >"${work}/diff" 2>&1; then
	echo "verify-chart-version-bumped: ${version} matches the published chart"
	echo "verify-chart-version-bumped: pass"
	exit 0
fi

{
	echo "::error::chart contents changed but version ${version} is already published; bump version in ${chart_path}/Chart.yaml (published chart versions are immutable)."
	echo "Files that differ from the published ${version}:"
	sed -e "s#${work}/remote-x#published#g" -e "s#${work}/local-x#local#g" "${work}/diff"
} >&2
exit 1
