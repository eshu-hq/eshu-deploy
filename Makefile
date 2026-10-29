# eshu-deploy local checks

# `make ci` runs the same lane bodies the PR gates run
# (.github/workflows/ci.yml calls these scripts, so local and CI judge with
# the same validators). Run it before pushing; CI must never be the first
# test of an unproven change.
#
# Required local tools (CONTRIBUTING.md): python3 with yamllint, shellcheck,
# helm, docker, kubectl is bootstrapped. Pinned downloads (actionlint,
# kubeconform, kubectl) land in .tools/ (gitignored) on first run.

.PHONY: ci ci-lint ci-charts ci-compose ci-manifests ci-go help

ci: ci-lint ci-charts ci-compose ci-manifests ci-go ## Run all PR lanes locally.

ci-lint: ## YAML lint/format, actionlint, shellcheck.
	bash scripts/ci/lint.sh

ci-charts: ## helm lint + template render checks.
	bash scripts/ci/charts.sh

ci-compose: ## docker compose config per documented invocation.
	bash scripts/ci/compose.sh

ci-manifests: ## kustomize builds + helm renders + pinned-schema kubeconform.
	bash scripts/ci/manifests.sh

ci-go: ## gofmt, vet, tidy check, govulncheck, Go contract suite (pinned toolchain).
	bash scripts/ci/go.sh

help: ## List targets.
	@grep -E '^[a-z-]+: ## ' $(MAKEFILE_LIST) | sort | awk -F ': ## ' '{printf "%-14s %s\n", $$1, $$2}'
