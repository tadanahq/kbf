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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// file is one embedded copy: relPath is relative to
// tools/internal/embedded/data/ (always slash-separated, e.g.
// "playbooks/core-business/manifest.yaml"), content is what gets written
// there, header included where one applies.
type file struct {
	relPath string
	content []byte
}

// corePlaybooks are the only playbooks/ directories that get embedded:
// the public core layer, never a vertical (verticals are teaching
// content, examples/, not something a consumer's own playbook should
// silently compose against).
var corePlaybooks = []string{"core-business", "core-operations", "core-services"}

// skills are the skill directories that get embedded, one subtree each.
var skills = []string{"kbf-authoring"}

// sourceFiles reads every file this batch embeds from root (the repo
// root) and returns them as data/-relative file values, sorted by
// relPath so output is deterministic run to run.
func sourceFiles(root string) ([]file, error) {
	var files []file

	for _, name := range corePlaybooks {
		src := filepath.Join(root, "playbooks", name)
		got, err := walkTree(src, filepath.Join("playbooks", name))
		if err != nil {
			return nil, fmt.Errorf("playbooks/%s: %w", name, err)
		}
		files = append(files, got...)
	}

	for _, name := range skills {
		src := filepath.Join(root, "skills", name)
		got, err := walkTree(src, filepath.Join("skills", name))
		if err != nil {
			return nil, fmt.Errorf("skills/%s: %w", name, err)
		}
		files = append(files, got...)
	}

	specMD, err := specMarkdown(filepath.Join(root, "spec"), "spec")
	if err != nil {
		return nil, fmt.Errorf("spec: %w", err)
	}
	files = append(files, specMD...)

	primitivesMD, err := specMarkdown(filepath.Join(root, "spec", "primitives"), filepath.Join("spec", "primitives"))
	if err != nil {
		return nil, fmt.Errorf("spec/primitives: %w", err)
	}
	files = append(files, primitivesMD...)

	sort.Slice(files, func(i, j int) bool { return files[i].relPath < files[j].relPath })
	return files, nil
}

// walkTree copies every regular file under src recursively, tagging each
// with a DO-NOT-EDIT header (YAML files get a leading comment line; any
// other extension is copied verbatim, since a playbook or skill in v0
// never contains a format without a comment syntax this cheap to inject).
// relBase is the data/-relative prefix every found file's relPath is
// rooted at; for playbooks/ and skills/ this is also, unmodified, the
// real repo-root-relative source path (data/ mirrors the repo root's own
// layout exactly), which is what withHeader cites in its DO-NOT-EDIT note.
func walkTree(src, relBase string) ([]file, error) {
	var files []file
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(filepath.Join(relBase, rel))
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, file{relPath: relPath, content: withHeader(relPath, data)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// specMarkdown copies every *.md file directly inside dir (non-recursive:
// spec/ and spec/primitives/ are handled as two separate calls, exactly
// once each) with a markdown DO-NOT-EDIT header injected after
// frontmatter. relBase is the data/-relative prefix.
func specMarkdown(dir, relBase string) ([]file, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []file
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		relPath := filepath.ToSlash(filepath.Join(relBase, e.Name()))
		files = append(files, file{relPath: relPath, content: withHeader(relPath, data)})
	}
	return files, nil
}

// withHeader prepends (YAML) or injects after frontmatter (Markdown) a
// one-line DO-NOT-EDIT note naming relPath, which (data/ mirroring the
// repo root's own layout exactly) also doubles as the repo-root-relative
// path a maintainer should actually edit. Any other extension is
// returned unchanged: nothing embedded in v0 needs that case, and
// silently corrupting an unknown format would be worse than a missing
// note.
func withHeader(relPath string, content []byte) []byte {
	note := fmt.Sprintf("DO NOT EDIT: generated from %s by scripts/embedsync. Edit the source, then run `make embed-sync`.", relPath)
	switch filepath.Ext(relPath) {
	case ".yaml", ".yml":
		return append([]byte("# "+note+"\n"), content...)
	case ".md":
		return []byte(injectAfterFrontmatter(string(content), note))
	default:
		return content
	}
}

// injectAfterFrontmatter inserts an HTML comment carrying note right
// after a document's closing frontmatter delimiter (the second line that
// is exactly "---"), so it never lands before the frontmatter and breaks
// OKF parsing. Every file this program hands to it opens with
// frontmatter (verified once, by hand, across all 14 synced docs: each
// has exactly two bare "---" lines, both the delimiter, none a markdown
// horizontal rule); a document with none found is prepended instead of
// silently dropping the note.
func injectAfterFrontmatter(content, note string) string {
	lines := strings.Split(content, "\n")
	dashes := 0
	for i, line := range lines {
		if strings.TrimSpace(line) != "---" {
			continue
		}
		dashes++
		if dashes != 2 {
			continue
		}
		before := strings.Join(lines[:i+1], "\n")
		after := strings.Join(lines[i+1:], "\n")
		return before + "\n\n<!-- " + note + " -->\n" + after
	}
	return "<!-- " + note + " -->\n\n" + content
}
