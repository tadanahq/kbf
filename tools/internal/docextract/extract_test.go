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

package docextract_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// attributionPattern matches a repo-relative-looking path ending in .yaml
// or .yml: the one stable signal across every way spec/ docs name the real
// file an example is copied from ("Copied from
// `packages/universal-core/ontology/metrics.yaml`:", "The root package,
// `packages/universal-core/manifest.yaml`:", "...from
// `examples/cafe-demo/install/slots.yaml`:", or a plain comment inside the
// fence itself, "# See examples/cafe-demo/ontology/entities.yaml for the
// real file."). Wording and backtick-quoting both vary; the path shape does
// not, so that is what this matches on rather than a fixed phrase.
var attributionPattern = regexp.MustCompile(`([A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)+\.ya?ml)`)

// findMarkdownFiles walks dir for *.md files, sorted for determinism.
func findMarkdownFiles(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	return files
}

// extractYAMLBlocks parses mdFile's markdown AST and returns every fenced
// code block tagged ```yaml (not a bare ``` fence, which package-format.md
// uses for a plain directory-tree illustration, not YAML data).
func extractYAMLBlocks(t *testing.T, mdFile string) []yamlBlock {
	t.Helper()
	source, err := os.ReadFile(mdFile)
	if err != nil {
		t.Fatalf("reading %s: %v", mdFile, err)
	}

	doc := goldmark.DefaultParser().Parse(text.NewReader(source))
	relFile, err := filepath.Rel(specDir, mdFile)
	if err != nil {
		relFile = mdFile
	}

	var blocks []yamlBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		fence, ok := n.(*ast.FencedCodeBlock)
		if !ok || string(fence.Language(source)) != "yaml" {
			return ast.WalkContinue, nil
		}

		lines := fence.Lines()
		line := 1
		if lines.Len() > 0 {
			line = offsetToLine(source, lines.At(0).Start)
		}
		content := string(lines.Value(source))

		blocks = append(blocks, yamlBlock{
			mdFile:   filepath.Join("spec", relFile),
			line:     line,
			content:  content,
			attrPath: findAttribution(source, fence, content),
		})
		return ast.WalkContinue, nil
	})
	return blocks
}

// findAttribution looks for a repo-relative path in two places, in order:
// the paragraph immediately before the fence ("Copied from `path`:"), then
// the fence's own leading comment lines ("# See path for the real file.",
// as entity.md and metric.md's "common mistakes" fragment examples use).
func findAttribution(source []byte, fence *ast.FencedCodeBlock, content string) string {
	if prev := fence.PreviousSibling(); prev != nil {
		if m := attributionPattern.FindStringSubmatch(extractText(source, prev)); m != nil {
			return m[1]
		}
	}
	if m := attributionPattern.FindStringSubmatch(leadingComments(content)); m != nil {
		return m[1]
	}
	return ""
}

// leadingComments returns content's leading run of `#`-prefixed lines,
// stopping at the first line that isn't one.
func leadingComments(content string) string {
	var b strings.Builder
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			break
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// stripLeadingComments removes content's leading `#`-prefixed lines before
// the real-file comparison: those lines are doc-only annotations pointing
// a reader at the source file, never part of the file itself.
func stripLeadingComments(content string) string {
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
		i++
	}
	return strings.Join(lines[i:], "\n")
}

// extractText pulls a block node's rendered text, however it stores it:
// most block types expose Text(source); a few (blockquote, list items)
// don't, and doc examples do not currently use those right before a fence,
// so returning "" for anything else is correct, not a gap.
func extractText(source []byte, n ast.Node) string {
	if t, ok := n.(interface{ Text([]byte) []byte }); ok {
		return string(t.Text(source))
	}
	return ""
}

// offsetToLine converts a byte offset into source to a 1-based line
// number, for error messages that point a reader at the right place.
func offsetToLine(source []byte, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(source); i++ {
		if source[i] == '\n' {
			line++
		}
	}
	return line
}

// subtestName turns a block's location into a stable, readable subtest
// name (t.Run sanitizes spaces itself, but slashes and colons read badly
// in `go test -run`, so this avoids them).
func subtestName(b yamlBlock) string {
	name := strings.ReplaceAll(b.mdFile, "/", "_")
	return name + "_L" + strconv.Itoa(b.line)
}
