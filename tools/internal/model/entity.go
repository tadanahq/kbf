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

// Entity is a thing the business tracks with identity: an order, a
// customer, a shift. It is the anchor other primitives point at (relations
// join entities, metrics grain by them, actions target them), so its
// identity and meaning must be unambiguous before anything else is checked.
type Entity struct {
	// Kind discriminates this document as an entity; always "entity".
	Kind Kind `yaml:"kind" json:"kind" jsonschema:"enum=entity"`
	// Name is the kebab-case identifier other primitives reference by
	// (relation from/to, metric grain, action on). Unique within a namespace:
	// KBF003.
	Name string `yaml:"name" json:"name"`
	// Meaning is the one-sentence definition an agent grounds itself in
	// before acting; required because a field name alone is not a contract.
	Meaning string `yaml:"meaning" json:"meaning"`
	// Identity lists the key(s) that resolve a real-world instance to one
	// row. Empty identity means the entity cannot be resolved: KBF004.
	Identity []string `yaml:"identity" json:"identity"`
	// Resolution names the strategy that turns raw source keys into the
	// identity above. Opaque free text in v0 (see design.md); named,
	// checkable strategies are gate-era work, not a v0 concern.
	Resolution string `yaml:"resolution" json:"resolution"`
	// Tier is this entity's governance tier: structural, glossary, or
	// instance. Empty tier is KBF010; a non-empty but unrecognized value is
	// KBF002.
	Tier string `yaml:"tier" json:"tier"`
	// Synonyms maps a locale to alternate names an agent might see in
	// natural language. Glossary-tier: an extending package may add
	// synonyms to a core entity without forking it (KBF008).
	Synonyms map[string][]string `yaml:"synonyms,omitempty" json:"synonyms,omitempty"`
	// Attributes are the entity's typed fields, each declaring the source
	// slot it fills from. Optional: an entity can exist before its
	// attributes are fully mapped.
	Attributes []Attribute `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	// States lists the lifecycle values this entity can hold, when it has
	// real state (an order has states; a supplier-purchase line may not).
	// Optional by design.
	States []string `yaml:"states,omitempty" json:"states,omitempty"`
}

// Attribute is one typed field on an entity, sourced from a system slot.
// A separate struct (not inlined) because schemagen needs it as its own
// named definition, and both Entity and future package content reference it.
type Attribute struct {
	// Name is the attribute's identifier within its owning entity.
	Name string `yaml:"name" json:"name"`
	// Type is the attribute's data type: text, number, boolean, date and
	// similar are expected. Free text in v0: design.md does not mark this
	// field as a closed vocabulary, unlike cardinality/additivity/risk/tier.
	Type string `yaml:"type" json:"type"`
	// Slot is the source-system slot this attribute fills from, e.g.
	// "pos.catalog". Must be declared in the package's install/slots.yaml:
	// KBF012. Empty means intentionally unmapped.
	Slot string `yaml:"slot,omitempty" json:"slot,omitempty"`
}
