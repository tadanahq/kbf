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

	"github.com/tadanahq/kbf/tools/internal/lint"
)

var lintFormat string

var lintCmd = &cobra.Command{
	Use:   "lint <path> [path...]",
	Short: "Validate one or more KBF playbooks against the meta-model and semantic rules.",
	Long: `lint loads one or more playbook directories, resolves manifest/builds-on
across all of them together, and checks structure (KBF001-004, KBF010, KBF011,
KBF013) then semantics (KBF005-009, KBF012). Pass a playbook's own path and
every playbook it builds on together: builds-on is resolved by manifest name
across exactly the paths given on this command line, not by filesystem
convention.

Exits 1 if any rule fires.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLint,
}

func init() {
	lintCmd.Flags().StringVar(&lintFormat, "format", "human", `output format: "human" or "json"`)
	rootCmd.AddCommand(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	result, err := lint.Run(args)
	if err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	switch lintFormat {
	case "human":
		if _, err := fmt.Fprint(cmd.OutOrStdout(), lint.RenderHuman(result)); err != nil {
			return err
		}
	case "json":
		out, err := lint.RenderJSON(result)
		if err != nil {
			return fmt.Errorf("render json: %w", err)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(out)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown --format %q: want human or json", lintFormat)
	}

	if len(result.Findings) > 0 {
		return exitError{code: 1}
	}
	return nil
}
