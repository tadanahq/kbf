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

package main

import (
	"strings"
	"testing"

	"github.com/tadanahq/kbf/tools/internal/embedded"
)

// TestDocsList checks the no-argument form: one name per line, every
// embedded doc represented, nothing else.
func TestDocsList(t *testing.T) {
	stdout, _, err := runCLI(t, "docs")
	if err != nil {
		t.Fatalf("docs: %v", err)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	want := embedded.DocNames()
	if len(lines) != len(want) {
		t.Fatalf("got %d line(s), want %d\ngot: %v\nwant: %v", len(lines), len(want), lines, want)
	}
	for i, name := range want {
		if lines[i] != name {
			t.Errorf("line %d = %q, want %q", i, lines[i], name)
		}
	}
}

// TestDocsPrint checks the named form: the doc's real content on stdout,
// byte for byte what embedded.Doc itself returns.
func TestDocsPrint(t *testing.T) {
	for _, name := range []string{"onboarding", "cli", "primitives/entity"} {
		t.Run(name, func(t *testing.T) {
			stdout, _, err := runCLI(t, "docs", name)
			if err != nil {
				t.Fatalf("docs %s: %v", name, err)
			}
			want, ok := embedded.Doc(name)
			if !ok {
				t.Fatalf("embedded.Doc(%q) not found (test itself is broken)", name)
			}
			if stdout != want {
				t.Errorf("docs %s printed content that does not match embedded.Doc", name)
			}
			if !strings.Contains(stdout, "type: spec-doc") {
				t.Errorf("docs %s: missing OKF frontmatter in printed output", name)
			}
		})
	}
}

// TestDocsUnknownName: an unrecognized name is a teachable error, not a
// silent empty print or a panic.
func TestDocsUnknownName(t *testing.T) {
	_, _, err := runCLI(t, "docs", "not-a-real-doc")
	if err == nil {
		t.Fatal("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "not-a-real-doc") {
		t.Errorf("error = %q, want it to name the bad argument", err.Error())
	}
	if !strings.Contains(err.Error(), "onboarding") {
		t.Errorf("error = %q, want it to list valid names", err.Error())
	}
}
