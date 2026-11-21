// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"fmt"
	"strings"
	"testing"
)

// The #10031 remote-e2e driver needs a go-reachability compose shape before
// eshu-scanner-worker publishes its first go-reachability release
// (eshu-hq/eshu-deploy#40). The fragment below stays a skeleton: the seed
// image resolves from a required environment variable so no digest is
// invented, and the reconcile grants mirror the scanner RequiredGrants set
// the eshu core admission enforces.

const (
	goReachabilityFragment           = "fragments/docker-compose.scanner-go-reachability.yaml"
	goReachabilityWorkerService      = "collector-go-reachability"
	goReachabilitySeedService        = "collector-go-reachability-seed"
	goReachabilityReconcileService   = "collector-go-reachability-reconcile"
	goReachabilityFragmentInstanceID = "remote-e2e-go-reachability"
	goReachabilitySeedImageEnv       = "${ESHU_SCANNER_GOREACHABILITY_SEED_IMAGE:?set ESHU_SCANNER_GOREACHABILITY_SEED_IMAGE after the scanner-worker go-reachability release}"
	goReachabilityFragmentTODO       = "TODO(eshu-hq/eshu-deploy#40)"
)

// TestRemoteE2EComposeGoReachabilityFragmentRunsExtensionHost proves the
// skeleton shape the #10031 driver will include: the worker keeps the
// component-extension command and trust identity, and both one-shots must
// exit 0 before it starts.
func TestRemoteE2EComposeGoReachabilityFragmentRunsExtensionHost(t *testing.T) {
	t.Parallel()

	doc := readComposeDocument(t, goReachabilityFragment)
	worker := requireComposeService(t, doc, goReachabilityWorkerService)
	if got, want := fmt.Sprint(worker.Command), "[/usr/local/bin/eshu-collector-component-extension]"; got != want {
		t.Fatalf("worker command = %s, want %s", got, want)
	}
	assertComposeEnv(t, worker, "ESHU_COMPONENT_HOME", "/components")
	assertComposeEnv(t, worker, "ESHU_COMPONENT_TRUST_MODE", "allowlist")
	assertComposeEnv(t, worker, "ESHU_COMPONENT_ALLOW_IDS", "eshu.scanner.go-reachability")
	assertComposeEnv(t, worker, "ESHU_COMPONENT_ALLOW_PUBLISHERS", "eshu-hq")
	assertComposeEnv(t, worker, "ESHU_COMPONENT_COLLECTOR_INSTANCE_ID", goReachabilityFragmentInstanceID)
	assertComposeNamedVolume(t, worker, "go_reachability_component_home", "/components")
	for _, dependency := range []string{goReachabilitySeedService, goReachabilityReconcileService} {
		assertComposeDependency(t, worker, dependency)
	}
	assertComposeDependencyCondition(t, worker, goReachabilitySeedService, "service_completed_successfully")
	assertComposeDependencyCondition(t, worker, goReachabilityReconcileService, "service_completed_successfully")

	seed := requireComposeService(t, doc, goReachabilitySeedService)
	if seed.Image != goReachabilitySeedImageEnv {
		t.Fatalf("seed image = %q, want the required release-gated env %q", seed.Image, goReachabilitySeedImageEnv)
	}
	if strings.Contains(seed.Image, "@sha256:") {
		t.Fatalf("seed image pins a digest before the release exists: %q", seed.Image)
	}
	// The image entrypoint already is the component binary, so argv must
	// start at the seed subcommand: repeating the binary path shifts it
	// to argv[1] and the binary falls through to Runner mode on empty
	// stdin instead of seeding (proven by the live bring-up).
	if got := fmt.Sprint(seed.Command); !strings.HasPrefix(got, "[seed ") {
		t.Fatalf("seed command = %s, want argv starting at the seed subcommand", got)
	}
	assertComposeScriptContains(t, seed, "--image-ref")
	assertComposeNamedVolume(t, seed, "go_reachability_component_home", "/components")
	if seed.User != "0:0" {
		t.Fatalf("seed user = %q, want 0:0: a fresh named volume is root-owned", seed.User)
	}

	reconcile := requireComposeService(t, doc, goReachabilityReconcileService)
	assertComposeEnv(t, reconcile, "COMPONENT_ID", "eshu.scanner.go-reachability")
	assertComposeEnv(t, reconcile, "COMPONENT_VERSION", "0.1.0")
	assertComposeEnv(t, reconcile, "INSTANCE_ID", goReachabilityFragmentInstanceID)
	assertComposeEnv(t, reconcile, "ALLOW_IDS", "eshu.scanner.go-reachability")
	assertComposeEnv(t, reconcile, "ALLOW_PUBLISHERS", "eshu-hq")
	assertComposeDependency(t, reconcile, goReachabilitySeedService)
	assertComposeDependencyCondition(t, reconcile, goReachabilitySeedService, "service_completed_successfully")
	for _, grant := range []string{
		"--kind 'vulnerability.go_module_evidence' --schema-version '1.0.0' --scope 'scanner_worker'",
		"--kind 'vulnerability.go_call_reachability' --schema-version '1.0.0' --scope 'scanner_worker'",
		"--kind 'vulnerability.warning' --schema-version '1.0.0' --scope 'scanner_worker'",
		"--kind 'scanner_worker.warning' --schema-version '1.0.0' --scope 'scanner_worker'",
	} {
		assertComposeScriptContains(t, reconcile, grant)
	}
	assertComposeScriptContains(t, reconcile, "allowed_clone_hosts")
}

