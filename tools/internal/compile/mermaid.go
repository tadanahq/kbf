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

package compile

import (
	"fmt"
	"strings"
)

// ToMermaid renders a Graph as a mermaid flowchart: entities as
// rectangle nodes, relations as solid labeled edges, actions as circle
// nodes with a dashed labeled edge into the entity they target (a
// visually distinct "annotation", not a peer connection between two
// entities). g is assumed already sorted (BuildGraph does this); ToMermaid
// does not re-sort, so it stays a pure formatter.
//
// Output is plain, un-fenced mermaid source: valid as the body of a
// GitHub Markdown ```mermaid fence, or as a standalone .mmd file for any
// mermaid-aware editor. Every node id is sanitized (hyphens to
// underscores; mermaid ids may not contain them unquoted) while the
// visible label keeps the real kebab-case name, so the rendered diagram
// still reads exactly like the ontology's own vocabulary.
func ToMermaid(g Graph) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")

	if len(g.Entities) > 0 {
		b.WriteString("  %% entities\n")
		for _, name := range g.Entities {
			fmt.Fprintf(&b, "  %s[%q]\n", nodeID(name), name)
		}
	}

	if len(g.Relations) > 0 {
		b.WriteString("  %% relations\n")
		for _, r := range g.Relations {
			fmt.Fprintf(&b, "  %s -->|%s| %s\n", nodeID(r.From), r.Verb, nodeID(r.To))
		}
	}

	if len(g.Actions) > 0 {
		b.WriteString("  %% actions\n")
		for _, a := range g.Actions {
			id := "action_" + nodeID(a.Name)
			fmt.Fprintf(&b, "  %s((%q))\n", id, a.Name)
			fmt.Fprintf(&b, "  %s -.->|%s| %s\n", id, a.Approval, nodeID(a.On))
		}
	}

	return b.String()
}

// nodeID turns a kebab-case element name into a mermaid-safe bare
// identifier. Mermaid node ids are alphanumeric-and-underscore in
// practice; a hyphen is parsed as part of the `-->` edge syntax rather
// than an id character the moment it is unquoted, so every id here is
// underscore-joined while quoted labels (`[%q]`) keep the original name.
func nodeID(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
