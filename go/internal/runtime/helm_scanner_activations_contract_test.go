// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Scanner seed digests pinned below mirror the eshu-side admission pins
// (artifactDigests in go/internal/collector/scannerworker/component on
// eshu-hq/eshu origin/main, Refs eshu-hq/eshu#9501). A scanner republish
// moves those pins and these entries together; a test that pins one side
// against a stale copy of the other fails here instead of pulling stale
// analyzers in QA.
const (
	scannerImageDigest     = "sha256:f92fabfc560a7703cb92ac227b90ce42dc45522f4d25533d342c5a140e7bd372"
	scannerSBOMDigest      = "sha256:cf0e10d62bd4ce0c6bfdf0de1cf3260565b67ed5189fe9544d9284906c9ecd90"
	scannerOSRuntimeDigest = "sha256:3dbca86bcb8f068a3fb50fcb6ad6a5806f32b9dc3e567dfb9fd7e74f60821e31"
)

var scannerSeedImagePattern = regexp.MustCompile(`^ghcr\.io/eshu-hq/scanner-[a-z]+@sha256:[0-9a-f]{64}$`)

type scannerActivationCase struct {
	instance    string
	componentID string
	binary      string
	image       string
	// grants lists one reconcile-script fragment per RequiredGrants entry
	// the QA activation must carry. The grant script is the admission truth
	// at runtime, so the fragments pin the kinds, schema versions, scope,
	// and lifetimes the entries claim.
	grants []string
}

// renderScannerActivationsQA renders the chart with the committed QA
// activation entries (helm/eshu/ci/scanner-activations.values.yaml).
func renderScannerActivationsQA(t *testing.T) []helmManifest {
	t.Helper()

	valuesPath := deployChartPath(t, filepath.Join("ci", "scanner-activations.values.yaml"))
	if _, err := os.Stat(valuesPath); err != nil {
		t.Fatalf("scanner QA values missing: %v", err)
	}
	return renderHelmChart(t, "-f", valuesPath)
}

