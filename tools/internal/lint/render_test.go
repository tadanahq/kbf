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

package lint_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/tadanahq/kbf/tools/internal/lint"
)

// Force a fixed, colorless output profile for the whole test binary. The
// real `kbf` CLI leaves this on lipgloss's normal terminal auto-detection
// (color in a TTY, plain when piped); tests need one fixed profile so the
// golden files are exact byte comparisons regardless of whether the test
// runs in a terminal, a CI runner, or under `go test -v` capture.
func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

var update = flag.Bool("update", false, "write golden files instead of comparing against them")

func TestRenderHumanGolden(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
	}{
		{"clean", []string{"testdata/clean"}},
		{"structural", []string{"testdata/structural"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := mustRun(t, c.paths...)
			checkGolden(t, filepath.Join("testdata", "golden", c.name+".human.txt"), lint.RenderHuman(result))
		})
	}
}

func TestRenderJSONGolden(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
	}{
		{"clean", []string{"testdata/clean"}},
		{"structural", []string{"testdata/structural"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := mustRun(t, c.paths...)
			out, err := lint.RenderJSON(result)
			if err != nil {
				t.Fatalf("RenderJSON: %v", err)
			}
			checkGolden(t, filepath.Join("testdata", "golden", c.name+".json"), string(out)+"\n")
		})
	}
}

// checkGolden compares got against the contents of goldenFile. Run with
// -update to (re)write the golden file from the current output, the
// standard Go idiom for this kind of test.
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
