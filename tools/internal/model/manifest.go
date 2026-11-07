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

package model

// Manifest is a playbook's identity card: manifest.yaml at the playbook root.
// It names the playbook, the spec version it targets, and what it builds on.
// Every playbook has exactly one manifest; the linter reads it before
// touching any ontology content, since builds-on drives cross-playbook
// resolution (KBF006, KBF007, KBF008).
type Manifest struct {
	// Name is the playbook's identifier, e.g. "core-business". Doubles as
	// its namespace: element uniqueness (KBF003) and fork detection
	// (KBF008) are scoped by this name.
	Name string `yaml:"name" json:"name"`
	// Version is the playbook's own version, independent of the spec
	// version it targets: a playbook can change while the spec stands
	// still, and the other way around.
	Version string `yaml:"version" json:"version"`
	// Spec is the KBF spec version this playbook's content was authored
	// against, e.g. "v0".
	Spec string `yaml:"spec" json:"spec"`
	// BuildsOn lists the playbook(s) this one composes: zero or more
	// parent names, resolved as a DAG closure (transitive, deduped by
	// name), not a single linear chain. Two entries can legitimately
	// share a deeper ancestor (the diamond case: two core playbooks both
	// building on the same root, then a third composing both of them)
	// without that ancestor being loaded twice. Empty means this playbook
	// builds on nothing: for layer: core that is what makes it a root
	// (root-ness is derived from an empty BuildsOn, never declared as its
	// own value); layer: vertical always needs at least one entry
	// (KBF013). Every name must resolve, either within the set of
	// playbooks passed to `kbf lint` or from the core playbooks embedded
	// in kbf when a name isn't passed at all (a local path of the same
	// name always wins over the embedded copy), and the closure must not
	// cycle: both KBF011.
	BuildsOn []string `yaml:"builds-on" json:"builds-on"`
	// Layer states this playbook's place in the taxonomy: core (a
	// foundation playbook other playbooks compose; builds-on only core
	// playbooks, empty allowed) or vertical (a business-specific leaf;
	// builds-on at least one core or vertical playbook, so a vertical may
	// itself be composed by another vertical). This is what makes "core
	// playbook" a checkable claim instead of a naming convention nobody
	// enforces: KBF013 cross-checks Layer against BuildsOn and against
	// Name (core playbooks are named "core-...", vertical playbooks are
	// not). Layer is a governance-taxonomy field the same way an entity's
	// or metric's tier is: an empty layer is KBF010, a value other than
	// core or vertical is KBF002, and a valid value that still doesn't
	// line up with builds-on or name is KBF013.
	Layer string `yaml:"layer" json:"layer"`
}
