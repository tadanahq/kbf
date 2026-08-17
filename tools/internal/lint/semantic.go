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
	"sort"
	"strings"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// recognizedDimensions are grain tokens that are valid without being a
// declared entity: spec/primitives/metric.md, "`business-date` is the
// standard time dimension." Everything else in a metric's grain must
// resolve to an entity.
var recognizedDimensions = map[string]bool{"business-date": true}

// semanticFindings runs the rules that need cross-references resolved:
// KBF003's cross-playbook variant (two different closure members
// declaring the same identity), KBF005 (metric completeness, grouped
// here per tasks.md's batch split, not structural.go's), KBF006
// (relation endpoints), KBF007 (verb legality), KBF008 (fork detection),
// KBF009 (dangling references), and KBF012 (slot declarations without a
// matching attribute). closure is pkg's full composition (Universe.
// Closure); empty means pkg builds on nothing. A missing or cyclic link
// partway through the closure still leaves it populated with whatever
// resolved before that point (Closure's own doc comment), so every check
// below just uses closure as given without re-deriving how complete it
// is: Run already recorded KBF011 for the gap itself.
func semanticFindings(pkg *Package, closure []*Package) []Finding {
	var findings []Finding

	entities := knownNames(pkg, closure, model.KindEntity)
	composedVerbs := controlledVocabulary(closure)
	minted := mintedEntities(pkg, closure)
	anyName := knownNames(pkg, closure, model.KindEntity, model.KindMetric, model.KindAction, model.KindRelation)

	findings = append(findings, checkCrossPlaybookCollisions(closure)...)

	for _, e := range pkg.Elements {
		if _, member, matched := matchInClosure(e, closure); matched {
			findings = append(findings, checkFork(e, member)...)
			continue // KBF008 is the whole story for a matched element; see structuralFindings.
		}
		findings = append(findings, checkGrainAndVocabulary(e, composedVerbs, minted)...)
		findings = append(findings, checkReferences(e, entities, anyName)...)
	}

	findings = append(findings, checkSlotUsage(pkg, closure)...)
	return findings
}

// checkCrossPlaybookCollisions is KBF003's cross-playbook variant: two
// DIFFERENT playbooks in pkg's own composition closure declaring the
// same element identity. Composition has no resolution order — unlike
// the old single-parent chain, where "nearest ancestor wins" gave every
// name exactly one meaning, a DAG closure has no such tie-break once
// more than one immediate parent exists — so an identity declared by
// more than one closure member is never resolvable and is always an
// error, whether or not pkg itself also touches that identity (that
// separate case, pkg redeclaring something in its closure, is KBF008).
// Symmetric: every closure member that participates in a given collision
// gets its own finding in its own file, the same way a composition cycle
// reports against every package on the cycle, not just one.
func checkCrossPlaybookCollisions(closure []*Package) []Finding {
	type occurrence struct {
		pkg  *Package
		elem Element
	}
	byIdentity := map[string][]occurrence{}
	for _, member := range closure {
		for _, e := range member.Elements {
			if !matchable[e.Kind] {
				continue
			}
			key := string(e.Kind) + "\x00" + identityKey(e)
			byIdentity[key] = append(byIdentity[key], occurrence{member, e})
		}
	}

	var findings []Finding
	for _, occs := range byIdentity {
		involved := map[string]bool{}
		for _, o := range occs {
			involved[o.pkg.Manifest.Name] = true
		}
		if len(involved) < 2 {
			continue // one playbook declaring itself twice is ordinary KBF003, checkDuplicateNames' job
		}
		for _, o := range occs {
			var others []string
			for name := range involved {
				if name != o.pkg.Manifest.Name {
					others = append(others, name)
				}
			}
			sort.Strings(others)
			findings = append(findings, Finding{
				Rule: KBF003, File: o.elem.File, Line: o.elem.Line, Element: o.elem.Name(),
				Message: fmt.Sprintf("%s %s is also declared by %s: composition has no resolution order, so identity must be unique across the closure", o.elem.Kind, describeKey(o.elem), strings.Join(others, ", ")),
				Fix:     "rename one, or remove the duplicate from whichever playbook shouldn't own it",
			})
		}
	}
	return findings
}

