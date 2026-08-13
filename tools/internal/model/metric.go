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

// Metric is a named, computable number: a formula at a grain, with
// declared additivity so an agent (or a downstream aggregation) never sums
// something that should have been averaged. Thresholds are glossary-tier by
// definition: a business owner can retune them without forking the metric.
type Metric struct {
	// Kind discriminates this document as a metric; always "metric".
	Kind Kind `yaml:"kind" json:"kind" jsonschema:"enum=metric"`
	// Name is the kebab-case identifier other primitives reference by
	// (competency-question expects). Unique within a namespace: KBF003.
	Name string `yaml:"name" json:"name"`
	// Formula is the computation, kept as an opaque expression string in
	// v0: parsing and validating expression semantics is gate-era work,
	// out of scope for config-phase tooling.
	Formula string `yaml:"formula" json:"formula"`
	// Grain lists the entity name(s) this metric is computed per. Empty
	// grain makes the metric un-aggregatable: KBF005. Each entry must
	// resolve to a declared entity: KBF009.
	Grain []string `yaml:"grain" json:"grain"`
	// Additivity states whether the metric can be summed across its grain,
	// summed only across some dimensions, or never summed: additive,
	// semi-additive, or non-additive. Required alongside Grain: KBF005.
	Additivity string `yaml:"additivity" json:"additivity"`
	// Unit is the metric's unit of measure (ratio, currency, count, days,
	// ...). Free text in v0: units are too numerous for a closed vocabulary.
	Unit string `yaml:"unit" json:"unit"`
	// Thresholds maps a threshold name to its value, e.g. warn-below: 0.60.
	// Glossary-tier: an extending package may retune thresholds on a core
	// metric without forking it (KBF008).
	Thresholds map[string]float64 `yaml:"thresholds,omitempty" json:"thresholds,omitempty"`
	// Tier is this metric's governance tier: structural, glossary, or
	// instance. Empty tier is KBF010; a non-empty but unrecognized value is
	// KBF002.
	Tier string `yaml:"tier" json:"tier"`
}
