// ABOUTME: Tests that m3start.txt records the M3 starting counts against the oracle's pinned interpreter.
// ABOUTME: Later criteria compare against that file, so a count measured under another perl must fail here.

package parseoracle_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// m3startPath is the append-only record of the M3 starting counts.
var m3startPath = filepath.Join("testdata", "ratchet", "m3start.txt")

// readM3Start reads m3start.txt's `key = value` lines, ignoring comments.
// A key written twice is an error: the file is append-only, so each key is
// recorded once, and keeping either value would hide the other.
func readM3Start(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	keys := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if _, dup := keys[key]; dup {
			return nil, fmt.Errorf("%s records %s twice: the file is append-only, so a key is written once", path, key)
		}
		keys[key] = strings.TrimSpace(value)
	}
	return keys, nil
}

// m3StartProblems checks the m3start file at path against the pin at
// pinPath: the starting keys are present and the interpreter is the pin's.
func m3StartProblems(path, pinPath string) []string {
	keys, err := readM3Start(path)
	if err != nil {
		return []string{err.Error()}
	}
	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		return []string{err.Error()}
	}
	var problems []string
	for _, key := range []string{"interpreter", "compiled", "graded"} {
		if keys[key] == "" {
			problems = append(problems, fmt.Sprintf("%s records no %s", path, key))
		}
	}
	if keys["interpreter"] != pin.Interpreter {
		problems = append(problems, fmt.Sprintf("%s was measured under %q but %s pins %s",
			path, keys["interpreter"], pinPath, pin.Interpreter))
	}
	return problems
}

// TestM3StartRecordsThePin: the starting counts were measured under the
// perl corpus.pin names.
func TestM3StartRecordsThePin(t *testing.T) {
	for _, p := range m3StartProblems(m3startPath, filepath.Join("testdata", "corpus.pin")) {
		t.Error(p)
	}
}

// TestM3StartRejectsInterpreterSkew: an m3start whose interpreter is not
// the pin's fails, so counts from another perl cannot pass as the start.
func TestM3StartRejectsInterpreterSkew(t *testing.T) {
	dir := t.TempDir()
	pin := filepath.Join(dir, "corpus.pin")
	start := filepath.Join(dir, "m3start.txt")
	if err := os.WriteFile(pin, []byte("interpreter = 5.044000\nrevision = deadbeef\nthreads = define\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(start, []byte("interpreter = 5.042000\ncompiled = 1\ngraded = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if problems := m3StartProblems(start, pin); len(problems) == 0 {
		t.Error("an m3start measured under 5.042000 passed against a 5.044000 pin")
	}
	if err := os.WriteFile(start, []byte("interpreter = 5.044000\ncompiled = 1\ngraded = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if problems := m3StartProblems(start, pin); len(problems) != 0 {
		t.Errorf("premise: a matching m3start must pass, got %v", problems)
	}
}

// TestM3StartRejectsARepeatedKey: m3start.txt is append-only, so a key
// written twice is two answers to one question. Reading either one would
// hide the other, so the file fails instead.
func TestM3StartRejectsARepeatedKey(t *testing.T) {
	dir := t.TempDir()
	pin := filepath.Join(dir, "corpus.pin")
	start := filepath.Join(dir, "m3start.txt")
	if err := os.WriteFile(pin, []byte("interpreter = 5.044000\nrevision = deadbeef\nthreads = define\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := "interpreter = 5.044000\ncompiled = 1\ngraded = 1\ncompiled = 2\n"
	if err := os.WriteFile(start, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if problems := m3StartProblems(start, pin); len(problems) == 0 {
		t.Error("an m3start recording compiled twice passed")
	}
}
