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

// realChains: both demo chains this repo ships, dogfooded the same way
// the linter is (design.md, coverage_test.go). design.md is explicit that
// both exist so "one shared multi-vertical core" is checkable, not
// asserted: universal-core is never exercised alone as "the" example,
// only as the common ancestor of two visibly different businesses, so
// both get the same golden protection, not just cafe-demo. Each chain is
// three levels (layered-packages restructure, 2026-08-13): all three
// paths must be passed or the leaf's own extends fails to resolve
// (KBF011), same as any other package whose ancestor isn't among the
// loaded paths.
var realChains = []struct {
	name          string // also the golden file's basename
	paths         []string
	wantEntities  int
	wantRelations int
	wantActions   int
}{
	{
		name: "cafe-demo",
		paths: []string{
			filepath.Join("..", "..", "..", "examples", "cafe-demo"),
			filepath.Join("..", "..", "..", "packages", "operations-core"),
			filepath.Join("..", "..", "..", "packages", "universal-core"),
		},
		// 9 entities: universal-core's 7 (organization, customer,
		// offering, transaction, employee, supplier, purchase) plus
		// operations-core's 2 new ones (location, shift). cafe-demo
		// introduces no entity of its own (its product fragment is a
		// synonym-only override of offering).
		wantEntities: 9,
		// 17 relations: universal-core's own, plus operations-core's 6
		// (belongs-to and responsible-for reused on new pairs, plus the 4
		// minted verbs located-at/staffed-by/works-at/sells), plus
		// cafe-demo's one new (name,from,to) triple (location belongs-to
		// location) that does not collide with anything up the chain.
		wantRelations: 17,
		// 4 actions: all from universal-core; operations-core and
		// cafe-demo add none.
		wantActions: 4,
	},
	{
		name: "studio-demo",
		paths: []string{
			filepath.Join("..", "..", "..", "examples", "studio-demo"),
			filepath.Join("..", "..", "..", "packages", "services-core"),
			filepath.Join("..", "..", "..", "packages", "universal-core"),
		},
		// 9 entities: universal-core's 7 plus services-core's 2 new ones
		// (engagement, deliverable). studio-demo introduces none of its
		// own.
		wantEntities: 9,
		// 15 relations: universal-core's own, plus services-core's 5
		// (places, contains, derived-from, responsible-for all reused on
		// new pairs; services-core mints no new verb at all), plus
		// studio-demo's one new triple (customer belongs-to customer).
		wantRelations: 15,
		wantActions:   4,
	},
}

func TestBuildGraphAgainstRealPackages(t *testing.T) {
	for _, c := range realChains {
		t.Run(c.name, func(t *testing.T) {
			universe, findings, err := lint.Load(c.paths)
			if err != nil {
				t.Fatalf("lint.Load: %v", err)
			}
			if len(findings) != 0 {
				t.Fatalf("lint.Load produced findings, want none: %+v", findings)
			}

			g := compile.BuildGraph(universe)
			if len(g.Entities) != c.wantEntities {
				t.Errorf("got %d entities, want %d", len(g.Entities), c.wantEntities)
			}
			if len(g.Relations) != c.wantRelations {
				t.Errorf("got %d relations, want %d", len(g.Relations), c.wantRelations)
			}
			if len(g.Actions) != c.wantActions {
				t.Errorf("got %d actions, want %d", len(g.Actions), c.wantActions)
			}

			// Determinism: building the graph twice from the same universe
			// must produce byte-identical mermaid output (design.md:
			// "everything deterministic: sorted iteration everywhere").
			first := compile.ToMermaid(g)
			second := compile.ToMermaid(compile.BuildGraph(universe))
			if first != second {
				t.Error("ToMermaid(BuildGraph(u)) is not deterministic across repeated calls")
			}
		})
	}
}

func TestMermaidGolden(t *testing.T) {
	for _, c := range realChains {
		t.Run(c.name, func(t *testing.T) {
			universe, _, err := lint.Load(c.paths)
			if err != nil {
				t.Fatalf("lint.Load: %v", err)
			}
			got := compile.ToMermaid(compile.BuildGraph(universe))

			goldenFile := filepath.Join("testdata", "golden", c.name+".mmd")
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
		})
	}
}
