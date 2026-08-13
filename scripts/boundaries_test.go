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
	"testing"
)

// plant writes content to name under a fresh temp directory and returns
// the directory, so each case scans an isolated tree.
func plant(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestScanCatchesBlocklistTerm(t *testing.T) {
	// A synthetic term, not a real one: see the package comment on why the
	// production blocklist is never populated with real names in this file.
	dir := plant(t, "notes.md", "the customer is Example-Confidential-Co and they pay a lot\n")

	got, err := scan(dir, []string{"Example-Confidential-Co"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d violations, want 1: %+v", len(got), got)
	}
	if got[0].line != 1 {
		t.Errorf("line = %d, want 1", got[0].line)
	}
	if got[0].rule != "blocklist: Example-Confidential-Co" {
		t.Errorf("rule = %q, want it to name the term", got[0].rule)
	}
}

func TestScanBlocklistIsCaseInsensitive(t *testing.T) {
	dir := plant(t, "notes.md", "EXAMPLE-CONFIDENTIAL-CO uses this\n")
	got, err := scan(dir, []string{"example-confidential-co"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d violations, want 1 (case-insensitive match): %+v", len(got), got)
	}
}

func TestScanCatchesPrivatePath(t *testing.T) {
	dir := plant(t, "example.go", "// see /Users/someone/notes.txt for details\n")
	got, err := scan(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].rule != "private filesystem path" {
		t.Fatalf("got %+v, want one private-filesystem-path violation", got)
	}
}

func TestScanCatchesPrice(t *testing.T) {
	dir := plant(t, "notes.md", "the retainer is $2,500 per month\n")
	got, err := scan(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].rule != "price" {
		t.Fatalf("got %+v, want one price violation", got)
	}
}

func TestScanCleanTreePasses(t *testing.T) {
	dir := plant(t, "notes.md", "this repository has nothing to hide\n")
	got, err := scan(dir, []string{"Example-Confidential-Co"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %d violations on a clean tree, want 0: %+v", len(got), got)
	}
}

func TestScanSkipsGitAndBuildDirs(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{".git", "bin", "dist"} {
		path := filepath.Join(dir, sub, "leftover.txt")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("Example-Confidential-Co\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := scan(dir, []string{"Example-Confidential-Co"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %d violations inside .git/bin/dist, want 0 (those dirs are skipped): %+v", len(got), got)
	}
}

// TestRealRepoIsClean runs the actual production blocklist (empty today,
// see the package comment) plus the heuristics against this repository's
// real content: the same check `make boundaries` runs, folded into `go
// test` too so it fails fast in the same place as everything else.
func TestRealRepoIsClean(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	got, err := scan(repoRoot, blocklist)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("repo has %d boundaries violation(s):", len(got))
		for _, v := range got {
			t.Errorf("  %s:%d: %s: %s", v.file, v.line, v.rule, v.text)
		}
	}
}
