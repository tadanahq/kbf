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
	"os"
	"path/filepath"
	"sort"

	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// loadPackage reads one package directory: manifest.yaml, everything under
// ontology/ and evals/ (position-aware, KBF-coded on trouble), and
// install/slots.yaml. It returns a genuine error only when root itself
// isn't a readable directory; every content problem becomes a Finding so
// one bad file never aborts linting the rest of the package. Per-document
// decoding (turning one AST node into a typed Element) lives in decode.go.
func loadPackage(root string) (*Package, []Finding, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("%s: not a directory", root)
	}

	pkg := &Package{Root: root}
	var findings []Finding

	manifest, mFile, mLine, mFindings := loadManifest(root)
	pkg.Manifest, pkg.ManifestFile, pkg.ManifestLine = manifest, mFile, mLine
	findings = append(findings, mFindings...)

	for _, dir := range []string{"ontology", "evals"} {
		elems, f := loadElements(filepath.Join(root, dir))
		pkg.Elements = append(pkg.Elements, elems...)
		findings = append(findings, f...)
	}

	slots, sFile, sFindings := loadSlots(root)
	pkg.Slots, pkg.SlotsFile = slots, sFile
	findings = append(findings, sFindings...)

	return pkg, findings, nil
}

// loadManifest reads manifest.yaml. A missing file is KBF011, not a Go
// error: the caller still gets a Package back (with Manifest == nil) so it
// can keep loading the rest and report every problem in one lint run.
func loadManifest(root string) (m *model.Manifest, file string, line int, findings []Finding) {
	file = filepath.Join(root, "manifest.yaml")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, file, 1, []Finding{{
			Rule: KBF011, File: file, Line: 1, Element: root,
			Message: "manifest.yaml is missing or unreadable: " + err.Error(),
			Fix:     "add a manifest.yaml with name, version, spec, builds-on, and layer",
		}}
	}

	astFile, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return nil, file, errorLine(err, 1), []Finding{malformedYAML(file, err)}
	}
	if len(astFile.Docs) == 0 || astFile.Docs[0].Body == nil {
		return nil, file, 1, []Finding{{Rule: KBF011, File: file, Line: 1, Element: root, Message: "manifest.yaml is empty", Fix: "add name, version, spec, builds-on, and layer"}}
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
func loadSlots(root string) (slots []SlotRow, file string, findings []Finding) {
	file = filepath.Join(root, "install", "slots.yaml")
	data, err := os.ReadFile(file)
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

// loadElements walks dir (recursively, sorted, so output is deterministic)
// for *.yaml/*.yml files and decodes every `---`-separated document in
// each. A missing dir (e.g. a package with no evals/ yet) is not an error.
func loadElements(dir string) ([]Element, []Finding) {
	paths := yamlFiles(dir)
	var elements []Element
	var findings []Finding
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			findings = append(findings, malformedYAML(path, err))
			continue
		}
		astFile, err := parser.ParseBytes(data, parser.ParseComments)
		if err != nil {
			findings = append(findings, malformedYAML(path, err))
			continue
		}
		for _, doc := range astFile.Docs {
			if doc.Body == nil {
				continue // a stray leading `---` with nothing after it
			}
			elem, f := decodeDocument(path, doc.Body)
			findings = append(findings, f...)
			if elem != nil {
				elements = append(elements, *elem)
			}
		}
	}
	return elements, findings
}

// yamlFiles returns every .yaml/.yml file under dir, sorted, or nil if dir
// doesn't exist.
func yamlFiles(dir string) []string {
	var paths []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // a missing dir just yields zero files
		}
		if ext := filepath.Ext(path); ext == ".yaml" || ext == ".yml" {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	return paths
}
