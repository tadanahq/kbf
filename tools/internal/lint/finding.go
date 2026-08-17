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

// Package lint validates a KBF package (or an extends chain of them)
// against internal/model: structural shape first, then semantic rules that
// need cross-references resolved. Rule ids are public API (design.md): once
// assigned, a rule id's meaning does not change.
package lint

import "sort"

// Rule ids. Comments are the one-line rule statement; full semantics for
// the non-obvious ones (KBF007/KBF008/KBF009/KBF010/KBF012 in particular)
// are recorded in design.md's "Implementation clarifications" because the
// capsule's original examples underspecified them.
const (
	KBF001 = "KBF001" // unknown field
	KBF002 = "KBF002" // bad (or missing) enum value, including an unrecognized `kind`
	KBF003 = "KBF003" // duplicate name in namespace (relations: duplicate (name,from,to))
	KBF004 = "KBF004" // entity without identity
	KBF005 = "KBF005" // metric without grain and/or additivity
	KBF006 = "KBF006" // relation endpoint (from/to) not a declared entity
	KBF007 = "KBF007" // relation verb outside the controlled vocabulary
	KBF008 = "KBF008" // fork of a core element (redefinition in an extending package)
	KBF009 = "KBF009" // dangling cross-reference (grain, on, expects)
	KBF010 = "KBF010" // missing governance tier (or risk, for actions)
	KBF011 = "KBF011" // manifest missing or invalid
	KBF012 = "KBF012" // attribute slot reference without a matching install/slots.yaml declaration
	KBF013 = "KBF013" // layer/taxonomy consistency: extends-target layer, name-prefix convention
)

// Finding is one rule violation. The field set and JSON keys are exactly
// design.md's documented `--format json` contract: authoring agents parse
// this shape, so it does not grow or change casually.
type Finding struct {
	Rule    string `json:"id"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Element string `json:"element"`
	Message string `json:"message"`
	Fix     string `json:"fix"`
}

// sortFindings orders findings by file, then line, then rule id: the same
// order every time regardless of map iteration or load order, so both
// renderers and any test asserting on output are deterministic. Stable,
// not just sorted: a relation's `from` and `to` can both dangle on the
// same line under the same rule (a self-relation to an undeclared
// entity), and only a stable sort keeps those in the order the checks
// actually ran instead of leaving it to sort.Slice's unspecified tie
// order.
func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Rule < b.Rule
	})
}
