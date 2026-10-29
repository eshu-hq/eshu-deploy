#!/usr/bin/env bash
# compose lane: docker compose config validation per Compose file.
#
# Each case mirrors a documented invocation (docs/run-locally/docker-compose.md
# and the file headers): standalone stacks validate alone, overlays validate
# combined with the base they layer on, and remote E2E entrypoints validate
# with the example env file. Standalone validation of an overlay is
# meaningless by design (it is a partial service set), so overlays appear only
# in their documented combos. compose/fragments/* has no in-tree consumer;
# yamllint judges its syntax in the lint lane until a consumer lands, at which
# point the consumer's combo case is added here.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"

# The e2e stack's `extends:` entries anchor at the checkout so the same file
# renders from any project directory (eshu CI drives it with its own root).
export ESHU_DEPLOY_CHECKOUT="${ESHU_DEPLOY_CHECKOUT:-${repo_root}}"

command -v docker >/dev/null 2>&1 || { printf 'compose: docker is required (see CONTRIBUTING.md)\n' >&2; exit 1; }

check() {
  printf 'compose: docker compose %s config\n' "$*"
  docker compose "$@" config -q
}

# Standalone stacks.
check -f compose/docker-compose.yaml
check -f compose/docker-compose.neo4j.yml
check -f compose/docker-compose.live-backend-neo4j.yml
check -f compose/docker-compose.live-backend-nornicdb.yml
check -f compose/docker-compose.demo.yaml
check -f compose/docker-compose.demo.corpus.yaml
check -f compose/docker-compose.e2e.yaml

# Documented overlays on their bases.
check -f compose/docker-compose.e2e.yaml -f compose/docker-compose.e2e.mutation.yaml
check -f compose/docker-compose.yaml -f compose/docker-compose.telemetry.yml
check -f compose/docker-compose.neo4j.yml -f compose/docker-compose.telemetry.yml
check -f compose/docker-compose.yaml -f compose/docker-compose.tier2-tfstate.yaml
check -f compose/docker-compose.yaml -f compose/docker-compose.tier2-tfstate-v25.yaml
check -f compose/docker-compose.yaml -f compose/docker-compose.read-api-latency-gate.yaml
check -f compose/docker-compose.neo4j.yml -f compose/docker-compose.read-api-latency-gate.yaml

# Remote E2E entrypoint and its opt-in overlays (example env file only).
check --env-file .env.remote-e2e.example -f compose/docker-compose.remote-e2e.yaml
check --env-file .env.remote-e2e.example -f compose/docker-compose.remote-e2e.yaml \
  -f compose/docker-compose.remote-e2e.observability.yaml
check --env-file .env.remote-e2e.example -f compose/docker-compose.remote-e2e.yaml \
  -f compose/docker-compose.remote-e2e.pprof.yaml
check --env-file .env.remote-e2e.example -f compose/docker-compose.remote-e2e.yaml \
  -f compose/docker-compose.remote-e2e.observability.yaml \
  -f compose/docker-compose.remote-e2e.observability.pprof.yaml

printf 'compose: pass\n'
