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
	"bytes"
	"encoding/json"
)

// jsonReport is the stable `--format json` shape from design.md:
// `{rules: [{id, file, line, element, message, fix}]}`. Authoring agents
// parse this; it does not change casually once published.
type jsonReport struct {
	Rules []Finding `json:"rules"`
}

// RenderJSON renders a Result as the stable JSON contract. Findings is
// always a real (possibly empty) slice, never nil, so a clean run renders
// `{"rules":[]}` and not `{"rules":null}`: a parser should not have to
// special-case the zero-findings case. Uses an Encoder with HTML-escaping
// off: fix hints routinely contain `<key>`-style placeholders, and
// json.Marshal's default HTML-safe escaping would turn every `<`/`>` into
// `<`/`>`, which is meant for embedding JSON in HTML, not for a
// CLI's output.
func RenderJSON(r Result) ([]byte, error) {
	rules := r.Findings
	if rules == nil {
		rules = []Finding{}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(jsonReport{Rules: rules}); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
