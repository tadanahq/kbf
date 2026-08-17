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

// Package model holds the canonical structs for the KBF meta-model: the
// seven primitives an ontology is authored from. This package is the single
// source of truth. schema/*.yaml is generated from it (internal/schemagen);
// the loader (internal/lint) decodes YAML into these same types, so a field
// added here is a field the schema publishes and the linter understands,
// with no second place to update.
//
// Every exported field carries a doc comment stating why the field exists,
// not just its shape: those comments double as the generated JSON Schema's
// `description` (see internal/schemagen), so they are read by editors and
// authoring agents, not only by people reading this source file.
package model

// Kind discriminates which primitive a YAML document (or list entry)
// decodes as. It is the first thing the loader reads before picking a target
// struct, and the only field every element kind shares.
type Kind string

// The five element kinds an ontology document can declare. Slot mapping and
// namespace are operational primitives too, but they are not `kind:`-tagged
// documents: slot mapping is a plain row list (SlotMapping) and namespace is
// derived from the owning package's manifest name, never authored directly.
const (
	KindEntity             Kind = "entity"
	KindRelation           Kind = "relation"
	KindMetric             Kind = "metric"
	KindAction             Kind = "action"
	KindCompetencyQuestion Kind = "competency-question"
)

// GovernanceTiers is the default governance-tier vocabulary from
// project-standards.md, used by Entity.Tier and Metric.Tier. Relation.Tier
// and Action.Risk deliberately use their own, smaller vocabularies: see
// RelationTiers and ActionRisks.
var GovernanceTiers = []string{"structural", "glossary", "instance"}

// RelationTiers is Relation.Tier's vocabulary: who populates the relation,
// not who governs changes to its definition. Distinct from GovernanceTiers
// on purpose; see design.md's "Implementation clarifications".
var RelationTiers = []string{"source-synced", "client-configured"}

// ActionRisks is Action.Risk's vocabulary: Action has no separate governance
// tier field because risk plays that role (project-architecture.md calls it
// "risk tier").
var ActionRisks = []string{"auto", "confirm"}

// Cardinalities is Relation.Cardinality's vocabulary. Four values, not
// three: design.md's original example listed only one-to-one/one-to-many/
// many-to-many, but from/to are directional and fixed by the relation, so
// "many locations belong-to one organization" is genuinely many-to-one,
// not restatable as one-to-many without reversing from/to. Found dogfooding
// against playbooks/core-universal, which uses many-to-one 7 times for
// exactly this shape (child-to-parent relations); see design.md.
var Cardinalities = []string{"one-to-one", "one-to-many", "many-to-one", "many-to-many"}

// Additivities is Metric.Additivity's vocabulary.
var Additivities = []string{"additive", "semi-additive", "non-additive"}

// Layers is Manifest.Layer's vocabulary: core (a foundation playbook other
// playbooks compose; may have an empty BuildsOn, in which case it is a
// root — root-ness is derived, never its own layer value) or vertical (a
// business-specific leaf; always composes at least one other playbook,
// core or vertical). KBF013 (internal/lint/taxonomy.go) checks a
// playbook's Layer for consistency against its BuildsOn and Name, not
// just against this vocabulary; a bad-but-nonempty value is KBF002, same
// treatment as every other closed-vocabulary field.
var Layers = []string{"core", "vertical"}
