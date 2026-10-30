// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"strings"
	"testing"
)

// TestHelmExtractedApiDisabledByDefault is the eshu-hq/eshu#9633 review P2
// regression: the extracted-api Deployment/Service/probe surface must not
// render (and must not move any workload) until the cutover window flips
// extractedApi.enabled on.
func TestHelmExtractedApiDisabledByDefault(t *testing.T) {
	t.Parallel()

	manifests := renderHelmChart(t)
	if helmManifestExists(manifests, "Deployment", "eshu-extracted-api") {
		t.Fatal("default render includes extracted-api Deployment, want none")
	}
	if helmManifestExists(manifests, "Service", "eshu-extracted-api") {
		t.Fatal("default render includes extracted-api Service, want none")
	}
}

func requireExtractedAPIEnabled(t *testing.T, args ...string) []helmManifest {
	t.Helper()

	full := append([]string{"--set", "extractedApi.enabled=true"}, args...)
	return renderHelmChart(t, full...)
}

// TestHelmExtractedApiEnabledRendersMatchedWorkloads is the eshu-hq/eshu#9633
// review P2 regression: the enabled path must render a Deployment and a
// Service whose selector matches the Deployment pod labels.
func TestHelmExtractedApiEnabledRendersMatchedWorkloads(t *testing.T) {
	t.Parallel()

	manifests := requireExtractedAPIEnabled(t)

	deployment := requireHelmManifest(t, manifests, "Deployment", "eshu-extracted-api")
	service := requireHelmManifest(t, manifests, "Service", "eshu-extracted-api")

	podLabels := helmMap(helmMap(helmMap(deployment["spec"])["template"])["metadata"])["labels"]
	matchLabels := helmMap(helmMap(deployment["spec"])["selector"])["matchLabels"]
	svcSelector := helmMap(service["spec"])["selector"]

	for key, want := range helmMapToStrings(matchLabels) {
		if got := helmMapToStrings(podLabels)[key]; got != want {
			t.Fatalf("pod label %q = %q, want matchLabels value %q", key, got, want)
		}
		if got := helmMapToStrings(svcSelector)[key]; got != want {
			t.Fatalf("service selector %q = %q, want matchLabels value %q", key, got, want)
		}
	}
}

// TestHelmExtractedApiImage is the eshu-hq/eshu#9633 review P2 regression:
// the eshu.extractedApiImage helper must render repository:tag, append
// @digest only when set, and fall back to the chart default pin.
func TestHelmExtractedApiImage(t *testing.T) {
	t.Parallel()

	const (
		repository = "ghcr.io/eshu-hq/eshu-api"
		tag        = "0.2.0"
	)
	digest := "sha256:" + strings.Repeat("a", 64)

	deployment := requireHelmManifest(t, requireExtractedAPIEnabled(t), "Deployment", "eshu-extracted-api")
	container := requireHelmContainer(t, deployment, "extracted-api")
	if got := container["image"]; got != repository+":"+tag {
		t.Fatalf("default extracted-api image = %#v, want %#v", got, repository+":"+tag)
	}

	withDigest := requireHelmManifest(t,
		requireExtractedAPIEnabled(t, "--set", "extractedApi.image.digest="+digest),
		"Deployment", "eshu-extracted-api")
	pinned := requireHelmContainer(t, withDigest, "extracted-api")
	if got := pinned["image"]; got != repository+":"+tag+"@"+digest {
		t.Fatalf("pinned extracted-api image = %#v, want digest appended", got)
	}
}

