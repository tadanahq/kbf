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

// corePlaybooks are exactly the playbooks embedded: the public core
// layer, never a vertical (examples/ is teaching content, not something
// a consumer's own playbook should silently compose against). Keep this
// list in sync with scripts/embedsync/source.go's own corePlaybooks;
// `make embed-freshness` is what actually enforces the two never drift,
// not this comment.
var corePlaybooks = []string{"core-business", "core-operations", "core-services"}

// CorePlaybookNames returns the embedded core playbooks' manifest names,
// sorted.
func CorePlaybookNames() []string {
	names := append([]string(nil), corePlaybooks...)
	sort.Strings(names)
	return names
}

// Playbook returns an fs.FS rooted at name's embedded playbook directory
// (manifest.yaml, ontology/, evals/, install/), or ok=false if name isn't
// one of the embedded core playbooks. This is lint.PlaybookSource's shape
// exactly: cmd/kbf passes embedded.Playbook to
// lint.LoadWithEmbedded/RunWithEmbedded unchanged, as the fallback
// consulted for a builds-on name no local path already provided.
func Playbook(name string) (fsys fs.FS, ok bool) {
	for _, n := range corePlaybooks {
		if n != name {
			continue
		}
		sub, err := fs.Sub(FS, "data/playbooks/"+name)
		if err != nil {
			return nil, false
		}
		return sub, true
	}
	return nil, false
}
