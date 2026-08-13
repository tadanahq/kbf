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

// Package docextract holds a single test: every ```yaml fenced block in
// spec/**/*.md must be real, not aspirational prose. project-overview.md's
// acceptance criteria says so directly ("every doc's YAML examples lint
// green via a doc-extraction test"), and the content batch's own tasks.md
// notes record the same intent by hand (byte-diffing every example
// against the file it claims to be copied from) before this test existed
// to make that permanent and automatic. The markdown-walking machinery
// that pulls blocks out of a .md file lives in extract_test.go; this file
// is just what the test asserts about them.
package docextract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

// specDir is spec/ at the repo root, relative to this test file's own
// package (tools/internal/docextract).
const specDir = "../../../spec"

// yamlBlock is one extracted fenced code block.
type yamlBlock struct {
	mdFile   string
	line     int
	content  string
	attrPath string // repo-relative path this block claims to be copied from, if any
}

func TestSpecYAMLBlocksAreReal(t *testing.T) {
	mdFiles := findMarkdownFiles(t, specDir)
	if len(mdFiles) == 0 {
		t.Fatalf("no markdown files found under %s", specDir)
	}

	var blocks []yamlBlock
	for _, mdFile := range mdFiles {
		blocks = append(blocks, extractYAMLBlocks(t, mdFile)...)
	}
	if len(blocks) == 0 {
		t.Fatal("no ```yaml fenced blocks found anywhere under spec/ - extraction is broken, or the docs have none, either way this test cannot do its job")
	}

	attributed, bare := 0, 0
	for _, b := range blocks {
		t.Run(subtestName(b), func(t *testing.T) {
			// Baseline, every block: must be syntactically valid YAML. A
			// typo here is exactly the kind of doc/reality drift this test
			// exists to catch.
			var v any
			if err := yaml.Unmarshal([]byte(b.content), &v); err != nil {
				t.Fatalf("%s:%d: invalid YAML: %v\n---\n%s", b.mdFile, b.line, err, b.content)
			}

			if b.attrPath == "" {
				return // an inline illustrative snippet, not tied to a real file
			}

			real, err := os.ReadFile(filepath.Join("..", "..", "..", b.attrPath))
			if err != nil {
				t.Fatalf("%s:%d: claims to be copied from %s, but that file does not exist: %v", b.mdFile, b.line, b.attrPath, err)
			}
			// A substring match, not whole-file equality: several examples
			// are one row or one document lifted out of a larger multi-row
			// or multi-document file (slot-mapping.md's single slots.yaml
			// row is the clearest case). Leading `#` comment lines are
			// doc-only annotations (e.g. "# See path for the real file.")
			// that never appear in the real file, so they're stripped
			// before comparing; what matters is that the actual YAML
			// content still appears in the file, verbatim.
			compare := strings.TrimRight(stripLeadingComments(b.content), "\n")
			if !strings.Contains(string(real), compare) {
				t.Errorf("%s:%d: content does not appear verbatim in %s (the doc has drifted from the real file, or was never copied from it)\n--- doc says ---\n%s\n--- real file ---\n%s", b.mdFile, b.line, b.attrPath, compare, string(real))
			}
		})
		if b.attrPath != "" {
			attributed++
		} else {
			bare++
		}
	}
	t.Logf("%d yaml blocks checked (%d verified against a real file, %d syntax-only)", len(blocks), attributed, bare)
}
