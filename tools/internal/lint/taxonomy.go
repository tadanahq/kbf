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

// corePrefix is the name prefix KBF013 requires of root/base playbooks and
// forbids on vertical ones: conventions.md's "core playbooks" category.
const corePrefix = "core-"

// extendsLayers is the set of ancestor layers each non-root Manifest.Layer
// value may legally extend. Both base and vertical currently accept the
// same set (root or base); kept as a map, not a single shared constant,
// so the rule stays easy to diverge later if a narrower policy (e.g.
// "vertical may only extend base") is ever adopted.
var extendsLayers = map[string][]string{
	"base":     {"root", "base"},
	"vertical": {"root", "base"},
}

// checkTaxonomy is KBF013: a playbook's Layer must be consistent with what
// it Extends and how it is Name-d, not just a self-declared label. Layer's
// own value is validated first, the same way every other closed-vocabulary
// field is (checkCompleteness/checkEnums): empty is KBF010, a non-empty
// value outside model.Layers is KBF002. KBF013 itself (checkExtendsLayer,
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
			Fix:     "add layer: root, base, or vertical",
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
	findings = append(findings, checkExtendsLayer(pkg, m, pkgs)...)
	findings = append(findings, checkNamePrefix(pkg, m)...)
	return findings
}

// checkExtendsLayer is KBF013's extends-target rule: layer root must have
// extends: null; layer base or vertical must extend a playbook whose own
// Layer is root or base. Whether the extends name resolves at all is
// KBF011's job (checkManifest, runs first); this only compares layers
// once a parent is actually resolvable, so a dangling extends is never
// double-reported under two different rule ids. A parent that resolves
// but has an empty or invalid Layer of its own still fails the comparison
// here (its own Layer problem is reported separately, in its own file, by
// this same function running on that package).
func checkExtendsLayer(pkg *Package, m *model.Manifest, pkgs map[string]*Package) []Finding {
	hasExtends := m.Extends != nil && *m.Extends != ""

	if m.Layer == "root" {
		if hasExtends {
			return []Finding{taxonomyFinding(pkg, "extends",
				fmt.Sprintf("layer root must have extends: null, got extends: %s", *m.Extends),
				"set extends: null, or change layer to base or vertical")}
		}
		return nil
	}

	if !hasExtends {
		return []Finding{taxonomyFinding(pkg, "extends",
			fmt.Sprintf("layer %s must extend a root or base playbook", m.Layer),
			"add extends: <parent-name>, or change layer to root")}
	}
	parent, ok := pkgs[*m.Extends]
	if !ok || parent.Manifest == nil {
		return nil // KBF011 already reported the unresolved extends
	}
	if !contains(extendsLayers[m.Layer], parent.Manifest.Layer) {
		return []Finding{taxonomyFinding(pkg, "extends",
			fmt.Sprintf("layer %s must extend a root or base playbook, but %q has layer %q", m.Layer, parent.Manifest.Name, parent.Manifest.Layer),
			"extend a root or base playbook instead, or fix the parent's layer")}
	}
	return nil
}

// checkNamePrefix is KBF013's name-prefix rule: root and base playbooks
// (together, "core playbooks": spec/conventions.md) are named "core-...";
// vertical playbooks are not. The prefix is how a reader tells the
// foundation layer from the business-specific one at a glance, so the
// linter holds the naming convention to the same account as the
// extends-shape rule above, not just documents it.
func checkNamePrefix(pkg *Package, m *model.Manifest) []Finding {
	hasPrefix := strings.HasPrefix(m.Name, corePrefix)
	switch {
	case m.Layer != "vertical" && !hasPrefix:
		return []Finding{taxonomyFinding(pkg, "name",
			fmt.Sprintf("layer %s requires a name starting with %q, got %q", m.Layer, corePrefix, m.Name),
			fmt.Sprintf("rename to %s%s, or change layer to vertical", corePrefix, m.Name))}
	case m.Layer == "vertical" && hasPrefix:
		return []Finding{taxonomyFinding(pkg, "name",
			fmt.Sprintf("layer vertical must not have a name starting with %q, got %q", corePrefix, m.Name),
			"rename without the core- prefix, or change layer to root or base")}
	}
	return nil
}

// taxonomyFinding builds one KBF013 finding at pkg's manifest position.
func taxonomyFinding(pkg *Package, element, msg, fix string) Finding {
	return Finding{Rule: KBF013, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: element, Message: msg, Fix: fix}
}
