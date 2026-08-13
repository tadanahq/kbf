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

// Command kbf is the reference CLI for the KBF Ontology Spec: v0 is
// config-phase tooling only (schema generation and linting). See
// project-overview.md for what's in and out of v0.
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	err := rootCmd.Execute()
	if err == nil {
		return
	}
	var exit exitError
	if errors.As(err, &exit) {
		os.Exit(exit.code)
	}
	// Anything that isn't an explicit exitError is unexpected (bad flags,
	// I/O failure): print it, since no command printed it already.
	fmt.Fprintln(os.Stderr, "kbf:", err)
	os.Exit(1)
}
