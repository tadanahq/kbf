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
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	fileStyle    = lipgloss.NewStyle().Bold(true)
	ruleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	headerStyle  = lipgloss.NewStyle().Bold(true)
	dimStyle     = lipgloss.NewStyle().Faint(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
)

// RenderHuman renders a Result as one lipgloss table per file (design.md:
// "lipgloss table grouped by file"), each row a finding: line, rule,
// element, message, fix. Findings arrive pre-sorted by file then line
// (Result comes from Run, which calls sortFindings), so grouping only
// needs to notice where the file changes, never re-sort. When
// r.EmbeddedUsed is non-empty (RunWithEmbedded resolved at least one
// builds-on name from the embedded core playbooks rather than a local
// path), a footer line names which: --format json carries no equivalent
// field, by design, so this is the only place that information surfaces.
func RenderHuman(r Result) string {
	if len(r.Findings) == 0 {
		return successStyle.Render("kbf lint: no findings") + "\n" + embeddedFooter(r.EmbeddedUsed)
	}

	groups := groupByFile(r.Findings)
	var b strings.Builder
	for _, g := range groups {
		b.WriteString(fileStyle.Render(g.file))
		b.WriteString("\n")
		b.WriteString(renderTable(g.findings))
		b.WriteString("\n\n")
	}
	b.WriteString(dimStyle.Render(fmt.Sprintf("%d finding(s) across %d file(s)", len(r.Findings), len(groups))))
	b.WriteString("\n")
	b.WriteString(embeddedFooter(r.EmbeddedUsed))
	return b.String()
}

// embeddedFooter renders the "resolved from embedded: ..." line, or ""
// when names is empty: no trailing blank line to keep clean of, since
// callers already end their own output with "\n" before appending this.
func embeddedFooter(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return dimStyle.Render("resolved from embedded: "+strings.Join(names, ", ")) + "\n"
}

// renderTable builds one file's findings table.
func renderTable(findings []Finding) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(dimStyle).
		Headers("LINE", "RULE", "ELEMENT", "MESSAGE", "FIX").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col == 1 {
				return ruleStyle
			}
			return lipgloss.NewStyle()
		})
	for _, f := range findings {
		t.Row(strconv.Itoa(f.Line), f.Rule, f.Element, f.Message, f.Fix)
	}
	return t.Render()
}

// fileGroup is one file's consecutive run of findings.
type fileGroup struct {
	file     string
	findings []Finding
}

// groupByFile splits an already file-sorted findings slice into
// consecutive per-file runs, preserving order.
func groupByFile(findings []Finding) []fileGroup {
	var groups []fileGroup
	for _, f := range findings {
		if len(groups) == 0 || groups[len(groups)-1].file != f.File {
			groups = append(groups, fileGroup{file: f.File})
		}
		groups[len(groups)-1].findings = append(groups[len(groups)-1].findings, f)
	}
	return groups
}
