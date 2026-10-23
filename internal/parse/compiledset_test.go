// ABOUTME: Reads compiled.txt, the perl.git t/ files the oracle's pinned perl compiles, recorded by path.
// ABOUTME: Checks the set leaves out perlRejects and agrees with the count m3start.txt records.

package parse_test

import (
	"fmt"
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
	n, err := m3StartValue(filepath.Join(ratchetDir, "m3start.txt"), key)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// m3StartValue is the integer the m3start file at path records under key.
func m3StartValue(path, key string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, found := "", false
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(k) != key {
			continue
		}
		if found {
			return 0, fmt.Errorf("%s records %s twice: the file is append-only, so a key is written once", path, key)
		}
		value, found = strings.TrimSpace(v), true
	}
	if !found {
		return 0, fmt.Errorf("%s records no %s", path, key)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s's %s is not a count: %w", path, key, err)
	}
	return n, nil
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

// TestM3StartValueRejectsARepeatedKey: m3start.txt is append-only, so a key
// written twice is two answers to one question, and reading either would
// hide the other.
func TestM3StartValueRejectsARepeatedKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m3start.txt")
	if err := os.WriteFile(path, []byte("compiled = 1\ngraded = 1\ncompiled = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := m3StartValue(path, "compiled"); err == nil {
		t.Errorf("a repeated compiled key read as %d", n)
	}
	if n, err := m3StartValue(path, "graded"); err != nil || n != 1 {
		t.Errorf("graded = %d, %v; want 1", n, err)
	}
}
