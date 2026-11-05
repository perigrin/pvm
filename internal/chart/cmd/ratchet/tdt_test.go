// ABOUTME: Tests the ratchet's tdt keys on toy aligned pairs, a toy token table and toy reviewed lists; no fixtures are read.
// ABOUTME: Covers each key's value, the edges of mismatch and aligned, stale reviewed entries, missing fixtures and exact families.

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

var (
	toyNames = []string{"WORD", "NUM", "SEMI"}
	toyDecls = []grammar.ClassDecl{
		{Terminal: "WORD", Key: grammar.TokenClass{Kind: ptr(lexer.Word)}},
		{Terminal: "NUM", Key: grammar.TokenClass{Kind: ptr(lexer.Number)}},
		{Terminal: "SEMI", Key: grammar.TokenClass{Kind: ptr(lexer.Semicolon)}},
		{Terminal: "NUM", Key: grammar.TokenClass{Kind: ptr(lexer.Word)}}, // a word is also NUM: a multi-terminal set
	}
	toyFams    = map[string]bool{"brace-semicolon": true}
	semiPair   = "SEMI Kind=Semicolon"
	numWordKey = "NUM Kind=Word"
)

func ptr[T any](v T) *T { return &v }

// toy is `foo 42;` (foo 0-3, 42 4-6, ; 6-7) with the given fixture lines.
func toy(fixture string) []tdtFile {
	return []tdtFile{{"toy.pl", []byte("foo 42;"), "# capture: whole\n" + fixture}}
}

