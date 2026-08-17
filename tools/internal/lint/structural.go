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
	"strings"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// structuralFindings runs the rules that check one package's own shape:
// KBF003 (duplicate identity) and, per element, KBF002 (enum values) and
// KBF004/KBF010 (required fields). An unrecognized `kind` is also KBF002,
// but that happens during loadPackage since it gates whether an element
// can be decoded at all; this only checks enum-shaped *fields* on an
// already-decoded element. Completeness (KBF004/KBF010, not KBF002: an
// empty enum field is merely absent, not a bad value, so it never fires on
// a fragment) is skipped for an element that matches something anywhere in
// closure (pkg's full composition, Universe.Closure): it is an override
// fragment or a fork, and either way KBF008 (semantic.go) is the rule
// that speaks to it, not a second, redundant "missing tier" complaint.
// See design.md's "Implementation clarifications".
func structuralFindings(pkg *Package, closure []*Package) []Finding {
	findings := checkDuplicateNames(pkg)

	for _, e := range pkg.Elements {
		findings = append(findings, checkEnums(e)...)
		if _, _, matched := matchInClosure(e, closure); matched {
			continue
		}
		findings = append(findings, checkCompleteness(e)...)
	}
	return findings
}

// checkEnums is KBF002 for already-decoded elements: every closed-
// vocabulary field (tier, cardinality, additivity, risk) that is
// non-empty must hold one of its kind's allowed values. Empty is not
// checked here: an absent value is KBF010 (entity/relation/metric tier,
// action risk) or simply not required (relation cardinality/additivity
// have no "missing" rule of their own in the 12-rule set), never KBF002.
func checkEnums(e Element) []Finding {
	bad := func(field, value string, allowed []string) Finding {
		return Finding{
			Rule: KBF002, File: e.File, Line: e.Line, Element: value,
			Message: fmt.Sprintf("%s %q is not one of %s", field, value, strings.Join(allowed, ", ")),
			Fix:     fmt.Sprintf("set %s to one of: %s", field, strings.Join(allowed, ", ")),
		}
	}

	var findings []Finding
	switch v := e.Value.(type) {
	case *model.Entity:
		if v.Tier != "" && !contains(model.GovernanceTiers, v.Tier) {
			findings = append(findings, bad("tier", v.Tier, model.GovernanceTiers))
		}
	case *model.Metric:
		if v.Tier != "" && !contains(model.GovernanceTiers, v.Tier) {
			findings = append(findings, bad("tier", v.Tier, model.GovernanceTiers))
		}
		if v.Additivity != "" && !contains(model.Additivities, v.Additivity) {
			findings = append(findings, bad("additivity", v.Additivity, model.Additivities))
		}
	case *model.Relation:
		if v.Tier != "" && !contains(model.RelationTiers, v.Tier) {
			findings = append(findings, bad("tier", v.Tier, model.RelationTiers))
		}
		if v.Cardinality != "" && !contains(model.Cardinalities, v.Cardinality) {
			findings = append(findings, bad("cardinality", v.Cardinality, model.Cardinalities))
		}
	case *model.Action:
		if v.Risk != "" && !contains(model.ActionRisks, v.Risk) {
			findings = append(findings, bad("risk", v.Risk, model.ActionRisks))
		}
	}
	return findings
}

// contains reports whether values holds want. The enum vocabularies in
// internal/model are all 2-3 entries, so a linear scan reads clearer than
// building a set for a one-shot check.
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

// checkManifest validates manifest.yaml itself: KBF011. A missing file was
// already reported during load (pkg.Manifest == nil); this only runs when
// there is a manifest to check the content of. Each unresolved builds-on
// entry gets its own finding, same as every other "each problem is
// independently fixable" rule in this file.
func checkManifest(pkg *Package, pkgs map[string]*Package) []Finding {
	if pkg.Manifest == nil {
		return nil
	}
	m := pkg.Manifest
	var findings []Finding
	add := func(element, msg, fix string) {
		findings = append(findings, Finding{Rule: KBF011, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: element, Message: msg, Fix: fix})
	}
	if m.Name == "" {
		add("name", "manifest is missing name", "add a package name")
	}
	if m.Version == "" {
		add("version", "manifest is missing version", "add a package version")
	}
	if m.Spec == "" {
		add("spec", "manifest is missing spec", "add the KBF spec version this package targets, e.g. v0")
	}
	for _, name := range m.BuildsOn {
		if _, ok := pkgs[name]; !ok {
			add("builds-on", fmt.Sprintf("builds-on %q does not resolve", name), "pass that playbook's path to kbf lint too")
		}
	}
	return findings
}

// checkDuplicateNames is KBF003: two elements of the same kind sharing an
// identityKey within one package. Competency questions have no name and
// are exempt (matchable[...] is false for that kind).
func checkDuplicateNames(pkg *Package) []Finding {
	seen := map[model.Kind]map[string]Element{}
	var findings []Finding
	for _, e := range pkg.Elements {
		if !matchable[e.Kind] {
			continue
		}
		if seen[e.Kind] == nil {
			seen[e.Kind] = map[string]Element{}
		}
		key := identityKey(e)
		first, dup := seen[e.Kind][key]
		if !dup {
			seen[e.Kind][key] = e
			continue
		}
		findings = append(findings, Finding{
			Rule: KBF003, File: e.File, Line: e.Line, Element: e.Name(),
			Message: fmt.Sprintf("duplicate %s %q, first declared at %s:%d", e.Kind, describeKey(e), first.File, first.Line),
			Fix:     "rename one, or remove the duplicate",
		})
	}
	return findings
}

// describeKey renders identityKey for a human: the verb alone is
// ambiguous for a relation, so name the full (name, from, to) triple.
func describeKey(e Element) string {
	if r, ok := e.Value.(*model.Relation); ok {
		return fmt.Sprintf("%s: %s -> %s", r.Name, r.From, r.To)
	}
	return e.Name()
}

// checkCompleteness is KBF004 (entity identity) and KBF010 (governance
// tier, or risk for actions): the fields every *standalone* element of
// that kind must carry.
func checkCompleteness(e Element) []Finding {
	var findings []Finding
	switch v := e.Value.(type) {
	case *model.Entity:
		if len(v.Identity) == 0 {
			findings = append(findings, Finding{Rule: KBF004, File: e.File, Line: e.Line, Element: v.Name, Message: "entity has no identity key(s)", Fix: "add identity: [<key>, ...]"})
		}
		if v.Tier == "" {
			findings = append(findings, Finding{Rule: KBF010, File: e.File, Line: e.Line, Element: v.Name, Message: "entity has no governance tier", Fix: "add tier: structural, glossary, or instance"})
		}
	case *model.Relation:
		if v.Tier == "" {
			findings = append(findings, Finding{Rule: KBF010, File: e.File, Line: e.Line, Element: v.Name, Message: "relation has no tier", Fix: "add tier: source-synced or client-configured"})
		}
	case *model.Metric:
		if v.Tier == "" {
			findings = append(findings, Finding{Rule: KBF010, File: e.File, Line: e.Line, Element: v.Name, Message: "metric has no governance tier", Fix: "add tier: structural, glossary, or instance"})
		}
	case *model.Action:
		if v.Risk == "" {
			findings = append(findings, Finding{Rule: KBF010, File: e.File, Line: e.Line, Element: v.Name, Message: "action has no risk", Fix: "add risk: auto or confirm"})
		}
	}
	return findings
}
