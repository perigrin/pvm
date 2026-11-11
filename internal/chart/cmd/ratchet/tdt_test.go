// ABOUTME: Tests the ratchet's tdt keys on toy aligned pairs, a toy token table and toy reviewed lists; no fixtures are read.
// ABOUTME: Covers each key's value, the edges of mismatch and aligned, stale reviewed entries, missing fixtures and exact families.

package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
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
	return []tdtFile{{Path: "toy.pl", Src: []byte("foo 42;"), Fixture: "# capture: whole\n" + fixture}}
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
	partial := []tdtFile{{Path: "toy.pl", Src: []byte("foo 42;"), Fixture: "# capture: refused: perl exited 255\n"}}
	wantKeys(t, keysOf(t, partial), map[string]string{"tdt-aligned": "0.0000", "tdt-unaligned-unlisted": "0"})
}

func TestTDTKeysEdgeCases(t *testing.T) {
	// foo is aligned to NUM, inside the word's set {WORD, NUM}: no mismatch; and it produces (NUM, Word).
	wantKeys(t, keysOf(t, toy("0 3 NUM\n4 6 NUM\n6 7 SEMI\n")), map[string]string{"tdt-aligned": "1.0000", "tdt-mismatch": "0", "tdt-unseen": "1"})
	// A number aligned to WORD is outside its set {NUM}; a token no class admits has the empty set.
	wantKeys(t, keysOf(t, toy("0 3 WORD\n4 6 WORD\n6 7 SEMI\n")), map[string]string{"tdt-mismatch": "1"})
	files := []tdtFile{{Path: "v.pl", Src: []byte("$x"), Fixture: "# capture: whole\n0 2 WORD\n"}}
	wantKeys(t, keysOf(t, files), map[string]string{"tdt-aligned": "1.0000", "tdt-mismatch": "1"})
	wantKeys(t, keysOf(t, toy("")), map[string]string{"tdt-aligned": "0.0000"})
	wantKeys(t, keysOf(t, nil), map[string]string{"tdt-aligned": "0.0000", "tdt-mismatch": "0"})
}

