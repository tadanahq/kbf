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

// Package conformance is the thin runner design.md promises: it walks
// conformance/ and lints each fixture's input/ packages, checking the
// outcome against expect.yaml. The fixtures themselves are plain YAML,
// no Go (project-standards.md: "conformance is data"), so another KBF
// implementation can prove compliance without depending on this file.
package conformance_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	yaml "github.com/goccy/go-yaml"

	"github.com/tadanahq/kbf/tools/internal/lint"
)

// conformanceDir is conformance/ at the repo root, relative to this test
// file's own package (tools/internal/conformance).
const conformanceDir = "../../../conformance"

// expectation is expect.yaml's shape: `ok: true` for a fixture that must
// lint clean, or `ok: false` plus the rule ids every one of which must
// fire for a fixture that must not.
type expectation struct {
	OK    bool     `yaml:"ok"`
	Rules []string `yaml:"rules"`
}

// TestConformance runs every fixture under conformance/ as its own
// subtest, so `go test -run TestConformance/<fixture>` isolates one case
// and a single broken fixture doesn't hide failures in the others.
func TestConformance(t *testing.T) {
	entries, err := os.ReadDir(conformanceDir)
	if err != nil {
		t.Fatalf("reading %s: %v", conformanceDir, err)
	}

	var fixtures []string
	for _, e := range entries {
		if e.IsDir() {
			fixtures = append(fixtures, e.Name())
		}
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no fixtures found under conformance/")
	}

	var valid, invalid int
	for _, name := range fixtures {
		dir := filepath.Join(conformanceDir, name)
		ok := t.Run(name, func(t *testing.T) { runFixture(t, dir) })
		if !ok {
			continue
		}
		exp := readExpectation(t, dir)
		if exp.OK {
			valid++
		} else {
			invalid++
		}
	}

	// The capsule requires at least 6 of each (overview.md's acceptance
	// criteria): assert it here so a fixture accidentally deleted from
	// conformance/ fails loudly instead of just quietly shrinking coverage.
	if valid < 6 {
		t.Errorf("only %d valid fixtures ran clean, want at least 6", valid)
	}
	if invalid < 6 {
		t.Errorf("only %d invalid fixtures ran clean, want at least 6", invalid)
	}
}

// runFixture lints dir/input's packages and checks the outcome against
// dir/expect.yaml.
func runFixture(t *testing.T, dir string) {
	t.Helper()
	exp := readExpectation(t, dir)

	inputDir := filepath.Join(dir, "input")
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("reading %s: %v", inputDir, err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			paths = append(paths, filepath.Join(inputDir, e.Name()))
		}
	}
	sort.Strings(paths) // deterministic load order; also alphabetizes parent before child by convention

	result, err := lint.Run(paths)
	if err != nil {
		t.Fatalf("lint.Run(%v): %v", paths, err)
	}

	if exp.OK {
		if len(result.Findings) != 0 {
			t.Errorf("expected a clean lint, got %d finding(s):", len(result.Findings))
			for _, f := range result.Findings {
				t.Errorf("  %s:%d %s: %s", f.File, f.Line, f.Rule, f.Message)
			}
		}
		return
	}

	if len(exp.Rules) == 0 {
		t.Fatalf("expect.yaml has ok: false but no rules listed; name at least one rule id that must fire")
	}
	got := map[string]bool{}
	for _, f := range result.Findings {
		got[f.Rule] = true
	}
	for _, want := range exp.Rules {
		if !got[want] {
			t.Errorf("expected rule %s to fire, it didn't. Findings: %+v", want, result.Findings)
		}
	}
}

func readExpectation(t *testing.T, dir string) expectation {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "expect.yaml"))
	if err != nil {
		t.Fatalf("reading expect.yaml: %v", err)
	}
	var exp expectation
	if err := yaml.Unmarshal(data, &exp); err != nil {
		t.Fatalf("parsing expect.yaml: %v", err)
	}
	return exp
}
