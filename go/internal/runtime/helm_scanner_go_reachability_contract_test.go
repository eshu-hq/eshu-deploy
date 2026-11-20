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

// The go-reachability scanner activation ships as a skeleton until
// eshu-scanner-worker publishes its first go-reachability release
// (eshu-hq/eshu-deploy#40): the seed image digest cannot be pinned yet, so the
// fourth entry in the scanner example values stays commented, and the
// component Deployment gets a dedicated least-privilege egress NetworkPolicy
// now — never an allow-all egress as done.

const (
	goReachabilityInstanceID   = "scanner-go-reachability"
	goReachabilityComponentID  = "eshu.scanner.go-reachability"
	goReachabilityPolicyName   = "eshu-cx-scanner-go-reachability-egress"
	goReachabilitySkeletonTODO = "TODO(eshu-hq/eshu-deploy#40)"
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

// TestScannerActivationsFourthEntryStaysSkeleton pins the digest-gated
// skeleton: the go-reachability entry is present and names its release, but
// stays commented until eshu-scanner-worker publishes the image — and no
// invented digest ships in its place.
func TestScannerActivationsFourthEntryStaysSkeleton(t *testing.T) {
	t.Parallel()

	raw := readRepositoryFile(t, deployRepositoryRoot(t), "helm/eshu/ci/scanner-activations.values.yaml")
	if !strings.Contains(raw, goReachabilitySkeletonTODO) {
		t.Fatalf("scanner example values miss the skeleton marker %q", goReachabilitySkeletonTODO)
	}
	for _, want := range []string{
		goReachabilityInstanceID,
		goReachabilityComponentID,
		"eshu-scanner-component-go-reachability",
		"scanner-go-reachability",
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("scanner example values miss skeleton reference %q", want)
		}
	}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(line, "scanner-go-reachability@sha256:") {
			t.Fatalf("scanner example values pin an uncommented go-reachability digest, want the TODO skeleton: %q", line)
		}
		if strings.Contains(line, "instanceId: "+goReachabilityInstanceID) {
			t.Fatalf("go-reachability activation is live before its release, want the commented skeleton: %q", line)
		}
	}
}
