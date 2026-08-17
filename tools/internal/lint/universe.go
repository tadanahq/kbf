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

import "sort"

// Universe is every package loaded together for one kbf invocation, keyed
// by manifest name plus the original command-line order. internal/lint
// owns loading (project-architecture.md), so internal/coverage and
// internal/compile both build on Load and Universe instead of re-parsing
// YAML themselves: one loader, three consumers.
type Universe struct {
	Packages map[string]*Package
	Order    []*Package
}

// Load reads every package rooted at each of paths (position-aware; see
// loadPackage) into one Universe. Each path is loaded independently, then
// all of them together form the set builds-on resolves against (design.md:
// "kbf lint takes one or more package paths ... resolved by name against
// that set, not by filesystem convention"). Load returns a Go error only
// for a path that isn't a readable directory; every content problem,
// including an unresolvable builds-on entry, is a Finding for the caller
// to decide what to do with (lint fails on it; coverage/compile mostly
// don't care and just resolve what they can).
func Load(paths []string) (*Universe, []Finding, error) {
	u := &Universe{Packages: make(map[string]*Package, len(paths))}
	var findings []Finding

	for _, p := range paths {
		pkg, loadFindings, err := loadPackage(p)
		if err != nil {
			return nil, nil, err
		}
		findings = append(findings, loadFindings...)
		u.Order = append(u.Order, pkg)
		if pkg.Manifest != nil && pkg.Manifest.Name != "" {
			u.Packages[pkg.Manifest.Name] = pkg
		}
	}
	return u, findings, nil
}

// Closure resolves pkg's full composition: every playbook reachable by
// following BuildsOn transitively, deduped by manifest name, pkg itself
// never included. This is a DAG walk, not a single-parent chain walk: a
// playbook can compose more than one immediate parent (the diamond case
// — two core playbooks both building on the same root, then a third
// composing both of them — design.md's "Implementation clarifications"),
// and an ancestor reached through two different paths is still one
// instance in the result, not two. The returned slice is sorted by name
// for determinism (composition has no "nearest first" notion once more
// than one immediate parent is possible; every rule that used to walk a
// chain nearest-ancestor-first now treats the closure as an unordered
// set instead — see semantic.go's checkCrossPlaybookCollisions for what
// that changes about identity conflicts).
//
// resolved is false the moment a BuildsOn name isn't in the universe: the
// existing "missing parent" case (KBF011), just possibly discovered
// several hops in rather than on pkg's own immediate list. cycle is true
// if the walk would revisit a package name already on the current
// composition path (not merely "already seen anywhere": a diamond
// re-visits a shared ancestor legitimately, that is not a cycle). Either
// way, closure still holds whatever was resolved before the problem,
// since partial context beats none.
func (u *Universe) Closure(pkg *Package) (closure []*Package, resolved bool, cycle bool) {
	seen := map[string]*Package{}
	resolved = true

	var walk func(current *Package, onPath map[string]bool)
	walk = func(current *Package, onPath map[string]bool) {
		if current.Manifest == nil {
			return
		}
		for _, name := range current.Manifest.BuildsOn {
			parent, ok := u.Packages[name]
			if !ok {
				resolved = false
				continue
			}
			if onPath[name] {
				cycle = true
				continue
			}
			if _, already := seen[name]; already {
				continue // diamond merge: this ancestor's own subtree was already walked
			}
			seen[name] = parent
			closure = append(closure, parent)

			nextPath := make(map[string]bool, len(onPath)+1)
			for k := range onPath {
				nextPath[k] = true
			}
			nextPath[name] = true
			walk(parent, nextPath)
		}
	}

	startPath := map[string]bool{}
	if pkg.Manifest != nil {
		startPath[pkg.Manifest.Name] = true
	}
	walk(pkg, startPath)

	sort.Slice(closure, func(i, j int) bool { return closure[i].Manifest.Name < closure[j].Manifest.Name })
	return closure, resolved, cycle
}

// IsLeaf reports whether pkg is not a build-on target of any other package
// in this universe. coverage and compile report on leaf packages (the
// ones actually being evaluated or rendered); a parent given alongside
// them is resolution context, not a second subject. A package linted
// alone (no children in the set) is trivially its own leaf.
func (u *Universe) IsLeaf(pkg *Package) bool {
	if pkg.Manifest == nil {
		return true
	}
	for _, other := range u.Order {
		if other == pkg || other.Manifest == nil {
			continue
		}
		for _, name := range other.Manifest.BuildsOn {
			if name == pkg.Manifest.Name {
				return false
			}
		}
	}
	return true
}

// Result is one `kbf lint` run's outcome: every finding across every
// package passed in, sorted deterministically.
type Result struct {
	Findings []Finding
}

// Run lints the packages rooted at each of paths: see Load for how the
// universe is built and Closure for how composition is resolved.
func Run(paths []string) (Result, error) {
	universe, findings, err := Load(paths)
	if err != nil {
		return Result{}, err
	}

	for _, pkg := range universe.Order {
		findings = append(findings, checkManifest(pkg, universe.Packages)...)
		findings = append(findings, checkTaxonomy(pkg, universe.Packages)...)
		if pkg.Manifest == nil {
			continue // already flagged: no manifest, nothing coherent to check
		}
		closure, _, cycle := universe.Closure(pkg)
		if cycle {
			findings = append(findings, Finding{
				Rule: KBF011, File: pkg.ManifestFile, Line: pkg.ManifestLine, Element: "builds-on",
				Message: "composition closure cycles back to a package already on this playbook's own path",
				Fix:     "break the cycle: one of these playbooks must stop building on something that, directly or indirectly, builds on it back",
			})
		}
		findings = append(findings, structuralFindings(pkg, closure)...)
		findings = append(findings, semanticFindings(pkg, closure)...)
	}

	sortFindings(findings)
	return Result{Findings: findings}, nil
}
