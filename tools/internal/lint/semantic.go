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
// vocabulary), KBF008 (fork detection), KBF009 (dangling references), and
// KBF012 (slot declarations without a matching attribute). root is the
// resolved extends-root (see extendsRoot); nil means extends was set but
// unresolvable, already reported as KBF011, so nothing here is checked
// against wrongly-empty context.
func semanticFindings(pkg *Package, root *Package, pkgs map[string]*Package) []Finding {
	var findings []Finding

	entities := knownNames(pkg, root, model.KindEntity)
	vocabulary := controlledVocabulary(root)
	anyName := knownNames(pkg, root, model.KindEntity, model.KindMetric, model.KindAction, model.KindRelation)

	child := isChildPackage(pkg, root)
	for _, e := range pkg.Elements {
		if child {
			if _, matched := matchExtendsRoot(e, root); matched {
				findings = append(findings, checkFork(e, root)...)
				continue // KBF008 is the whole story for a matched element; see structuralFindings.
			}
		}
		findings = append(findings, checkGrainAndVocabulary(e, vocabulary)...)
		findings = append(findings, checkReferences(e, entities, anyName)...)
	}

	findings = append(findings, checkSlotUsage(pkg, root)...)
	return findings
}

// checkGrainAndVocabulary is KBF005 (metric grain/additivity presence) and
// KBF007 (relation verb in the controlled vocabulary). Grain's *entries*
// (are they declared entities?) is a reference check: see checkReferences.
func checkGrainAndVocabulary(e Element, vocabulary map[string]bool) []Finding {
	switch v := e.Value.(type) {
	case *model.Metric:
		if len(v.Grain) == 0 || v.Additivity == "" {
			return []Finding{{Rule: KBF005, File: e.File, Line: e.Line, Element: v.Name, Message: "metric has no grain and/or additivity", Fix: "add grain: [<entity>, ...] and additivity: additive|semi-additive|non-additive"}}
		}
	case *model.Relation:
		if !vocabulary[v.Name] {
			return []Finding{{Rule: KBF007, File: e.File, Line: e.Line, Element: v.Name, Message: fmt.Sprintf("verb %q is outside the controlled vocabulary", v.Name), Fix: "use an existing verb, or add it to universal-core via RFC first"}}
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
// root by identity: legitimate (entity/metric setting only their
// glossary-eligible field) or a fork (everything else).
func checkFork(e Element, root *Package) []Finding {
	if (e.Kind == model.KindEntity || e.Kind == model.KindMetric) && isGlossaryOnly(e) {
		return nil
	}
	return []Finding{{
		Rule: KBF008, File: e.File, Line: e.Line, Element: e.Name(),
		Message: fmt.Sprintf("%s %s already exists in %s: this redeclares it instead of extending it", e.Kind, describeKey(e), root.Manifest.Name),
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
// root when pkg is a child: an install configures the whole effective
// ontology, not just what the child itself redeclares). This is the
// direction spec/primitives/slot-mapping.md documents; the reverse (an
// attribute slot with no row) is deliberately not a v0 lint error, see
// design.md.
func checkSlotUsage(pkg *Package, root *Package) []Finding {
	used := map[string]bool{}
	collect := func(p *Package) {
		if p == nil {
			return
		}
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
	if isChildPackage(pkg, root) {
		collect(root)
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
// in pkg, plus root's too when pkg is a genuine child (a child's content
// may reference an element it inherits rather than redeclares).
func knownNames(pkg, root *Package, kinds ...model.Kind) map[string]bool {
	want := make(map[model.Kind]bool, len(kinds))
	for _, k := range kinds {
		want[k] = true
	}
	names := map[string]bool{}
	collect := func(p *Package) {
		if p == nil {
			return
		}
		for _, e := range p.Elements {
			if want[e.Kind] {
				names[e.Name()] = true
			}
		}
	}
	collect(pkg)
	if isChildPackage(pkg, root) {
		collect(root)
	}
	return names
}

// controlledVocabulary is the set of distinct relation verbs declared in
// root: design.md's "Implementation clarifications" (KBF007). Passing pkg
// itself as root (extendsRoot's result for a package with no parent) means
// "self", matching "or self when linting universal-core itself".
func controlledVocabulary(root *Package) map[string]bool {
	verbs := map[string]bool{}
	if root == nil {
		return verbs
	}
	for _, e := range root.Elements {
		if e.Kind == model.KindRelation {
			verbs[e.Name()] = true
		}
	}
	return verbs
}
