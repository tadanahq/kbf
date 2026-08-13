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

// Package schemagen turns internal/model's canonical structs into the
// published JSON Schema files under schema/. The struct is the single
// source of truth: field doc comments become schema descriptions (via
// jsonschema.Reflector.AddGoComments), so there is exactly one place to
// update a field's meaning. `kbf schema --check` (cmd/kbf) diffs this
// package's output against the committed files; CI fails on drift.
package schemagen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/invopop/jsonschema"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// modulePath must match go.mod's module line: it is the import-path prefix
// AddGoComments needs to match reflected types back to their source
// comments (see reflect_comments.go in invopop/jsonschema).
const modulePath = "github.com/tadanahq/kbf/tools"

// baseSchemaID is the URL prefix generated schemas publish under. Points at
// the file as it will actually be servable (raw.githubusercontent.com, not
// github.com, which serves HTML) once this repo is public.
const baseSchemaID = "https://raw.githubusercontent.com/tadanahq/kbf/main/schema/"

// File is one generated schema file: its name under schema/ and rendered
// YAML bytes.
type File struct {
	Name  string
	Bytes []byte
}

// Generate reflects internal/model into the two published schema files.
func Generate() ([]File, error) {
	ontology, err := ontologySchema()
	if err != nil {
		return nil, fmt.Errorf("ontology schema: %w", err)
	}
	manifest, err := manifestSchema()
	if err != nil {
		return nil, fmt.Errorf("manifest schema: %w", err)
	}
	return []File{
		{Name: "ontology.schema.yaml", Bytes: ontology},
		{Name: "manifest.schema.yaml", Bytes: manifest},
	}, nil
}

// newReflector returns a Reflector wired to internal/model's source so
// struct and field doc comments flow into the generated descriptions.
func newReflector() (*jsonschema.Reflector, error) {
	base, dir, err := modelImportComponents()
	if err != nil {
		return nil, err
	}
	r := &jsonschema.Reflector{BaseSchemaID: jsonschema.ID(baseSchemaID)}
	if err := r.AddGoComments(base, dir, jsonschema.WithFullComment()); err != nil {
		return nil, fmt.Errorf("reading model doc comments: %w", err)
	}
	// Go doc comments wrap at ~77 columns; that wrapping is source
	// formatting, not schema content. Collapse it so descriptions read as
	// single flowed paragraphs in an editor tooltip instead of carrying
	// literal newlines into the YAML.
	for k, v := range r.CommentMap {
		r.CommentMap[k] = strings.Join(strings.Fields(v), " ")
	}
	return r, nil
}

// modelImportComponents splits internal/model's true import path into a
// (base, dir) pair that AddGoComments can recombine correctly regardless of
// the process's current working directory.
//
// AddGoComments walks `dir` (resolved relative to the process's actual
// CWD by filepath.Walk) and keys extracted comments by path.Join(base,
// dir-of-each-walked-file). For that key to match reflect.Type.PkgPath()
// at lookup time, base+dir must reassemble to the real import path exactly
// - so dir must be CWD-relative, and base must be whatever's left after
// removing dir's suffix from the import path. `kbf schema` can run from
// repo root or from tools/ (Makefile vs `go run` during development), so
// this is computed, not hardcoded, and fails loudly rather than silently
// generating schema with empty descriptions if the two ever can't align.
func modelImportComponents() (base, dir string, err error) {
	_, thisFile, _, _ := runtime.Caller(0)
	absModelDir := filepath.Join(filepath.Dir(thisFile), "..", "model")

	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("getwd: %w", err)
	}
	rel, err := filepath.Rel(cwd, absModelDir)
	if err != nil {
		return "", "", fmt.Errorf("relativize model dir: %w", err)
	}
	rel = filepath.ToSlash(rel)

	fullImportPath := modulePath + "/internal/model"
	if !strings.HasSuffix(fullImportPath, rel) {
		return "", "", fmt.Errorf("cannot align model dir %q (relative to cwd %q) with import path %q; run kbf from the repo root or tools/", rel, cwd, fullImportPath)
	}
	base = strings.TrimSuffix(strings.TrimSuffix(fullImportPath, rel), "/")
	return base, rel, nil
}

// ontologySchema builds the oneOf union of the five element kinds. Go has
// no union type, so this is assembled by hand from five independent
// reflections rather than reflecting a single Go type: see design.md's
// "Implementation clarifications" for why the union lives here, not in
// internal/model.
func ontologySchema() ([]byte, error) {
	r, err := newReflector()
	if err != nil {
		return nil, err
	}

	elements := []any{
		&model.Entity{},
		&model.Relation{},
		&model.Metric{},
		&model.Action{},
		&model.CompetencyQuestion{},
	}

	defs := jsonschema.Definitions{}
	oneOf := make([]*jsonschema.Schema, 0, len(elements))
	for _, e := range elements {
		s := r.Reflect(e)
		for name, def := range s.Definitions {
			defs[name] = def
		}
		oneOf = append(oneOf, &jsonschema.Schema{Ref: s.Ref})
	}

	root := &jsonschema.Schema{
		Version:     jsonschema.Version,
		ID:          jsonschema.ID(baseSchemaID + "ontology.schema.yaml"),
		Title:       "KBF Ontology Element",
		Description: "One KBF ontology element document: entity, relation, metric, action, or competency-question, discriminated by `kind`.",
		OneOf:       oneOf,
		Definitions: defs,
	}
	return toYAML(root)
}

// manifestSchema publishes Manifest as the file's root shape (matching
// manifest.yaml) and folds SlotMapping's definition in alongside it, since
// both describe package installation config. SlotMapping is not itself a
// root schema in v0: install/slots.yaml has no direct yaml-language-server
// binding yet (documented gap, see design.md).
func manifestSchema() ([]byte, error) {
	r, err := newReflector()
	if err != nil {
		return nil, err
	}

	manifest := r.Reflect(&model.Manifest{})
	slot := r.Reflect(&model.SlotMapping{})
	for name, def := range slot.Definitions {
		manifest.Definitions[name] = def
	}

	manifest.ID = jsonschema.ID(baseSchemaID + "manifest.schema.yaml")
	manifest.Title = "KBF Package Manifest"
	manifest.Description = "Identity and extension declaration for a KBF package (manifest.yaml)."
	return toYAML(manifest)
}

// toYAML renders a schema through JSON first: invopop/jsonschema marshals
// itself via `json` tags, and goccy/go-yaml's JSONToYAML keeps key order
// stable (encoding/json sorts map keys; struct fields keep declaration
// order), so output is reproducible run to run: required for --check.
func toYAML(s *jsonschema.Schema) ([]byte, error) {
	j, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal schema to json: %w", err)
	}
	y, err := goyaml.JSONToYAML(j)
	if err != nil {
		return nil, fmt.Errorf("convert schema json to yaml: %w", err)
	}
	return y, nil
}
