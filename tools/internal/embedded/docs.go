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
	"path"
	"sort"
	"strings"
)

// docsRoot is where embedded spec docs live, one directory below FS's
// own root.
const docsRoot = "data/spec"

// DocNames returns every embedded spec doc's name: its path under
// data/spec/, without the .md extension (e.g. "onboarding",
// "primitives/entity"), sorted. Derived by walking the embedded tree
// rather than a hardcoded list, so it can never drift from what
// scripts/embedsync actually copied in.
func DocNames() []string {
	var names []string
	_ = fs.WalkDir(FS, docsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // a root walk error here would mean data/spec itself is missing from the build, which embed-freshness guards against, not this
		}
		if path.Ext(p) != ".md" {
			return nil
		}
		rel := strings.TrimPrefix(p, docsRoot+"/")
		names = append(names, strings.TrimSuffix(rel, ".md"))
		return nil
	})
	sort.Strings(names)
	return names
}

// Doc returns name's embedded content (see DocNames for the name shape),
// or ok=false if name isn't an embedded doc.
func Doc(name string) (content string, ok bool) {
	data, err := fs.ReadFile(FS, docsRoot+"/"+name+".md")
	if err != nil {
		return "", false
	}
	return string(data), true
}
