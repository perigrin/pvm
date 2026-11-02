// ABOUTME: Tests the -DT aligner on hand-written streams: in-order matching only, and invented tokens folded without shifting.
// ABOUTME: Runs no perl; the streams are written in tdt-capture.sh's form.

package main

import (
	"os"
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
