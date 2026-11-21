// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The go-reachability scanner activation went live with the v0.1.0
// scanner-worker release (eshu-hq/eshu-deploy#40): the fourth entry in the
// scanner example values pins that release's digest, and the component
// Deployment gets a dedicated least-privilege egress NetworkPolicy — never
// an allow-all egress as done.

const (
	goReachabilityInstanceID  = "scanner-go-reachability"
	goReachabilityComponentID = "eshu.scanner.go-reachability"
	goReachabilityPolicyName  = "eshu-cx-scanner-go-reachability-egress"
	// goReachabilityHelmSeedDigest pins the v0.1.0 multi-arch release,
	// verified against the ghcr.io manifest. The digest-only seed image and
	// the in-image seed binary below must stay identical to the fourth
	// activation entry the render test pins.
	goReachabilityHelmSeedDigest = "sha256:8e4059a2acbd020a761862a8196e25567b7119bb97ef42a228432c59f5e39284"
	goReachabilityHelmSeedImage  = "ghcr.io/eshu-hq/scanner-go-reachability@sha256:8e4059a2acbd020a761862a8196e25567b7119bb97ef42a228432c59f5e39284"
)

// goReachabilityEgressValues renders the chart with the dedicated
// go-reachability egress policy enabled on top of the given CIDR lists. Empty
// CIDR lists stay fail-closed: no rule renders for that destination.
func goReachabilityEgressValues(t *testing.T, extra string) string {
	t.Helper()

	values := `
contentStore:
  dsn: postgresql://eshu:secret@postgres:5432/eshu
neo4j:
  auth:
    secretName: ""
scannerGoReachability:
  enabled: true
` + extra
	path := filepath.Join(t.TempDir(), "go-reachability-egress-values.yaml")
	if err := os.WriteFile(path, []byte(values), 0o600); err != nil {
		t.Fatalf("write go-reachability egress values: %v", err)
	}
	return path
}

// TestHelmScannerGoReachabilityEgressDefaultOff proves the skeleton default:
// no dedicated egress policy renders until the operator enables it.
func TestHelmScannerGoReachabilityEgressDefaultOff(t *testing.T) {
	t.Parallel()

	manifests := renderHelmChart(t)
	if helmManifestExists(manifests, "NetworkPolicy", goReachabilityPolicyName) {
		t.Fatalf("default chart render included %s", goReachabilityPolicyName)
	}
}

// TestHelmScannerGoReachabilityEgressHonorsNetworkPolicyToggle proves the
// global networkPolicy.enabled=false kill switch suppresses the dedicated
// policy even when the component enables it.
func TestHelmScannerGoReachabilityEgressHonorsNetworkPolicyToggle(t *testing.T) {
	t.Parallel()

	path := goReachabilityEgressValues(t, `networkPolicy:
  enabled: false
`)
	manifests := renderHelmChart(t, "-f", path)
	if helmManifestExists(manifests, "NetworkPolicy", goReachabilityPolicyName) {
		t.Fatal("go-reachability egress policy rendered with networkPolicy.enabled=false")
	}
}

// TestHelmScannerGoReachabilityEgressRendersLeastPrivilege proves the
// dedicated policy selects only the go-reachability activation pods, carries
// Egress alone, keeps DNS, and encodes the four scanner destinations (git
// host, module proxy, checksum DB, vuln DB) as TCP 443 rules — never the
// allow-all `- {}` stub.
func TestHelmScannerGoReachabilityEgressRendersLeastPrivilege(t *testing.T) {
	t.Parallel()

	path := goReachabilityEgressValues(t, `  egress:
    gitClone:
      to:
        - ipBlock:
            cidr: 140.82.112.0/20
    moduleProxy:
      to:
        - ipBlock:
            cidr: 142.250.0.0/16
    checksumDB:
      to:
        - ipBlock:
            cidr: 142.250.0.0/16
    vulnDB:
      to:
        - ipBlock:
            cidr: 142.250.0.0/16
`)
	manifests := renderHelmChart(t, "-f", path)
	policy := requireHelmManifest(t, manifests, "NetworkPolicy", goReachabilityPolicyName)

	spec := helmMap(policy["spec"])
	if len(spec) == 0 {
		t.Fatalf("policy %s has no spec map: %#v", goReachabilityPolicyName, policy)
	}
	selector := helmMap(helmMap(spec["podSelector"])["matchLabels"])
	if selector["app.kubernetes.io/component"] != "component-extension-collector" {
		t.Fatalf("policy selector component = %q, want component-extension-collector", selector["app.kubernetes.io/component"])
	}
	if selector["eshu.dev/component-instance"] != goReachabilityInstanceID {
		t.Fatalf("policy selector instance = %q, want %q", selector["eshu.dev/component-instance"], goReachabilityInstanceID)
	}
	types := helmStringSlice(spec["policyTypes"])
	if len(types) != 1 || types[0] != "Egress" {
		t.Fatalf("policy types = %#v, want [Egress] only", types)
	}
	egress := helmMapSlice(spec["egress"])
	if len(egress) == 0 {
		t.Fatalf("policy %s renders no egress rules", goReachabilityPolicyName)
	}
	for _, rule := range egress {
		if _, open := rule["to"]; !open {
			t.Fatalf("policy %s renders an allow-all egress rule: %#v", goReachabilityPolicyName, rule)
		}
	}
	body := helmManifestText(t, manifests, "NetworkPolicy", goReachabilityPolicyName)
	for _, want := range []string{"140.82.112.0/20", "142.250.0.0/16", "53", "443"} {
		if !strings.Contains(body, want) {
			t.Fatalf("policy %s omits egress destination %q", goReachabilityPolicyName, want)
		}
	}
}

