// ABOUTME: Tests the -DT aligner on hand-written streams: in-order matching only, and invented tokens folded without shifting.
// ABOUTME: Runs no perl; the streams are written in tdt-capture.sh's form.

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// aligned returns each fixture line's fields after the span, keyed by the
// pvm token's text, and each unmatched run keyed by `unmatched <byte>`.
func aligned(t *testing.T, src, stream string) map[string]string {
	t.Helper()
	toks, marker := parseStream(stream)
	if marker != "" {
		t.Fatalf("marker %q", marker)
	}
	lines, _ := align([]byte(src), toks)
	got := map[string]string{}
	for _, l := range lines {
		f := strings.Fields(l)
		if f[0] == "unmatched" {
			got["unmatched "+f[1]] = strings.Join(f[2:], " ")
			continue
		}
		s, err1 := strconv.Atoi(f[0])
		e, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil {
			t.Fatalf("line %q", l)
		}
		got[src[s:e]] = strings.Join(f[2:], " ")
	}
	return got
}

// TestAlignRefusesReorder: tokens whose texts pair only if reordered stay
// unaligned; the same streams in order pair fully.
func TestAlignRefusesReorder(t *testing.T) {
	const src = "foo bar;"
	got := aligned(t, src, "BAREWORD\tPV(\"bar\"\\0)\nBAREWORD\tPV(\"foo\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n")
	if got["foo"] != "-" || got["bar"] != "-" {
		t.Errorf("reordered: foo %q, bar %q; want both unaligned (-)", got["foo"], got["bar"])
	}
	if got[";"] != "PERLY_SEMICOLON eof-semicolon PERLY_SEMICOLON" {
		t.Errorf("reordered: ; %q, want the ; and toke.c's at end of file", got[";"])
	}
	got = aligned(t, src, "BAREWORD\tPV(\"foo\"\\0)\nBAREWORD\tPV(\"bar\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n")
	if got[";"] != "PERLY_SEMICOLON eof-semicolon PERLY_SEMICOLON" {
		t.Errorf("in order: ; %q, want the ; and toke.c's at end of file", got[";"])
	}
	for _, w := range []string{"foo", "bar"} {
		if got[w] != "BAREWORD" {
			t.Errorf("in order: %s %q, want BAREWORD", w, got[w])
		}
	}
}

// TestAlignInsertedSemicolonDoesNotShift: the ; toke.c inserts before }
// folds onto the } and every later token pairs with its own text.
func TestAlignInsertedSemicolonDoesNotShift(t *testing.T) {
	const src = "{ f } g;"
	got := aligned(t, src, "PERLY_BRACE_OPEN\nBAREWORD\tPV(\"f\"\\0)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\n"+
		"BAREWORD\tPV(\"g\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n")
	want := map[string]string{
		"{": "PERLY_BRACE_OPEN",
		"f": "BAREWORD",
		"}": "PERLY_BRACE_CLOSE brace-semicolon PERLY_SEMICOLON",
		"g": "BAREWORD",
		";": "PERLY_SEMICOLON eof-semicolon PERLY_SEMICOLON",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: %q, want %q", k, got[k], w)
		}
	}
	if len(got) != len(want) {
		t.Errorf("lines %q, want only %q", got, want)
	}
}

