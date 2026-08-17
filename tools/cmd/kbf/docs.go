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
	"strings"

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/embedded"
)

var docsCmd = &cobra.Command{
	Use:   "docs [name]",
	Short: "Read the embedded spec, no clone required.",
	Long: `docs serves the prose spec straight out of the binary: with no argument, it
lists every embedded doc's name; with a name, it prints that doc to
stdout. Names are a doc's path under spec/ without the .md extension, e.g.
"onboarding", "cli", "primitives/entity".

This is how an agent or a human reads spec/onboarding.md, spec/cli.md, or
any primitive doc without cloning this repository: pipe the output
anywhere (a pager, a file, another tool's stdin).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func runDocs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		for _, name := range embedded.DocNames() {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), name); err != nil {
				return err
			}
		}
		return nil
	}

	name := args[0]
	content, ok := embedded.Doc(name)
	if !ok {
		return fmt.Errorf("docs: %q is not an embedded doc\navailable: %s", name, strings.Join(embedded.DocNames(), ", "))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), content)
	return err
}