// A PRIVATEREF stands for a name the noamp family folds PERLY_AMPERSAND onto and for a constant sub's BAREWORD, and a
// KW_SUB_named for a statement word the phaser family folds KW_SUB_named or KW_SUB_named_sig onto; for nothing else.
func TestTDTCallFit(t *testing.T) {
	names := []string{"PRIVATEREF", "SEMI", "FUNC1", "KW_SUB_named"}
	privateref := []grammar.ClassDecl{
		{Terminal: "PRIVATEREF", Key: grammar.TokenClass{Kind: ptr(lexer.Word)}},
		{Terminal: "SEMI", Key: grammar.TokenClass{Kind: ptr(lexer.Semicolon)}},
	}
	func1 := []grammar.ClassDecl{ // the word's only class is not PRIVATEREF
		{Terminal: "FUNC1", Key: grammar.TokenClass{Kind: ptr(lexer.Word)}},
		{Terminal: "SEMI", Key: grammar.TokenClass{Kind: ptr(lexer.Semicolon)}},
	}
	kwSub := []grammar.ClassDecl{
		{Terminal: "KW_SUB_named", Key: grammar.TokenClass{Kind: ptr(lexer.Word)}},
		{Terminal: "SEMI", Key: grammar.TokenClass{Kind: ptr(lexer.Semicolon)}},
	}
	for _, c := range []struct {
		decls                  []grammar.ClassDecl
		line, mismatch, unseen string
	}{
		{privateref, "0 3 BAREWORD noamp PERLY_AMPERSAND", "0", "0"},
		{privateref, "0 3 PRIVATEREF", "0", "0"},
		{privateref, "0 3 BAREWORD other PERLY_AMPERSAND", "1", "1"},
		{privateref, "0 3 BAREWORD scope-sub noamp NOAMP", "1", "1"},
		{privateref, "0 3 BAREWORD noamp", "1", "1"},
		{privateref, "0 3 BAREWORD", "1", "1"},
		{func1, "0 3 BAREWORD noamp PERLY_AMPERSAND", "1", "1"},
		{privateref, "0 3 BAREWORD constant-sub", "0", "0"},
		{privateref, "0 3 THING constant-sub", "1", "1"},
		{func1, "0 3 BAREWORD constant-sub", "1", "1"},
		{kwSub, "0 3 BAREWORD phaser KW_SUB_named", "0", "0"},
		{kwSub, "0 3 BAREWORD phaser KW_SUB_named_sig", "0", "0"},
		{kwSub, "0 3 BAREWORD phaser", "1", "1"},
		{kwSub, "0 3 BAREWORD noamp KW_SUB_named", "1", "1"},
		{privateref, "0 3 BAREWORD phaser KW_SUB_named", "1", "1"},
	} {
		files := []tdtFile{{Path: "c.pl", Src: []byte("foo;"), Fixture: "# capture: whole\n" + c.line + "\n3 4 SEMI\n"}}
		m, err := tdtKeys(c.decls, names, files, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if m["tdt-mismatch"] != c.mismatch || m["tdt-unseen"] != c.unseen {
			t.Errorf("%q: tdt-mismatch %s, tdt-unseen %s; want %s, %s", c.line, m["tdt-mismatch"], m["tdt-unseen"], c.mismatch, c.unseen)
		}
	}
}

// A listed byte skips its token's comparison, and a { listed skips the } that closes it; a byte where no token starts
// is an error.
func TestTDTSkipped(t *testing.T) {
	fx := "0 3 WORD\n4 6 WORD\n6 7 SEMI\n" // 42 aligned to WORD, outside {NUM}
	files := toy(fx)
	wantKeys(t, keysOf(t, files), map[string]string{"tdt-mismatch": "1"})
	files[0].Skip = []int{4}
	wantKeys(t, keysOf(t, files), map[string]string{"tdt-mismatch": "0", "tdt-aligned": "1.0000"})
	unaligned := toy("0 3 WORD\n4 6 -\n6 7 SEMI\n") // 42 unaligned, in no family
	wantKeys(t, keysOf(t, unaligned), map[string]string{"tdt-unaligned-unlisted": "1"})
	unaligned[0].Skip = []int{4}
	wantKeys(t, keysOf(t, unaligned), map[string]string{"tdt-unaligned-unlisted": "0"})
	files[0].Skip = []int{5}
	if _, err := tdtKeys(toyDecls, toyNames, files, toyFams, nil); err == nil || !strings.Contains(err.Error(), "byte 5") {
		t.Errorf("a skip inside a token: error %v, want one naming byte 5", err)
	}
	src := []byte("$x = { a => { } } + { };")
	got, err := skipped(tdtFile{Path: "b.pl", Src: src, Skip: []int{5}}, lexer.Tokenize(src))
	if err != nil {
		t.Fatal(err)
	}
	if want := map[int]bool{5: true, 16: true}; !maps.Equal(got, want) {
		t.Errorf("skipped = %v, want %v", got, want)
	}
}

func TestTDTReviewedListRefusesStaleEntries(t *testing.T) {
	fx := "0 3 WORD\n4 6 NUM\n6 7 -\n"
	for _, bad := range []string{"SEMI Kind=Number", "NUM *", "*", "WORD Kind=Word extra"} {
		if _, err := tdtKeys(toyDecls, toyNames, toy(fx), toyFams, []string{bad}); err == nil || !strings.Contains(err.Error(), bad) {
			t.Errorf("reviewed entry %q: error %v, want one naming it", bad, err)
		}
	}
	// NUM is seen for the word in this fixture, so listing it is an error: a reason that no longer holds.
	if _, err := tdtKeys(toyDecls, toyNames, toy("0 3 NUM\n4 6 NUM\n6 7 -\n"), toyFams, []string{numWordKey}); err == nil || !strings.Contains(err.Error(), "is seen") {
		t.Errorf("a seen pair on the list: error %v, want one saying it is seen", err)
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

// absentPhrase marks a reviewed tdt-unseen reason that claims the construct is absent from the corpora.
const absentPhrase = "no corpus token matches the key"

// corpusFile is a corpus file tokenized once, with what a claim's own token must have in it.
type corpusFile struct {
	src   []byte
	raw   []lexer.Token
	kinds map[lexer.Kind]bool
	texts map[string]bool // of the words, operators and closing brackets, which are the tokens a key's Text names
}

// mayAdmit is false only where a table compiled from c alone admits no token of the file: the token's own Kind is the
// key's, and, for a word, operator or closing bracket, its Text is the raw text, or CORE:: and the raw text (chart/lex reads CORE::kw as kw), or isa for
// grammar.PlainIsa. It may be true for a file that admits nothing; it is never false for one that does.
func (f corpusFile) mayAdmit(c grammar.TokenClass) bool {
	if c.Kind != nil && !f.kinds[*c.Kind] {
		return false
	}
	if c.Kind == nil || *c.Kind != lexer.Word && *c.Kind != lexer.Operator && *c.Kind != lexer.CloseBracket ||
		c.Text == nil || *c.Text == "" || strings.Contains(*c.Text, "\x00") && *c.Text != grammar.PlainIsa {
		return true
	}
	return f.texts[*c.Text] || f.texts["CORE::"+*c.Text] || *c.Text == grammar.PlainIsa && f.texts["isa"]
}

// falseAbsences returns, for each reviewed entry whose reason claims absentPhrase, the corpus tokens its key admits
// compiled by itself (as tdtKeys reads a key for seen): a claim with any is false. A table shared between claims would
// change what each matches (a named word is no longer an OtherWord), so each claim has its own; the cost is held down by
// tokenizing each file once and mapping it only for the claims it could hold a token of (corpusFile.mayAdmit). srcs is
// every corpus file's source, whole capture or not. checked is how many absence claims it read, lines trimmed as listLines trims them for the
// ratchet, so a run that checks none can be told from one that found nothing false.
func falseAbsences(t *testing.T, decls []grammar.ClassDecl, reviewed string, srcs map[string][]byte) (out []string, checked int) {
	t.Helper()
	byKey := map[string]grammar.ClassDecl{}
	for _, d := range decls {
		if k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key)); byKey[k].Terminal == "" {
			byKey[k] = d
		}
	}
	files := map[string]corpusFile{}
	reason := ""
	for _, l := range strings.Split(reviewed, "\n") {
		l = strings.TrimSpace(l)
		if r, ok := strings.CutPrefix(l, "reason:"); ok {
			reason = r
			continue
		}
		d, ok := byKey[l]
		if !ok || !strings.Contains(reason, absentPhrase) {
			continue
		}
		checked++
		if checked == 1 { // tokenized once, and only when a claim is to be checked
			for p, src := range srcs {
				f := corpusFile{src, lexer.Tokenize(src), map[lexer.Kind]bool{}, map[string]bool{}}
				for _, tk := range f.raw {
					f.kinds[tk.Kind] = true
					if tk.Kind == lexer.Word || tk.Kind == lexer.Operator || tk.Kind == lexer.CloseBracket {
						f.texts[string(src[tk.Start:tk.End])] = true
					}
				}
				files[p] = f
			}
		}
		c := d.Key
		c.Terminals = []int{0} // one class: which terminal it admits cannot change whether it admits a token
		tt, err := grammar.CompileTokens([]grammar.TokenClass{c})
		if err != nil {
			t.Fatal(err)
		}
		for p, f := range files {
			if !f.mayAdmit(c) {
				continue
			}
			s, _ := lex.Map(f.src, f.raw, tt)
			for _, tk := range s.Tokens {
				if tk.TerminalSet != 0 {
					out = append(out, fmt.Sprintf("%s admits %s %d", l, p, f.raw[tk.Index].Start))
				}
			}
		}
	}
	return out, checked
}

// TestTDTAbsentReasonsHold: a reviewed tdt-unseen entry whose reason says no corpus token matches its key really has
// none, over every corpus file (T1 included); about 5 s. A synthetic false claim is caught.
func TestTDTAbsentReasonsHold(t *testing.T) {
	src := map[string][]byte{"toy.pl": []byte("foo 42;")}
	if got, n := falseAbsences(t, toyDecls, "reason: x "+absentPhrase+"\n"+numWordKey+"\n", src); len(got) != 1 || n != 1 {
		t.Errorf("a false absence claim: %q from %d checked, want it caught once", got, n)
	}
	// A trailing space or CR, which listLines trims for the ratchet, is checked all the same.
	if got, n := falseAbsences(t, toyDecls, "reason: x "+absentPhrase+" \r\n"+numWordKey+" \r\n", src); len(got) != 1 || n != 1 {
		t.Errorf("a claim with trailing space: %q from %d checked, want it caught once", got, n)
	}
	if got, n := falseAbsences(t, toyDecls, "reason: unaligned\n"+numWordKey+"\n", src); len(got) != 0 || n != 0 {
		t.Errorf("a reason that claims no absence was checked: %q from %d", got, n)
	}
	t.Chdir("../../../..")
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	_, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := os.ReadFile(reviewedFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := holdAbsences(t, g.Classes, string(reviewed), srcs); err != nil {
		t.Error(err)
	}
	// An empty list holds no claim, which must fail rather than pass vacuously.
	if err := holdAbsences(t, g.Classes, "", srcs); err == nil || !strings.Contains(err.Error(), "checked no") {
		t.Errorf("an empty list: %v, want a failure saying it checked no claim", err)
	}
}

// holdAbsences is falseAbsences as a verdict: an error naming every false claim, or one when no claim was checked.
func holdAbsences(t *testing.T, decls []grammar.ClassDecl, reviewed string, srcs map[string][]byte) error {
	t.Helper()
	bad, n := falseAbsences(t, decls, reviewed, srcs)
	if n == 0 {
		return fmt.Errorf("unseen-reviewed.txt: checked no %q claim, so the scan held nothing", absentPhrase)
	}
	if len(bad) > 0 {
		return fmt.Errorf("unseen-reviewed.txt claims absence, but:\n%s", strings.Join(bad, "\n"))
	}
	return nil
}

// TestTDTAbsentReasonsCatchFalseClaimOverCorpus: over the committed grammar and the real corpus, an absence claim for
// a key the corpus does contain (every corpus file ends statements with a semicolon) is caught and the key is named.
func TestTDTAbsentReasonsCatchFalseClaimOverCorpus(t *testing.T) {
	t.Chdir("../../../..")
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	_, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	const key = "PERLY_SEMICOLON Kind=Semicolon"
	err = holdAbsences(t, g.Classes, "reason: x "+absentPhrase+"\n"+key+"\n", srcs)
	if err == nil || !strings.Contains(err.Error(), key+" admits ") {
		t.Errorf("a false claim over the corpus: %v, want it caught and naming %q", err, key)
	}
}

// refAbsences is the reference falseAbsences is held to: each claim judged by a table compiled from it alone, over
// every file, with nothing shared between claims.
func refAbsences(t *testing.T, decls []grammar.ClassDecl, names []string, reviewed string, srcs map[string][]byte) (out []string) {
	t.Helper()
	ids := map[string]int{}
	for i, n := range names {
		ids[n] = i
	}
	byKey := map[string]grammar.ClassDecl{}
	for _, d := range decls {
		if k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key)); byKey[k].Terminal == "" {
			byKey[k] = d
		}
	}
	raws := map[string][]lexer.Token{}
	for p, src := range srcs {
		raws[p] = lexer.Tokenize(src)
	}
	reason := ""
	for _, l := range strings.Split(reviewed, "\n") {
		l = strings.TrimSpace(l)
		if r, ok := strings.CutPrefix(l, "reason:"); ok {
			reason = r
			continue
		}
		d, ok := byKey[l]
		if !ok || !strings.Contains(reason, absentPhrase) {
			continue
		}
		c := d.Key
		c.Terminals = []int{ids[d.Terminal]}
		tt, err := grammar.CompileTokens([]grammar.TokenClass{c})
		if err != nil {
			t.Fatal(err)
		}
		for p, raw := range raws {
			s, _ := lex.Map(srcs[p], raw, tt)
			for _, tk := range s.Tokens {
				if tk.TerminalSet != 0 {
					out = append(out, fmt.Sprintf("%s admits %s %d", l, p, raw[tk.Index].Start))
				}
			}
		}
	}
	return out
}

