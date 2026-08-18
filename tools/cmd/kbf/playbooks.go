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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tadanahq/kbf/tools/internal/embedded"
)

var (
	pinTo    string
	pinForce bool
)

var playbooksCmd = &cobra.Command{
	Use:   "playbooks",
	Short: "Manage the playbooks yours builds on.",
}

var playbooksPinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin the embedded core playbooks to disk.",
	Long: `pin writes the playbooks yours can build on (core-business,
core-operations, core-services) to real directories under --to, exactly
as embedded: a DO-NOT-EDIT header and all, since a pinned copy is meant
to be inspected and, if you choose to keep editing it locally from here,
treated as your own from that point on. "Embedded" never means "hidden":
this command is how you get the same content lint/coverage/compile fall
back to onto disk, where you can read, diff, or fork it.

Refuses to overwrite an existing directory; pass --force to replace it.`,
	Args: cobra.NoArgs,
	RunE: runPlaybooksPin,
}

func init() {
	playbooksPinCmd.Flags().StringVar(&pinTo, "to", "playbooks", "directory the pinned playbooks are written under")
	playbooksPinCmd.Flags().BoolVar(&pinForce, "force", false, "overwrite existing directories")
	playbooksCmd.AddCommand(playbooksPinCmd)
	rootCmd.AddCommand(playbooksCmd)
}

func runPlaybooksPin(cmd *cobra.Command, _ []string) error {
	names := embedded.CorePlaybookNames()

	if !pinForce {
		var existing []string
		for _, name := range names {
			if _, err := os.Stat(filepath.Join(pinTo, name)); err == nil {
				existing = append(existing, filepath.Join(pinTo, name))
			}
		}
		if len(existing) > 0 {
			return fmt.Errorf("playbooks pin: already exists: %v (pass --force to overwrite)", existing)
		}
	}

	for _, name := range names {
		fsys, ok := embedded.Playbook(name)
		if !ok {
			return fmt.Errorf("playbooks pin: %q is not embedded (this should not happen: CorePlaybookNames and Playbook disagree)", name)
		}
		dest := filepath.Join(pinTo, name)
		if err := writeTree(fsys, dest); err != nil {
			return fmt.Errorf("pin %s: %w", name, err)
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "wrote", dest); err != nil {
			return err
		}
	}
	return nil
}

// writeTree copies every file in fsys to dest, preserving its relative
// path, creating directories as needed.
func writeTree(fsys fs.FS, dest string) error {
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dest, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
