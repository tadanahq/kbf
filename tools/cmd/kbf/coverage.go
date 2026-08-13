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

	"github.com/tadanahq/kbf/tools/internal/coverage"
	"github.com/tadanahq/kbf/tools/internal/lint"
)

var coverageFormat string

var coverageCmd = &cobra.Command{
	Use:   "coverage <path> [path...]",
	Short: "Report static slot-mapping completeness: declared, mapped, unmapped.",
	Long: `coverage loads one or more package directories the same way lint does
(extends resolved by manifest name across exactly the paths given), then
reports install/slots.yaml completeness for each leaf package: a package
given only as extends context (e.g. universal-core alongside cafe-demo) is
not itself reported on, since its slots.yaml is a template by definition.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCoverage,
}

func init() {
	coverageCmd.Flags().StringVar(&coverageFormat, "format", "human", `output format: "human" or "json"`)
	rootCmd.AddCommand(coverageCmd)
}

func runCoverage(cmd *cobra.Command, args []string) error {
	universe, _, err := lint.Load(args)
	if err != nil {
		return fmt.Errorf("coverage: %w", err)
	}
	reports := coverage.Compute(universe)

	switch coverageFormat {
	case "human":
		if _, err := fmt.Fprint(cmd.OutOrStdout(), coverage.RenderHuman(reports)); err != nil {
			return err
		}
	case "json":
		out, err := coverage.RenderJSON(reports)
		if err != nil {
			return fmt.Errorf("render json: %w", err)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(out)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown --format %q: want human or json", coverageFormat)
	}
	return nil
}