// TestRemoteE2EComposeGoReachabilityFragmentStaysSkeleton proves the fragment
// names its release gate and ships no invented digest anywhere.
func TestRemoteE2EComposeGoReachabilityFragmentStaysSkeleton(t *testing.T) {
	t.Parallel()

	raw := readRepositoryFile(t, deployRepositoryRoot(t), "compose/"+goReachabilityFragment)
	if !strings.Contains(raw, goReachabilityFragmentTODO) {
		t.Fatalf("go-reachability fragment misses the skeleton marker %q", goReachabilityFragmentTODO)
	}
	if !strings.Contains(raw, "scanner-worker") {
		t.Fatalf("go-reachability fragment does not name the release it waits on")
	}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(line, "scanner-go-reachability@sha256:") {
			t.Fatalf("fragment pins a go-reachability digest before the release exists: %q", line)
		}
	}
}

// TestRemoteE2EComposeGoReachabilityFragmentBuildContextsRooted proves every
// build context in the fragment resolves from ESHU_REPO_ROOT: a bare
// `context: .` renders to the compose/ directory, which has no Dockerfile,
// so `up --build` fails with "failed to read dockerfile" (proven live on
// the Wave-3 bring-up). compose config --quiet renders without checking
// Dockerfile existence, so only this assertion guards the regression.
func TestRemoteE2EComposeGoReachabilityFragmentBuildContextsRooted(t *testing.T) {
	t.Parallel()

	for _, file := range []string{
		"compose/" + goReachabilityFragment,
		"compose/docker-compose.remote-e2e.runtime.yaml",
	} {
		raw := readRepositoryFile(t, deployRepositoryRoot(t), file)
		for _, line := range strings.Split(raw, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if trimmed == "context: ." {
				t.Fatalf("%s: bare build context resolves to compose/ with no Dockerfile: %q", file, line)
			}
		}
	}
}

// TestRemoteE2EComposeGoReachabilityFragmentDeclaresProcessCommand proves the
// reconcile activation carries the extension host's required process.command:
// without it the worker crash-loops at startup and no claim can run.
func TestRemoteE2EComposeGoReachabilityFragmentDeclaresProcessCommand(t *testing.T) {
	t.Parallel()

	raw := readRepositoryFile(t, deployRepositoryRoot(t), "compose/"+goReachabilityFragment)
	if !strings.Contains(raw, "process:") {
		t.Fatal("reconcile activation has no process: block for the extension host")
	}
	if !strings.Contains(raw, "command: /components/runtime/remote-e2e-go-reachability/collector") {
		t.Fatal("reconcile activation does not run the seeded runtime collector")
	}
}
