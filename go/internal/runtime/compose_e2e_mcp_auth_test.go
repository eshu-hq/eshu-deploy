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
