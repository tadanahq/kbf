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
	"fmt"
	"sort"
	"strings"
)

// checkCrossPlaybookCollisions is KBF003's cross-playbook variant: two
// DIFFERENT playbooks in pkg's own composition closure declaring the
// same element identity. Composition has no resolution order: a DAG closure
// has no tie-break that gives a name one meaning once more than one
// immediate parent exists, so an identity declared by more than one closure
// member is never resolvable and is always an error, whether or not pkg
// itself also touches that identity (that separate case, pkg redeclaring
// something in its closure, is KBF008). Symmetric: every closure member
// that participates in a given collision gets its own finding in its own
// file, the same way a composition cycle reports against every playbook on
// the cycle, not just one.
func checkCrossPlaybookCollisions(closure []*Package) []Finding {
	type occurrence struct {
		pkg  *Package
		elem Element
	}
	byIdentity := map[string][]occurrence{}
	for _, member := range closure {
		for _, e := range member.Elements {
			if !matchable[e.Kind] {
				continue
			}
			key := string(e.Kind) + "\x00" + identityKey(e)
			byIdentity[key] = append(byIdentity[key], occurrence{member, e})
		}
	}

	var findings []Finding
	for _, occs := range byIdentity {
		involved := map[string]bool{}
		for _, o := range occs {
			involved[o.pkg.Manifest.Name] = true
		}
		if len(involved) < 2 {
			continue // one playbook declaring itself twice is ordinary KBF003, checkDuplicateNames' job
		}
		for _, o := range occs {
			var others []string
			for name := range involved {
				if name != o.pkg.Manifest.Name {
					others = append(others, name)
				}
			}
			sort.Strings(others)
			findings = append(findings, Finding{
				Rule: KBF003, File: o.elem.File, Line: o.elem.Line, Element: o.elem.Name(),
				Message: fmt.Sprintf("%s %s is also declared by %s: composition has no resolution order, so identity must be unique across the closure", o.elem.Kind, describeKey(o.elem), strings.Join(others, ", ")),
				Fix:     "rename one, or remove the duplicate from whichever playbook shouldn't own it",
			})
		}
	}
	return findings
}
