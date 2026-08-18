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

func resetPinFlags() {
	pinTo = "builds-on"
	pinForce = false
}

// TestPlaybooksPin checks the default destination (builds-on, the
// install-repo layout), that all three core playbooks land, and that
// the result lints clean standalone (no embedded fallback needed:
// everything pin writes is now local).
func TestPlaybooksPin(t *testing.T) {
	resetPinFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "playbooks", "pin"); err != nil {
		t.Fatalf("playbooks pin: %v", err)
	}

	var paths []string
	for _, name := range embedded.CorePlaybookNames() {
		dir := filepath.Join("builds-on", name)
		if _, err := os.Stat(filepath.Join(dir, "manifest.yaml")); err != nil {
			t.Errorf("%s/manifest.yaml: %v", dir, err)
		}
		paths = append(paths, dir)
	}

	result, err := lint.Run(paths) // no fallback: every playbook is local now
	if err != nil {
		t.Fatalf("lint.Run: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("pinned playbooks do not lint clean: %+v", result.Findings)
	}
}

// TestPlaybooksPinIdempotencyAndForce mirrors skill install's own
// guard: a second pin without --force is refused and changes nothing;
// --force overwrites.
func TestPlaybooksPinIdempotencyAndForce(t *testing.T) {
	resetPinFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "playbooks", "pin"); err != nil {
		t.Fatalf("first pin: %v", err)
	}

	marker := filepath.Join("builds-on", "core-business", "manifest.yaml")
	if err := os.WriteFile(marker, []byte("locally edited"), 0o644); err != nil {
		t.Fatal(err)
	}

	resetPinFlags()
	if _, _, err := runCLI(t, "playbooks", "pin"); err == nil {
		t.Fatal("expected the second pin to be refused without --force, it wasn't")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention \"already exists\"", err.Error())
	}
	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "locally edited" {
		t.Error("a rejected reinstall modified the existing file")
	}

	resetPinFlags()
	if _, _, err := runCLI(t, "playbooks", "pin", "--force"); err != nil {
		t.Fatalf("forced pin: %v", err)
	}
	got, err = os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "locally edited" {
		t.Error("--force did not overwrite the existing file")
	}
}

// TestPlaybooksPinTo checks --to writes under a custom directory
// instead of the default ./builds-on.
func TestPlaybooksPinTo(t *testing.T) {
	resetPinFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "playbooks", "pin", "--to", "third_party/kbf"); err != nil {
		t.Fatalf("playbooks pin --to: %v", err)
	}
	if _, err := os.Stat(filepath.Join("third_party", "kbf", "core-business", "manifest.yaml")); err != nil {
		t.Errorf("--to third_party/kbf: %v", err)
	}
}
