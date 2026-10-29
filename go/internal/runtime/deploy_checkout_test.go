// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package runtime

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// deployRepositoryRoot reports the eshu-deploy repository root, three
// directories above this test file (go/internal/runtime). Deploy-asset
// tests address chart, Compose, and docs files relative to it.
func deployRepositoryRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test filename")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "helm", "eshu", "Chart.yaml")); err != nil {
		t.Fatalf("deploy repository root %s has no helm/eshu/Chart.yaml: %v", root, err)
	}
	return root
}

// deployChartPath joins a chart-relative path under the deploy repo root.
func deployChartPath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(deployRepositoryRoot(t), "helm", "eshu", name)
}

// deployComposePath joins a compose-relative path under the deploy repo root.
func deployComposePath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(deployRepositoryRoot(t), "compose", name)
}

// deployDocPath joins a docs-relative path under the deploy repo root.
func deployDocPath(t *testing.T, name string) string {
	t.Helper()

	return filepath.Join(deployRepositoryRoot(t), "docs", name)
}
