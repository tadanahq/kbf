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
	"fmt"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// recognizedDimensions are grain tokens that are valid without being a
// declared entity: spec/primitives/metric.md, "`business-date` is the
// standard time dimension." Everything else in a metric's grain must
// resolve to an entity.
var recognizedDimensions = map[string]bool{"business-date": true}

// semanticFindings runs the rules that need cross-references resolved:
// KBF005 (metric completeness, grouped here per tasks.md's batch split,
// not structural.go's), KBF006 (relation endpoints), KBF007 (verb
// legality), KBF008 (fork detection), KBF009 (dangling references), and
// KBF012 (slot declarations without a matching attribute). chain is pkg's
// full ancestor line, nearest first (Universe.Chain); empty means pkg is a
// root package. A missing or cyclic link partway up the chain still
// leaves chain populated with whatever ancestors resolved before that
// point (Chain's own doc comment), so every check below just uses chain
// as given without re-deriving how complete it is: Run already recorded
// KBF011 for the gap itself.
func semanticFindings(pkg *Package, chain []*Package) []Finding {
	var findings []Finding

	entities := knownNames(pkg, chain, model.KindEntity)
	ancestorVerbs := controlledVocabulary(chain)
	minted := mintedEntities(pkg, chain)
	anyName := knownNames(pkg, chain, model.KindEntity, model.KindMetric, model.KindAction, model.KindRelation)

	for _, e := range pkg.Elements {
		if _, ancestor, matched := matchInChain(e, chain); matched {
			findings = append(findings, checkFork(e, ancestor)...)
			continue // KBF008 is the whole story for a matched element; see structuralFindings.
		}
		findings = append(findings, checkGrainAndVocabulary(e, ancestorVerbs, minted)...)
		findings = append(findings, checkReferences(e, entities, anyName)...)
	}

	findings = append(findings, checkSlotUsage(pkg, chain)...)
	return findings
}

// checkGrainAndVocabulary is KBF005 (metric grain/additivity presence) and
// KBF007 (relation verb legality, owner-adjudicated 2026-08-13: see
// design.md's "Implementation clarifications"). A relation's verb passes
// if EITHER an ancestor already declares it (ancestorVerbs: ordinary
// reuse), OR the relation touches an entity pkg itself mints (minted:
// introducing a new entity carries the right to name the relationships it
// participates in, without pre-clearing the verb upstream first).
// Evaluated per relation, not per package: minting a verb once does not
// blanket-legalize it for a *different*, fully-inherited pair elsewhere in
// the same package. Grain's *entries* (are they declared entities?) is a
// reference check: see checkReferences.
func checkGrainAndVocabulary(e Element, ancestorVerbs, minted map[string]bool) []Finding {
	switch v := e.Value.(type) {
	case *model.Metric:
		if len(v.Grain) == 0 || v.Additivity == "" {
			return []Finding{{Rule: KBF005, File: e.File, Line: e.Line, Element: v.Name, Message: "metric has no grain and/or additivity", Fix: "add grain: [<entity>, ...] and additivity: additive|semi-additive|non-additive"}}
		}
	case *model.Relation:
		if !ancestorVerbs[v.Name] && !minted[v.From] && !minted[v.To] {
			return []Finding{{Rule: KBF007, File: e.File, Line: e.Line, Element: v.Name, Message: fmt.Sprintf("verb %q is outside the controlled vocabulary", v.Name), Fix: "reuse a verb an ancestor already declares, mint it on a relation that touches an entity this package introduces, or add it upstream via RFC first"}}
		}
	}
	return nil
}

// checkReferences is KBF006 (relation endpoints) and KBF009 (every other
// named reference: metric grain, action on, competency-question expects).
func checkReferences(e Element, entities, anyName map[string]bool) []Finding {
	var findings []Finding
	dangling := func(rule, ref, fix string) {
		findings = append(findings, Finding{Rule: rule, File: e.File, Line: e.Line, Element: ref, Message: fmt.Sprintf("%q does not resolve to a declared element", ref), Fix: fix})
	}

	switch v := e.Value.(type) {
	case *model.Relation:
		if v.From != "" && !entities[v.From] {
			dangling(KBF006, v.From, "declare the entity, or fix the typo in from")
		}
		if v.To != "" && !entities[v.To] {
			dangling(KBF006, v.To, "declare the entity, or fix the typo in to")
		}
	case *model.Metric:
		for _, g := range v.Grain {
			if !entities[g] && !recognizedDimensions[g] {
				dangling(KBF009, g, "declare the entity, use business-date, or fix the typo in grain")
			}
		}
	case *model.Action:
		if v.On != "" && !entities[v.On] {
			dangling(KBF006, v.On, "declare the entity, or fix the typo in on")
		}
	case *model.CompetencyQuestion:
		for _, want := range v.Expects {
			if !anyName[want] {
				dangling(KBF009, want, "reference an entity, relation, metric, or action name that exists")
			}
		}
	}
	return findings
}

