// ABOUTME: Reads compiled.txt, the perl.git t/ files the oracle's pinned perl compiles, recorded by path.
// ABOUTME: Checks the set leaves out perlRejects and agrees with the count m3start.txt records.

package parse_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ratchetDir holds the oracle's recorded sets and starting counts.
var ratchetDir = filepath.Join("..", "parseoracle", "testdata", "ratchet")

// readCompiledSet returns compiled.txt's lines, one corpus path each, in
// file order. It is the one reader of that file.
func readCompiledSet(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(ratchetDir, "compiled.txt"))
	if err != nil {
		t.Fatalf("reading the compiled set: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

// m3StartCount is the integer m3start.txt records under key.
func m3StartCount(t *testing.T, key string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(ratchetDir, "m3start.txt"))
	if err != nil {
		t.Fatalf("reading m3start.txt: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if k, v, ok := strings.Cut(line, "="); ok && strings.TrimSpace(k) == key {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				t.Fatalf("m3start.txt's %s is not a count: %v", key, err)
			}
			return n
		}
	}
	t.Fatalf("m3start.txt records no %s", key)
	return 0
}

// TestCompiledSetExcludesPerlRejects: no file perl.git's ratchet treats as
// rejected is in the compiled set, even where `perl -c` exits 0 on it; the
// set has no blank or duplicate line; and its size is m3start's compiled.
func TestCompiledSetExcludesPerlRejects(t *testing.T) {
	paths := readCompiledSet(t)
	seen := map[string]bool{}
	for i, path := range paths {
		switch {
		case strings.TrimSpace(path) == "":
			t.Errorf("compiled.txt line %d is blank", i+1)
		case seen[path]:
			t.Errorf("compiled.txt lists %s twice", path)
		}
		seen[path] = true
		if _, rejected := perlRejects[path]; rejected {
			t.Errorf("compiled.txt lists %s, which perlRejects excludes", path)
		}
	}
	if want := m3StartCount(t, "compiled"); len(paths) != want {
		t.Errorf("compiled.txt has %d lines, m3start.txt records compiled = %d", len(paths), want)
	}
}
