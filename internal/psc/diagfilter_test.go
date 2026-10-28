// ABOUTME: Tests for §6.8.4 through the LSP path: didOpen and didChange in, publishDiagnostics out.
// ABOUTME: One syntax error is one diagnostic, the 40-byte cooldown and cap of 20 hold, and type diagnostics are gated.

package psc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
)

const untitledURI = "untitled:Untitled-1"

// lspSession is a handler with no debounce, driven through its message
// handling as an editor drives it.
type lspSession struct {
	t   *testing.T
	h   *handler
	out *bytes.Buffer
}

func newLSPSession(t *testing.T) *lspSession {
	h, out := newTestHandler(t)
	h.initialized = true
	return &lspSession{t: t, h: h, out: out}
}

func (s *lspSession) notify(method string, params any) {
	raw, err := json.Marshal(params)
	require.NoError(s.t, err)
	require.NoError(s.t, s.h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: method, Params: raw}))
}

// published is the publishDiagnostics notification for uri sent since the
// last call, once its parse has landed.
func (s *lspSession) published(uri string) lspPublishDiagnosticsParams {
	s.t.Helper()
	settle(s.h, uri)
	for s.out.Len() > 0 {
		msg := readResponse(s.t, s.out)
		if msg.Method != "textDocument/publishDiagnostics" {
			continue
		}
		var p lspPublishDiagnosticsParams
		require.NoError(s.t, json.Unmarshal(msg.Params, &p))
		if p.URI == uri {
			return p
		}
	}
	s.t.Fatalf("no publishDiagnostics for %s", uri)
	return lspPublishDiagnosticsParams{}
}

func (s *lspSession) open(uri, text string) []lspDiagnostic {
	s.t.Helper()
	s.notify("textDocument/didOpen", lspDidOpenParams{TextDocument: lspTextDocumentItem{URI: uri, LanguageID: "perl", Version: 1, Text: text}})
	return s.published(uri).Diagnostics
}

func (s *lspSession) change(uri string, version int, text string) []lspDiagnostic {
	s.t.Helper()
	s.notify("textDocument/didChange", lspDidChangeParams{
		TextDocument:   lspVersionedTextDocumentIdentifier{URI: uri, Version: version},
		ContentChanges: []lspTextDocumentContentChangeEvent{{Text: text}},
	})
	return s.published(uri).Diagnostics
}

// publishedFor opens text as an untitled buffer and returns what was published.
func publishedFor(t *testing.T, text string) []lspDiagnostic {
	t.Helper()
	return newLSPSession(t).open(untitledURI, text)
}

func codes(diags []lspDiagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Code
	}
	return out
}

// farApart is n independent call_arity errors (`push();`, which perl refuses)
// each on its own line, with a comment between them that puts more than
// errorCooldownBytes between one error's end and the next one's start.
func farApart(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "push(); # %s %d\n", strings.Repeat("x", errorCooldownBytes), i)
	}
	return b.String()
}

// tenSubs is ten named subs after an opening line, each clean Perl.
func tenSubs(indent string) string {
	var b strings.Builder
	for i := range 10 {
		fmt.Fprintf(&b, "%ssub b%d { my @a; push(@a, %d); return scalar @a }\n", indent, i, i)
	}
	return b.String()
}

func TestParseDiagnosticsAntiCascade(t *testing.T) {
	for name, src := range map[string]string{
		// The subs close sub a's block at column 0: one Missing `}`.
		"column-0 subs": "sub a {\n    my $x = 1;\n\n" + tenSubs(""),
		// Indented, they nest in sub a, and the package's `}` closes it:
		// one unclosed_opener refusal holding the Missing `}`.
		"nested subs": "package P {\n  sub a {\n    my $x = 1;\n\n" + tenSubs("  ") + "}\n",
		// Nothing closes it: unclosed at end of input.
		"open at EOF": "package P;\nsub a {\n    my $x = 1;\n" + tenSubs("    "),
	} {
		got := publishedFor(t, src)
		assert.Len(t, got, 1, "%s: one missing `}` publishes one diagnostic in total: %v", name, codes(got))
	}

	// Cooldown: five errors all starting within 40 bytes of the first's end
	// are one cascade.
	got := publishedFor(t, strings.Repeat("push();\n", 5))
	assert.Len(t, got, 1, "adjacent errors within the cooldown: %v", codes(got))

	// Cap: thirty independent errors publish twenty.
	assert.Len(t, publishedFor(t, farApart(30)), maxParseDiagnostics)
}

