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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/embedded"
)

var (
	initBuildsOn []string
	initLayer    string
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Scaffold a new playbook: manifest.yaml, empty ontology/evals, a template slots.yaml.",
	Long: `init creates <name>/ with a manifest.yaml (spec: v0, version: 0.1.0),
empty ontology/ and evals/ directories, and an install/slots.yaml template,
the starting point spec/onboarding.md's step 1 describes: pick the
composition closure, then start the competency-question interview before
modeling anything.

--layer defaults to "vertical" (a business-specific playbook, the normal
case for a new business's own ontology); pass --layer core only when you
are authoring a new foundation playbook other playbooks will compose. A
vertical must build on at least one playbook, so --builds-on is required
unless --layer core is given with an empty closure on purpose (a new
root). Refuses to run if <name> already exists: init never overwrites.

The scaffold lints clean against kbf's embedded core playbooks with no
other paths on the command line, e.g.:

  kbf init my-business --builds-on core-operations
  kbf lint my-business`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringSliceVar(&initBuildsOn, "builds-on", nil, "comma-separated playbook names this one composes")
	initCmd.Flags().StringVar(&initLayer, "layer", "vertical", `"core" or "vertical"`)
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]

	if initLayer != "core" && initLayer != "vertical" {
		return fmt.Errorf("init: --layer %q is not core or vertical", initLayer)
	}
	if initLayer == "vertical" && len(initBuildsOn) == 0 {
		return fmt.Errorf("init: --layer vertical needs --builds-on (a vertical must build on at least one playbook)\n"+
			"embedded cores available: %s\n"+
			"pass one or more, e.g. --builds-on %s", strings.Join(embedded.CorePlaybookNames(), ", "), embedded.CorePlaybookNames()[0])
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("init: %s already exists; init never overwrites", name)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("init: %s: %w", name, err)
	}

	if err := scaffold(name, initBuildsOn, initLayer); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "wrote", name); err != nil {
		return err
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "next: kbf lint %s\n", name)
	return err
}

// scaffold writes name/manifest.yaml, name/ontology/, name/evals/, and
// name/install/slots.yaml. buildsOn is rendered as a flow-style YAML
// list (`[a, b]`, or `[]` for none) rather than block style, since an
// empty block list has no clean YAML spelling and a scaffold with zero
// entries (a fresh root core playbook) is a real, intended case.
func scaffold(dir string, buildsOn []string, layer string) error {
	for _, sub := range []string{"ontology", "evals", "install"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return err
		}
	}

	manifest := fmt.Sprintf(
		"name: %s\nversion: 0.1.0\nspec: v0\nbuilds-on: [%s]\nlayer: %s\n",
		dir, strings.Join(buildsOn, ", "), layer,
	)
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0o644); err != nil {
		return err
	}

	slots := `# One row per attribute slot in this playbook's composed closure:
# {slot: <domain.concept>, source: <system, or "" if not yet mapped>}.
# kbf coverage fills in the declared/mapped picture once ontology/
# attributes exist; run it after the entity interview (spec/onboarding.md
# step 7), not before, since there is nothing to map yet on a fresh init.
`
	return os.WriteFile(filepath.Join(dir, "install", "slots.yaml"), []byte(slots), 0o644)
}
