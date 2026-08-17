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
	"io/fs"
	"path/filepath"
	"sort"

	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// loadPackage reads one package from fsys: an fs.FS rooted at the
// package's own directory, either a real disk directory (os.DirFS, for a
// locally-passed path) or a sub-tree of the embedded core playbooks
// (composition fallback, see embedded.go). Both sources share this one
// loader (project-architecture.md: "internal/lint owns loading"), so an
// embedded package is validated exactly as strictly as a local one, never
// a second, looser code path. display is the path used to build every
// human-facing file reference (Finding.File, Package.Root): a real disk
// path for a local package, "embedded:<name>" for one resolved from
// fallback, so a reader can always tell which is which even though both
// parse identically. It returns a genuine error only when fsys's root
// itself isn't readable as a directory; every content problem becomes a
// Finding, so one bad file never aborts loading the rest of the package.
func loadPackage(fsys fs.FS, display string) (*Package, []Finding, error) {
	info, err := fs.Stat(fsys, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", display, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("%s: not a directory", display)
	}

	pkg := &Package{Root: display}
	var findings []Finding

	manifest, mFile, mLine, mFindings := loadManifest(fsys, display)
	pkg.Manifest, pkg.ManifestFile, pkg.ManifestLine = manifest, mFile, mLine
	findings = append(findings, mFindings...)

	for _, dir := range []string{"ontology", "evals"} {
		elems, f := loadElements(fsys, dir, display)
		pkg.Elements = append(pkg.Elements, elems...)
		findings = append(findings, f...)
	}

	slots, sFile, sFindings := loadSlots(fsys, display)
	pkg.Slots, pkg.SlotsFile = slots, sFile
	findings = append(findings, sFindings...)

	return pkg, findings, nil
}

// loadManifest reads manifest.yaml from fsys's root. A missing file is
// KBF011, not a Go error: the caller still gets a Package back (with
// Manifest == nil) so it can keep loading the rest and report every
// problem in one lint run.
func loadManifest(fsys fs.FS, display string) (m *model.Manifest, file string, line int, findings []Finding) {
	file = filepath.Join(display, "manifest.yaml")
	data, err := fs.ReadFile(fsys, "manifest.yaml")
	if err != nil {
		return nil, file, 1, []Finding{{
			Rule: KBF011, File: file, Line: 1, Element: display,
			Message: "manifest.yaml is missing or unreadable: " + err.Error(),
			Fix:     "add a manifest.yaml with name, version, spec, builds-on, and layer",
		}}
	}

	astFile, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return nil, file, errorLine(err, 1), []Finding{malformedYAML(file, err)}
	}
	if len(astFile.Docs) == 0 || astFile.Docs[0].Body == nil {
		return nil, file, 1, []Finding{{Rule: KBF011, File: file, Line: 1, Element: display, Message: "manifest.yaml is empty", Fix: "add name, version, spec, builds-on, and layer"}}
	}
	node := astFile.Docs[0].Body
	line = node.GetToken().Position.Line

	var manifest model.Manifest
	if err := yaml.NodeToValue(node, &manifest); err != nil {
		return nil, file, errorLine(err, line), []Finding{malformedYAML(file, err)}
	}
	return &manifest, file, line, nil
}

// loadSlots reads install/slots.yaml: a flat list with no `kind:`
// discriminator, one row per line, position-aware so KBF012 can point at
// the specific row. A missing file is not an error: a package with no
// attributes to map legitimately has nothing to declare.
func loadSlots(fsys fs.FS, display string) (slots []SlotRow, file string, findings []Finding) {
	file = filepath.Join(display, "install", "slots.yaml")
	data, err := fs.ReadFile(fsys, "install/slots.yaml")
	if err != nil {
		return nil, file, nil
	}

	astFile, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return nil, file, []Finding{malformedYAML(file, err)}
	}
	if len(astFile.Docs) == 0 || astFile.Docs[0].Body == nil {
		return nil, file, nil
	}
	if _, onlyComments := astFile.Docs[0].Body.(*ast.CommentGroupNode); onlyComments {
		// A file holding nothing but a header comment (kbf init's own
		// install/slots.yaml scaffold, before any row exists) parses to
		// exactly this node type, not an empty SequenceNode: legitimately
		// zero rows, not a shape error.
		return nil, file, nil
	}
	seq, ok := astFile.Docs[0].Body.(*ast.SequenceNode)
	if !ok {
		return nil, file, []Finding{{
			Rule: KBF001, File: file, Line: astFile.Docs[0].Body.GetToken().Position.Line,
			Message: "install/slots.yaml must be a bare YAML list of {slot, source} rows",
			Fix:     "remove any wrapper key (e.g. slots:); the file's root must be the list itself",
		}}
	}

	for _, item := range seq.Values {
		var row model.SlotMapping
		if err := yaml.NodeToValue(item, &row); err != nil {
			findings = append(findings, malformedYAML(file, err))
			continue
		}
		slots = append(slots, SlotRow{SlotMapping: row, Line: item.GetToken().Position.Line})
	}
	return slots, file, findings
}

// loadElements walks dir ("ontology" or "evals", fs.FS-relative, always
// slash-separated regardless of host OS) within fsys, recursively and
// sorted so output is deterministic, for *.yaml/*.yml files, and decodes
// every `---`-separated document in each. A missing dir (e.g. a package
// with no evals/ yet) is not an error. display is the package's own
// human-facing root; each file's display path is display/<relative path>,
// built with filepath.Join so it renders with the host's native separator
// even though the fs.FS-relative path underneath is always slash-joined.
func loadElements(fsys fs.FS, dir, display string) ([]Element, []Finding) {
	relPaths := yamlFiles(fsys, dir)
	var elements []Element
	var findings []Finding
	for _, rel := range relPaths {
		displayPath := filepath.Join(display, rel)
		data, err := fs.ReadFile(fsys, rel)
		if err != nil {
			findings = append(findings, malformedYAML(displayPath, err))
			continue
		}
		astFile, err := parser.ParseBytes(data, parser.ParseComments)
		if err != nil {
			findings = append(findings, malformedYAML(displayPath, err))
			continue
		}
		for _, doc := range astFile.Docs {
			if doc.Body == nil {
				continue // a stray leading `---` with nothing after it
			}
			elem, f := decodeDocument(displayPath, doc.Body)
			findings = append(findings, f...)
			if elem != nil {
				elements = append(elements, *elem)
			}
		}
	}
	return elements, findings
}

// yamlFiles returns every .yaml/.yml file under dir within fsys
// (fs.FS-relative paths, e.g. "ontology/entities.yaml"), sorted, or nil
// if dir doesn't exist.
func yamlFiles(fsys fs.FS, dir string) []string {
	var paths []string
	_ = fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // a missing dir just yields zero files
		}
		if ext := filepath.Ext(p); ext == ".yaml" || ext == ".yml" {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	return paths
}
