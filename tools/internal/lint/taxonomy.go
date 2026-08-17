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

// corePrefix is the name prefix KBF013 requires of core playbooks and
// forbids on vertical ones: conventions.md's "core playbooks" category.
const corePrefix = "core-"

// buildsOnLayers is the set of playbook layers each Manifest.Layer value
// may legally build on. core may only build on other core playbooks: the
// foundation stays a closed, composable set, so nothing business-specific
// ever ends up load-bearing for something that claims to be universal.
// vertical may build on a core or another vertical: hybrids (a vertical
// composing two core playbooks) and layered verticals (a vertical
// composing another vertical) are both legal composition shapes.
var buildsOnLayers = map[string][]string{
	"core":     {"core"},
	"vertical": {"core", "vertical"},
}

// checkTaxonomy is KBF013: a playbook's Layer must be consistent with what
// it BuildsOn and how it is Name-d, not just a self-declared label. Layer's
// own value is validated first, the same way every other closed-vocabulary
// field is (checkCompleteness/checkEnums): empty is KBF010, a non-empty
// value outside model.Layers is KBF002. KBF013 itself (checkBuildsOnLayer,
// checkNamePrefix) only runs once Layer is a known-good value, since
// there is nothing coherent to cross-check an unrecognized layer against.
func checkTaxonomy(pkg *Package, pkgs map[string]*Package) []Finding {
	if pkg.Manifest == nil {
		return nil
	}
	m := pkg.Manifest

	if m.Layer == "" {
		return []Finding{{
			Rule: KBF010, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: "layer",
			Message: "manifest has no layer",
			Fix:     "add layer: core or vertical",
		}}
	}
	if !contains(model.Layers, m.Layer) {
		return []Finding{{
			Rule: KBF002, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: m.Layer,
			Message: fmt.Sprintf("layer %q is not one of %s", m.Layer, strings.Join(model.Layers, ", ")),
			Fix:     "set layer to one of: " + strings.Join(model.Layers, ", "),
		}}
	}

	var findings []Finding
	findings = append(findings, checkBuildsOnLayer(pkg, m, pkgs)...)
	findings = append(findings, checkNamePrefix(pkg, m)...)
	return findings
}

// checkBuildsOnLayer is KBF013's composition-target rule: a core playbook
// may have an empty BuildsOn (that is what makes it a root — root-ness is
// derived, never declared) or may build on other core playbooks only; a
// vertical playbook must build on at least one playbook, and every entry
// must be core or vertical (any resolved entry always satisfies that
// last part in practice, since core/vertical are the only two Layer
// values — the check still matters when an entry's own Layer is empty or
// invalid, which this reports on independently of that entry's own
// KBF010/KBF002). Whether each name resolves at all is KBF011's job
// (checkManifest, runs first); this only compares layers once a target
// is actually resolvable, so a dangling builds-on entry is never
// double-reported under two different rule ids. Each violating entry
// gets its own finding, same as every other "each problem is
// independently fixable" rule in this package.
func checkBuildsOnLayer(pkg *Package, m *model.Manifest, pkgs map[string]*Package) []Finding {
	if len(m.BuildsOn) == 0 {
		if m.Layer == "vertical" {
			return []Finding{taxonomyFinding(pkg, "builds-on",
				"layer vertical must build on at least one playbook",
				"add builds-on: [<playbook-name>, ...], or change layer to core")}
		}
		return nil // layer: core with no builds-on is a valid root.
	}

	var findings []Finding
	allowed := buildsOnLayers[m.Layer]
	for _, name := range m.BuildsOn {
		parent, ok := pkgs[name]
		if !ok || parent.Manifest == nil {
			continue // KBF011 already reported the unresolved builds-on entry
		}
		if !contains(allowed, parent.Manifest.Layer) {
			hint := "a vertical playbook may only build on core or vertical playbooks"
			if m.Layer == "core" {
				hint = "a core playbook may only build on core playbooks"
			}
			findings = append(findings, taxonomyFinding(pkg, "builds-on",
				fmt.Sprintf("layer %s builds on %q, whose layer is %q: %s", m.Layer, name, parent.Manifest.Layer, hint),
				"build on a playbook with an allowed layer instead, or fix that playbook's own layer"))
		}
	}
	return findings
}

// checkNamePrefix is KBF013's name-prefix rule: core playbooks (the
// foundation every vertical composes: spec/conventions.md) are named
// "core-..."; vertical playbooks are not.
func checkNamePrefix(pkg *Package, m *model.Manifest) []Finding {
	hasPrefix := strings.HasPrefix(m.Name, corePrefix)
	switch {
	case m.Layer == "core" && !hasPrefix:
		return []Finding{taxonomyFinding(pkg, "name",
			fmt.Sprintf("layer core requires a name starting with %q, got %q", corePrefix, m.Name),
			fmt.Sprintf("rename to %s%s, or change layer to vertical", corePrefix, m.Name))}
	case m.Layer == "vertical" && hasPrefix:
		return []Finding{taxonomyFinding(pkg, "name",
			fmt.Sprintf("layer vertical must not have a name starting with %q, got %q", corePrefix, m.Name),
			"rename without the core- prefix, or change layer to core")}
	}
	return nil
}

// taxonomyFinding builds one KBF013 finding at pkg's manifest position.
func taxonomyFinding(pkg *Package, element, msg, fix string) Finding {
	return Finding{Rule: KBF013, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: element, Message: msg, Fix: fix}
}
