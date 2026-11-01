// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"strings"
	"testing"
)

// TestComposeE2EMcpServerSetsAuthorizationServers pins that the E2E
// mcp-server service configures ESHU_AUTH_AUTHORIZATION_SERVERS (issue
// #9731, stack gap). Without it satellite discovery is 404 forever because
// DiscoveryConfigFromEnv needs both ESHU_AUTH_RESOURCE_URI and at least one
// issuer. The value defaults to the in-network mock IdP issuer so the E2E
// ShapeB discovery doc lists exactly that issuer.
func TestComposeE2EMcpServerSetsAuthorizationServers(t *testing.T) {
	t.Parallel()

	doc := readComposeDocument(t, "docker-compose.e2e.yaml")
	service := requireComposeService(t, doc, "mcp-server")
	raw, ok := service.Environment["ESHU_AUTH_AUTHORIZATION_SERVERS"]
	if !ok {
		t.Fatal("compose env ESHU_AUTH_AUTHORIZATION_SERVERS missing on mcp-server (issue #9731)")
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		t.Fatalf("compose env ESHU_AUTH_AUTHORIZATION_SERVERS = %#v, want non-empty issuer list", raw)
	}
	if !strings.Contains(value, "mock-oidc-idp:8080") {
		t.Fatalf("compose env ESHU_AUTH_AUTHORIZATION_SERVERS = %q, want it to name the mock IdP issuer (mock-oidc-idp:8080)", value)
	}
	// Pin the ESHU_E2E_MOCK_OIDC_ISSUER_URL override indirection, not just
	// its default: a hardcoded literal would still name the mock issuer
	// while silently breaking the stay-in-sync contract with mock-oidc-idp's
	// MOCK_OIDC_ISSUER_URL above.
	if !strings.Contains(value, "ESHU_E2E_MOCK_OIDC_ISSUER_URL") {
		t.Fatalf("compose env ESHU_AUTH_AUTHORIZATION_SERVERS = %q, want the ESHU_E2E_MOCK_OIDC_ISSUER_URL override indirection, not a hardcoded literal", value)
	}
}

// TestComposeE2ECoreSharesMcpResourceURI pins that the E2E core API service
// (eshu) configures the same ESHU_AUTH_RESOURCE_URI as the satellite
// (mcp-server) (issue #9731). The satellite verifies a dotted JWT at its own
// edge and then forwards the caller's Authorization header to core, where the
// OIDC bearer resolver is the only thing that can admit it. That resolver
// exists only when ESHU_AUTH_RESOURCE_URI is set; without it core answers 401
// "authentication is required" to every forwarded JWT on a core-served
// tools/call, even though the satellite's own checks (tools/list, denial
// shapes) all pass. The value is also the expected token audience, so the two
// services must agree exactly.
func TestComposeE2ECoreSharesMcpResourceURI(t *testing.T) {
	t.Parallel()

	doc := readComposeDocument(t, "docker-compose.e2e.yaml")
	mcp := requireComposeService(t, doc, "mcp-server")
	core := requireComposeService(t, doc, "eshu")

	mcpRaw, ok := mcp.Environment["ESHU_AUTH_RESOURCE_URI"]
	if !ok {
		t.Fatal("compose env ESHU_AUTH_RESOURCE_URI missing on mcp-server")
	}
	coreRaw, ok := core.Environment["ESHU_AUTH_RESOURCE_URI"]
	if !ok {
		t.Fatal("compose env ESHU_AUTH_RESOURCE_URI missing on eshu: core has no OIDC bearer resolver and rejects every JWT the satellite forwards (issue #9731)")
	}
	if coreRaw != mcpRaw {
		t.Fatalf("eshu ESHU_AUTH_RESOURCE_URI = %#v, want it identical to mcp-server's %#v (the token audience must match on both sides)", coreRaw, mcpRaw)
	}
}
