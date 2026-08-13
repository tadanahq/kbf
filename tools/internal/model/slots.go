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

// SlotMapping is one row of a package's install/slots.yaml: a declared slot
// and the source system it fills from, if known yet. The file itself is a
// plain list of these rows (no `kind:` wrapper): static coverage is the
// share of rows with a non-empty Source. Every attribute Slot reference
// (model.Attribute.Slot) must name a slot declared here: KBF012.
type SlotMapping struct {
	// Slot is the slot identifier, e.g. "pos.catalog". Referenced by
	// Attribute.Slot elsewhere in the package.
	Slot string `yaml:"slot" json:"slot"`
	// Source is the source system this slot is mapped to. Empty means
	// declared but not yet mapped: counted as unmapped by `kbf coverage`.
	Source string `yaml:"source" json:"source"`
}
