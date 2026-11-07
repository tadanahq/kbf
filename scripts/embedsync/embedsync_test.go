// Copyright 2026 The kbf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildFakeRoot lays down the minimal subset of a real repo root
// sourceFiles actually reads: one file per embedded category (a
// playbooks/core-* manifest, a skills/kbf-authoring file, a spec/*.md and
// a spec/primitives/*.md doc, both carrying real frontmatter). corePlaybooks
// and skills are package-level, fixed name lists (not test-parameterized),
// so the fixture's directory names must match them exactly for
// sourceFiles to find anything at all.
func buildFakeRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range corePlaybooks {
		write(filepath.Join("playbooks", name, "manifest.yaml"), "name: "+name+"\nversion: 0.1.0\n")
	}
	for _, name := range skills {
		write(filepath.Join("skills", name, "SKILL.md"), "---\nname: "+name+"\n---\n\nbody\n")
	}
	write(filepath.Join("spec", "onboarding.md"), "---\ntype: spec-doc\n---\n\nbody\n")
	write(filepath.Join("spec", "primitives", "entity.md"), "---\ntype: spec-doc\n---\n\nbody\n")
	return root
}

func TestSyncThenCheckIsClean(t *testing.T) {
	root := buildFakeRoot(t)
	files, err := sourceFiles(root)
	if err != nil {
		t.Fatalf("sourceFiles: %v", err)
	}
	dataDir := filepath.Join(root, "tools", "internal", "embedded", "data")
	if err := writeFiles(dataDir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}

	stale, err := checkFiles(dataDir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected a clean check right after sync, got: %v", stale)
	}
}

// TestCheckCatchesStaleSource is embed-freshness's first direction: a
// source file changes and data/ is not re-synced, --check must fail on
// that file specifically.
func TestCheckCatchesStaleSource(t *testing.T) {
	root := buildFakeRoot(t)
	files, err := sourceFiles(root)
	if err != nil {
		t.Fatalf("sourceFiles: %v", err)
	}
	dataDir := filepath.Join(root, "tools", "internal", "embedded", "data")
	if err := writeFiles(dataDir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}

	onboarding := filepath.Join(root, "spec", "onboarding.md")
	if err := os.WriteFile(onboarding, []byte("---\ntype: spec-doc\n---\n\nchanged body, not re-synced\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	freshFiles, err := sourceFiles(root)
	if err != nil {
		t.Fatalf("sourceFiles (after edit): %v", err)
	}
	stale, err := checkFiles(dataDir, freshFiles)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if !anyContains(stale, "spec/onboarding.md") {
		t.Errorf("expected spec/onboarding.md flagged stale, got: %v", stale)
	}
}

// TestCheckCatchesOrphan is embed-freshness's other direction: a file
// sitting in data/ that no source file justifies anymore (a hand-edit, or
// a source file deleted without re-syncing) must also fail --check, not
// just missing/stale content.
func TestCheckCatchesOrphan(t *testing.T) {
	root := buildFakeRoot(t)
	files, err := sourceFiles(root)
	if err != nil {
		t.Fatalf("sourceFiles: %v", err)
	}
	dataDir := filepath.Join(root, "tools", "internal", "embedded", "data")
	if err := writeFiles(dataDir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}

	orphan := filepath.Join(dataDir, "spec", "deleted-doc.md")
	if err := os.WriteFile(orphan, []byte("---\ntype: spec-doc\n---\n\nno source justifies this anymore\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stale, err := checkFiles(dataDir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if !anyContains(stale, "spec/deleted-doc.md") {
		t.Errorf("expected spec/deleted-doc.md flagged orphaned, got: %v", stale)
	}
}

// TestCheckCatchesMissing: a file a fresh sync would produce, but data/
// doesn't have at all (never synced, or deleted by hand).
func TestCheckCatchesMissing(t *testing.T) {
	root := buildFakeRoot(t)
	files, err := sourceFiles(root)
	if err != nil {
		t.Fatalf("sourceFiles: %v", err)
	}
	dataDir := filepath.Join(root, "tools", "internal", "embedded", "data")
	if err := writeFiles(dataDir, files); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	if err := os.Remove(filepath.Join(dataDir, "spec", "onboarding.md")); err != nil {
		t.Fatal(err)
	}

	stale, err := checkFiles(dataDir, files)
	if err != nil {
		t.Fatalf("checkFiles: %v", err)
	}
	if !anyContains(stale, "spec/onboarding.md") {
		t.Errorf("expected spec/onboarding.md flagged missing, got: %v", stale)
	}
}

func anyContains(lines []string, substr string) bool {
	for _, l := range lines {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}