// TestAlignSample: a fresh alignment of the committed sample's stream is
// the committed sample.tdt, in which the variable, the interpolated quote,
// the `=>` word and the heredoc (marker and body) are aligned.
func TestAlignSample(t *testing.T) {
	src, err := os.ReadFile("../../testdata/tdt-sample.pl")
	if err != nil {
		t.Fatal(err)
	}
	stream, err := os.ReadFile("../../testdata/tdt-sample.expected")
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("../../../testdata/tdt/sample.tdt")
	if err != nil {
		t.Fatal(err)
	}
	toks, _ := parseStream(string(stream))
	lines, _ := align(src, toks)
	var want []string
	for _, l := range strings.Split(strings.TrimSpace(string(golden)), "\n") {
		if !strings.HasPrefix(l, "#") {
			want = append(want, l)
		}
	}
	if got := strings.Join(lines, "\n"); got != strings.Join(want, "\n") {
		t.Errorf("alignment differs from sample.tdt:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
	}
	got := aligned(t, string(src), string(stream))
	for k, w := range map[string]string{
		"$name":               "PERLY_DOLLAR variable PRIVATEREF",
		"qq{hello, $name}":    "FUNC quote SUBLEXSTART THING ADDOP PERLY_DOLLAR PRIVATEREF SUBLEXEND",
		"greeting":            "BAREWORD",
		`<<"EOT"`:             "FUNC quote SUBLEXSTART PERLY_DOLLAR PRIVATEREF PERLY_BRACE_OPEN BAREWORD PERLY_SEMICOLON PERLY_BRACE_CLOSE ADDOP THING SUBLEXEND",
		"$h{greeting}\nEOT\n": "FUNC heredoc",
	} {
		if got[k] != w {
			t.Errorf("%q: %q, want %q", k, got[k], w)
		}
	}
}

// TestClearFixturesKeepsOtherFiles: clearing an output directory removes
// only *.tdt files, and a directory holding go.mod or .git is refused with
// nothing removed.
func TestClearFixturesKeepsOtherFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(p string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fixture, other := filepath.Join(dir, "internal", "a.pl.tdt"), filepath.Join(dir, "internal", "a.go")
	write(fixture)
	write(other)
	if err := clearFixtures(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(fixture); !os.IsNotExist(err) {
		t.Errorf("%s survived clearing (%v)", fixture, err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Errorf("%s was removed: %v", other, err)
	}

	for _, marker := range []string{"go.mod", ".git"} {
		repo := t.TempDir()
		write(filepath.Join(repo, marker))
		f := filepath.Join(repo, "T1", "b.t.tdt")
		write(f)
		if err := clearFixtures(repo); err == nil || !strings.Contains(err.Error(), marker) {
			t.Errorf("clearFixtures on a directory holding %s: %v, want a refusal naming it", marker, err)
		}
		if _, err := os.Stat(f); err != nil {
			t.Errorf("refused clear still removed %s: %v", f, err)
		}
	}
}

// TestAlignRequireVersion: the version yyl_require forces through
// force_version is a forced-version, not a constant sub's value.
func TestAlignRequireVersion(t *testing.T) {
	got := aligned(t, "require v5.10.0;", "KW_REQUIRE\nBAREWORD\tPVMG(\"\\5\\n\\0\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n")
	if w := "BAREWORD forced-version"; got["v5.10.0"] != w {
		t.Errorf("v5.10.0: %q, want %q", got["v5.10.0"], w)
	}
	if w := "KW_REQUIRE"; got["require"] != w {
		t.Errorf("require: %q, want %q", got["require"], w)
	}
}

// TestAlignReturnIsUnreported: OLDLOP never reports the return keyword's
// token, so pvm's keyword `return` is unreported-return and the
// scope-decided sub after it pairs with its own word; `return` as a word
// (an auto-quoted key, a subscript, a method name) pairs with toke.c's token.
func TestAlignReturnIsUnreported(t *testing.T) {
	got := aligned(t, "return run_callback { 1 };",
		"LSTOPSUB\top_rv2cv\nPERLY_BRACE_OPEN\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n")
	if w := "- unreported-return"; got["return"] != w {
		t.Errorf("return: %q, want %q", got["return"], w)
	}
	if w := "LSTOPSUB scope-sub"; got["run_callback"] != w {
		t.Errorf("run_callback: %q, want %q", got["run_callback"], w)
	}

	for _, c := range []struct{ src, stream, want string }{
		{"(return => 1);", "PERLY_PAREN_OPEN\nBAREWORD\tPV(\"return\"\\0)\nPERLY_COMMA\nTHING\tIV(1)\nPERLY_PAREN_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "BAREWORD"},
		{"$h{return};", "PERLY_DOLLAR\nBAREWORD\tPV(\"main::h\"\\0)\nPERLY_BRACE_OPEN\nBAREWORD\tPV(\"return\"\\0)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "BAREWORD"},
		{"$o->return;", "PERLY_DOLLAR\nBAREWORD\tPV(\"main::o\"\\0)\nARROW\nMETHCALL0\tPV(\"return\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "METHCALL0"},
	} {
		got := aligned(t, c.src, c.stream)
		if got["return"] != c.want {
			t.Errorf("%s: return %q, want %q", c.src, got["return"], c.want)
		}
		for k, v := range got {
			if strings.HasPrefix(k, "unmatched") {
				t.Errorf("%s: %s %s, want no unmatched run", c.src, k, v)
			}
		}
	}

	// The keyword with an auto-quoted `return` right after its opener: the
	// key's BAREWORD must not pull the keyword ahead of the opener's unit.
	const tail = "BAREWORD\tPV(\"return\"\\0)\nPERLY_COMMA\nTHING\tIV(1)\n"
	for _, c := range []struct{ src, stream, open, want string }{
		{"return { return => 1 };", "HASHBRACK\n" + tail + "PERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "{", "HASHBRACK"},
		{"return (return => 1);", "PERLY_PAREN_OPEN\n" + tail + "PERLY_PAREN_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "(", "PERLY_PAREN_OPEN"},
		{"return [ return => 1 ];", "PERLY_BRACKET_OPEN\n" + tail + "PERLY_BRACKET_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "[", "PERLY_BRACKET_OPEN"},
		{"CORE::return { return => 1 };", "HASHBRACK\n" + tail + "PERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n", "{", "HASHBRACK"},
	} {
		got := aligned(t, c.src, c.stream)
		if got[c.open] != c.want {
			t.Errorf("%s: %s %q, want %q", c.src, c.open, got[c.open], c.want)
		}
		if got["return"] != "BAREWORD" { // the key, the later of the two
			t.Errorf("%s: key return %q, want BAREWORD", c.src, got["return"])
		}
		kw := c.src[:strings.IndexByte(c.src, ' ')]
		if kw != "return" && got[kw] != "- unreported-return" {
			t.Errorf("%s: %s %q, want - unreported-return", c.src, kw, got[kw])
		}
		for k, v := range got {
			if strings.HasPrefix(k, "unmatched") {
				t.Errorf("%s: %s %s, want no unmatched run", c.src, k, v)
			}
		}
	}
}

// TestAlignAmendment2: the families TDTALIGN amendment 2 adds. A constant called with () is one toke.c token, so its
// () is constant-parens and the comma after it still pairs; a string constant's word is constant-string; the my or
// our before sub, and an our sub's package-qualified name, are lexical-sub. Streams are the pinned perl's -DT.
func TestAlignAmendment2(t *testing.T) {
	for _, c := range []struct {
		src, stream string
		want        map[string]string
	}{
		{"sub F () { 1 } my $x = F(), 2;",
			"KW_SUB_named\nBAREWORD\tPV(\"F\"\\0)\nPROTOTYPE\tPVIV(\"\"\\0)\nPERLY_BRACE_OPEN\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nKW_MY\nPERLY_DOLLAR\nPRIVATEREF\t$x\nASSIGNOP\top_null\nBAREWORD\tIV(1)\nPERLY_COMMA\nTHING\tIV(2)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			map[string]string{"(": "- constant-parens", ")": "- constant-parens", ",": "PERLY_COMMA"}},
		// Matched rather than passed over, the () would tie the comma and the ) at one distance and lose both.
		{"sub F () { 1 } g(F(),);",
			"KW_SUB_named\nBAREWORD\tPV(\"F\"\\0)\nPROTOTYPE\tPVIV(\"\"\\0)\nPERLY_BRACE_OPEN\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_AMPERSAND\nBAREWORD\tPV(\"g\"\\0)\nPERLY_PAREN_OPEN\nBAREWORD\tIV(1)\nPERLY_COMMA\nPERLY_PAREN_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			map[string]string{",": "PERLY_COMMA", "F": "BAREWORD constant-sub"}},
		{"sub C () { \"v\" } my $x = C . 1;",
			"KW_SUB_named\nBAREWORD\tPV(\"C\"\\0)\nPROTOTYPE\tPVIV(\"\"\\0)\nPERLY_BRACE_OPEN\nTHING\tPV(\"v\"\\0)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nKW_MY\nPERLY_DOLLAR\nPRIVATEREF\t$x\nASSIGNOP\top_null\nBAREWORD\tPV(\"v\"\\0)\nADDOP\top_concat\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			map[string]string{"C": "- constant-string", ".": "ADDOP"}},
		{"my sub f { 1 } our sub g; f(&g);",
			"KW_SUB_named\nPRIVATEREF\t&f\nPERLY_BRACE_OPEN\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nKW_SUB_named\nBAREWORD\tPV(\"main::g\"\\0)\nPERLY_SEMICOLON\nPERLY_AMPERSAND\nPRIVATEREF\top_padany\nPERLY_PAREN_OPEN\nPERLY_AMPERSAND\nBAREWORD\tPV(\"main::g\"\\0)\nPERLY_PAREN_CLOSE\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			map[string]string{"my": "- lexical-sub", "our": "- lexical-sub"}},
	} {
		got := aligned(t, c.src, c.stream)
		for k, w := range c.want {
			if got[k] != w {
				t.Errorf("%s: %q is %q, want %q", c.src, k, got[k], w)
			}
		}
	}
	// A word whose unit is not a constant's value stays unexplained.
	if got := aligned(t, "foo;", "THING\tIV(3)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n"); got["foo"] != "-" {
		t.Errorf("foo: %q, want unaligned with no family", got["foo"])
	}
}

// TestRelabel: a listed file's unaligned lines inside its span take its family, and nothing else changes.
func TestRelabel(t *testing.T) {
	lines := []string{"333 334 PERLY_PAREN_OPEN", "334 341 -", "343 344 - label", "353 361 -"}
	relabel("T1/attribute_method_named_sub_deparse.t", lines)
	want := []string{"333 334 PERLY_PAREN_OPEN", "334 341 - attribute-resync", "343 344 - label", "353 361 -"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("relabel = %q, want %q", lines, want)
	}
	other := []string{"0 3 -"}
	relabel("T1/other.t", other)
	if other[0] != "0 3 -" {
		t.Errorf("an unlisted file changed: %q", other)
	}
}

// TestAlignAmendment3: TDTALIGN amendment 3's matching rules, each pairing a token the aligner left unaligned: ~.
// under the bitwise feature is PERLY_TILDE; Legacy'Package is toke.c's Legacy::Package; $main_x keeps its name, as
// main:: alone is a qualifier; whitespace may follow the sigil; and a heredoc body opened inside an interpolated string
// is that string's. Streams are the pinned perl's -DT.
func TestAlignAmendment3(t *testing.T) {
	for _, c := range []struct{ src, stream, tok, want string }{
		{"use feature 'bitwise'; my $s = ~. $x;",
			"KW_USE_or_NO\nBAREWORD\nBAREWORD\tPV(\"feature\"\\0)\nTHING\tPV(\"bitwise\"\\0)\nPERLY_SEMICOLON\nKW_MY\nPERLY_DOLLAR\nPRIVATEREF\t$s\nASSIGNOP\top_null\nPERLY_TILDE\nPERLY_DOLLAR\nBAREWORD\tPV(\"x\"\\0)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			"~.", "PERLY_TILDE"},
		{"package Legacy'Package; 1;",
			"KW_PACKAGE\nBAREWORD\nBAREWORD\tPV(\"Legacy::Package\"\\0)\nPERLY_SEMICOLON\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			"Legacy'Package", "BAREWORD"},
		{"my $main_x = 1;",
			"KW_MY\nPERLY_DOLLAR\nPRIVATEREF\t$main_x\nASSIGNOP\top_null\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			"$main_x", "PERLY_DOLLAR variable PRIVATEREF"},
		{"$ x = 1;",
			"PERLY_DOLLAR\nBAREWORD\tPV(\"x\"\\0)\nASSIGNOP\top_null\nTHING\tIV(1)\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			"$ x", "PERLY_DOLLAR variable BAREWORD"},
		{"my $v = \"x @{[ <<EOT ]} x\";\nHERE\nEOT\n",
			"KW_MY\nPERLY_DOLLAR\nPRIVATEREF\t$v\nASSIGNOP\top_null\nFUNC\top_stringify\nSUBLEXSTART\nTHING\tPV(\"x \"\\0)\nADDOP\top_concat\nFUNC\top_join\nPERLY_PAREN_OPEN\nPERLY_DOLLAR\nBAREWORD\tPV(\"\\\"\"\\0)\nPERLY_COMMA\nPERLY_SNAIL\nPERLY_BRACE_OPEN\nPERLY_BRACKET_OPEN\nFUNC\top_scalar\nSUBLEXSTART\nTHING\tPV(\"HERE\\n\"\\0)\nSUBLEXEND\nPERLY_BRACKET_CLOSE\nPERLY_SEMICOLON\nPERLY_BRACE_CLOSE\nPERLY_PAREN_CLOSE\nADDOP\top_concat\nTHING\tPV(\" x\"\\0)\nSUBLEXEND\nPERLY_SEMICOLON\nPERLY_SEMICOLON\n",
			"HERE\nEOT\n", "FUNC heredoc"},
	} {
		if got := aligned(t, c.src, c.stream)[c.tok]; got != c.want {
			t.Errorf("%q: %q is %q, want %q", c.src, c.tok, got, c.want)
		}
	}
}