func TestParseDiagnosticsMissingBraceOneTotal(t *testing.T) {
	for _, src := range []string{
		"sub a {\n  my $x = 1;\n",
		"my $x;\nif ($x) { print \"a\";\n",
		"sub f {\n  if (1) {\n    print 1;\n",
	} {
		got := publishedFor(t, src)
		if assert.Len(t, got, 1, "%q: %v", src, codes(got)) {
			assert.Contains(t, []string{string(parse.UnclosedBrace), string(parse.UnclosedOpener), CodeMissingToken}, got[0].Code, src)
		}
	}
}

func TestParseDiagnosticsCapAtTwenty(t *testing.T) {
	got := publishedFor(t, farApart(25))
	require.Len(t, got, 20)
	for i, d := range got {
		assert.Equal(t, string(parse.CallArity), d.Code)
		assert.Equal(t, i, d.Range.Start.Line, "the first twenty are published, in order")
	}
}

func TestParseDiagnosticsCooldownBoundary(t *testing.T) {
	// The first error is the statement `push();`, bytes [0, 7); the second
	// starts gap bytes after its end, the bytes between a comment.
	at := func(gap int) string {
		return "push();#" + strings.Repeat("x", gap-2) + "\npush();\n"
	}
	for gap, want := range map[int]int{39: 1, 40: 2} {
		src := at(gap)
		require.Equal(t, 7+gap, strings.LastIndex(src, "push"), "fixture")
		got := publishedFor(t, src)
		assert.Len(t, got, want, "second error %d bytes after the first's end: %v", gap, codes(got))
	}

	var f diagFilter
	assert.True(t, f.admit(infer.Diagnostic{StartByte: 0, EndByte: 6}))
	assert.False(t, f.admit(infer.Diagnostic{StartByte: 45, EndByte: 50}), "39 bytes after")
	assert.True(t, f.admit(infer.Diagnostic{StartByte: 46, EndByte: 50}), "40 bytes after")
}

func TestParseDiagnosticsFilterFreshPerPublish(t *testing.T) {
	s := newLSPSession(t)
	assert.Len(t, s.open(untitledURI, farApart(25)), 20)
	assert.Len(t, s.change(untitledURI, 2, farApart(25)+"\n"), 20, "a later publication starts a fresh count")
	assert.Len(t, s.open("untitled:Untitled-2", farApart(25)), 20, "another document starts a fresh count")

	// The cooldown does not carry either: an error at the start of a fresh
	// publication is published though the last one ended close before it.
	assert.Len(t, s.open("untitled:Untitled-3", "push();\n"), 1)
	assert.Len(t, s.change("untitled:Untitled-3", 2, "push();\n\n"), 1)
}

func TestParseDiagnosticsClearedWhenFixed(t *testing.T) {
	s := newLSPSession(t)
	require.Len(t, s.open(untitledURI, "my $x = ;\n"), 1)
	s.notify("textDocument/didChange", lspDidChangeParams{
		TextDocument:   lspVersionedTextDocumentIdentifier{URI: untitledURI, Version: 2},
		ContentChanges: []lspTextDocumentContentChangeEvent{{Text: "my $x = 1;\n"}},
	})
	settle(s.h, untitledURI)
	raw := s.out.String()
	assert.Contains(t, raw, `"diagnostics":[]`, "an empty list is published, not null: %s", raw)
	assert.Empty(t, s.published(untitledURI).Diagnostics)
}

func TestTypeDiagnosticsGatedOnParseErrors(t *testing.T) {
	const undef = "my $u = undef;\n"
	const typeErr = "my $w = $u + 1;\n"
	for _, c := range []struct {
		name, src string
		want      []string
	}{
		{"operand missing after a mistyped operand", undef + "my $y = $u + 1 + ;\n", []string{"missing_operand"}},
		{"elsewhere is still checked", undef + typeErr + "my $y = $u + 1 + ;\n", []string{infer.CodeCoercionMismatch, "missing_operand"}},
		{"inside a sub with an unclosed brace", "package P {\n  sub a {\n" + undef + typeErr + "\n" + tenSubs("  ") + "}\n", []string{string(parse.UnclosedOpener)}},
		{"control: no parse error", undef + typeErr, []string{infer.CodeCoercionMismatch}},
	} {
		assert.Equal(t, c.want, codes(publishedFor(t, c.src)), c.name)
	}
}

