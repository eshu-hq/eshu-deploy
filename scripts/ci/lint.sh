#!/usr/bin/env bash
# lint lane: YAML lint/format, actionlint, shellcheck over the whole tree.
#
# No new-code-only mode: every file is judged, findings are fixed, and the
# only exclusion is helm/eshu/templates/** in .yamllint.yml (Go template
# directives do not parse as YAML; helm lint/template/kubeconform judge them
# instead, with the reason stated in that config).
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
# shellcheck source=versions.sh
. scripts/ci/versions.sh

command -v python3 >/dev/null 2>&1 || { printf 'lint: python3 is required (see CONTRIBUTING.md)\n' >&2; exit 1; }
if ! python3 -m yamllint --version >/dev/null 2>&1; then
  printf 'lint: yamllint is required: pip install "yamllint==%s"\n' "${YAMLLINT_VERSION}" >&2
  exit 1
fi
command -v shellcheck >/dev/null 2>&1 || { printf 'lint: shellcheck is required (see CONTRIBUTING.md)\n' >&2; exit 1; }

printf 'lint: yamllint %s\n' "$(python3 -m yamllint --version)"
python3 -m yamllint -s .

tools_dir="${repo_root}/.tools"
mkdir -p "${tools_dir}"
actionlint_bin="${tools_dir}/actionlint"
if [ ! -x "${actionlint_bin}" ]; then
  printf 'lint: downloading actionlint %s\n' "${ACTIONLINT_VERSION}"
  curl -sSL -o "${tools_dir}/actionlint.tar.gz" \
    "https://github.com/rhysd/actionlint/releases/download/v${ACTIONLINT_VERSION}/actionlint_${ACTIONLINT_VERSION}_linux_amd64.tar.gz"
  printf '%s  %s\n' "${ACTIONLINT_SHA256}" "${tools_dir}/actionlint.tar.gz" | sha256sum -c -
  tar -xzf "${tools_dir}/actionlint.tar.gz" -C "${tools_dir}" actionlint
  rm -f "${tools_dir}/actionlint.tar.gz"
fi
"${actionlint_bin}" -no-color .github/workflows/*.yml

printf 'lint: shellcheck %s\n' "$(shellcheck --version | awk '/^version:/{print $2}')"
# SC1090/SC1091: sourced version files resolve at lane runtime, not statically.
shellcheck -S warning -e SC1090,SC1091 --format gcc scripts/*.sh scripts/ci/*.sh

printf 'lint: JSON syntax (grafana dashboards)\n'
for dashboard in grafana/dashboards/*.json; do
  python3 -m json.tool "${dashboard}" >/dev/null
done

printf 'lint: pass\n'
