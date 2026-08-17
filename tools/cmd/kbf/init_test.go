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

	"github.com/tadanahq/kbf/tools/internal/embedded"
	"github.com/tadanahq/kbf/tools/internal/lint"
)

// resetInitFlags restores init's package-level flag vars to their
// documented default: see testhelpers_test.go's runCLI doc comment for
// why this is necessary between tests that share rootCmd.
func resetInitFlags() {
	initBuildsOn = nil
	initLayer = "vertical"
}

// TestInitScaffoldGolden checks the exact bytes kbf init writes: a golden
// test, not just a "did a file appear" check, since the header comment's
// wording and the manifest's field order are both things a later edit
// could silently drift without either failing.
func TestInitScaffoldGolden(t *testing.T) {
	resetInitFlags()
	t.Chdir(t.TempDir())

	if _, stderr, err := runCLI(t, "init", "demo-biz", "--builds-on", "core-operations"); err != nil {
		t.Fatalf("init: %v (stderr: %s)", err, stderr)
	}

	manifest, err := os.ReadFile(filepath.Join("demo-biz", "manifest.yaml"))
	if err != nil {
		t.Fatalf("reading manifest.yaml: %v", err)
	}
	checkGolden(t, filepath.Join(testdataDir, "golden", "init-manifest.yaml"), string(manifest))

	slots, err := os.ReadFile(filepath.Join("demo-biz", "install", "slots.yaml"))
	if err != nil {
		t.Fatalf("reading install/slots.yaml: %v", err)
	}
	checkGolden(t, filepath.Join(testdataDir, "golden", "init-slots.yaml"), string(slots))

	for _, dir := range []string{"ontology", "evals"} {
		info, err := os.Stat(filepath.Join("demo-biz", dir))
		if err != nil {
			t.Errorf("%s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s exists but is not a directory", dir)
		}
	}
}

// TestInitLintsCleanWithEmbeddedOnly is the batch's own proof
// requirement made permanent: a freshly-inited playbook must lint clean
// using embedded resolution, with zero other paths on the command line.
func TestInitLintsCleanWithEmbeddedOnly(t *testing.T) {
	resetInitFlags()
	t.Chdir(t.TempDir())

	if _, stderr, err := runCLI(t, "init", "demo-biz", "--builds-on", "core-operations"); err != nil {
		t.Fatalf("init: %v (stderr: %s)", err, stderr)
	}

	result, err := lint.RunWithEmbedded([]string{"demo-biz"}, embedded.Playbook)
	if err != nil {
		t.Fatalf("lint.RunWithEmbedded: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected a clean lint, got %d finding(s):", len(result.Findings))
		for _, f := range result.Findings {
			t.Errorf("  %s:%d %s: %s", f.File, f.Line, f.Rule, f.Message)
		}
	}
	wantEmbedded := []string{"core-business", "core-operations"}
	if !equalStrings(result.EmbeddedUsed, wantEmbedded) {
		t.Errorf("EmbeddedUsed = %v, want %v", result.EmbeddedUsed, wantEmbedded)
	}
}

// TestInitCoreLayerWithEmptyBuildsOn checks the other legal shape: a new
// root core playbook, layer: core, no --builds-on at all, correctly
// prefixed so KBF013's name check also stays silent.
func TestInitCoreLayerWithEmptyBuildsOn(t *testing.T) {
	resetInitFlags()
	t.Chdir(t.TempDir())

	if _, stderr, err := runCLI(t, "init", "core-my-foundation", "--layer", "core"); err != nil {
		t.Fatalf("init: %v (stderr: %s)", err, stderr)
	}

	result, err := lint.RunWithEmbedded([]string{"core-my-foundation"}, embedded.Playbook)
	if err != nil {
		t.Fatalf("lint.RunWithEmbedded: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected a clean lint, got %d finding(s): %+v", len(result.Findings), result.Findings)
	}
}

// TestInitRequiresBuildsOnForVertical is the guard the batch's own
// dispatch spelled out explicitly: --layer vertical (the default) with
// no --builds-on must not silently default to something; it must error,
// naming the embedded cores as the fix.
func TestInitRequiresBuildsOnForVertical(t *testing.T) {
	resetInitFlags()
	t.Chdir(t.TempDir())

	_, _, err := runCLI(t, "init", "demo-biz")
	if err == nil {
		t.Fatal("expected an error, got none")
	}
	// rootCmd sets SilenceErrors: true (main.go is what prints a
	// RunE-returned error, not Execute itself), so the message to check
	// is the returned error, not anything written to stderr.
	msg := err.Error()
	for _, want := range []string{"--builds-on", "core-business", "core-operations", "core-services"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q; got:\n%s", want, msg)
		}
	}
	if _, err := os.Stat("demo-biz"); err == nil {
		t.Error("demo-biz should not have been created after a rejected init")
	}
}

// TestInitRefusesExistingDirectory: init never overwrites, no --force
// escape hatch (unlike vendor and skill install), since a scaffold
// overwrite could destroy real authored content sitting in ontology/.
func TestInitRefusesExistingDirectory(t *testing.T) {
	resetInitFlags()
	t.Chdir(t.TempDir())

	if _, stderr, err := runCLI(t, "init", "demo-biz", "--builds-on", "core-operations"); err != nil {
		t.Fatalf("first init: %v (stderr: %s)", err, stderr)
	}
	resetInitFlags()
	if err := os.WriteFile(filepath.Join("demo-biz", "sentinel"), []byte("do not lose me"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := runCLI(t, "init", "demo-biz", "--builds-on", "core-operations")
	if err == nil {
		t.Fatal("expected the second init to be refused, it wasn't")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention \"already exists\"", err.Error())
	}
	if _, err := os.Stat(filepath.Join("demo-biz", "sentinel")); err != nil {
		t.Error("sentinel file was lost: init overwrote an existing directory")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
