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

// TestKBF013Valid is the composition happy path (testdata/taxonomy/
// core-root, core-child, vertical-on-core, vertical-on-vertical, hybrid):
// a root (core, empty builds-on), a core building on that core, a
// vertical building on a core, a vertical building on another vertical,
// and a vertical building on two cores at once (the diamond/hybrid
// shape). Must produce zero KBF013 (and zero KBF010/KBF002 on layer)
// findings; the branch tests below each load this set plus the one extra
// package under test, so a failure is never masked by a defect elsewhere
// in the baseline.
func TestKBF013Valid(t *testing.T) {
	result := mustRun(t,
		"testdata/taxonomy/core-root",
		"testdata/taxonomy/core-child",
		"testdata/taxonomy/vertical-on-core",
		"testdata/taxonomy/vertical-on-vertical",
		"testdata/taxonomy/hybrid",
	)
	for _, rule := range []string{lint.KBF013, lint.KBF010, lint.KBF002} {
		if got := find(result.Findings, rule); len(got) != 0 {
			t.Errorf("%s unexpectedly present on the valid composition set: %+v", rule, got)
		}
	}
}

// TestKBF013CoreMustNotBuildOnVertical is KBF013's builds-on-target rule,
// core branch: a core playbook may only build on other core playbooks.
func TestKBF013CoreMustNotBuildOnVertical(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/vertical-on-core", "testdata/taxonomy/core-on-vertical-bad")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	f := got[0]
	if f.Element != "builds-on" {
		t.Errorf("KBF013 element = %q, want builds-on", f.Element)
	}
	if !strings.Contains(f.File, "core-on-vertical-bad") {
		t.Errorf("KBF013 file %q, want the core-on-vertical-bad package's manifest", f.File)
	}
	if !strings.Contains(f.Message, "tax-vertical") || !strings.Contains(f.Message, "vertical") {
		t.Errorf("KBF013 message %q does not name the target and its actual layer", f.Message)
	}
}

// TestKBF013VerticalMustBuildOnSomething is KBF013's builds-on-target
// rule, the other direction: layer: vertical with an empty BuildsOn has
// no derived meaning (unlike core, where empty means root) and must
// build on at least one playbook.
func TestKBF013VerticalMustBuildOnSomething(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/vertical-empty-bad")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if got[0].Element != "builds-on" {
		t.Errorf("KBF013 element = %q, want builds-on", got[0].Element)
	}
	if !strings.Contains(got[0].Message, "must build on") {
		t.Errorf("KBF013 message %q does not explain the missing builds-on", got[0].Message)
	}
}

// TestKBF013SkipsWhenBuildsOnUnresolved confirms the builds-on-layer
// check degrades the same way every other closure-dependent rule does: a
// dangling builds-on entry is KBF011's story alone, not a second,
// confusing KBF013 finding about a parent layer there is no way to know.
func TestKBF013SkipsWhenBuildsOnUnresolved(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/dangling-builds-on")
	if got := find(result.Findings, lint.KBF013); len(got) != 0 {
		t.Errorf("KBF013 fired on a dangling builds-on entry, want silence (KBF011 alone should speak to it): %+v", got)
	}
	got := find(result.Findings, lint.KBF011)
	if len(got) != 1 {
		t.Fatalf("KBF011: got %d findings, want 1: %+v", len(got), result.Findings)
	}
}

// TestKBF013NamePrefixRequired is KBF013's name-prefix rule, core branch:
// layer: core requires a name matching ^core-.
func TestKBF013NamePrefixRequired(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/bad-prefix-core")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	f := got[0]
	if f.Element != "name" {
		t.Errorf("KBF013 element = %q, want name", f.Element)
	}
	if !strings.Contains(f.Fix, "core-") {
		t.Errorf("KBF013 fix %q does not mention the core- prefix", f.Fix)
	}
}

// TestKBF013NamePrefixForbidden is KBF013's name-prefix rule, the other
// direction: layer: vertical must NOT match ^core-.
func TestKBF013NamePrefixForbidden(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/core-root", "testdata/taxonomy/bad-prefix-vertical")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	f := got[0]
	if f.Element != "name" {
		t.Errorf("KBF013 element = %q, want name", f.Element)
	}
	if !strings.Contains(f.Message, "must not") {
		t.Errorf("KBF013 message %q does not state the prohibition", f.Message)
	}
}

// TestKBF010MissingLayer confirms an absent layer joins KBF010's scope
// (empty Entity/Relation/Metric.Tier, Action.Risk) rather than being a
// KBF013 case: there is nothing to cross-check an absent value against.
func TestKBF010MissingLayer(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/no-layer")
	got := find(result.Findings, lint.KBF010)
	if len(got) != 1 {
		t.Fatalf("KBF010: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if got[0].Element != "layer" {
		t.Errorf("KBF010 element = %q, want layer", got[0].Element)
	}
	if got := find(result.Findings, lint.KBF013); len(got) != 0 {
		t.Errorf("KBF013 fired on a package with no layer at all, want silence: %+v", got)
	}
}

// TestKBF002BadLayerValue confirms a non-empty, unrecognized layer value
// is KBF002 (the same treatment as a bad cardinality/additivity/origin/
// approval value), not KBF013: an unrecognized layer has nothing coherent
// to cross-check either.
func TestKBF002BadLayerValue(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/bad-layer-value")
	got := find(result.Findings, lint.KBF002)
	if len(got) != 1 {
		t.Fatalf("KBF002: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if !strings.Contains(got[0].Message, "bogus") {
		t.Errorf("KBF002 message %q does not name the bad value", got[0].Message)
	}
	if got := find(result.Findings, lint.KBF013); len(got) != 0 {
		t.Errorf("KBF013 fired on a package with an invalid layer value, want silence: %+v", got)
	}
}
