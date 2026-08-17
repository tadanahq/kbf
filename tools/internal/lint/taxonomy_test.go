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

// TestKBF013Valid is the three-layer happy path (testdata/taxonomy/
// root, base, vertical): a root, a base extending it, and a vertical
// extending the base, name-prefixed exactly as spec/conventions.md's
// "core playbooks" vs "vertical playbooks" split documents. The branch
// tests below each load this trio (or a subset of it) plus the one extra
// package under test, so a failure is never masked by a defect elsewhere
// in the baseline.
func TestKBF013Valid(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/root", "testdata/taxonomy/base", "testdata/taxonomy/vertical")
	for _, rule := range []string{lint.KBF013, lint.KBF010, lint.KBF002} {
		if got := find(result.Findings, rule); len(got) != 0 {
			t.Errorf("%s unexpectedly present on the valid 3-layer chain: %+v", rule, got)
		}
	}
}

// TestKBF013RootMustNotExtend is KBF013's extends-target rule, root
// branch: layer: root with a non-null extends.
func TestKBF013RootMustNotExtend(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/root", "testdata/taxonomy/root-bad-extends")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	f := got[0]
	if f.Element != "extends" {
		t.Errorf("KBF013 element = %q, want extends", f.Element)
	}
	if !strings.Contains(f.File, "root-bad-extends") {
		t.Errorf("KBF013 file %q, want the root-bad-extends package's manifest", f.File)
	}
	if !strings.Contains(f.Message, "extends: null") {
		t.Errorf("KBF013 message %q does not say what a root's extends must be", f.Message)
	}
}

// TestKBF013BaseMustExtend is KBF013's extends-target rule, the other
// direction: layer: base (or vertical; both share the same requirement)
// with extends: null has nothing to check a parent's layer against
// because there is no parent at all.
func TestKBF013BaseMustExtend(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/base-no-extends")
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if got[0].Element != "extends" {
		t.Errorf("KBF013 element = %q, want extends", got[0].Element)
	}
	if !strings.Contains(got[0].Message, "must extend") {
		t.Errorf("KBF013 message %q does not explain the missing extends", got[0].Message)
	}
}

// TestKBF013WrongParentLayer is KBF013's extends-target rule when extends
// resolves to a real playbook whose own layer is not root or base:
// core-tax-base-bad-parent extends tax-vertical, a leaf, which no base or
// vertical playbook may do.
func TestKBF013WrongParentLayer(t *testing.T) {
	result := mustRun(t,
		"testdata/taxonomy/root",
		"testdata/taxonomy/base",
		"testdata/taxonomy/vertical",
		"testdata/taxonomy/base-bad-parent",
	)
	got := find(result.Findings, lint.KBF013)
	if len(got) != 1 {
		t.Fatalf("KBF013: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	f := got[0]
	if !strings.Contains(f.File, "base-bad-parent") {
		t.Errorf("KBF013 file %q, want the base-bad-parent package's manifest", f.File)
	}
	if !strings.Contains(f.Message, "tax-vertical") || !strings.Contains(f.Message, "vertical") {
		t.Errorf("KBF013 message %q does not name the parent and its actual layer", f.Message)
	}
}

// TestKBF013SkipsWhenExtendsUnresolved confirms the extends-layer check
// degrades the same way every other chain-dependent rule does: a
// dangling extends is KBF011's story alone, not a second, confusing
// KBF013 finding about a parent layer there is no way to know.
func TestKBF013SkipsWhenExtendsUnresolved(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/dangling-extends")
	if got := find(result.Findings, lint.KBF013); len(got) != 0 {
		t.Errorf("KBF013 fired on a dangling extends, want silence (KBF011 alone should speak to it): %+v", got)
	}
	got := find(result.Findings, lint.KBF011)
	if len(got) != 1 {
		t.Fatalf("KBF011: got %d findings, want 1: %+v", len(got), result.Findings)
	}
}

// TestKBF013NamePrefixRequired is KBF013's name-prefix rule, root/base
// branch: layer: root (or base) requires a name matching ^core-.
func TestKBF013NamePrefixRequired(t *testing.T) {
	result := mustRun(t, "testdata/taxonomy/bad-prefix-root")
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
	result := mustRun(t, "testdata/taxonomy/root", "testdata/taxonomy/bad-prefix-vertical")
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
// is KBF002 (the same treatment as a bad tier/cardinality/additivity/
// risk value), not KBF013: an unrecognized layer has nothing coherent to
// cross-check either.
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
