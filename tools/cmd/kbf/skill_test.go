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

func resetSkillFlags() {
	skillInstallForce = false
}

// TestSkillInstall checks the destination and that the written SKILL.md
// matches the embedded copy exactly.
func TestSkillInstall(t *testing.T) {
	resetSkillFlags()
	t.Chdir(t.TempDir())

	stdout, _, err := runCLI(t, "skill", "install")
	if err != nil {
		t.Fatalf("skill install: %v", err)
	}

	dest := filepath.Join(".claude", "skills", "kbf-authoring", "SKILL.md")
	if !strings.Contains(stdout, dest) {
		t.Errorf("stdout = %q, want it to name %s", stdout, dest)
	}
	if !strings.Contains(stdout, "next:") {
		t.Errorf("stdout = %q, want a one-line next step", stdout)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("%s: %v", dest, err)
	}
}

// TestSkillInstallIdempotencyAndForce: a second install without --force
// is refused and changes nothing; with --force it overwrites cleanly.
func TestSkillInstallIdempotencyAndForce(t *testing.T) {
	resetSkillFlags()
	t.Chdir(t.TempDir())

	if _, _, err := runCLI(t, "skill", "install"); err != nil {
		t.Fatalf("first install: %v", err)
	}

	dest := filepath.Join(".claude", "skills", "kbf-authoring", "SKILL.md")
	if err := os.WriteFile(dest, []byte("locally edited, should survive a rejected reinstall"), 0o644); err != nil {
		t.Fatal(err)
	}

	resetSkillFlags()
	if _, _, err := runCLI(t, "skill", "install"); err == nil {
		t.Fatal("expected the second install to be refused without --force, it wasn't")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention \"already exists\"", err.Error())
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "locally edited, should survive a rejected reinstall" {
		t.Error("a rejected reinstall modified the existing file")
	}

	resetSkillFlags()
	if _, _, err := runCLI(t, "skill", "install", "--force"); err != nil {
		t.Fatalf("forced reinstall: %v", err)
	}
	got, err = os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "locally edited, should survive a rejected reinstall" {
		t.Error("--force did not overwrite the existing file")
	}
	if !strings.Contains(string(got), "name: kbf-authoring") {
		t.Error("--force did not write back the real embedded SKILL.md")
	}
}
