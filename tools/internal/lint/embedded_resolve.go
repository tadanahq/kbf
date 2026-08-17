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

import (
	"io/fs"
	"sort"
)

// PlaybookSource resolves a playbook name to an fs.FS rooted at its
// directory (containing manifest.yaml, ontology/, evals/, install/), for
// composition fallback beyond what was explicitly passed on the command
// line. ok is false if name isn't available from this source.
// internal/embedded's core-playbook lookup is the only implementation in
// v0 (project-decisions.md: "batteries-included binary"); the type lives
// here, not there, so internal/lint never imports internal/embedded and
// gains a dependency it doesn't otherwise need, only a func value shaped
// like one of its exports.
type PlaybookSource func(name string) (fsys fs.FS, ok bool)

// LoadWithEmbedded is Load, plus one more source of playbooks: any
// builds-on name still missing once paths are loaded is resolved from
// fallback, transitively, before being reported unresolved (KBF011).
// Local paths always win: fallback is consulted only for a name paths
// never provided at all, never to override or shadow a package already
// loaded under that name, so a local checkout of core-operations (say,
// mid-edit, not yet matching the published embedded copy) is always what
// resolves, not the embedded fallback sitting behind it. This is the kbf
// CLI's own loader (cmd/kbf's lint/coverage/compile all call this, never
// Load, so embedded resolution is one behavior shared by all three, not
// three separate opt-ins); conformance and internal/lint's own tests use
// Load, unaffected by fallback existing at all.
func LoadWithEmbedded(paths []string, fallback PlaybookSource) (*Universe, []Finding, error) {
	u, findings, err := Load(paths)
	if err != nil {
		return nil, nil, err
	}
	if fallback != nil {
		u.resolveEmbedded(fallback)
	}
	return u, findings, nil
}

// RunWithEmbedded is Run, built on LoadWithEmbedded instead of Load: see
// that doc comment for the precedence rule embedded fallback follows.
func RunWithEmbedded(paths []string, fallback PlaybookSource) (Result, error) {
	universe, findings, err := LoadWithEmbedded(paths, fallback)
	if err != nil {
		return Result{}, err
	}
	return runOn(universe, findings), nil
}

// resolveEmbedded pulls in, from fallback, every builds-on name a
// currently-loaded package needs but doesn't have, repeated to a fixed
// point: a package resolved this way can itself have builds-on entries
// that also need resolving (core-operations needing core-business is
// exactly this, one extra hop). Each pass that resolves nothing new ends
// the loop, so a name fallback genuinely doesn't have is left for KBF011
// to report exactly as it always has, not retried forever.
func (u *Universe) resolveEmbedded(fallback PlaybookSource) {
	// A plain `defer sort.Strings(u.EmbeddedNames)` would evaluate
	// u.EmbeddedNames as the deferred call's argument immediately, while
	// it is still nil (defer evaluates arguments at the defer statement,
	// not at return): wrapped in a closure so the field is read when the
	// deferred call actually runs, after every pass below has appended
	// to it.
	defer func() { sort.Strings(u.EmbeddedNames) }()
	for {
		missing := u.missingNames()
		if len(missing) == 0 {
			return
		}
		progress := false
		for _, name := range missing {
			fsys, ok := fallback(name)
			if !ok {
				continue // not embedded either; KBF011 fires on it downstream, same as always
			}
			pkg, _, err := loadPackage(fsys, "embedded:"+name)
			if err != nil {
				continue // an embedded playbook that fails to even load is not this name's problem to explain differently; leave it unresolved
			}
			u.Packages[name] = pkg
			u.Order = append(u.Order, pkg)
			u.EmbeddedNames = append(u.EmbeddedNames, name)
			progress = true
		}
		if !progress {
			return
		}
	}
}

// missingNames returns every distinct builds-on name a loaded package
// references that isn't itself in u.Packages yet, sorted.
func (u *Universe) missingNames() []string {
	need := map[string]bool{}
	for _, pkg := range u.Order {
		if pkg.Manifest == nil {
			continue
		}
		for _, name := range pkg.Manifest.BuildsOn {
			if _, ok := u.Packages[name]; !ok {
				need[name] = true
			}
		}
	}
	names := make([]string, 0, len(need))
	for n := range need {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