// checkGrainAndVocabulary is KBF005 (metric grain/additivity presence) and
// KBF007 (relation verb legality, owner-adjudicated 2026-08-13: see
// design.md's "Implementation clarifications"). A relation's verb passes
// if EITHER a closure member already declares it (composedVerbs: ordinary
// reuse), OR the relation touches an entity pkg itself mints (minted:
// introducing a new entity carries the right to name the relationships it
// participates in, without pre-clearing the verb upstream first).
// Evaluated per relation, not per package: minting a verb once does not
// blanket-legalize it for a *different*, fully-inherited pair elsewhere in
// the same package. Grain's *entries* (are they declared entities?) is a
// reference check: see checkReferences.
func checkGrainAndVocabulary(e Element, composedVerbs, minted map[string]bool) []Finding {
	switch v := e.Value.(type) {
	case *model.Metric:
		if len(v.Grain) == 0 || v.Additivity == "" {
			return []Finding{{Rule: KBF005, File: e.File, Line: e.Line, Element: v.Name, Message: "metric has no grain and/or additivity", Fix: "add grain: [<entity>, ...] and additivity: additive|semi-additive|non-additive"}}
		}
	case *model.Relation:
		if !composedVerbs[v.Name] && !minted[v.From] && !minted[v.To] {
			return []Finding{{Rule: KBF007, File: e.File, Line: e.Line, Element: v.Name, Message: fmt.Sprintf("verb %q is outside the controlled vocabulary", v.Name), Fix: "reuse a verb a composed playbook already declares, mint it on a relation that touches an entity this package introduces, or add it upstream via RFC first"}}
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
// member by identity: legitimate (entity/metric setting only their
// glossary-eligible field) or a fork (everything else). member is
// whichever package in the closure matchInClosure actually found the
// collision in — composition has no "nearest ancestor" to prefer, so when
// exactly one closure member declares the identity (the well-formed
// case) this is simply that member; when more than one does, that is a
// separate, explicit finding (checkCrossPlaybookCollisions above).
func checkFork(e Element, member *Package) []Finding {
	if (e.Kind == model.KindEntity || e.Kind == model.KindMetric) && isGlossaryOnly(e) {
		return nil
	}
	return []Finding{{
		Rule: KBF008, File: e.File, Line: e.Line, Element: e.Name(),
		Message: fmt.Sprintf("%s %s already exists in %s: this redeclares it instead of composing it", e.Kind, describeKey(e), member.Manifest.Name),
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
// every package in closure: an install configures the whole effective
// ontology, not just what pkg itself redeclares). This is the direction
// spec/primitives/slot-mapping.md documents; the reverse (an attribute
// slot with no row) is deliberately not a v0 lint error, see design.md.
func checkSlotUsage(pkg *Package, closure []*Package) []Finding {
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
	for _, member := range closure {
		collect(member)
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
// in pkg, plus every package in closure (pkg's full composition): a
// package's content may reference an element from anywhere in its
// closure, not just an immediate parent.
func knownNames(pkg *Package, closure []*Package, kinds ...model.Kind) map[string]bool {
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
	for _, member := range closure {
		collect(member)
	}
	return names
}

// controlledVocabulary is KBF007 condition (a): the set of relation verbs
// any member of closure already declares (design.md's "Implementation
// clarifications" — a composed playbook adding a verb legalizes it for
// everything that composes it, directly or transitively, not just an
// immediate child). Empty when closure is empty: a root has nothing to
// compose vocabulary from, so every one of its own relations instead
// passes through condition (b) (mintedEntities), which it always does —
// a root necessarily declares every entity it references, since it
// builds on nothing.
func controlledVocabulary(closure []*Package) map[string]bool {
	verbs := map[string]bool{}
	for _, member := range closure {
		for _, e := range member.Elements {
			if e.Kind == model.KindRelation {
				verbs[e.Name()] = true
			}
		}
	}
	return verbs
}

// mintedEntities is KBF007 condition (b): the set of entity names pkg
// declares that do NOT already exist, by name, anywhere in its closure —
// genuinely new concepts this package introduces, as opposed to a
// glossary override or fork of a composed entity (same name, but not
// new: KBF008 handles whether that redeclaration is itself legitimate).
// Only a genuinely new entity carries minting rights (owner adjudication,
// design.md, 2026-08-13): a relation that only touches an entity already
// declared somewhere in the closure gets no self-inclusion, even if pkg
// happens to redeclare that same name on its own copy.
func mintedEntities(pkg *Package, closure []*Package) map[string]bool {
	minted := map[string]bool{}
	for _, e := range pkg.Elements {
		if e.Kind == model.KindEntity {
			minted[e.Name()] = true
		}
	}
	for _, member := range closure {
		for _, e := range member.Elements {
			if e.Kind == model.KindEntity {
				delete(minted, e.Name())
			}
		}
	}
	return minted
}