func keysOf(t *testing.T, files []tdtFile, reviewed ...string) map[string]string {
	t.Helper()
	m, err := tdtKeys(toyDecls, toyNames, files, toyFams, reviewed)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func wantKeys(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestTDTKeys(t *testing.T) {
	fx := "0 3 WORD\n4 6 NUM\n6 7 -\n"
	wantKeys(t, keysOf(t, toy(fx)), map[string]string{
		"tdt-aligned": "0.6667", "tdt-mismatch": "0",
		"tdt-unseen": "2", "tdt-unseen-unlisted": "2", "tdt-unaligned-unlisted": "1",
	})
	// SEMI was never produced (unaligned), as was NUM for a word; listing one leaves one.
	wantKeys(t, keysOf(t, toy(fx), semiPair), map[string]string{"tdt-unseen": "2", "tdt-unseen-unlisted": "1"})
	// A fixture span that matches no token leaves the token unaligned.
	wantKeys(t, keysOf(t, toy("0 2 WORD\n4 6 NUM\n6 7 SEMI\n")), map[string]string{"tdt-aligned": "0.6667", "tdt-unaligned-unlisted": "1"})
	// A capture that is not whole has no alignment and counts for nothing.
	partial := []tdtFile{{"toy.pl", []byte("foo 42;"), "# capture: refused: perl exited 255\n"}}
	wantKeys(t, keysOf(t, partial), map[string]string{"tdt-aligned": "0.0000", "tdt-unaligned-unlisted": "0"})
}

func TestTDTKeysEdgeCases(t *testing.T) {
	// foo is aligned to NUM, inside the word's set {WORD, NUM}: no mismatch; and it produces (NUM, Word).
	wantKeys(t, keysOf(t, toy("0 3 NUM\n4 6 NUM\n6 7 SEMI\n")), map[string]string{"tdt-aligned": "1.0000", "tdt-mismatch": "0", "tdt-unseen": "1"})
	// A number aligned to WORD is outside its set {NUM}; a token no class admits has the empty set.
	wantKeys(t, keysOf(t, toy("0 3 WORD\n4 6 WORD\n6 7 SEMI\n")), map[string]string{"tdt-mismatch": "1"})
	files := []tdtFile{{"v.pl", []byte("$x"), "# capture: whole\n0 2 WORD\n"}}
	wantKeys(t, keysOf(t, files), map[string]string{"tdt-aligned": "1.0000", "tdt-mismatch": "1"})
	wantKeys(t, keysOf(t, toy("")), map[string]string{"tdt-aligned": "0.0000"})
	wantKeys(t, keysOf(t, nil), map[string]string{"tdt-aligned": "0.0000", "tdt-mismatch": "0"})
}

func TestTDTReviewedListRefusesStaleEntries(t *testing.T) {
	fx := "0 3 WORD\n4 6 NUM\n6 7 -\n"
	for _, bad := range []string{"SEMI Kind=Number", "NUM *", "*", "WORD Kind=Word extra"} {
		if _, err := tdtKeys(toyDecls, toyNames, toy(fx), toyFams, []string{bad}); err == nil || !strings.Contains(err.Error(), bad) {
			t.Errorf("reviewed entry %q: error %v, want one naming it", bad, err)
		}
	}
	m := keysOf(t, toy(fx), semiPair, numWordKey)
	wantKeys(t, m, map[string]string{"tdt-unseen": "2", "tdt-unseen-unlisted": "0"})
}

func TestTDTKeysMissingFixtureIsAnError(t *testing.T) {
	dir := t.TempDir()
	src := map[string][]byte{"a.pl": []byte("foo 42;"), "b.pl": []byte("foo;")}
	files := []File{{"a.pl", "x"}, {"b.pl", "x"}}
	if err := os.WriteFile(filepath.Join(dir, "a.pl.tdt"), []byte("# capture: whole\n0 3 WORD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTDT(dir, files, src); err == nil || !strings.Contains(err.Error(), "b.pl") {
		t.Fatalf("loadTDT with b.pl's fixture missing: error %v, want one naming b.pl", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.pl.tdt"), []byte("# capture: whole\n0 3 WORD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadTDT(dir, files, src)
	if err != nil || len(got) != 2 {
		t.Fatalf("loadTDT with both fixtures = %v, %v; want both files", got, err)
	}
	wantKeys(t, keysOf(t, got), map[string]string{"tdt-aligned": "0.4000"}) // 2 of 5 tokens
}

func TestTDTUnalignedFamilyExact(t *testing.T) {
	for _, c := range []struct{ line, want string }{
		{"6 7 - brace-semicolon", "0"}, // listed exactly
		{"6 7 - Brace-Semicolon", "1"}, // a different spelling
		{"6 7 - brace_semicolon", "1"},
		{"6 7 - unlisted", "1"}, // not listed at all
		{"6 7 -", "1"},          // no family
	} {
		got := keysOf(t, toy("0 3 WORD\n4 6 NUM\n"+c.line+"\n"))["tdt-unaligned-unlisted"]
		if got != c.want {
			t.Errorf("%q: tdt-unaligned-unlisted = %s, want %s", c.line, got, c.want)
		}
	}
}

func TestTDTFixturesDefault(t *testing.T) {
	def := t.TempDir()
	if got := fixturesDir("", def); got != def {
		t.Errorf("unset TDT_FIXTURES reads %q, want the default %q", got, def)
	}
	if got := fixturesDir("/x", def); got != "/x" {
		t.Errorf("a set TDT_FIXTURES reads %q, want it", got)
	}
	err := measureTDT(filepath.Join(def, "nope"), nil, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "TDT_FIXTURES") {
		t.Errorf("a path that is no directory: error %v, want one naming TDT_FIXTURES", err)
	}
}

// Two classes that differ only in what they exclude are two keys, and the key names the exclusion, so a
// reviewed-list entry cannot stand for both.
func TestKeyTextRendersNegatedParts(t *testing.T) {
	word := lexer.Word
	paren := grammar.Neighbour{Kind: lexer.Operator, Text: "("}
	base := grammar.TokenClass{Kind: &word, Text: ptr("x")}
	keys := map[string]grammar.TokenClass{}
	for name, mod := range map[string]func(*grammar.TokenClass){
		"plain":        func(*grammar.TokenClass) {},
		"notnext":      func(c *grammar.TokenClass) { c.NotNext = []grammar.Neighbour{paren} },
		"notprev1":     func(c *grammar.TokenClass) { c.NotPrev1 = []grammar.Neighbour{paren} },
		"nottext":      func(c *grammar.TokenClass) { c.NotText = []string{"if"} },
		"nottext2":     func(c *grammar.TokenClass) { c.NotText = []string{"if", "my"} },
		"prev2":        func(c *grammar.TokenClass) { c.Prev2 = &paren },
		"notprev1b":    func(c *grammar.TokenClass) { c.NotPrev1 = []grammar.Neighbour{paren, {Kind: lexer.Word, Text: "ref"}} },
		"notprev2":     func(c *grammar.TokenClass) { c.NotPrev2 = []grammar.Neighbour{paren} },
		"notnextident": func(c *grammar.TokenClass) { c.NotNextIdent = true },
		"nextname":     func(c *grammar.TokenClass) { c.NextName = true },
		"notnextname":  func(c *grammar.TokenClass) { c.NotNextName = true },
		"blanks":       func(c *grammar.TokenClass) { c.Blanks = ptr(true) },
		"noblanks":     func(c *grammar.TokenClass) { c.Blanks = ptr(false) },
		"lone":         func(c *grammar.TokenClass) { c.Lone = true },
		"prev1var":     func(c *grammar.TokenClass) { c.Prev1 = &grammar.Neighbour{Kind: lexer.Variable} },
		"prev1array":   func(c *grammar.TokenClass) { c.Prev1 = &grammar.Neighbour{Kind: lexer.Variable, Array: true} },
	} {
		c := base
		mod(&c)
		keys[name] = c
	}
	seen := map[string]string{}
	for name, c := range keys {
		k := keyText(c)
		if other, dup := seen[k]; dup {
			t.Errorf("%s and %s are both %q", name, other, k)
		}
		seen[k] = name
	}
	long := base
	for i := range 300 {
		long.NotText = append(long.NotText, strconv.Itoa(i))
	}
	if k := keyText(long); len(k) > 100 || keyText(long) != k {
		t.Errorf("a long NotText renders %d bytes: %q", len(k), k)
	}
}
