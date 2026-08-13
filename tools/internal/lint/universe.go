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
// all of them together form the set extends resolves against (design.md:
// "kbf lint takes one or more package paths ... resolved by name against
// that set, not by filesystem convention"). Load returns a Go error only
// for a path that isn't a readable directory; every content problem,
// including an unresolvable extends, is a Finding for the caller to
// decide what to do with (lint fails on it; coverage/compile mostly don't
// care and just resolve what they can).
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

// ExtendsRoot resolves pkg's single extends hop (v0: depth 1) to the
// package its content is checked or resolved against. Returns pkg itself
// for a root package (Extends == nil). Returns nil when Extends is set but
// doesn't resolve within the universe: lint's checkManifest records that
// as KBF011; coverage/compile simply have nothing more to inherit from.
func (u *Universe) ExtendsRoot(pkg *Package) *Package {
	if pkg.Manifest == nil || pkg.Manifest.Extends == nil {
		return pkg
	}
	parent, ok := u.Packages[*pkg.Manifest.Extends]
	if !ok {
		return nil
	}
	return parent
}

// IsLeaf reports whether pkg is not the extends-root of any other package
// in this universe. coverage and compile report on leaf packages (the
// ones actually being evaluated or rendered); a parent given alongside
// them is resolution context, not a second subject. A package linted
// alone (no children in the set) is trivially its own leaf.
func (u *Universe) IsLeaf(pkg *Package) bool {
	for _, other := range u.Order {
		if other == pkg {
			continue
		}
		if other.Manifest != nil && other.Manifest.Extends != nil && pkg.Manifest != nil && *other.Manifest.Extends == pkg.Manifest.Name {
			return false
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
// universe is built and extends resolved.
func Run(paths []string) (Result, error) {
	universe, findings, err := Load(paths)
	if err != nil {
		return Result{}, err
	}

	for _, pkg := range universe.Order {
		findings = append(findings, checkManifest(pkg, universe.Packages)...)
	}
	for _, pkg := range universe.Order {
		if pkg.Manifest == nil {
			continue // already flagged: no manifest, nothing coherent to check
		}
		root := universe.ExtendsRoot(pkg)
		findings = append(findings, structuralFindings(pkg, root)...)
		findings = append(findings, semanticFindings(pkg, root, universe.Packages)...)
	}

	sortFindings(findings)
	return Result{Findings: findings}, nil
}
