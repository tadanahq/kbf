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

// Package coverage computes static slot-mapping completeness (design.md:
// "table of slots by entity: declared / mapped / unmapped"). It builds on
// internal/lint's loader (lint.Load) rather than parsing YAML itself:
// internal/lint owns loading (project-architecture.md).
package coverage

import (
	"sort"

	"github.com/tadanahq/kbf/tools/internal/lint"
	"github.com/tadanahq/kbf/tools/internal/model"
)

// Row is one install/slots.yaml entry, resolved back to the entity whose
// attribute declared it.
type Row struct {
	Entity string `json:"entity"`
	Slot   string `json:"slot"`
	Source string `json:"source"`
	Mapped bool   `json:"mapped"`
}

// Report is one package's coverage: every slot it declares, plus the
// share that are mapped. Declared/Mapped are carried explicitly (not left
// for a consumer to recompute from len(Rows)) because the JSON shape is a
// stable interface an agent parses: the summary should not require
// counting.
type Report struct {
	Package  string `json:"package"`
	Rows     []Row  `json:"rows"`
	Declared int    `json:"declared"`
	Mapped   int    `json:"mapped"`
}

// Compute returns one Report per leaf package in u (see
// lint.Universe.IsLeaf): a parent given only for extends resolution, e.g.
// universal-core alongside cafe-demo, is not itself reported on, since its
// slots.yaml is a template by definition (every source empty) and would
// only be noise next to the package actually being evaluated. Linting a
// root package alone still reports on it: with no children in the
// universe, it is trivially its own leaf.
func Compute(u *lint.Universe) []Report {
	var reports []Report
	for _, pkg := range u.Order {
		if pkg.Manifest == nil || !u.IsLeaf(pkg) {
			continue
		}
		reports = append(reports, computeOne(u, pkg))
	}
	return reports
}

// computeOne builds pkg's report. The attribute-to-entity lookup spans
// pkg plus every package in its extends chain (an install configures the
// resolved ontology, the same union KBF012 checks against: a package
// three layers deep still needs its great-grandparent's attributes
// resolvable), but the rows themselves come only from pkg's own
// install/slots.yaml: by convention a leaf package's slots.yaml is
// already the full resolved list, not just its own additions
// (examples/cafe-demo's has all 26 rows, not the 0 new ones it declares).
func computeOne(u *lint.Universe, pkg *lint.Package) Report {
	entityOf := attributeEntities(pkg)
	chain, _, _ := u.Chain(pkg)
	for _, ancestor := range chain {
		for slot, entity := range attributeEntities(ancestor) {
			if _, ok := entityOf[slot]; !ok {
				entityOf[slot] = entity
			}
		}
	}

	name := ""
	if pkg.Manifest != nil {
		name = pkg.Manifest.Name
	}
	report := Report{Package: name}
	for _, row := range pkg.Slots {
		mapped := row.Source != ""
		report.Rows = append(report.Rows, Row{
			Entity: entityOf[row.Slot],
			Slot:   row.Slot,
			Source: row.Source,
			Mapped: mapped,
		})
		report.Declared++
		if mapped {
			report.Mapped++
		}
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		a, b := report.Rows[i], report.Rows[j]
		if a.Entity != b.Entity {
			return a.Entity < b.Entity
		}
		return a.Slot < b.Slot
	})
	return report
}

// attributeEntities maps every declared attribute slot in pkg to the name
// of the entity that declares it.
func attributeEntities(pkg *lint.Package) map[string]string {
	m := map[string]string{}
	for _, e := range pkg.Elements {
		entity, ok := e.Value.(*model.Entity)
		if !ok {
			continue
		}
		for _, attr := range entity.Attributes {
			if attr.Slot != "" {
				m[attr.Slot] = entity.Name
			}
		}
	}
	return m
}