// TestHelmScannerGoReachabilityEgressEmptyDestinationsFailClosed proves an
// enabled policy with no destination CIDRs keeps DNS only: every unnamed
// destination stays denied instead of opening up.
func TestHelmScannerGoReachabilityEgressEmptyDestinationsFailClosed(t *testing.T) {
	t.Parallel()

	path := goReachabilityEgressValues(t, "")
	manifests := renderHelmChart(t, "-f", path)
	policy := requireHelmManifest(t, manifests, "NetworkPolicy", goReachabilityPolicyName)

	spec := helmMap(policy["spec"])
	if len(spec) == 0 {
		t.Fatalf("policy %s has no spec map: %#v", goReachabilityPolicyName, policy)
	}
	egress := helmMapSlice(spec["egress"])
	if len(egress) == 0 {
		t.Fatalf("policy %s renders no egress rules, want DNS at least", goReachabilityPolicyName)
	}
	for _, rule := range egress {
		if _, open := rule["to"]; !open {
			t.Fatalf("policy %s renders an allow-all egress rule: %#v", goReachabilityPolicyName, rule)
		}
	}
	body := helmManifestText(t, manifests, "NetworkPolicy", goReachabilityPolicyName)
	if !strings.Contains(body, "53") {
		t.Fatalf("policy %s omits the DNS rule, want DNS kept while destinations fail closed", goReachabilityPolicyName)
	}
	if strings.Contains(body, "443") {
		t.Fatalf("policy %s renders a 443 destination rule with no CIDRs configured, want fail-closed deny", goReachabilityPolicyName)
	}
}

// TestHelmScannerGoReachabilityQAPolicyRendersWithActivation proves the
// policy ships together with the live activation: the committed QA values
// enable scannerGoReachability, so the dedicated egress policy renders
// (fail-closed with DNS only until the operator populates the destination
// CIDR lists at QA deploy time).
func TestHelmScannerGoReachabilityQAPolicyRendersWithActivation(t *testing.T) {
	t.Parallel()

	manifests := renderScannerActivationsQA(t)
	policy := requireHelmManifest(t, manifests, "NetworkPolicy", goReachabilityPolicyName)

	spec := helmMap(policy["spec"])
	types := helmStringSlice(spec["policyTypes"])
	if len(types) != 1 || types[0] != "Egress" {
		t.Fatalf("QA policy types = %#v, want [Egress] only", types)
	}
	body := helmManifestText(t, manifests, "NetworkPolicy", goReachabilityPolicyName)
	if !strings.Contains(body, "53") {
		t.Fatalf("QA policy omits the DNS rule:\n%s", body)
	}
	if strings.Contains(body, "443") {
		t.Fatalf("QA policy renders a 443 destination with no committed CIDRs, want fail-closed deny:\n%s", body)
	}
}

// helmManifestText re-encodes one rendered manifest so tests can assert on
// destination strings (CIDRs, ports) without walking every nested map.
func helmManifestText(t *testing.T, manifests []helmManifest, kind, name string) string {
	t.Helper()

	manifest := requireHelmManifest(t, manifests, kind, name)
	raw, err := yaml.Marshal(map[string]any(manifest))
	if err != nil {
		t.Fatalf("encode %s %s: %v", kind, name, err)
	}
	return string(raw)
}

// TestScannerActivationsFourthEntryIsLive pins the digest-gated live
// entry: the go-reachability activation is uncommented, pins the v0.1.0
// release digest (no placeholder, no second digest), names the in-image
// seed binary the render test asserts, and stays allowlisted.
func TestScannerActivationsFourthEntryIsLive(t *testing.T) {
	t.Parallel()

	raw := readRepositoryFile(t, deployRepositoryRoot(t), "helm/eshu/ci/scanner-activations.values.yaml")
	if strings.Contains(raw, "TODO(eshu-hq/eshu-deploy#40)") {
		t.Fatal("scanner example values still carry the go-reachability skeleton marker, want the live entry")
	}
	for _, want := range []string{
		"instanceId: " + goReachabilityInstanceID,
		"componentId: " + goReachabilityComponentID,
		"eshu-scanner-component-go-reachability",
		goReachabilityHelmSeedImage,
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("scanner example values miss live go-reachability reference %q", want)
		}
	}
	if got := strings.Count(raw, "scanner-go-reachability@sha256:"); got != 1 {
		t.Fatalf("scanner example values pin %d go-reachability digests, want exactly the v0.1.0 pin", got)
	}
	if !strings.Contains(raw, goReachabilityComponentID) || !strings.Contains(raw, "allowIds:") {
		t.Fatal("scanner example values miss the allowIds list carrying the go-reachability component")
	}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(line, "<TODO") {
			t.Fatalf("scanner example values ship an unreplaced digest placeholder: %q", line)
		}
	}
}