// TestTDTAbsentReasonsMatchPerClaimScan: falseAbsences reports exactly the admissions a table compiled from each claim
// alone reports, over the real corpus, for the keys a shared table once masked (an OtherWord or IdentQuote neighbour,
// which another claim's named word or quote changes the meaning of) and every 40th key besides.
func TestTDTAbsentReasonsMatchPerClaimScan(t *testing.T) {
	t.Chdir("../../../..")
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		t.Fatal(err)
	}
	_, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	names := symbolNames(tab)
	var b strings.Builder
	b.WriteString("reason: x " + absentPhrase + "\n")
	for i, d := range g.Classes {
		k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key))
		masked := strings.Contains(k, grammar.OtherWord) || strings.Contains(k, grammar.IdentQuote)
		if masked && (strings.HasPrefix(k, "METHCALL") || strings.HasPrefix(k, "MULOP")) || i%8 == 0 {
			b.WriteString(k + "\n")
		}
	}
	want := refAbsences(t, g.Classes, names, b.String(), srcs)
	got, n := falseAbsences(t, g.Classes, b.String(), srcs)
	if n < 20 || len(want) == 0 {
		t.Fatalf("checked %d claims, the reference found %d admissions: the sample proves nothing", n, len(want))
	}
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("falseAbsences reports %d admissions, a table per claim reports %d; first difference:\n%s", len(got), len(want), firstDiff(got, want))
	}
}

