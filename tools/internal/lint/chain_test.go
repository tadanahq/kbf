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

// TestChainThreeLevelValid is the recursive-extends happy path: a
// grandparent -> parent -> child chain where the child references
// entities two layers up (KBF006/009), reuses two verbs the grandparent
// (the only true root, hence the only layer allowed to establish one)
// declared: one already reused once by the middle layer for its own new
// pair, proving reuse composes across depth, not just a single hop
// (KBF007), and maps a slot declared on a grandparent attribute (KBF012).
// All of it must resolve; none of it was reachable under the old
// single-hop ExtendsRoot.
func TestChainThreeLevelValid(t *testing.T) {
	result := mustRun(t,
		"testdata/chain/grandparent",
		"testdata/chain/parent",
		"testdata/chain/child-valid",
	)
	if len(result.Findings) != 0 {
		t.Errorf("3-level valid chain produced findings, want none: %+v", result.Findings)
	}
}

// TestChainVerbNotInheritedBySibling is design.md's "not for siblings of
// another chain", and owner item (2) of the 2026-08-13 KBF007 adjudication
// ("invalid: new verb over two inherited entities fires KBF007"):
// chain-cousin extends chain-other-root, a genuinely unrelated root, not
// chain-grandparent's line at all. assembled-from is real (chain-
// grandparent declares it) and chain-grandparent/chain-parent are loaded
// in the very same universe, but chain-cousin's own chain never passes
// through either, so KBF007's ancestor union (scoped to a package's
// actual ancestors) must not see it. chain-cousin also declares no
// entities of its own — both endpoints (gadget, widget) are inherited
// from chain-other-root — so the owner-adjudicated own-entity minting
// right does not open a back door either: merely being loaded together in
// one kbf lint invocation, or redeclaring an inherited entity's name, is
// not the same as genuinely introducing something new.
func TestChainVerbNotInheritedBySibling(t *testing.T) {
	result := mustRun(t,
		"testdata/chain/grandparent",
		"testdata/chain/parent",
		"testdata/chain/other-root",
		"testdata/chain/cousin-invalid-verb",
	)
	got := find(result.Findings, lint.KBF007)
	if len(got) != 1 {
		t.Fatalf("KBF007: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if got[0].Element != "assembled-from" {
		t.Errorf("KBF007 element = %q, want assembled-from", got[0].Element)
	}
	if !strings.Contains(got[0].File, "cousin") {
		t.Errorf("KBF007 file %q, want the cousin package's relations.yaml", got[0].File)
	}
	// chain-parent and chain-grandparent are unaffected by the cousin's
	// violation: only its own file should carry a finding.
	for _, f := range result.Findings {
		if !strings.Contains(f.File, "cousin") {
			t.Errorf("unexpected finding outside the cousin package: %+v", f)
		}
	}
}

// TestChainMintOnOwnEntity is owner item (1) of the 2026-08-13 KBF007
// adjudication ("valid: base playbook mints verb on own-entity pair"),
// mirroring playbooks/core-operations minting located-at/staffed-by/
// works-at/sells on pairs that touch location or shift. chain-mint-own-
// entity extends chain-grandparent directly and mints stored-at, a verb
// no ancestor declares, on a pair that touches warehouse (its own new
// entity): must pass. The same package reuses stored-at on a second,
// fully-inherited pair (material -> site, both chain-grandparent's): must
// still fail, proving minting is evaluated per relation, not granted to
// the whole package once the verb has been used anywhere in it.
func TestChainMintOnOwnEntity(t *testing.T) {
	result := mustRun(t,
		"testdata/chain/grandparent",
		"testdata/chain/mint-own-entity",
	)
	got := find(result.Findings, lint.KBF007)
	if len(got) != 1 {
		t.Fatalf("KBF007: got %d findings, want 1 (only the fully-inherited pair): %+v", len(got), result.Findings)
	}
	if got[0].Element != "stored-at" {
		t.Errorf("KBF007 element = %q, want stored-at", got[0].Element)
	}
	if !strings.Contains(got[0].File, "mint-own-entity") {
		t.Errorf("KBF007 file %q, want the mint-own-entity package's relations.yaml", got[0].File)
	}
}

// TestChainForkAgainstGrandparent is design.md's "nearest ancestor that
// has the identity": chain-child-fork extends chain-parent, which never
// touches chain-grandparent's `material` entity, so the fork must still
// be caught, reported against chain-grandparent by name specifically
// (not "chain-parent", which never declared it, and not a generic "the
// chain").
func TestChainForkAgainstGrandparent(t *testing.T) {
	result := mustRun(t,
		"testdata/chain/grandparent",
		"testdata/chain/parent",
		"testdata/chain/child-fork",
	)
	got := find(result.Findings, lint.KBF008)
	if len(got) != 1 {
		t.Fatalf("KBF008: got %d findings, want 1: %+v", len(got), result.Findings)
	}
	if got[0].Element != "material" {
		t.Errorf("KBF008 element = %q, want material", got[0].Element)
	}
	if !strings.Contains(got[0].Message, "chain-grandparent") {
		t.Errorf("KBF008 message %q does not name chain-grandparent specifically", got[0].Message)
	}
}

// TestChainCycle is a genuine A->B->C->A cycle: must fail as KBF011, not
// hang or overflow the stack. Every package on the cycle independently
// detects it starting from itself, so all three get their own finding.
func TestChainCycle(t *testing.T) {
	result := mustRun(t,
		"testdata/chain/cycle-a",
		"testdata/chain/cycle-b",
		"testdata/chain/cycle-c",
	)
	got := find(result.Findings, lint.KBF011)
	if len(got) != 3 {
		t.Fatalf("KBF011: got %d cycle findings, want 3 (one per package on the cycle): %+v", len(got), result.Findings)
	}
	files := map[string]bool{}
	for _, f := range got {
		if !strings.Contains(f.Message, "cycle") {
			t.Errorf("KBF011 message %q does not mention a cycle", f.Message)
		}
		files[f.File] = true
	}
	if len(files) != 3 {
		t.Errorf("cycle findings span %d distinct files, want 3 (one per package): %v", len(files), files)
	}
}
