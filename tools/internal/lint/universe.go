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

// Result is one `kbf lint` run's outcome: every finding across every
// package passed in, sorted deterministically.
type Result struct {
	Findings []Finding
}

// Run lints the packages rooted at each of paths. Each path is loaded
// independently, then all of them together form the universe extends
// resolves against (design.md: "kbf lint takes one or more package paths
// ... resolved by name against that set, not by filesystem convention").
// Run returns a Go error only for a path that isn't a readable directory;
// every content problem, including an unresolvable extends, is a Finding.
func Run(paths []string) (Result, error) {
	pkgs := make(map[string]*Package, len(paths))
	var order []*Package
	var findings []Finding

	for _, p := range paths {
		pkg, loadFindings, err := loadPackage(p)
		if err != nil {
			return Result{}, err
		}
		findings = append(findings, loadFindings...)
		order = append(order, pkg)
		if pkg.Manifest != nil && pkg.Manifest.Name != "" {
			pkgs[pkg.Manifest.Name] = pkg
		}
	}

	for _, pkg := range order {
		findings = append(findings, checkManifest(pkg, pkgs)...)
	}
	for _, pkg := range order {
		if pkg.Manifest == nil {
			continue // already flagged: no manifest, nothing coherent to check
		}
		root := extendsRoot(pkg, pkgs)
		findings = append(findings, structuralFindings(pkg, root)...)
		findings = append(findings, semanticFindings(pkg, root, pkgs)...)
	}

	sortFindings(findings)
	return Result{Findings: findings}, nil
}

// extendsRoot resolves pkg's single extends hop (v0: depth 1) to the
// package its content is checked against. Returns pkg itself for a root
// package (Extends == nil). Returns nil when Extends is set but doesn't
// resolve within the loaded set: checkManifest already records that as
// KBF011, so extends-relative rules simply have nothing to check against.
func extendsRoot(pkg *Package, pkgs map[string]*Package) *Package {
	if pkg.Manifest.Extends == nil {
		return pkg
	}
	parent, ok := pkgs[*pkg.Manifest.Extends]
	if !ok {
		return nil
	}
	return parent
}