func firstDiff(a, b []string) string {
	for i := range min(len(a), len(b)) {
		if a[i] != b[i] {
			return a[i] + "\nvs\n" + b[i]
		}
	}
	return fmt.Sprintf("one is a prefix of the other, %d vs %d", len(a), len(b))
}

// TestTDTAbsentReasonsClaimsAreIndependent: of two claims over one file, the false one is named and the true one is not,
// though the true one names a word ("a") that the false one's OtherWord neighbour stands for.
func TestTDTAbsentReasonsClaimsAreIndependent(t *testing.T) {
	word := ptr(lexer.Word)
	truth := grammar.ClassDecl{Terminal: "WORD", Key: grammar.TokenClass{Kind: word, Text: ptr("q"), Prev1: &grammar.Neighbour{Kind: lexer.Word, Text: "a"}}}
	lie := grammar.ClassDecl{Terminal: "WORD", Key: grammar.TokenClass{Kind: word, Prev1: &grammar.Neighbour{Kind: lexer.Word, Text: grammar.OtherWord}}}
	tk, lk := "WORD "+keyText(truth.Key), "WORD "+keyText(lie.Key)
	decls := []grammar.ClassDecl{truth, lie}
	src := map[string][]byte{"toy.pl": []byte("a b;")}
	reviewed := "reason: x " + absentPhrase + "\n" + tk + "\n" + lk + "\n"
	got, n := falseAbsences(t, decls, reviewed, src)
	want := refAbsences(t, decls, toyNames, reviewed, src)
	if n != 2 || len(want) != 1 || !slices.Equal(got, want) || !strings.HasPrefix(got[0], lk+" admits toy.pl ") {
		t.Errorf("got %q from %d claims, want %q (the false claim only)", got, n, want)
	}
}

// symbolNames is the grammar's symbol names, which a class's Terminals index.
func symbolNames(tab *grammar.Tables) []string {
	names := make([]string, len(tab.Symbols))
	for i, s := range tab.Symbols {
		names[i] = s.Name
	}
	return names
}
