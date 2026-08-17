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

// Package compile emits an ontology as a diagram: v0 ships one emitter,
// mermaid (design.md: "mermaid over any richer render: it previews
// natively in editors and GitHub"). Like internal/coverage, it builds on
// internal/lint's loader (lint.Load) rather than parsing YAML itself.
package compile

import (
	"sort"

	"github.com/tadanahq/kbf/tools/internal/lint"
	"github.com/tadanahq/kbf/tools/internal/model"
)

// Graph is the ontology map: entities as nodes, relations as labeled
// edges, actions as annotations (design.md's "Key flows"). Metrics carry
// no graph shape (a formula over entities, not a connection between
// them) and are not part of this output.
type Graph struct {
	Entities  []string
	Relations []RelationEdge
	Actions   []ActionEdge
}

// RelationEdge is one relation: From -verb-> To.
type RelationEdge struct {
	From, To, Verb string
}

// ActionEdge is one action, drawn as an annotation on the entity it
// targets: Name -approval-> On.
type ActionEdge struct {
	Name, On, Approval string
}

// BuildGraph unions every package in u: for the common case (one leaf
// package plus the playbook(s) it builds on, e.g. cafe-demo +
// core-business), this is exactly the resolved map a reader wants: the
// whole picture, not just the leaf's own incremental additions. Unlike
// coverage, compile has no reason to exclude a parent given as context:
// showing the same picture "again" from union is harmless (dedup is by
// identity), and omitting it would render a near-empty diagram for any
// package that mostly composes rather than adds, which defeats the point
// of a map.
func BuildGraph(u *lint.Universe) Graph {
	entities := map[string]bool{}
	relations := map[string]RelationEdge{}
	actions := map[string]ActionEdge{}

	for _, pkg := range u.Order {
		for _, e := range pkg.Elements {
			switch v := e.Value.(type) {
			case *model.Entity:
				entities[v.Name] = true
			case *model.Relation:
				edge := RelationEdge{From: v.From, To: v.To, Verb: v.Name}
				relations[v.From+"\x00"+v.Name+"\x00"+v.To] = edge
			case *model.Action:
				actions[v.Name] = ActionEdge{Name: v.Name, On: v.On, Approval: v.Approval}
			}
		}
	}

	g := Graph{Entities: keys(entities)}
	for _, e := range relations {
		g.Relations = append(g.Relations, e)
	}
	for _, a := range actions {
		g.Actions = append(g.Actions, a)
	}

	sort.Strings(g.Entities)
	sort.Slice(g.Relations, func(i, j int) bool {
		a, b := g.Relations[i], g.Relations[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.Verb != b.Verb {
			return a.Verb < b.Verb
		}
		return a.To < b.To
	})
	sort.Slice(g.Actions, func(i, j int) bool {
		a, b := g.Actions[i], g.Actions[j]
		if a.On != b.On {
			return a.On < b.On
		}
		return a.Name < b.Name
	})
	return g
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
