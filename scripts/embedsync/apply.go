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
	"sort"
)

// writeFiles wipes dataDir and writes every file fresh: a generated
// mirror is never edited in place or reconciled incrementally, so a
// deleted source (a spec doc renamed, a playbook file removed) can never
// leave a stale copy behind under the old name.
func writeFiles(dataDir string, files []file) error {
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("clear %s: %w", dataDir, err)
	}
	for _, f := range files {
		path := filepath.Join(dataDir, filepath.FromSlash(f.relPath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, f.content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// checkFiles compares files (what a fresh sync would produce) against
// dataDir's committed content, returning one human-readable line per
// difference: a file a fresh sync would produce that isn't committed, one
// that's committed but doesn't match, or one committed that no source
// file justifies anymore (an orphan a sync should have removed but
// didn't, e.g. from a hand-edit). A missing dataDir (nothing has ever
// been synced) is not a Go error: every wanted file is simply reported
// missing, the same as any other drift.
func checkFiles(dataDir string, files []file) ([]string, error) {
	want := map[string][]byte{}
	for _, f := range files {
		want[f.relPath] = f.content
	}

	got := map[string][]byte{}
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dataDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got[filepath.ToSlash(rel)] = data
		return nil
	})
	if err != nil {
		return nil, err
	}

	var stale []string
	for relPath, wantContent := range want {
		gotContent, ok := got[relPath]
		switch {
		case !ok:
			stale = append(stale, fmt.Sprintf("%s: missing from tools/internal/embedded/data", relPath))
		case !bytes.Equal(gotContent, wantContent):
			stale = append(stale, fmt.Sprintf("%s: stale, does not match source", relPath))
		}
	}
	for relPath := range got {
		if _, ok := want[relPath]; !ok {
			stale = append(stale, fmt.Sprintf("%s: orphaned in tools/internal/embedded/data, no source file justifies it anymore", relPath))
		}
	}
	sort.Strings(stale)
	return stale, nil
}
