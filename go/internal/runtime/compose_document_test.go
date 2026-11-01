// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func readComposeDocument(t *testing.T, name string) composeDocument {
	t.Helper()

	return readComposeDocumentWithIncludes(t, name, map[string]bool{})
}

func readComposeDocumentWithIncludes(t *testing.T, name string, seen map[string]bool) composeDocument {
	t.Helper()

	cleanName := filepath.Clean(name)
	if seen[cleanName] {
		return composeDocument{}
	}
	seen[cleanName] = true

	raw, err := os.ReadFile(deployComposePath(t, cleanName))
	if err != nil {
		t.Fatalf("read %s: %v", cleanName, err)
	}

	var doc composeDocument
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", cleanName, err)
	}
	if doc.Services == nil {
		doc.Services = map[string]composeService{}
	}
	for _, included := range doc.Include {
		includedName := filepath.Join(filepath.Dir(cleanName), included)
		includedDoc := readComposeDocumentWithIncludes(t, includedName, seen)
		for serviceName, service := range includedDoc.Services {
			if _, exists := doc.Services[serviceName]; exists {
				t.Fatalf("compose service %q declared more than once through %s", serviceName, cleanName)
			}
			doc.Services[serviceName] = service
		}
	}
	return doc
}
