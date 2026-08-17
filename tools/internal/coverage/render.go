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

package coverage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	packageStyle  = lipgloss.NewStyle().Bold(true)
	headerStyle   = lipgloss.NewStyle().Bold(true)
	mappedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	unmappedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle      = lipgloss.NewStyle().Faint(true)
)

// RenderHuman renders one lipgloss table per package, each row a slot,
// followed by a declared/mapped summary line. When embeddedUsed is
// non-empty (the universe resolved at least one builds-on name from the
// embedded core playbooks rather than a local path: lint.LoadWithEmbedded,
// see cmd/kbf's runCoverage), a footer line names which: --format json
// carries no equivalent field, by design, so this is the only place that
// information surfaces.
func RenderHuman(reports []Report, embeddedUsed []string) string {
	if len(reports) == 0 {
		return dimStyle.Render("kbf coverage: nothing to report (no leaf package in the given paths)") + "\n" + embeddedFooter(embeddedUsed)
	}

	var b strings.Builder
	for _, r := range reports {
		b.WriteString(packageStyle.Render(r.Package))
		b.WriteString("\n")
		b.WriteString(renderTable(r))
		b.WriteString("\n")
		pct := 0
		if r.Declared > 0 {
			pct = r.Mapped * 100 / r.Declared
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d/%d slots mapped (%d%%)", r.Mapped, r.Declared, pct)))
		b.WriteString("\n\n")
	}
	b.WriteString(embeddedFooter(embeddedUsed))
	return b.String()
}

// embeddedFooter renders the "resolved from embedded: ..." line, or ""
// when names is empty. Mirrors internal/lint's own embeddedFooter
// (render_human.go); not shared code, since sharing it would mean one of
// these two packages importing the other for a single one-line format
// string, a worse trade than the tiny duplication.
func embeddedFooter(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return dimStyle.Render("resolved from embedded: "+strings.Join(names, ", ")) + "\n"
}

func renderTable(r Report) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(dimStyle).
		Headers("ENTITY", "SLOT", "SOURCE", "STATUS").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			// Data rows are 0-indexed (HeaderRow is -1, not a row 0 data
			// rows start after): r.Rows[row], not r.Rows[row-1].
			if col == 3 && row >= 0 && row < len(r.Rows) {
				if r.Rows[row].Mapped {
					return mappedStyle
				}
				return unmappedStyle
			}
			return lipgloss.NewStyle()
		})
	for _, row := range r.Rows {
		status := "unmapped"
		if row.Mapped {
			status = "mapped"
		}
		t.Row(row.Entity, row.Slot, row.Source, status)
	}
	return t.Render()
}

// RenderJSON renders reports as the stable `--format json` shape:
// `{"packages": [{package, rows, declared, mapped}, ...]}`. Reports (and
// each report's Rows) are always real, possibly-empty slices, never nil,
// so a package with no slots renders `"rows":[]`, not `"rows":null`.
func RenderJSON(reports []Report) ([]byte, error) {
	if reports == nil {
		reports = []Report{}
	}
	for i := range reports {
		if reports[i].Rows == nil {
			reports[i].Rows = []Row{}
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Packages []Report `json:"packages"`
	}{Packages: reports}); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
