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

// Command embedsync is tools/'s batteries-included data pipeline: the
// copy-and-check pattern project-decisions.md's "batteries-included
// binary" entry calls for, mirrored from how `kbf schema` already
// generates schema/ and verifies it with --check. go:embed cannot reach
// outside the tools/ Go module, so the public core playbooks
// (playbooks/core-*), the authoring skill (skills/kbf-authoring/), and
// the prose spec (spec/*.md, spec/primitives/*.md) all live at the repo
// root, one level above tools/: this program is what copies them in.
// Run without flags to sync tools/internal/embedded/data/ from the repo
// root; run with --check to verify the committed copy still matches
// (CI's freshness gate, wired into `make check` as `embed-freshness`,
// same shape as schema-freshness).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := flag.String("root", ".", "repo root (source of truth for what gets embedded)")
	check := flag.Bool("check", false, "verify tools/internal/embedded/data matches source; do not write")
	flag.Parse()

	files, err := sourceFiles(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "embedsync:", err)
		os.Exit(2)
	}
	dataDir := filepath.Join(*root, "tools", "internal", "embedded", "data")

	if *check {
		stale, err := checkFiles(dataDir, files)
		if err != nil {
			fmt.Fprintln(os.Stderr, "embedsync:", err)
			os.Exit(2)
		}
		if len(stale) > 0 {
			for _, s := range stale {
				fmt.Fprintln(os.Stderr, s)
			}
			fmt.Fprintln(os.Stderr, "run `make embed-sync` to regenerate")
			os.Exit(1)
		}
		fmt.Println("embedded data is current")
		return
	}

	if err := writeFiles(dataDir, files); err != nil {
		fmt.Fprintln(os.Stderr, "embedsync:", err)
		os.Exit(2)
	}
	fmt.Printf("synced %d file(s) into %s\n", len(files), dataDir)
}
