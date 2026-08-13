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

package compile_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/tadanahq/kbf/tools/internal/compile"
	"github.com/tadanahq/kbf/tools/internal/lint"
)

var update = flag.Bool("update", false, "write golden files instead of comparing against them")

// realPackages: see internal/coverage/coverage_test.go for why this is
// dogfooded against the real content this repo ships, not a synthetic
// fixture.
var realPackages = []string{
	filepath.Join("..", "..", "..", "examples", "cafe-demo"),
	filepath.Join("..", "..", "..", "packages", "universal-core"),
}

func TestBuildGraphAgainstRealPackages(t *testing.T) {
	universe, findings, err := lint.Load(realPackages)
	if err != nil {
		t.Fatalf("lint.Load: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("lint.Load produced findings, want none: %+v", findings)
	}

	g := compile.BuildGraph(universe)
	if len(g.Entities) != 9 {
		t.Errorf("got %d entities, want 9 (universal-core's own; cafe-demo's product fragment is a synonym-only override, not a new entity)", len(g.Entities))
	}
	// universal-core declares 15 relations; cafe-demo adds one new
	// (name,from,to) triple (location belongs-to location) that does not
	// collide with anything in the parent, so the union is 16, not 15.
	if len(g.Relations) != 16 {
		t.Errorf("got %d relations, want 16", len(g.Relations))
	}
	if len(g.Actions) != 4 {
		t.Errorf("got %d actions, want 4 (all from universal-core; cafe-demo adds none)", len(g.Actions))
	}

	// Determinism: building the graph twice from the same universe must
	// produce byte-identical mermaid output (design.md: "everything
	// deterministic: sorted iteration everywhere").
	first := compile.ToMermaid(g)
	second := compile.ToMermaid(compile.BuildGraph(universe))
	if first != second {
		t.Error("ToMermaid(BuildGraph(u)) is not deterministic across repeated calls")
	}
}

func TestMermaidGolden(t *testing.T) {
	universe, _, err := lint.Load(realPackages)
	if err != nil {
		t.Fatalf("lint.Load: %v", err)
	}
	got := compile.ToMermaid(compile.BuildGraph(universe))

	goldenFile := filepath.Join("testdata", "golden", "cafe-demo.mmd")
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
		t.Errorf("mermaid output does not match %s\n--- got ---\n%s\n--- want ---\n%s", goldenFile, got, string(want))
	}
}
