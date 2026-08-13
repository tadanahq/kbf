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

// CompetencyQuestion is the acceptance test for an ontology: a question a
// business owner would ask, plus the elements the answer must draw on. It
// has no Name (nothing else references a competency question by name) and
// no governance Tier: it is an operational element, not owned content.
type CompetencyQuestion struct {
	// Kind discriminates this document as a competency question; always
	// "competency-question".
	Kind Kind `yaml:"kind" json:"kind" jsonschema:"enum=competency-question"`
	// Question is the natural-language question a business owner would
	// actually ask.
	Question string `yaml:"question" json:"question"`
	// Expects lists the element name(s) (metrics, entities) the answer
	// must use. Each entry must resolve to a declared element: KBF009.
	Expects []string `yaml:"expects" json:"expects"`
}
