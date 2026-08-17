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

package lint_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/tadanahq/kbf/tools/internal/lint"
)

// fakeFallback is a synthetic PlaybookSource (testing/fstest.MapFS, not
// the real internal/embedded package): precedence is a property of the
// resolution logic itself, not of any particular playbook's content, so
// a minimal, hermetic fixture that never has to change when the real
// embedded playbooks do is the more maintainable test.
func fakeFallback(playbooks map[string]fstest.MapFS) lint.PlaybookSource {
	return func(name string) (fs.FS, bool) {
		fsys, ok := playbooks[name]
		return fsys, ok
	}
}

func manifestYAML(name string, buildsOn []string, layer string) string {
	list := ""
	for i, n := range buildsOn {
		if i > 0 {
			list += ", "
		}
		list += n
	}
	return "name: " + name + "\nversion: 0.1.0\nspec: v0\nbuilds-on: [" + list + "]\nlayer: " + layer + "\n"
}

// writeLocalPackage materializes a minimal, valid, standalone package
// (manifest only, no ontology content, layer: core with the given
// builds-on) under dir, real files on disk: Load (the local side of
// resolution) is hardcoded to os.DirFS, so the local half of a
// precedence test needs a real directory, unlike the embedded half.
func writeLocalPackage(t *testing.T, dir, name string, buildsOn []string, manifestOverride string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "install"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := manifestOverride
	if content == "" {
		content = manifestYAML(name, buildsOn, "core")
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestLoadWithEmbeddedFallsBackOnlyWhenLocalIsMissing is the base case:
// a local vertical builds on a name no local path provides at all, and
// fallback is what resolves it.
func TestLoadWithEmbeddedFallsBackOnlyWhenLocalIsMissing(t *testing.T) {
	root := t.TempDir()
	vertical := filepath.Join(root, "my-biz")
	writeLocalPackage(t, vertical, "my-biz", []string{"core-thing"}, manifestYAML("my-biz", []string{"core-thing"}, "vertical"))

	fallback := fakeFallback(map[string]fstest.MapFS{
		"core-thing": {
			"manifest.yaml": &fstest.MapFile{Data: []byte(manifestYAML("core-thing", nil, "core"))},
		},
	})

	u, findings, err := lint.LoadWithEmbedded([]string{vertical}, fallback)
	if err != nil {
		t.Fatalf("LoadWithEmbedded: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("unexpected load findings: %+v", findings)
	}
	if _, ok := u.Packages["core-thing"]; !ok {
		t.Fatal("core-thing was not resolved from fallback")
	}
	if got := u.EmbeddedNames; len(got) != 1 || got[0] != "core-thing" {
		t.Errorf("EmbeddedNames = %v, want [core-thing]", got)
	}
}

// TestLoadWithEmbeddedLocalOverridesFallback is the precedence guarantee
// itself: when a local path AND fallback both offer the same name, the
// local one is what resolves, never silently shadowed by fallback. Proven
// by giving the two conflicting, individually-identifiable content (a
// deliberately invalid field only the local copy has) and checking the
// resulting finding names the local file, not a fallback-sourced one, and
// that the name never appears in EmbeddedNames.
func TestLoadWithEmbeddedLocalOverridesFallback(t *testing.T) {
	root := t.TempDir()
	vertical := filepath.Join(root, "my-biz")
	writeLocalPackage(t, vertical, "my-biz", []string{"core-thing"}, manifestYAML("my-biz", []string{"core-thing"}, "vertical"))

	localOverride := filepath.Join(root, "local-core-thing")
	// Deliberately missing `spec:` so a resolved-locally finding is
	// unambiguous: the fallback copy below is valid and would produce no
	// finding at all if it were used instead.
	brokenManifest := "name: core-thing\nversion: 0.1.0\nbuilds-on: []\nlayer: core\n"
	writeLocalPackage(t, localOverride, "core-thing", nil, brokenManifest)

	fallback := fakeFallback(map[string]fstest.MapFS{
		"core-thing": {
			"manifest.yaml": &fstest.MapFile{Data: []byte(manifestYAML("core-thing", nil, "core"))},
		},
	})

	u, _, err := lint.LoadWithEmbedded([]string{vertical, localOverride}, fallback)
	if err != nil {
		t.Fatalf("LoadWithEmbedded: %v", err)
	}
	if len(u.EmbeddedNames) != 0 {
		t.Errorf("EmbeddedNames = %v, want none: core-thing was passed locally", u.EmbeddedNames)
	}

	result, err := lint.RunWithEmbedded([]string{vertical, localOverride}, fallback)
	if err != nil {
		t.Fatalf("RunWithEmbedded: %v", err)
	}
	var sawMissingSpec bool
	for _, f := range result.Findings {
		if f.Rule == lint.KBF011 && f.File == filepath.Join(localOverride, "manifest.yaml") {
			sawMissingSpec = true
		}
	}
	if !sawMissingSpec {
		t.Errorf("expected KBF011 against %s (the local override), got: %+v", filepath.Join(localOverride, "manifest.yaml"), result.Findings)
	}
}

// TestLoadWithoutFallbackIsUnchanged: passing a nil fallback (what plain
// Load/Run always do) behaves exactly as if fallback did not exist, the
// guarantee conformance and every other internal/lint test rely on.
func TestLoadWithoutFallbackIsUnchanged(t *testing.T) {
	root := t.TempDir()
	vertical := filepath.Join(root, "my-biz")
	writeLocalPackage(t, vertical, "my-biz", []string{"core-thing"}, manifestYAML("my-biz", []string{"core-thing"}, "vertical"))

	u, _, err := lint.LoadWithEmbedded([]string{vertical}, nil)
	if err != nil {
		t.Fatalf("LoadWithEmbedded with nil fallback: %v", err)
	}
	if len(u.EmbeddedNames) != 0 {
		t.Errorf("EmbeddedNames = %v, want none with a nil fallback", u.EmbeddedNames)
	}
	if _, ok := u.Packages["core-thing"]; ok {
		t.Error("core-thing resolved even though fallback was nil")
	}
}
