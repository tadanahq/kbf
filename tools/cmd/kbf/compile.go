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

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/compile"
	"github.com/tadanahq/kbf/tools/internal/lint"
)

var compileTo string

var compileCmd = &cobra.Command{
	Use:   "compile <path> [path...]",
	Short: "Render an ontology's map: entities, relations, and actions.",
	Long: `compile loads one or more playbook directories the same way lint does, then
renders the union of everything loaded: entities as nodes, relations as
labeled edges, actions as annotations on the entity they target. v0 ships
one emitter: --to mermaid. Output is deterministic (sorted) and is plain
mermaid source, valid as-is inside a GitHub Markdown ` + "```mermaid" + ` fence.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCompile,
}

func init() {
	compileCmd.Flags().StringVar(&compileTo, "to", "mermaid", `emitter: "mermaid" (the only one v0 ships)`)
	rootCmd.AddCommand(compileCmd)
}

func runCompile(cmd *cobra.Command, args []string) error {
	if compileTo != "mermaid" {
		return fmt.Errorf("unknown --to %q: v0 ships mermaid only", compileTo)
	}

	universe, _, err := lint.Load(args)
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	graph := compile.BuildGraph(universe)

	if _, err := fmt.Fprint(cmd.OutOrStdout(), compile.ToMermaid(graph)); err != nil {
		return err
	}
	return nil
}
