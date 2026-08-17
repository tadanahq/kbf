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

// Manifest is a package's identity card: manifest.yaml at the package root.
// It names the package, the spec version it targets, and what it extends.
// Every package has exactly one manifest; the linter reads it before
// touching any ontology content, since extends drives cross-package
// resolution (KBF006, KBF007, KBF008).
type Manifest struct {
	// Name is the package's identifier, e.g. "core-universal". Doubles as
	// its namespace: element uniqueness (KBF003) and fork detection
	// (KBF008) are scoped by this name.
	Name string `yaml:"name" json:"name"`
	// Version is the package's own version, independent of the spec
	// version it targets (project-standards.md: spec versioning is
	// independent of tool/package versioning).
	Version string `yaml:"version" json:"version"`
	// Spec is the KBF spec version this package's content was authored
	// against, e.g. "v0".
	Spec string `yaml:"spec" json:"spec"`
	// Extends names the parent package this one extends, or null for a
	// root package (only core-universal is root in v0). A pointer so the
	// loader can tell "null" (root, valid) apart from "missing key"
	// (invalid): both decode a Go string to "", but only the former is a
	// legal manifest. Non-null values must resolve within the set of
	// packages passed to `kbf lint`: KBF011.
	Extends *string `yaml:"extends" json:"extends" jsonschema:"nullable"`
}
