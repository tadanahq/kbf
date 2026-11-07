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

import "github.com/spf13/cobra"

// exitError lets a command set a specific process exit code (e.g. 1 for
// "lint found violations") without cobra printing its own "Error: ..." on
// top of the command's already-rendered output. Every command prints
// whatever the user should see itself (findings table, JSON, a plain
// message); it returns exitError only to carry the exit code back to main.
type exitError struct{ code int }

func (e exitError) Error() string { return "" }

var rootCmd = &cobra.Command{
	Use:   "kbf",
	Short: "kbf validates and generates artifacts for the KBF Ontology Spec.",
	Long: `kbf is the reference tool for the KBF Ontology Spec.

v0 is config-phase tooling: it helps an authoring agent, or a human, create
and validate a KBF ontology playbook. It never touches a business's live data
and runs no queries: files in, files out. Runtime enforcement (validating
agent-generated queries against an ontology) is on the roadmap, not in v0.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}
