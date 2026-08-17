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

// realChains: both demo chains this repo ships, dogfooded the same way
// the linter is (design.md and tasks.md's "Implementation
// clarifications"). design.md is explicit that both exist so "one shared
// multi-vertical core" is checkable, not asserted, so both get the same
// golden protection, not just cafe-demo. Each chain is three levels
// (layered-packages restructure, 2026-08-13): all three paths must be
// passed or the leaf's own extends fails to resolve (KBF011).
var realChains = []struct {
	name     string // also the golden files' basename
	paths    []string
	declared int
	mapped   int
	unmapped []string
}{
	{
		name: "cafe-demo",
		paths: []string{
			filepath.Join("..", "..", "..", "examples", "cafe-demo"),
			filepath.Join("..", "..", "..", "packages", "operations-core"),
			filepath.Join("..", "..", "..", "packages", "universal-core"),
		},
		// 27 slots: universal-core's 21 + operations-core's 6 (location
		// x3, shift x3), the full resolved ontology across the chain.
		declared: 27,
		mapped:   24,
		unmapped: []string{"crm.customer-contact", "crm.customer-joined-date", "crm.customer-name"},
	},
	{
		name: "studio-demo",
		paths: []string{
			filepath.Join("..", "..", "..", "examples", "studio-demo"),
			filepath.Join("..", "..", "..", "packages", "services-core"),
			filepath.Join("..", "..", "..", "packages", "universal-core"),
		},
		// 29 slots: universal-core's 21 + services-core's 8 (engagement
		// x5, deliverable x3).
		declared: 29,
		mapped:   27,
		unmapped: []string{"purchasing.supplier-contact", "purchasing.supplier-name"},
	},
}

func TestComputeAgainstRealPackages(t *testing.T) {
	for _, c := range realChains {
		t.Run(c.name, func(t *testing.T) {
			universe, findings, err := lint.Load(c.paths)
			if err != nil {
				t.Fatalf("lint.Load: %v", err)
			}
			if len(findings) != 0 {
				t.Fatalf("lint.Load produced findings, want none (packages/examples are expected to lint clean): %+v", findings)
			}

			reports := coverage.Compute(universe)
			if len(reports) != 1 {
				t.Fatalf("got %d reports, want 1 (only the demo is a leaf; its base/universal-core ancestors are extends-context, not a subject): %+v", len(reports), reports)
			}

			r := reports[0]
			if r.Package != c.name {
				t.Errorf("report.Package = %q, want %s", r.Package, c.name)
			}
			if r.Declared != c.declared {
				t.Errorf("Declared = %d, want %d", r.Declared, c.declared)
			}
			if r.Mapped != c.mapped {
				t.Errorf("Mapped = %d, want %d", r.Mapped, c.mapped)
			}

			unmapped := map[string]bool{}
			for _, row := range r.Rows {
				if !row.Mapped {
					unmapped[row.Slot] = true
				}
			}
			for _, slot := range c.unmapped {
				if !unmapped[slot] {
					t.Errorf("expected %s to be unmapped, it wasn't", slot)
				}
			}
			if len(unmapped) != len(c.unmapped) {
				t.Errorf("got %d unmapped slots, want exactly %d: %v", len(unmapped), len(c.unmapped), unmapped)
			}
		})
	}
}

func TestRenderGolden(t *testing.T) {
	for _, c := range realChains {
		t.Run(c.name, func(t *testing.T) {
			universe, _, err := lint.Load(c.paths)
			if err != nil {
				t.Fatalf("lint.Load: %v", err)
			}
			reports := coverage.Compute(universe)

			checkGolden(t, filepath.Join("testdata", "golden", c.name+".human.txt"), coverage.RenderHuman(reports))

			out, err := coverage.RenderJSON(reports)
			if err != nil {
				t.Fatalf("RenderJSON: %v", err)
			}
			checkGolden(t, filepath.Join("testdata", "golden", c.name+".json"), string(out)+"\n")
		})
	}
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
	if reports[0].Declared != 21 {
		t.Errorf("Declared = %d, want 21", reports[0].Declared)
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
