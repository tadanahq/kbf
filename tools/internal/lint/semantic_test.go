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

func TestKBF005MetricIncomplete(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF005)
	if len(got) != 0 {
		t.Fatalf("KBF005 unexpectedly present: %+v (this fixture pair does not exercise it directly; a metric fragment's empty grain is meant to be skipped, see KBF008 test)", got)
	}
}

func TestKBF006BadEndpoint(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF006)
	if len(got) != 1 {
		t.Fatalf("KBF006: got %d findings, want 1: %+v", len(got), got)
	}
	if got[0].Element != "nonexistent-entity" {
		t.Errorf("KBF006 element = %q, want nonexistent-entity", got[0].Element)
	}
}

func TestKBF007VerbOutsideVocabulary(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF007)
	if len(got) != 1 {
		t.Fatalf("KBF007: got %d findings, want 1: %+v", len(got), got)
	}
	if got[0].Element != "invented-verb" {
		t.Errorf("KBF007 element = %q, want invented-verb", got[0].Element)
	}
}

func TestKBF008ForkAndLegitimateOverride(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF008)

	elements := map[string]bool{}
	for _, f := range got {
		elements[f.Element] = true
	}
	// widget (entity) and widget-count (metric) both set a non-glossary
	// field on top of a parent match: forks.
	for _, want := range []string{"widget", "widget-count"} {
		if !elements[want] {
			t.Errorf("KBF008: no fork finding for %q; got %+v", want, got)
		}
	}
	// gizmo (entity, synonyms only) and gizmo-ratio (metric, thresholds
	// only) are legitimate glossary overrides: must NOT appear.
	for _, mustNotFork := range []string{"gizmo", "gizmo-ratio"} {
		if elements[mustNotFork] {
			t.Errorf("KBF008: %q flagged as a fork, but it only sets its glossary-eligible field", mustNotFork)
		}
	}
	if len(got) != 2 {
		t.Errorf("KBF008: got %d findings, want exactly 2 (widget, widget-count): %+v", len(got), got)
	}

	// The legitimate overrides must not trip completeness rules either:
	// their emptiness is inherited from the parent, not a violation.
	for _, rule := range []string{lint.KBF004, lint.KBF010, lint.KBF005} {
		for _, f := range find(result.Findings, rule) {
			if f.Element == "gizmo" || f.Element == "gizmo-ratio" {
				t.Errorf("%s fired on legitimate override %q: %+v", rule, f.Element, f)
			}
		}
	}
}

func TestKBF009DanglingReferences(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF009)

	elements := map[string]bool{}
	for _, f := range got {
		elements[f.Element] = true
	}
	if !elements["nonexistent-entity"] {
		t.Errorf("KBF009: no finding for orphan-metric's dangling grain entry; got %+v", got)
	}
	if !elements["totally-made-up-name"] {
		t.Errorf("KBF009: no finding for the competency question's dangling expects entry; got %+v", got)
	}
	if len(got) != 2 {
		t.Errorf("KBF009: got %d findings, want exactly 2: %+v", len(got), got)
	}
}

func TestKBF012OrphanSlot(t *testing.T) {
	result := mustRun(t, "testdata/semantic/parent", "testdata/semantic/child")
	got := find(result.Findings, lint.KBF012)
	if len(got) != 1 {
		t.Fatalf("KBF012: got %d findings, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Element != "catalog.orphan-slot" {
		t.Errorf("KBF012 element = %q, want catalog.orphan-slot", f.Element)
	}
	if !strings.Contains(f.File, "child") {
		t.Errorf("KBF012 file %q, want the child package's slots.yaml", f.File)
	}
	// The inherited slot (catalog.widget-label, used by the parent's
	// widget entity, not redeclared by the child) must not be flagged:
	// an install configures the resolved ontology, not just local content.
	for _, other := range got {
		if other.Element == "catalog.widget-label" {
			t.Errorf("KBF012 incorrectly fired on an inherited, matched slot: %+v", other)
		}
	}
}
