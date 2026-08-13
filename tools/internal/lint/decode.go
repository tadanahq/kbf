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
	"reflect"
	"strings"

	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"

	"github.com/tadanahq/kbf/tools/internal/model"
)

// decodeDocument dispatches one YAML document to its target struct by
// `kind:`. It returns nil (no element) alongside findings when the
// document can't be classified or decoded at all; unknown fields are
// reported but do not by themselves prevent decoding the rest.
func decodeDocument(file string, body ast.Node) (*Element, []Finding) {
	m, ok := body.(*ast.MappingNode)
	if !ok {
		line := body.GetToken().Position.Line
		return nil, []Finding{{
			Rule: KBF001, File: file, Line: line,
			Message: fmt.Sprintf("expected a YAML mapping for an element document, got %s", body.Type()),
			Fix:     "each `---`-separated document must be a single `kind:`-tagged mapping",
		}}
	}

	kindNode, hasKind := findKey(m, "kind")
	line := m.GetToken().Position.Line
	if hasKind {
		line = kindNode.Key.GetToken().Position.Line
	}
	if !hasKind {
		return nil, []Finding{{Rule: KBF001, File: file, Line: line, Message: "document has no `kind` field", Fix: "add kind: entity|relation|metric|action|competency-question"}}
	}
	kindStr, _ := scalarString(kindNode.Value)

	var target any
	switch model.Kind(kindStr) {
	case model.KindEntity:
		target = &model.Entity{}
	case model.KindRelation:
		target = &model.Relation{}
	case model.KindMetric:
		target = &model.Metric{}
	case model.KindAction:
		target = &model.Action{}
	case model.KindCompetencyQuestion:
		target = &model.CompetencyQuestion{}
	default:
		return nil, []Finding{{
			Rule: KBF002, File: file, Line: line, Element: kindStr,
			Message: fmt.Sprintf("unknown kind %q", kindStr),
			Fix:     "kind must be one of entity, relation, metric, action, competency-question",
		}}
	}

	findings := unknownFieldFindings(file, m, target)
	if err := yaml.NodeToValue(m, target); err != nil {
		return nil, append(findings, malformedYAML(file, err))
	}
	return &Element{Kind: model.Kind(kindStr), File: file, Line: line, Node: m, Value: target}, findings
}

// unknownFieldFindings reports KBF001 for every mapping key that isn't one
// of target's yaml-tagged fields, each at that key's own position. Done by
// walking the AST directly (not goccy's DisallowUnknownField) so every
// unknown key is reported in one pass, each with its own line, instead of
// stopping at the first.
func unknownFieldFindings(file string, m *ast.MappingNode, target any) []Finding {
	known := yamlFieldNames(reflect.TypeOf(target).Elem())
	var findings []Finding
	for _, v := range m.Values {
		key, ok := scalarString(v.Key)
		if !ok || known[key] {
			continue
		}
		findings = append(findings, Finding{
			Rule: KBF001, File: file, Line: v.Key.GetToken().Position.Line, Element: key,
			Message: fmt.Sprintf("unknown field %q", key),
			Fix:     "remove it, or check spelling against the primitive's fields",
		})
	}
	return findings
}

// yamlFieldNames returns the set of a struct's top-level `yaml:` tag names.
func yamlFieldNames(t reflect.Type) map[string]bool {
	names := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("yaml")
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			names[name] = true
		}
	}
	return names
}

// findKey returns the mapping's *ast.MappingValueNode for key, if present.
func findKey(m *ast.MappingNode, key string) (*ast.MappingValueNode, bool) {
	for _, v := range m.Values {
		if s, ok := scalarString(v.Key); ok && s == key {
			return v, true
		}
	}
	return nil, false
}

// scalarString extracts a plain string value from a scalar YAML node
// (string, or any other scalar rendered as text: `kind: entity` and
// `kind: "entity"` must be treated alike).
func scalarString(n ast.Node) (string, bool) {
	if s, ok := n.(*ast.StringNode); ok {
		return s.Value, true
	}
	return "", false
}

// malformedYAML wraps a parse or decode error as a KBF001 finding,
// extracting a line number when the error carries position info.
func malformedYAML(file string, err error) Finding {
	return Finding{
		Rule: KBF001, File: file, Line: errorLine(err, 1),
		Message: "malformed YAML: " + err.Error(),
		Fix:     "fix the YAML syntax so the document parses",
	}
}

// errorLine extracts a line number from goccy's position-aware error
// types, falling back when the error carries none.
func errorLine(err error, fallback int) int {
	switch e := err.(type) {
	case *yaml.SyntaxError:
		if e.Token != nil {
			return e.Token.Position.Line
		}
	case *yaml.TypeError:
		if e.Token != nil {
			return e.Token.Position.Line
		}
	case *yaml.UnknownFieldError:
		if e.Token != nil {
			return e.Token.Position.Line
		}
	}
	return fallback
}
