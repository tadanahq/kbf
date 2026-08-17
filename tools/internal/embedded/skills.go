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

package embedded

import (
	"io/fs"
	"sort"
)

// skills are the embedded skill directory names. Keep in sync with
// scripts/embedsync/source.go's own skills list; `make embed-freshness`
// enforces the two never drift, not this comment.
var skills = []string{"kbf-authoring"}

// SkillNames returns the embedded skill directory names, sorted.
func SkillNames() []string {
	names := append([]string(nil), skills...)
	sort.Strings(names)
	return names
}

// Skill returns an fs.FS rooted at name's embedded skill directory
// (SKILL.md and anything else that ships alongside it), or ok=false if
// name isn't an embedded skill.
func Skill(name string) (fsys fs.FS, ok bool) {
	for _, n := range skills {
		if n != name {
			continue
		}
		sub, err := fs.Sub(FS, "data/skills/"+name)
		if err != nil {
			return nil, false
		}
		return sub, true
	}
	return nil, false
}
