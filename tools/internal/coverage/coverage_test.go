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

package coverage_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/tadanahq/kbf/tools/internal/coverage"
	"github.com/tadanahq/kbf/tools/internal/lint"
)

// Fixed, colorless profile for this test binary only: see
// internal/lint/render_test.go for why (the real CLI keeps normal
// terminal auto-detection).
func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

var update = flag.Bool("update", false, "write golden files instead of comparing against them")

// realPackages resolves the actual packages/examples this repo ships,
// relative to this test file: coverage is meant to be dogfooded against
// real content, the same way the linter is (see design.md and tasks.md's
// "Implementation clarifications").
var realPackages = []string{
	filepath.Join("..", "..", "..", "examples", "cafe-demo"),
	filepath.Join("..", "..", "..", "packages", "universal-core"),
}

func TestComputeAgainstRealPackages(t *testing.T) {
	universe, findings, err := lint.Load(realPackages)
	if err != nil {
		t.Fatalf("lint.Load: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("lint.Load produced findings, want none (packages/examples are expected to lint clean): %+v", findings)
	}

	reports := coverage.Compute(universe)
	if len(reports) != 1 {
		t.Fatalf("got %d reports, want 1 (only cafe-demo is a leaf; universal-core is extends-context, not a subject): %+v", len(reports), reports)
	}

	r := reports[0]
	if r.Package != "cafe-demo" {
		t.Errorf("report.Package = %q, want cafe-demo", r.Package)
	}
	if r.Declared != 26 {
		t.Errorf("Declared = %d, want 26", r.Declared)
	}
	if r.Mapped != 23 {
		t.Errorf("Mapped = %d, want 23 (3 crm.customer-* rows are deliberately unmapped)", r.Mapped)
	}

	unmapped := map[string]bool{}
	for _, row := range r.Rows {
		if !row.Mapped {
			unmapped[row.Slot] = true
		}
	}
	want := []string{"crm.customer-contact", "crm.customer-joined-date", "crm.customer-name"}
	for _, slot := range want {
		if !unmapped[slot] {
			t.Errorf("expected %s to be unmapped, it wasn't", slot)
		}
	}
	if len(unmapped) != len(want) {
		t.Errorf("got %d unmapped slots, want exactly %d: %v", len(unmapped), len(want), unmapped)
	}
}

func TestRenderGolden(t *testing.T) {
	universe, _, err := lint.Load(realPackages)
	if err != nil {
		t.Fatalf("lint.Load: %v", err)
	}
	reports := coverage.Compute(universe)

	checkGolden(t, filepath.Join("testdata", "golden", "cafe-demo.human.txt"), coverage.RenderHuman(reports))

	out, err := coverage.RenderJSON(reports)
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	checkGolden(t, filepath.Join("testdata", "golden", "cafe-demo.json"), string(out)+"\n")
}

func TestUniversalCoreAlone(t *testing.T) {
	universe, _, err := lint.Load([]string{filepath.Join("..", "..", "..", "packages", "universal-core")})
	if err != nil {
		t.Fatalf("lint.Load: %v", err)
	}
	reports := coverage.Compute(universe)
	if len(reports) != 1 {
		t.Fatalf("got %d reports, want 1: linted alone, universal-core is trivially its own leaf", len(reports))
	}
	if reports[0].Mapped != 0 {
		t.Errorf("Mapped = %d, want 0: universal-core's slots.yaml is a template, every source empty", reports[0].Mapped)
	}
	if reports[0].Declared != 26 {
		t.Errorf("Declared = %d, want 26", reports[0].Declared)
	}
}

func checkGolden(t *testing.T, goldenFile, got string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenFile, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("reading golden file %s: %v (run go test -run %s -update to create it)", goldenFile, err, t.Name())
	}
	if got != string(want) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s", goldenFile, got, string(want))
	}
}