// checkFork is KBF008 for one element already known to match something in
// ancestor by identity: legitimate (entity/metric setting only their
// glossary-eligible field) or a fork (everything else). ancestor is
// whichever package in the chain matchInChain actually found the
// collision in, nearest one first, which is not always pkg's immediate
// parent: a grandchild forking an element only its grandparent declares
// (the parent never touched it) reports against the grandparent, not a
// generic "the chain".
func checkFork(e Element, ancestor *Package) []Finding {
	if (e.Kind == model.KindEntity || e.Kind == model.KindMetric) && isGlossaryOnly(e) {
		return nil
	}
	return []Finding{{
		Rule: KBF008, File: e.File, Line: e.Line, Element: e.Name(),
		Message: fmt.Sprintf("%s %s already exists in %s: this redeclares it instead of extending it", e.Kind, describeKey(e), ancestor.Manifest.Name),
		Fix:     glossaryFixHint(e.Kind),
	}}
}

// glossaryFixHint names the one edit KBF008's fix would accept, when there
// is one.
func glossaryFixHint(kind model.Kind) string {
	fields := glossaryFields[kind]
	if len(fields) == 0 {
		return "remove this redeclaration; add a genuinely new element instead"
	}
	return fmt.Sprintf("keep only kind, name, and %s; every other field must be absent", fields[0])
}

// checkSlotUsage is KBF012: every install/slots.yaml row must match an
// attribute's slot reference somewhere in the resolved package (pkg, plus
// every package in chain: an install configures the whole effective
// ontology, not just what pkg itself redeclares). This is the direction
// spec/primitives/slot-mapping.md documents; the reverse (an attribute
// slot with no row) is deliberately not a v0 lint error, see design.md.
func checkSlotUsage(pkg *Package, chain []*Package) []Finding {
	used := map[string]bool{}
	collect := func(p *Package) {
		for _, e := range p.Elements {
			ent, ok := e.Value.(*model.Entity)
			if !ok {
				continue
			}
			for _, attr := range ent.Attributes {
				if attr.Slot != "" {
					used[attr.Slot] = true
				}
			}
		}
	}
	collect(pkg)
	for _, ancestor := range chain {
		collect(ancestor)
	}

	var findings []Finding
	for _, row := range pkg.Slots {
		if row.Slot != "" && !used[row.Slot] {
			findings = append(findings, Finding{
				Rule: KBF012, File: pkg.SlotsFile, Line: row.Line, Element: row.Slot,
				Message: fmt.Sprintf("slot %q has no matching attribute", row.Slot),
				Fix:     "remove the row, or add slot: " + row.Slot + " to the attribute it belongs to",
			})
		}
	}
	return findings
}

// knownNames collects the bare Name() of every element of the given kinds
// in pkg, plus every package in chain (pkg's full ancestor line): a
// package's content may reference an element it inherits from any layer
// above it, not just its immediate parent.
func knownNames(pkg *Package, chain []*Package, kinds ...model.Kind) map[string]bool {
	want := make(map[model.Kind]bool, len(kinds))
	for _, k := range kinds {
		want[k] = true
	}
	names := map[string]bool{}
	collect := func(p *Package) {
		for _, e := range p.Elements {
			if want[e.Kind] {
				names[e.Name()] = true
			}
		}
	}
	collect(pkg)
	for _, ancestor := range chain {
		collect(ancestor)
	}
	return names
}

// controlledVocabulary is KBF007 condition (a): the set of relation verbs
// any ancestor in chain already declares (design.md's "Implementation
// clarifications" — a base package adding a verb legalizes it for
// everything below it in the chain, not just its immediate child). Empty
// when chain is empty: a root has no ancestors to inherit vocabulary
// from, so every one of its own relations instead passes through
// condition (b) (mintedEntities), which it always does — a root
// necessarily declares every entity it references, since it has nothing
// to inherit from.
func controlledVocabulary(chain []*Package) map[string]bool {
	verbs := map[string]bool{}
	for _, p := range chain {
		for _, e := range p.Elements {
			if e.Kind == model.KindRelation {
				verbs[e.Name()] = true
			}
		}
	}
	return verbs
}

// mintedEntities is KBF007 condition (b): the set of entity names pkg
// declares that do NOT already exist, by name, in any ancestor —
// genuinely new concepts this package introduces, as opposed to a
// glossary override or fork of an inherited entity (same name, but not
// new: KBF008 handles whether that redeclaration is itself legitimate).
// Only a genuinely new entity carries minting rights (owner adjudication,
// design.md, 2026-08-13): a relation that only touches an entity already
// declared by an ancestor gets no self-inclusion, even if pkg happens to
// redeclare that same name on its own copy.
func mintedEntities(pkg *Package, chain []*Package) map[string]bool {
	minted := map[string]bool{}
	for _, e := range pkg.Elements {
		if e.Kind == model.KindEntity {
			minted[e.Name()] = true
		}
	}
	for _, ancestor := range chain {
		for _, e := range ancestor.Elements {
			if e.Kind == model.KindEntity {
				delete(minted, e.Name())
			}
		}
	}
	return minted
}
