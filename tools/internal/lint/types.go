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
	"github.com/goccy/go-yaml/ast"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// Element is one decoded ontology document: a typed model value plus enough
// position context for rules to report file:line, and the raw mapping node
// so rules can point at a specific field rather than only the element's
// start. Value is always a pointer: *model.Entity, *model.Relation,
// *model.Metric, *model.Action, or *model.CompetencyQuestion, matching Kind.
type Element struct {
	Kind  model.Kind
	File  string
	Line  int
	Node  *ast.MappingNode
	Value any
}

// Name returns the element's Name field, or "" for CompetencyQuestion,
// which has none (see design.md).
func (e Element) Name() string {
	switch v := e.Value.(type) {
	case *model.Entity:
		return v.Name
	case *model.Relation:
		return v.Name
	case *model.Metric:
		return v.Name
	case *model.Action:
		return v.Name
	default:
		return ""
	}
}

// Package is everything loaded from one package root: its manifest, the
// ontology/evals elements, and the install slot declarations. A Package is
// still "loaded" even when parts of it are invalid; invalid parts produce
// Findings during loadPackage and are simply absent here, so later rules
// never see a half-decoded value.
type Package struct {
	Root         string
	Manifest     *model.Manifest
	ManifestFile string
	ManifestLine int
	Elements     []Element
	Slots        []SlotRow
	SlotsFile    string
}

// SlotRow is one install/slots.yaml row plus its position, so KBF012 can
// report file:line the same way every other rule does.
type SlotRow struct {
	model.SlotMapping
	Line int
}
