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

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/embedded"
)

// skillName is the one skill v0 ships. install takes no argument (unlike
// vendor and docs, which name their targets): a bare kbf-authoring is the
// whole skill catalog today, so there is nothing to disambiguate yet.
const skillName = "kbf-authoring"

var skillInstallForce bool

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the kbf-authoring skill.",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Write the embedded kbf-authoring skill into .claude/skills/.",
	Long: `install writes the embedded kbf-authoring skill to ./.claude/skills/kbf-authoring/,
where Claude Code (and any agent runtime that reads the same convention)
picks it up automatically. No clone needed: the skill's own prerequisites
section reads spec/onboarding.md and spec/cli.md through "kbf docs", not
from a checked-out copy.

Refuses to overwrite an existing install; pass --force to replace it.`,
	Args: cobra.NoArgs,
	RunE: runSkillInstall,
}

func init() {
	skillInstallCmd.Flags().BoolVar(&skillInstallForce, "force", false, "overwrite an existing install")
	skillCmd.AddCommand(skillInstallCmd)
	rootCmd.AddCommand(skillCmd)
}

func runSkillInstall(cmd *cobra.Command, _ []string) error {
	dest := filepath.Join(".claude", "skills", skillName)

	if !skillInstallForce {
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("skill install: %s already exists (pass --force to overwrite)", dest)
		}
	}

	fsys, ok := embedded.Skill(skillName)
	if !ok {
		return fmt.Errorf("skill install: %q is not embedded (this should not happen)", skillName)
	}
	if err := writeTree(fsys, dest); err != nil {
		return fmt.Errorf("skill install: %w", err)
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "wrote", filepath.Join(dest, "SKILL.md")); err != nil {
		return err
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), "next: point your agent at this project and ask it to raise a new ontology; the skill reads spec/onboarding.md and spec/cli.md via `kbf docs`, no clone needed")
	return err
}