// renderHelmRaw returns the raw `helm template` bytes so renders can be
// compared byte for byte. It carries the same contract-safe password prefix
// as the manifest renders.
func renderHelmRaw(t *testing.T, args ...string) []byte {
	t.Helper()

	chartPath := filepath.Join(deployRepositoryRoot(t), "helm", "eshu")
	helmPath, err := exec.LookPath("helm")
	if err != nil {
		t.Skipf("helm binary not found in PATH: %v", err)
	}
	cmdArgs := append([]string{"template", "eshu", chartPath}, helmRenderArgs(args...)...)
	cmd := exec.Command(helmPath, cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("helm template: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("helm template: %v", err)
	}
	return output
}

// TestHelmScannerActivationsRenderSeedDeployments proves the #9501
// acceptance row: the three scanner activation entries render one worker
// Deployment each, and every seed init container runs the digest-only seed
// image with the in-image seed binary the scanner Dockerfiles ship.
func TestHelmScannerActivationsRenderSeedDeployments(t *testing.T) {
	t.Parallel()

	manifests := renderScannerActivationsQA(t)
	if helmManifestExists(manifests, "Deployment", "eshu-component-extension-collector") {
		t.Fatal("activations mode rendered the single legacy worker Deployment")
	}
	cases := []scannerActivationCase{
		{
			instance: "scanner-image", componentID: "eshu.scanner.image", binary: "eshu-scanner-component-image", image: "ghcr.io/eshu-hq/scanner-image@" + scannerImageDigest,
			grants: []string{
				"--kind 'scanner_worker.analysis' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'scanner_worker.warning' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'vulnerability.os_package' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'vulnerability.warning' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
			},
		},
		{
			instance: "scanner-sbom", componentID: "eshu.scanner.sbom", binary: "eshu-scanner-component-sbom", image: "ghcr.io/eshu-hq/scanner-sbom@" + scannerSBOMDigest,
			grants: []string{
				"--kind 'sbom.document' --schema-version '1.1.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'sbom.component' --schema-version '1.1.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'sbom.warning' --schema-version '1.1.0' --scope 'scanner_worker' --expires-in '720h'",
			},
		},
		{
			instance: "scanner-osruntime", componentID: "eshu.scanner.osruntime", binary: "eshu-scanner-component-osruntime", image: "ghcr.io/eshu-hq/scanner-osruntime@" + scannerOSRuntimeDigest,
			grants: []string{
				"--kind 'vulnerability.os_package' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
				"--kind 'vulnerability.warning' --schema-version '1.0.0' --scope 'scanner_worker' --expires-in '720h'",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.instance, func(t *testing.T) {
			t.Parallel()

			deployment := requireHelmManifest(t, manifests, "Deployment", "eshu-cx-"+tc.instance)
			inits := helmMapSlice(helmPodSpec(t, deployment)["initContainers"])
			if len(inits) != 2 {
				t.Fatalf("%s initContainers = %d, want seed then reconcile", tc.instance, len(inits))
			}
			seed, reconcile := inits[0], inits[1]
			if seed["name"] != "seed-"+tc.instance || reconcile["name"] != "reconcile-"+tc.instance {
				t.Fatalf("init order = %v, %v, want seed before reconcile", seed["name"], reconcile["name"])
			}
			if !scannerSeedImagePattern.MatchString(helmString(seed["image"])) {
				t.Fatalf("seed image = %v, want a digest-only scanner image ref", seed["image"])
			}
			if seed["image"] != tc.image {
				t.Fatalf("seed image = %v, want pinned %s", seed["image"], tc.image)
			}
			if seed["imagePullPolicy"] != "IfNotPresent" {
				t.Fatalf("seed pullPolicy = %v, want IfNotPresent", seed["imagePullPolicy"])
			}
			wantCommand := []string{
				"/usr/local/bin/" + tc.binary, "seed",
				"--dest", "/data/.eshu/components/runtime/" + tc.instance,
				"--image-ref", tc.image,
			}
			if got := helmStringSlice(seed["command"]); !stringSlicesEqual(got, wantCommand) {
				t.Fatalf("seed command = %#v, want %#v", got, wantCommand)
			}
			env := helmEnvByName(reconcile)
			assertHelmLiteralEnv(t, env, "COMPONENT_ID", tc.componentID)
			assertHelmLiteralEnv(t, env, "COMPONENT_VERSION", "0.1.0")
			assertHelmLiteralEnv(t, env, "INSTANCE_ID", tc.instance)
			script := strings.Join(helmStringSlice(reconcile["command"]), "\n")
			for _, grant := range tc.grants {
				if !strings.Contains(script, grant) {
					t.Fatalf("%s reconcile script missing grant %q:\n%s", tc.instance, grant, script)
				}
			}
			if got := strings.Count(script, "eshu component grant "); got != len(tc.grants) {
				t.Fatalf("%s reconcile script grants %d kinds, want %d", tc.instance, got, len(tc.grants))
			}
		})
	}
}

// TestHelmScannerActivationsConfigOmitsProcessArgs proves the arg-less
// scanner binaries take no CLI arguments: the rendered activation files run
// the seeded collector with no args block, unlike --sdk-stdio collectors.
func TestHelmScannerActivationsConfigOmitsProcessArgs(t *testing.T) {
	t.Parallel()

	manifests := renderScannerActivationsQA(t)
	configMap := requireHelmManifest(t, manifests, "ConfigMap", "eshu-component-activations")
	data := helmMap(configMap["data"])
	for _, instance := range []string{"scanner-image", "scanner-sbom", "scanner-osruntime"} {
		body := helmString(data[instance+".yaml"])
		if body == "" {
			t.Fatalf("activation file %s.yaml missing from the activations ConfigMap", instance)
		}
		if strings.Contains(body, "args:") {
			t.Fatalf("activation file %s.yaml carries process args, want the bare seeded collector:\n%s", instance, body)
		}
		if !strings.Contains(body, "/data/.eshu/components/runtime/"+instance+"/collector") {
			t.Fatalf("activation file %s.yaml does not run the seeded collector:\n%s", instance, body)
		}
	}
}

// TestHelmScannerActivationsEmptyStaysByteIdentical proves the disabled
// default: an explicit empty activations list renders byte-identical output
// to the bare default, so shipping the entries behind `-f` cannot drift the
// default install.
func TestHelmScannerActivationsEmptyStaysByteIdentical(t *testing.T) {
	t.Parallel()

	emptyPath := filepath.Join(t.TempDir(), "empty-activations.yaml")
	if err := os.WriteFile(emptyPath, []byte("componentExtensionCollector:\n  activations: []\n"), 0o600); err != nil {
		t.Fatalf("write empty activations values: %v", err)
	}
	deflt := renderHelmRaw(t)
	empty := renderHelmRaw(t, "-f", emptyPath)
	if !bytes.Equal(deflt, empty) {
		first := 0
		for first < len(deflt) && first < len(empty) && deflt[first] == empty[first] {
			first++
		}
		t.Fatalf("empty-activations render differs from default at byte %d (lengths %d vs %d)", first, len(deflt), len(empty))
	}
}
