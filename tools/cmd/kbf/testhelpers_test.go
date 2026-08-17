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
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "write golden files instead of comparing against them")

// testdataDir is this package's own testdata/ directory, captured once
// at load time (before any test calls t.Chdir, which several here do:
// init/vendor/skill install all act relative to CWD by design, the same
// way the real binary does). Golden-file paths are built from this, not
// from a bare "testdata/..." relative path, so they still resolve after
// the working directory has moved to a t.TempDir().
var testdataDir = func() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(wd, "testdata")
}()

// runCLI executes rootCmd with args, capturing stdout/stderr separately.
// rootCmd and every command's flag variables are package-level (cobra's
// usual shape), shared across every test in this package and with the
// real binary: callers that set a flag (init's --builds-on, vendor's
// --to, ...) must reset it themselves before the next test runs, since
// pflag never resets a var to its default between Execute calls on its
// own, only ever between one explicit --flag and the next.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return out.String(), errOut.String(), err
}

// checkGolden mirrors internal/coverage's and internal/compile's own
// checkGolden (each package keeps its own copy by established
// convention here, rather than a shared test-helper package): compares
// got against goldenFile, or writes it when -update is passed.
func checkGolden(t *testing.T, goldenFile, got string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenFile, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("reading golden file %s: %v (run go test -run %s -update to create it)", goldenFile, err, t.Name())
	}
	if got != string(want) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s", goldenFile, got, string(want))
	}
}
