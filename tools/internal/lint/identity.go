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

package lint

import (
	"reflect"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// identity.go holds the composition-relative matching shared by structural
// completeness rules (KBF004/KBF005/KBF010, which must not fire on a
// legitimate override fragment) and the semantic fork rule (KBF008, which
// decides whether a match is legitimate or a fork). See design.md's
// "Implementation clarifications": both were found-in-content requirements,
// not guesses, so they live together rather than each rule re-deriving them.

// glossaryFields names, per kind, the only field(s) an override fragment
// may set without being a fork. Kind and Name are always implicitly
// allowed (they are how the match was found in the first place). Relation
// and Action have no entry: any match is unconditionally a fork for them
// (design.md: "no glossary carve-out").
var glossaryFields = map[model.Kind][]string{
	model.KindEntity: {"Synonyms"},
	model.KindMetric: {"Thresholds"},
}

// matchable is the set of kinds that participate in identity matching at
// all. CompetencyQuestion has no Name and is never matched.
var matchable = map[model.Kind]bool{
	model.KindEntity:   true,
	model.KindRelation: true,
	model.KindMetric:   true,
	model.KindAction:   true,
}

// matchInClosure looks for elem's identity among closure's packages
// (Universe.Closure, sorted by name for determinism): (kind, name) for
// entity/metric/action, (kind, name, from, to) for relation. Returns the
// first match found, in that deterministic order, and which package it
// lives in. Composition has no "nearest ancestor" notion once more than
// one immediate parent is possible (design.md's "Implementation
// clarifications"): in the well-formed case at most one closure member
// declares a given identity anyway, so which one this function picks is
// unambiguous; if more than one does, that is itself a separate,
// explicit finding (checkCrossPlaybookCollisions, semantic.go, KBF003's
// cross-playbook variant), not something this function needs to
// adjudicate — it only needs one deterministic answer for KBF008's
// fork-vs-override check. closure is pkg's own composition (never
// including pkg itself), so there is no self-match case to guard against
// here.
func matchInClosure(elem Element, closure []*Package) (Element, *Package, bool) {
	if !matchable[elem.Kind] {
		return Element{}, nil, false
	}
	for _, member := range closure {
		for _, cand := range member.Elements {
			if cand.Kind == elem.Kind && sameIdentity(elem, cand) {
				return cand, member, true
			}
		}
	}
	return Element{}, nil, false
}

// sameIdentity compares two same-kind elements by identityKey.
func sameIdentity(a, b Element) bool {
	return identityKey(a) == identityKey(b)
}

// identityKey is an element's uniqueness/matching key within its kind: the
// (name, from, to) triple for relations (design.md: verbs recur by design,
// so bare name is not unique), plain name otherwise. Only meaningful
// between elements of the same Kind; callers compare kind separately.
func identityKey(e Element) string {
	if r, ok := e.Value.(*model.Relation); ok {
		return r.Name + "\x00" + r.From + "\x00" + r.To
	}
	return e.Name()
}

// isGlossaryOnly reports whether elem sets only Kind, Name, and its kind's
// glossary-eligible field(s); every other field must still be its Go zero
// value. Reflection, not a hand-written per-kind comparison, so adding a
// glossary-eligible field later means editing glossaryFields once, not a
// comparator per kind.
func isGlossaryOnly(elem Element) bool {
	allowed := map[string]bool{"Kind": true, "Name": true}
	for _, f := range glossaryFields[elem.Kind] {
		allowed[f] = true
	}
	rv := reflect.ValueOf(elem.Value).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		name := rt.Field(i).Name
		if allowed[name] {
			continue
		}
		if !rv.Field(i).IsZero() {
			return false
		}
	}
	return true
}