// TestHelmExtractedApiArgs is the eshu-hq/eshu#9633 review P2 regression:
// the container runs /eshu-api with --addr, and --core-base-url appears
// only when coreBaseURL is set (empty fails closed at startup).
func TestHelmExtractedApiArgs(t *testing.T) {
	t.Parallel()

	deployment := requireHelmManifest(t, requireExtractedAPIEnabled(t), "Deployment", "eshu-extracted-api")
	container := requireHelmContainer(t, deployment, "extracted-api")

	if command := helmStringSlice(container["command"]); !stringSlicesEqual(command, []string{"/eshu-api"}) {
		t.Fatalf("extracted-api command = %#v, want [/eshu-api]", command)
	}
	args := helmStringSlice(container["args"])
	if !stringSlicesEqual(args, []string{"--addr", ":8080"}) {
		t.Fatalf("extracted-api args without coreBaseURL = %#v, want [--addr :8080]", args)
	}

	withCore := requireHelmManifest(t,
		requireExtractedAPIEnabled(t, "--set", "extractedApi.coreBaseURL=https://core:8443"),
		"Deployment", "eshu-extracted-api")
	coreArgs := helmStringSlice(requireHelmContainer(t, withCore, "extracted-api")["args"])
	want := []string{"--addr", ":8080", "--core-base-url", "https://core:8443"}
	if !stringSlicesEqual(coreArgs, want) {
		t.Fatalf("extracted-api args with coreBaseURL = %#v, want %#v", coreArgs, want)
	}
}

// TestHelmExtractedApiProbes is the eshu-hq/eshu#9633 review P2 regression:
// liveness hits self-answered /health and readiness gates on core-reaching
// /readyz, matching the chart README.
func TestHelmExtractedApiProbes(t *testing.T) {
	t.Parallel()

	deployment := requireHelmManifest(t, requireExtractedAPIEnabled(t), "Deployment", "eshu-extracted-api")
	container := requireHelmContainer(t, deployment, "extracted-api")

	liveness := helmMap(helmMap(container["livenessProbe"])["httpGet"])
	if path := liveness["path"]; path != "/health" {
		t.Fatalf("liveness path = %#v, want /health", path)
	}
	readiness := helmMap(helmMap(container["readinessProbe"])["httpGet"])
	if path := readiness["path"]; path != "/readyz" {
		t.Fatalf("readiness path = %#v, want /readyz", path)
	}
}

// TestHelmExtractedApiSchemaRejects is the eshu-hq/eshu#9633 review P0/P2
// regression: the schema must reject malformed image references (bad digest,
// empty tag) instead of rendering broken images.
func TestHelmExtractedApiSchemaRejects(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"--set", "extractedApi.enabled=true", "--set", "extractedApi.image.digest=not-a-digest"},
		{"--set", "extractedApi.enabled=true", "--set-string", "extractedApi.image.tag="},
		{"--set", "extractedApi.enabled=true", "--set-string", "extractedApi.image.repository="},
	} {
		output := renderHelmChartFailure(t, args...)
		if !strings.Contains(output, "schema") {
			t.Fatalf("rejection for %v mentioned no schema violation:\n%s", args, output)
		}
	}
}

// TestHelmMCPServerImageFallbackAndPin guards the staged mcpServer.image
// knob: empty renders the shared monolith image, a set pin renders verbatim.
// The pin stays staged-only until eshu-hq/eshu#9521 (see values.yaml).
func TestHelmMCPServerImageFallbackAndPin(t *testing.T) {
	t.Parallel()

	deployment := requireHelmManifest(t, renderHelmChart(t), "Deployment", "eshu-mcp-server")
	container := requireHelmContainer(t, deployment, "mcp-server")
	fallback, ok := container["image"].(string)
	if !ok || !strings.HasPrefix(fallback, "ghcr.io/eshu-hq/eshu:") {
		t.Fatalf("default mcp-server image = %#v, want shared monolith image", container["image"])
	}

	const pin = "ghcr.io/eshu-hq/eshu-mcp:0.2.0@sha256:93390eefc87f67761d160f907c9a94092438f730cb59908572f9e72286e82346"
	pinned := requireHelmManifest(t,
		renderHelmChart(t, "--set-string", "mcpServer.image="+pin),
		"Deployment", "eshu-mcp-server")
	if got := requireHelmContainer(t, pinned, "mcp-server")["image"]; got != pin {
		t.Fatalf("pinned mcp-server image = %#v, want %#v", got, pin)
	}
}

func helmMapToStrings(raw any) map[string]string {
	out := map[string]string{}
	for key, value := range helmMap(raw) {
		text, ok := value.(string)
		if !ok {
			continue
		}
		out[key] = text
	}
	return out
}
