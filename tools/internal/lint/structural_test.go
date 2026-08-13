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
	"strings"
	"testing"

	"github.com/tadanahq/kbf/tools/internal/lint"
)

// find returns the findings matching rule, for assertions that need to
// inspect more than "did it fire" (message/fix content: design.md and
// AGENTS.md both require error messages to be product, not an
// afterthought, so tests hold that content to account).
func find(findings []lint.Finding, rule string) []lint.Finding {
	var out []lint.Finding
	for _, f := range findings {
		if f.Rule == rule {
			out = append(out, f)
		}
	}
	return out
}

// mustRun lints paths and fails the test on a genuine (non-lint) error:
// a bad path, not a rule violation.
func mustRun(t *testing.T, paths ...string) lint.Result {
	t.Helper()
	result, err := lint.Run(paths)
	if err != nil {
		t.Fatalf("lint.Run(%v): %v", paths, err)
	}
	return result
}

func TestCleanPackageHasNoFindings(t *testing.T) {
	result := mustRun(t, "testdata/clean")
	if len(result.Findings) != 0 {
		t.Errorf("clean fixture produced findings, want none: %+v", result.Findings)
	}
}

func TestKBF001UnknownField(t *testing.T) {
	result := mustRun(t, "testdata/structural")
	got := find(result.Findings, lint.KBF001)
	if len(got) != 1 {
		t.Fatalf("KBF001: got %d findings, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Element != "not_a_real_field" {
		t.Errorf("KBF001 element = %q, want the unknown field name", f.Element)
	}
	if !strings.Contains(f.Message, "not_a_real_field") {
		t.Errorf("KBF001 message %q does not name the field", f.Message)
	}
	if f.Fix == "" {
		t.Error("KBF001 finding has no fix hint")
	}
}

func TestKBF002BadEnumValue(t *testing.T) {
	result := mustRun(t, "testdata/structural")
	got := find(result.Findings, lint.KBF002)
	if len(got) != 1 {
		t.Fatalf("KBF002: got %d findings, want 1: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "sometimes") {
		t.Errorf("KBF002 message %q does not name the bad value", got[0].Message)
	}
}

func TestKBF003DuplicateName(t *testing.T) {
	result := mustRun(t, "testdata/structural")
	got := find(result.Findings, lint.KBF003)
	if len(got) != 1 {
		t.Fatalf("KBF003: got %d findings, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Element != "widget" {
		t.Errorf("KBF003 element = %q, want widget", f.Element)
	}
	if !strings.Contains(f.Message, "widget") {
		t.Errorf("KBF003 message %q does not name the duplicate", f.Message)
	}
}

func TestKBF004NoIdentity(t *testing.T) {
	result := mustRun(t, "testdata/structural")
	got := find(result.Findings, lint.KBF004)
	if len(got) != 1 {
		t.Fatalf("KBF004: got %d findings, want 1: %+v", len(got), got)
	}
	if got[0].Element != "sprocket" {
		t.Errorf("KBF004 element = %q, want sprocket", got[0].Element)
	}
}

func TestKBF010MissingGovernanceTier(t *testing.T) {
	result := mustRun(t, "testdata/structural")
	got := find(result.Findings, lint.KBF010)
	if len(got) != 1 {
		t.Fatalf("KBF010: got %d findings, want 1 (sprocket's entity tier): %+v", len(got), got)
	}
	if got[0].Element != "sprocket" {
		t.Errorf("KBF010 element = %q, want sprocket", got[0].Element)
	}
}

func TestKBF011ManifestMissingFields(t *testing.T) {
	result := mustRun(t, "testdata/manifest-missing-fields")
	got := find(result.Findings, lint.KBF011)
	if len(got) != 3 {
		t.Fatalf("KBF011: got %d findings, want 3 (name, version, spec): %+v", len(got), got)
	}
	elements := map[string]bool{}
	for _, f := range got {
		elements[f.Element] = true
	}
	for _, want := range []string{"name", "version", "spec"} {
		if !elements[want] {
			t.Errorf("KBF011: no finding for missing %q", want)
		}
	}
}

func TestKBF011UnresolvableExtends(t *testing.T) {
	result := mustRun(t, "testdata/manifest-bad-extends")
	got := find(result.Findings, lint.KBF011)
	if len(got) != 1 {
		t.Fatalf("KBF011: got %d findings, want 1: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "some-package-that-was-never-loaded") {
		t.Errorf("KBF011 message %q does not name the unresolved package", got[0].Message)
	}
	if !strings.Contains(got[0].Fix, "path") {
		t.Errorf("KBF011 fix %q does not point at passing the parent's path", got[0].Fix)
	}
}
