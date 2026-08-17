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

// Action is a verb an agent may execute against an entity: what it does,
// what it takes to let an agent do it unsupervised, and what it writes.
// Approval is Action's governance-equivalent field (project-architecture.md's
// "risk tier"); Action has no separate Tier field, see design.md.
type Action struct {
	// Kind discriminates this document as an action; always "action".
	Kind Kind `yaml:"kind" json:"kind" jsonschema:"enum=action"`
	// Name is the kebab-case identifier for this action, e.g.
	// "flag-for-review". Unique within a namespace: KBF003.
	Name string `yaml:"name" json:"name"`
	// On is the entity name this action targets. Must resolve to a
	// declared entity: KBF009.
	On string `yaml:"on" json:"on"`
	// Approval states whether an agent may take this action unsupervised
	// (automatic) or must get confirmation first (required). This is
	// Action's governance-equivalent tier: empty approval is KBF010.
	Approval string `yaml:"approval" json:"approval"`
	// Writes names what the action produces, e.g. "finding". Free text in
	// v0: the set of writable targets grows with the findings layer, not
	// something the linter should freeze early.
	Writes string `yaml:"writes" json:"writes"`
}
