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

// Command boundaries is this repo's public-hygiene gate (AGENTS.md,
// project-standards.md): no client names, no internal project references,
// no private paths, no prices, anywhere, fixtures and comments included.
// It scans every text file under --root for a blocklist of exact terms
// plus a small set of pattern heuristics, and fails (exit 1) on any hit.
//
// blocklist is deliberately empty here. This file is itself published to
// the public repo it scans, so writing a real client or internal project
// name into it as a literal string would be exactly the leak this scan
// exists to catch, the moment the file is committed: a blocklist that
// documents what must stay private, in public, defeats itself. The repo
// owner adds real entries directly, in a commit they review, each with a
// comment on why (project-standards.md: "maintained deliberately, never
// by pattern alone"). scan() takes the blocklist as a parameter for
// exactly this reason: boundaries_test.go exercises the mechanism with
// synthetic terms, without needing a populated production list to do it.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// blocklist: see the package comment for why this starts empty.
var blocklist = []string{}

// heuristics catch shapes, not specific words, so they carry no
// confidentiality risk of their own: an absolute path rooted in a
// personal machine, or a price. Deliberately narrow (project-standards.md:
// "a named blocklist plus heuristics") to keep false positives rare in a
// pure Go/YAML/Markdown repo that has no legitimate reason to contain
// either shape.
var heuristics = []struct {
	name string
	re   *regexp.Regexp
}{
	{"private filesystem path", regexp.MustCompile(`/(Users|home)/[A-Za-z0-9_.-]+/`)},
	{"price", regexp.MustCompile(`\$[0-9][0-9,.]*\b`)},
}

// skipDirs are never descended into: not repo content in the sense this
// scan cares about, .git especially can be large, and bin/ is where `go
// build` (with no -o) leaves an executable that would otherwise be
// scanned as garbage text the moment it exists on disk.
var skipDirs = map[string]bool{".git": true, "bin": true, "dist": true, "node_modules": true}

// binaryExt short-circuits without opening the file for extensions that
// are always binary. This is a fast path, not the real defense: scanFile
// also sniffs for a NUL byte, which catches any binary regardless of
// extension (including a stray compiled Go binary with no extension at
// all, found the hard way while building this scanner: `go build ./...`
// on a package main with no -o writes an executable named after the
// directory into the current directory).
var binaryExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
	".pdf": true, ".zip": true, ".tar": true, ".gz": true,
	".woff": true, ".woff2": true, ".ttf": true,
}

// selfTestFixture is scripts/boundaries_test.go's own path suffix. That
// file's entire job is to contain synthetic examples of what the
// heuristics catch (a fake private path, a fake price), so it is excluded
// here by identity, not by weakening the heuristics: every other file in
// the repo, tests included, stays fully in scope.
const selfTestFixture = "scripts/boundaries_test.go"

// violation is one line that tripped the blocklist or a heuristic.
type violation struct {
	file string
	line int
	rule string
	text string
}

func main() {
	root := flag.String("root", ".", "directory to scan")
	flag.Parse()

	violations, err := scan(*root, blocklist)
	if err != nil {
		fmt.Fprintln(os.Stderr, "boundaries:", err)
		os.Exit(2)
	}
	if len(violations) == 0 {
		fmt.Println("boundaries: clean")
		return
	}
	for _, v := range violations {
		fmt.Printf("%s:%d: %s: %s\n", v.file, v.line, v.rule, strings.TrimSpace(v.text))
	}
	fmt.Printf("boundaries: %d violation(s)\n", len(violations))
	os.Exit(1)
}

// scan walks root and returns every line matching a blocklist term (case-
// insensitive) or a heuristic pattern, sorted by file then line: the same
// order every run, so failures are easy to diff between CI runs. Takes
// blocklist as a parameter, not the package-level var, so
// boundaries_test.go can exercise the mechanism with synthetic terms
// without depending on (or requiring) a populated production list.
func scan(root string, blocklist []string) ([]violation, error) {
	var violations []violation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if binaryExt[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		if strings.HasSuffix(filepath.ToSlash(path), selfTestFixture) {
			return nil
		}
		found, err := scanFile(path, blocklist)
		if err != nil {
			return err
		}
		violations = append(violations, found...)
		return nil
	})
	return violations, err
}

// scanFile checks one file line by line against blocklist (case-
// insensitive) and the heuristics. Binary files are skipped: a NUL byte
// anywhere in the first 8000 bytes is the same sniff `git`/`grep -I` use,
// robust regardless of extension (binaryExt is only a fast path).
func scanFile(path string, blocklist []string) ([]violation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck // read-only scan; nothing to recover if close fails

	isBinary, err := looksBinary(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if isBinary {
		return nil, nil
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	var violations []violation
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		lower := strings.ToLower(line)
		for _, term := range blocklist {
			if strings.Contains(lower, strings.ToLower(term)) {
				violations = append(violations, violation{file: path, line: lineNo, rule: "blocklist: " + term, text: line})
			}
		}
		for _, h := range heuristics {
			if h.re.MatchString(line) {
				violations = append(violations, violation{file: path, line: lineNo, rule: h.name, text: line})
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return violations, nil
}

// looksBinary reports whether r's first 8000 bytes contain a NUL byte.
// Leaves the reader positioned after whatever it read; callers that go on
// to read the file again must Seek back to the start first.
func looksBinary(r io.Reader) (bool, error) {
	buf := make([]byte, 8000)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return false, err
	}
	return bytes.IndexByte(buf[:n], 0) != -1, nil
}
