// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// digestedImageRef requires a full 64-hex sha256 digest. A tag-only reference
// passes a regex but fails closed at deploy time when the registry serves a
// different image; the digest pins the artifact the gate actually exercised.
var digestedImageRef = regexp.MustCompile(`^[^:@\s]+:[^@\s]+@sha256:[0-9a-f]{64}$`)

// Shared compose and Helm assertion helpers for the deploy-asset contract
// tests. They are behavior-identical copies of the eshu-side helpers the
// extracted tests used before the #8871 in-tree removal; roots resolve
// against the deploy repository, so no ESHU_DEPLOY_CHECKOUT seam is needed
// here.

type composeDocument struct {
	Include  []string                  `yaml:"include"`
	Services map[string]composeService `yaml:"services"`
	Volumes  map[string]any            `yaml:"volumes"`
}

type composeService struct {
	Command     any            `yaml:"command"`
	Extends     any            `yaml:"extends"`
	Entrypoint  any            `yaml:"entrypoint"`
	Healthcheck map[string]any `yaml:"healthcheck"`
	Image       string         `yaml:"image"`
	Platform    string         `yaml:"platform"`
	Ports       []any          `yaml:"ports"`
	Profiles    []string       `yaml:"profiles"`
	Restart     string         `yaml:"restart"`
	DependsOn   any            `yaml:"depends_on"`
	Environment map[string]any `yaml:"environment"`
	Volumes     []any          `yaml:"volumes"`
}

func graphRuntimeServices() []string {
	return []string{
		"db-migrate",
		"bootstrap-index",
		"eshu",
		"mcp-server",
		"ingester",
		"resolution-engine",
		"projector",
	}
}

func readRepositoryFile(t *testing.T, root, relativePath string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(raw)
}

func requireComposeService(t *testing.T, doc composeDocument, name string) composeService {
	t.Helper()

	service, ok := doc.Services[name]
	if !ok {
		t.Fatalf("compose service %q missing", name)
	}
	return service
}

func assertComposeEnv(t *testing.T, service composeService, key, want string) {
	t.Helper()

	got, ok := service.Environment[key]
	if !ok {
		t.Fatalf("compose env %s missing", key)
	}
	if got != want {
		t.Fatalf("compose env %s = %#v, want %q", key, got, want)
	}
}

func assertComposeDependency(t *testing.T, service composeService, key string) {
	t.Helper()

	if !composeDependencyExists(service.DependsOn, key) {
		t.Fatalf("compose dependency %s missing", key)
	}
}

func composeDependencyExists(dependsOn any, key string) bool {
	switch dependencies := dependsOn.(type) {
	case map[string]any:
		_, ok := dependencies[key]
		return ok
	case []any:
		for _, dependency := range dependencies {
			if dependency == key {
				return true
			}
		}
	}
	return false
}

func assertComposeVolumeContains(t *testing.T, service composeService, want string) {
	t.Helper()

	for _, volume := range service.Volumes {
		if fmt.Sprint(volume) == want {
			return
		}
	}
	t.Fatalf("compose volume %q missing from %#v", want, service.Volumes)
}

func assertComposeScriptContains(t *testing.T, service composeService, want string) {
	t.Helper()

	body := fmt.Sprintf("%#v %#v", service.Entrypoint, service.Command)
	if !strings.Contains(body, want) {
		t.Fatalf("compose script missing %q in %s", want, body)
	}
}

func assertHelmReadOnlyVolumeMount(t *testing.T, container map[string]any, name, mountPath string) {
	t.Helper()

	for _, mount := range helmMapSlice(container["volumeMounts"]) {
		if helmString(mount["name"]) != name {
			continue
		}
		if got := helmString(mount["mountPath"]); got != mountPath {
			t.Fatalf("volume mount %s path = %q, want %q", name, got, mountPath)
		}
		if got, ok := mount["readOnly"].(bool); !ok || !got {
			t.Fatalf("volume mount %s readOnly = %#v, want true", name, mount["readOnly"])
		}
		return
	}
	t.Fatalf("volume mount %s missing", name)
}