func TestTypeDiagnosticsGatedAtSpanEdges(t *testing.T) {
	gates := []span{{start: 10, end: 20}, {start: 30, end: 30}}
	ty := func(start, end uint32) infer.Diagnostic { return infer.Diagnostic{StartByte: start, EndByte: end} }
	for _, c := range []struct {
		d    infer.Diagnostic
		kept bool
	}{
		{ty(5, 11), false},  // partly overlapping the start
		{ty(19, 25), false}, // partly overlapping the end
		{ty(12, 15), false}, // inside
		{ty(5, 10), true},   // ending exactly where the error starts
		{ty(20, 25), true},  // starting exactly where it ends
		{ty(29, 31), false}, // holding a zero-width error's byte
		{ty(25, 30), true},  // ending at a zero-width error
		{ty(30, 35), false}, // starting at a zero-width error
	} {
		assert.Equal(t, c.kept, len(gateTypes([]infer.Diagnostic{c.d}, gates)) == 1, "%+v", c.d)
	}

	// Through the LSP: the type error ends at the `;` before the broken
	// statement, which starts right after it, and is kept.
	got := publishedFor(t, "my $u = undef; my $w = $u + 1;my $y = ;\n")
	assert.Equal(t, []string{infer.CodeCoercionMismatch, "missing_operand"}, codes(got))
}

func TestTypeDiagnosticsNotGatedByCatchAll(t *testing.T) {
	src := "use Test::More;\nsubtest 'name' => sub { ok(1); };\nmy $u = undef; my $x = $u + 1;\n"
	var catchAll []*parse.Node
	var walk func(n *parse.Node)
	walk = func(n *parse.Node) {
		if n.Refused() && parse.CatchAll(n.Refusal) {
			catchAll = append(catchAll, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(parse.Parse([]byte(src)))
	require.Len(t, catchAll, 1, "fixture: the subtest is one catch-all")
	assert.Equal(t, parse.TrailingTokens, catchAll[0].Refusal)

	assert.Equal(t, []string{infer.CodeCoercionMismatch}, codes(publishedFor(t, src)),
		"the catch-all publishes nothing and gates nothing")
}

func TestParseDiagnosticsMissingOperandOneDiagnostic(t *testing.T) {
	for _, src := range []string{
		"my $x = ;\n",
		"sub f { 2 * }\n",
		"my $x = 2 *",
		"f(1 +",
	} {
		got := publishedFor(t, src)
		if assert.Len(t, got, 1, "%q: %v", src, codes(got)) && src != "f(1 +" {
			assert.Equal(t, "missing_operand", got[0].Code, src)
			assert.Equal(t, got[0].Range.Start, got[0].Range.End, "%q: zero-width", src)
		}
	}
	// `f(1 +` is one unclosed_opener holding the operand's Missing node.
	assert.Equal(t, []string{string(parse.UnclosedOpener)}, codes(publishedFor(t, "f(1 +")))

	for code := range parse.RefusalSites {
		quiet := code == parse.NotATerm || code == parse.TrailingTokens || code == parse.NotAnExpression
		assert.Equal(t, quiet, parse.CatchAll(code), "CatchAll(%s)", code)
	}
}

func TestParseDiagnosticMissingAtEOFRange(t *testing.T) {
	for _, src := range []string{
		"my $s = \"é😀\"; my $x = 1 +",
		"sub a {\n  my $s = \"é😀\";",
		"my $s = \"é😀\";\nmy $x = (1,",
	} {
		got := publishedFor(t, src)
		require.Len(t, got, 1, "%q: %v", src, codes(got))
		lines := strings.Split(src, "\n")
		last := lspPosition{Line: len(lines) - 1, Character: len(utf16.Encode([]rune(lines[len(lines)-1])))}
		r := got[0].Range
		assert.False(t, after(r.Start, r.End), "%q: start %v after end %v", src, r.Start, r.End)
		assert.False(t, after(r.End, last), "%q: end %v past the document's end %v", src, r.End, last)
		assert.Equal(t, last, r.End, "%q: the error reaches end of input, in UTF-16 columns", src)
	}
}

func after(a, b lspPosition) bool {
	return a.Line > b.Line || a.Line == b.Line && a.Character > b.Character
}
