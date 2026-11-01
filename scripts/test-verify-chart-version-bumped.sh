#!/usr/bin/env bash
# Hermetic self-test for verify-chart-version-bumped.sh, using a fake helm.
# Red: chart contents changed while the version is already published (the
# eshu-hq/eshu-deploy#8 failure mode) must fail and tell the author to bump.
# Green: identical contents under a published version, a bumped version that is
# not published yet, and a published version rebuilt from the same content with
# different archive timestamps must all pass.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
verifier="${repo_root}/scripts/ci/verify-chart-version-bumped.sh"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# make_chart <dir> <version> <values-line>: a minimal chart directory.
make_chart() {
	mkdir -p "$1/eshu"
	printf 'apiVersion: v2\nname: eshu\nversion: %s\n' "$2" >"$1/eshu/Chart.yaml"
	printf '%s\n' "$3" >"$1/eshu/values.yaml"
}

# The fake helm: `package` archives the chart dir into --destination with a
# caller-chosen mtime; `pull` copies the "published" archive or fails.
mkdir -p "${work}/bin"
cat >"${work}/bin/helm" <<'FAKE'
#!/usr/bin/env bash
set -euo pipefail
case "$1" in
package)
	chart="$2"; dest=""
	shift 2
	while [ $# -gt 0 ]; do
		[ "$1" = "--destination" ] && dest="$2"
		shift
	done
	version="$(awk '/^version:/ {print $2; exit}' "${chart}/Chart.yaml")"
	mkdir -p "${dest}"
	tar --sort=name --mtime="${FAKE_HELM_LOCAL_MTIME:-2026-01-01}" --owner=0 --group=0 \
		-czf "${dest}/eshu-${version}.tgz" -C "$(dirname "${chart}")" "$(basename "${chart}")"
	;;
pull)
	dest=""; want=""
	shift
	while [ $# -gt 0 ]; do
		[ "$1" = "--destination" ] && dest="$2"
		[ "$1" = "--version" ] && want="$2"
		shift
	done
	# Real helm pull fails when the destination directory does not exist.
	[ -d "${dest}" ] || { echo "Error: open ${dest}/eshu.tgz: no such file or directory" >&2; exit 1; }
	# A registry error that is not "version not found".
	[ -z "${FAKE_HELM_PULL_ERROR:-}" ] || { echo "${FAKE_HELM_PULL_ERROR}" >&2; exit 1; }
	# Only the published version exists, as in a real registry.
	[ -n "${FAKE_HELM_PUBLISHED:-}" ] && [ -f "${FAKE_HELM_PUBLISHED}" ] || { echo "not found" >&2; exit 1; }
	[ "$(basename "${FAKE_HELM_PUBLISHED}")" = "eshu-${want}.tgz" ] || { echo "version ${want} not found" >&2; exit 1; }
	cp "${FAKE_HELM_PUBLISHED}" "${dest}/$(basename "${FAKE_HELM_PUBLISHED}")"
	;;
esac
FAKE
chmod +x "${work}/bin/helm"

run_verifier() { # <chart parent dir> -> stdout/stderr in ${work}/out, ${work}/err
	PATH="${work}/bin:${PATH}" CHART_PATH="$1/eshu" CHART_REGISTRY="oci://example.invalid/charts" \
		"${verifier}" >"${work}/out" 2>"${work}/err"
}

# Published archive: version 1.0.0 with values "a".
make_chart "${work}/published" 1.0.0 "a: 1"
PATH="${work}/bin:${PATH}" helm package "${work}/published/eshu" --destination "${work}/pub-archive" >/dev/null
export FAKE_HELM_PUBLISHED="${work}/pub-archive/eshu-1.0.0.tgz"

# Red: same version, different content.
make_chart "${work}/changed" 1.0.0 "a: 2"
if run_verifier "${work}/changed"; then
	echo "expected the verifier to fail when content changed under a published version" >&2
	exit 1
fi
grep -Eq 'bump' "${work}/err" || { echo "expected a bump instruction, got:" >&2; cat "${work}/err" >&2; exit 1; }
grep -Eq '1\.0\.0' "${work}/err" || { echo "expected the version in the diagnostic, got:" >&2; cat "${work}/err" >&2; exit 1; }
grep -Eq 'values.yaml' "${work}/err" || { echo "expected the differing file named, got:" >&2; cat "${work}/err" >&2; exit 1; }

# Green: identical content, same version.
make_chart "${work}/same" 1.0.0 "a: 1"
run_verifier "${work}/same" || { echo "identical content under a published version must pass" >&2; cat "${work}/err" >&2; exit 1; }

# Green: identical content rebuilt with different archive timestamps.
FAKE_HELM_LOCAL_MTIME="2026-06-01" run_verifier "${work}/same" ||
	{ echo "a rebuild with different archive mtimes must still pass (compare contents, not bytes)" >&2; cat "${work}/err" >&2; exit 1; }

# Green: bumped version that is not published yet.
make_chart "${work}/bumped" 1.0.1 "a: 2"
run_verifier "${work}/bumped" || { echo "an unpublished bumped version must pass" >&2; cat "${work}/err" >&2; exit 1; }
grep -Eq 'not published' "${work}/out" || { echo "expected a not-published note, got:" >&2; cat "${work}/out" >&2; exit 1; }

# Green with a warning: the registry errors for a reason other than not-found.
FAKE_HELM_PULL_ERROR="dial tcp: connection refused" run_verifier "${work}/same" ||
	{ echo "a registry error must not fail the PR" >&2; cat "${work}/err" >&2; exit 1; }
grep -Eq 'could not read' "${work}/out" "${work}/err" ||
	{ echo "a registry error must be reported, not reported as 'not published'" >&2; cat "${work}/out" "${work}/err" >&2; exit 1; }
if grep -Eq 'not published' "${work}/out"; then
	echo "a registry error must not be described as 'not published'" >&2
	exit 1
fi

echo "test-verify-chart-version-bumped: pass"
