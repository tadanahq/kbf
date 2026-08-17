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

// TestKBF003CrossPlaybookCollision is KBF003's cross-playbook variant
// (Batch 7's composition mechanics): coll-root-a and coll-root-b both
// declare `widget`; neither builds on the other, but coll-composer
// builds on both, so from composer's own composition closure the
// identity is ambiguous. Composition has no resolution order (unlike the
// single-parent chain's "nearest ancestor wins"), so this is always an
// error, reported symmetrically against both colliding files.
func TestKBF003CrossPlaybookCollision(t *testing.T) {
	result := mustRun(t,
		"testdata/collision/root-a",
		"testdata/collision/root-b",
		"testdata/collision/composer",
	)
	got := find(result.Findings, lint.KBF003)
	if len(got) != 2 {
		t.Fatalf("KBF003: got %d findings, want 2 (one per colliding playbook): %+v", len(got), result.Findings)
	}

	var sawA, sawB bool
	for _, f := range got {
		if f.Element != "widget" {
			t.Errorf("KBF003 element = %q, want widget", f.Element)
		}
		switch {
		case strings.Contains(f.File, "root-a"):
			sawA = true
			if !strings.Contains(f.Message, "coll-root-b") {
				t.Errorf("KBF003 finding on root-a's file %q does not name coll-root-b", f.Message)
			}
		case strings.Contains(f.File, "root-b"):
			sawB = true
			if !strings.Contains(f.Message, "coll-root-a") {
				t.Errorf("KBF003 finding on root-b's file %q does not name coll-root-a", f.Message)
			}
		default:
			t.Errorf("KBF003 finding in unexpected file: %+v", f)
		}
	}
	if !sawA || !sawB {
		t.Errorf("KBF003 findings don't cover both colliding files: %+v", got)
	}
}

// TestKBF003NoCollisionWithoutComposition confirms merely loading two
// playbooks in the same kbf lint invocation is not the same as composing
// them: coll-root-a and coll-root-b both declare `widget`, but with
// neither building on the other and nothing composing them together,
// there is no closure in which the two ever meet.
func TestKBF003NoCollisionWithoutComposition(t *testing.T) {
	result := mustRun(t, "testdata/collision/root-a", "testdata/collision/root-b")
	if got := find(result.Findings, lint.KBF003); len(got) != 0 {
		t.Errorf("KBF003 fired without anything composing root-a and root-b together: %+v", got)
	}
}
