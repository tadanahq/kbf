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
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/schemagen"
)

var (
	schemaCheck bool
	schemaOut   string
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate the published JSON Schema files from internal/model.",
	Long: `schema reflects the canonical Go structs in internal/model into the two
published schema files (ontology.schema.yaml, manifest.schema.yaml).

Run without flags to (re)write them. Run with --check to verify the
committed files still match what internal/model would generate: this is
CI's freshness gate, and exits 1 naming any file that has drifted.`,
	RunE: runSchema,
}

func init() {
	schemaCmd.Flags().BoolVar(&schemaCheck, "check", false, "verify committed schema files match generated output; do not write")
	schemaCmd.Flags().StringVar(&schemaOut, "out", "schema", "directory the schema files are written to (or checked against)")
	rootCmd.AddCommand(schemaCmd)
}

func runSchema(cmd *cobra.Command, _ []string) error {
	files, err := schemagen.Generate()
	if err != nil {
		return fmt.Errorf("generate schema: %w", err)
	}

	if schemaCheck {
		return checkSchema(cmd, files)
	}
	return writeSchema(cmd, files)
}

func writeSchema(cmd *cobra.Command, files []schemagen.File) error {
	if err := os.MkdirAll(schemaOut, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", schemaOut, err)
	}
	for _, f := range files {
		path := filepath.Join(schemaOut, f.Name)
		if err := os.WriteFile(path, f.Bytes, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "wrote", path); err != nil {
			return err
		}
	}
	return nil
}

func checkSchema(cmd *cobra.Command, files []schemagen.File) error {
	stale := false
	for _, f := range files {
		path := filepath.Join(schemaOut, f.Name)
		committed, err := os.ReadFile(path)
		if err != nil {
			if _, printErr := fmt.Fprintf(cmd.ErrOrStderr(), "%s: missing (%v)\n", path, err); printErr != nil {
				return printErr
			}
			stale = true
			continue
		}
		if !bytes.Equal(committed, f.Bytes) {
			if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s: stale, does not match internal/model\n", path); err != nil {
				return err
			}
			stale = true
		}
	}
	if stale {
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "run `kbf schema` to regenerate"); err != nil {
			return err
		}
		return exitError{code: 1}
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "schema is current"); err != nil {
		return err
	}
	return nil
}
