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

func resetVendorFlags() {
	vendorTo = "playbooks"
	vendorForce = false
}

// TestVendor checks the default destination, that all three core
// playbooks land, and that the result lints clean standalone (no
// embedded fallback needed: everything vendor writes is now local).
func TestVendor(t *testing.T) {
	resetVendorFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "vendor"); err != nil {
		t.Fatalf("vendor: %v", err)
	}

	var paths []string
	for _, name := range embedded.CorePlaybookNames() {
		dir := filepath.Join("playbooks", name)
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
		t.Errorf("vendored playbooks do not lint clean: %+v", result.Findings)
	}
}

// TestVendorIdempotencyAndForce mirrors skill install's own guard: a
// second vendor without --force is refused and changes nothing; --force
// overwrites.
func TestVendorIdempotencyAndForce(t *testing.T) {
	resetVendorFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "vendor"); err != nil {
		t.Fatalf("first vendor: %v", err)
	}

	marker := filepath.Join("playbooks", "core-business", "manifest.yaml")
	if err := os.WriteFile(marker, []byte("locally edited"), 0o644); err != nil {
		t.Fatal(err)
	}

	resetVendorFlags()
	if _, _, err := runCLI(t, "vendor"); err == nil {
		t.Fatal("expected the second vendor to be refused without --force, it wasn't")
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

	resetVendorFlags()
	if _, _, err := runCLI(t, "vendor", "--force"); err != nil {
		t.Fatalf("forced vendor: %v", err)
	}
	got, err = os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "locally edited" {
		t.Error("--force did not overwrite the existing file")
	}
}

// TestVendorTo checks --to writes under a custom directory instead of
// the default ./playbooks.
func TestVendorTo(t *testing.T) {
	resetVendorFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "vendor", "--to", "third_party/kbf"); err != nil {
		t.Fatalf("vendor --to: %v", err)
	}
	if _, err := os.Stat(filepath.Join("third_party", "kbf", "core-business", "manifest.yaml")); err != nil {
		t.Errorf("--to third_party/kbf: %v", err)
	}
}
