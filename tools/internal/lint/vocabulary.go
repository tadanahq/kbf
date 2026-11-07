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

import "github.com/tadanahq/kbf/tools/internal/model"

// controlledVocabulary is KBF007 condition (a): the set of relation verbs
// any member of closure already declares. A composed playbook adding a
// verb legalizes it for everything that composes it, directly or
// transitively, not just an immediate child. Empty when closure is empty:
// a root has nothing to compose vocabulary from, so every one of its own
// relations instead passes through condition (b) (mintedEntities), which
// it always does: a root necessarily declares every entity it references,
// since it builds on nothing.
func controlledVocabulary(closure []*Package) map[string]bool {
	verbs := map[string]bool{}
	for _, member := range closure {
		for _, e := range member.Elements {
			if e.Kind == model.KindRelation {
				verbs[e.Name()] = true
			}
		}
	}
	return verbs
}

// mintedEntities is KBF007 condition (b): the set of entity names pkg
// declares that do NOT already exist, by name, anywhere in its closure:
// genuinely new concepts this playbook introduces, as opposed to a
// glossary override or fork of a composed entity (same name, but not
// new: KBF008 handles whether that redeclaration is itself legitimate).
// Only a genuinely new entity carries minting rights: a relation that only
// touches an entity already declared somewhere in the closure gets no
// self-inclusion, even if pkg happens to redeclare that same name on its
// own copy.
func mintedEntities(pkg *Package, closure []*Package) map[string]bool {
	minted := map[string]bool{}
	for _, e := range pkg.Elements {
		if e.Kind == model.KindEntity {
			minted[e.Name()] = true
		}
	}
	for _, member := range closure {
		for _, e := range member.Elements {
			if e.Kind == model.KindEntity {
				delete(minted, e.Name())
			}
		}
	}
	return minted
}
