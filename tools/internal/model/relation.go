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

// Relation is a directed, verb-named link between two entities: how they
// connect, at what cardinality, and through which join keys. The verb
// itself (Name) is drawn from a controlled vocabulary so agents can reason
// over relations without learning a new verb per package: see KBF007 and
// design.md's "Implementation clarifications" for where that vocabulary
// lives.
type Relation struct {
	// Kind discriminates this document as a relation; always "relation".
	Kind Kind `yaml:"kind" json:"kind" jsonschema:"enum=relation"`
	// Name is the relation's verb, e.g. "contains" or "staffed-by". Doubles
	// as the controlled-vocabulary token: KBF007 checks it against the set
	// of verbs already declared somewhere in the composition closure.
	Name string `yaml:"name" json:"name"`
	// From is the source entity's name. Must resolve within the package or
	// its composition closure: KBF006.
	From string `yaml:"from" json:"from"`
	// To is the target entity's name. Must resolve within the package or
	// its composition closure: KBF006.
	To string `yaml:"to" json:"to"`
	// Cardinality states the shape of the join, needed before any query
	// generation can be trusted to aggregate correctly: one-to-one,
	// one-to-many, or many-to-many.
	Cardinality string `yaml:"cardinality" json:"cardinality"`
	// Join lists the key(s) that implement the relation at the data layer.
	Join []string `yaml:"join" json:"join"`
	// Origin states who populates this relation: source-synced (from a
	// source system) or client-configured (per client install). This is
	// NOT the governance tier vocabulary (structural/glossary/instance)
	// that Entity and Metric use; see design.md's "Implementation
	// clarifications".
	Origin string `yaml:"origin" json:"origin"`
	// Temporal marks whether the relation carries a validity window (the
	// join can change over time and old rows still matter). Defaults to
	// false: most relations are point-in-time current-state only.
	Temporal bool `yaml:"temporal,omitempty" json:"temporal,omitempty"`
}
