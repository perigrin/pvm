// ABOUTME: Tests for didChange under parse.IncrementalRepair: range edits in UTF-16, ReparseIncremental, the advertised sync kind.
// ABOUTME: Each sequence of edits must leave the text, tree and diagnostics a fresh didOpen of the same text gives.

package psc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parse"
)

// flagGuard holds parse.IncrementalRepair at a test's value until every
// reparse that might read it has returned: a parse superseded by a later
// edit can still be running when the test ends, and restoring the flag
// under it is a race.
type flagGuard struct {
	mu      sync.Mutex
	cond    *sync.Cond
	running int
	closed  bool
}

// syncHandler is an initialized handler made, as a client meets it, with
// parse.IncrementalRepair set to on. reparses counts the parses compute
// re-parsed from a landed tree, with a module cache. The flag is restored once the server is
// shut down and no reparse is running.
func syncHandler(t *testing.T, on bool) (h *handler, out *bytes.Buffer, init lspInitializeResult, reparses *atomic.Int32) {
	t.Helper()
	was := parse.IncrementalRepair
	parse.IncrementalRepair = on
	h, out = newTestHandler(t)
	g := &flagGuard{}
	g.cond = sync.NewCond(&g.mu)
	reparses = &atomic.Int32{}
	reparse := h.server.compute
	h.server.compute = func(ctx context.Context, uri string, text []byte, prev *document, cache *parse.ModuleCache) (*document, error) {
		g.mu.Lock()
		if g.closed {
			g.mu.Unlock()
			return nil, context.Canceled
		}
		g.running++
		g.mu.Unlock()
		defer func() {
			g.mu.Lock()
			g.running--
			g.cond.Broadcast()
			g.mu.Unlock()
		}()
		if prev != nil && cache != nil {
			reparses.Add(1)
		}
		return reparse(ctx, uri, text, prev, cache)
	}
	t.Cleanup(func() {
		h.server.Shutdown()
		g.mu.Lock()
		g.closed = true
		for g.running > 0 {
			g.cond.Wait()
		}
		g.mu.Unlock()
		parse.IncrementalRepair = was
	})

	rawID := json.RawMessage(`1`)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &rawID, Method: "initialize", Params: json.RawMessage(`{}`)}))
	require.NoError(t, json.Unmarshal(readResponse(t, out).Result, &init))
	return h, out, init, reparses
}

// didOpen opens uri holding text at version 1 and waits for its parse.
func didOpen(t *testing.T, h *handler, uri, text string) {
	t.Helper()
	params, err := json.Marshal(lspDidOpenParams{TextDocument: lspTextDocumentItem{URI: uri, LanguageID: "perl", Version: 1, Text: text}})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: params}))
	settle(h, uri)
}

// didChange sends one didChange of uri at version holding changes, and
// waits for its parse.
func didChange(t *testing.T, h *handler, uri string, version int, changes ...lspTextDocumentContentChangeEvent) {
	t.Helper()
	params, err := json.Marshal(lspDidChangeParams{
		TextDocument:   lspVersionedTextDocumentIdentifier{URI: uri, Version: version},
		ContentChanges: changes,
	})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didChange", Params: params}))
	settle(h, uri)
}

// replace is the range change replacing the first old in text with new,
// its positions in UTF-16 units, and the text it leaves.
func replace(t *testing.T, text, old, new string) (lspTextDocumentContentChangeEvent, string) {
	t.Helper()
	i := strings.Index(text, old)
	require.GreaterOrEqual(t, i, 0, "%q is not in %q", old, text)
	d := NewDocument([]byte(text))
	r := lspRange{Start: d.ByteToPosition(uint32(i)), End: d.ByteToPosition(uint32(i + len(old)))}
	return lspTextDocumentContentChangeEvent{Range: &r, Text: new}, text[:i] + new + text[i+len(old):]
}

// published is the last diagnostics out holds for uri, draining it.
func published(t *testing.T, out *bytes.Buffer, uri string) []lspDiagnostic {
	t.Helper()
	var last []lspDiagnostic
	found := false
	for tr := newTransport(out, io.Discard); ; {
		msg, err := tr.readMessage()
		if err != nil {
			break
		}
		var p lspPublishDiagnosticsParams
		if msg.Method == "textDocument/publishDiagnostics" && json.Unmarshal(msg.Params, &p) == nil && p.URI == uri {
			last, found = p.Diagnostics, true
		}
	}
	require.True(t, found, "%s: nothing published", uri)
	return last
}

// matchesFreshOpen asserts uri in h holds text and that its tree and its
// latest diagnostics equal those a fresh server's didOpen of text gives.
func matchesFreshOpen(t *testing.T, h *handler, out *bytes.Buffer, uri, text string) {
	t.Helper()
	got := h.server.snapshot(uri)
	require.NotNil(t, got)
	require.Equal(t, text, string(got.text.Text), "the document")
	fresh, freshOut := newTestHandler(t)
	t.Cleanup(fresh.server.Shutdown)
	fresh.initialized = true
	didOpen(t, fresh, uri, text)
	want := fresh.server.snapshot(uri)
	assert.Empty(t, parse.TreeDiff(got.tree, want.tree), "the tree")
	assert.Equal(t, published(t, freshOut, uri), published(t, out, uri), "the diagnostics")
}

// multiByte is Perl with two- and three-byte UTF-8 and an astral character,
// which UTF-16 counts as two units, on the lines the edits touch.
const multiByte = "my $café = \"naïve 😀 → ok\";\n" +
	"sub greet {\n  my ($n) = @_;\n  return \"héllo $n\";\n}\n" +
	"print greet(\"wörld 😀\");\n" +
	"my $count = 1 + 2;\n"

func TestLSPSyncKindFollowsFlag(t *testing.T) {
	for on, want := range map[bool]int{false: 1, true: 2} {
		_, _, init, _ := syncHandler(t, on)
		assert.Equal(t, want, init.Capabilities.TextDocumentSync, "IncrementalRepair %v", on)
	}
}

func TestDidChangeUsesReparseIncrementalWhenFlagSet(t *testing.T) {
	const uri = "untitled:Untitled-1"
	h, out, _, reparses := syncHandler(t, true)
	didOpen(t, h, uri, multiByte)
	text := multiByte
	for v, e := range [][2]string{
		{"1 + 2", "1 + 2 + 3"},
		{"wörld 😀", "wörld 😀😀 ü"},
		{"héllo", "hé llo"},
		{"naïve 😀 →", "naïve →"},
		{"return", "return  "},
	} {
		c, next := replace(t, text, e[0], e[1])
		didChange(t, h, uri, v+2, c)
		text = next
		matchesFreshOpen(t, h, out, uri, text)
	}
	assert.Positive(t, reparses.Load(), "the owner re-parsed from its last tree")
	assert.Empty(t, parse.TreeDiff(h.server.snapshot(uri).tree, parse.Parse([]byte(text))), "control: the tree is Parse's")
}

func TestDidChangeRangeEditsMatchFullSync(t *testing.T) {
	const uri = "untitled:Untitled-2"
	ranged, rangedOut, _, _ := syncHandler(t, true)
	full, fullOut := newTestHandler(t)
	t.Cleanup(full.server.Shutdown)
	full.initialized = true // full sync: incremental unset
	didOpen(t, ranged, uri, multiByte)
	didOpen(t, full, uri, multiByte)

	text := multiByte
	for v, e := range [][2]string{
		{"😀 →", "😀😀 ←"},
		{"my ($n)", "my ($n, $ü)"},
		{"\n}\n", "\n  1;\n}\n"},
		{"print", "say"},
		{"wörld 😀", ""},
	} {
		c, next := replace(t, text, e[0], e[1])
		text = next
		didChange(t, ranged, uri, v+2, c)
		didChange(t, full, uri, v+2, lspTextDocumentContentChangeEvent{Text: text})
		r, f := ranged.server.snapshot(uri), full.server.snapshot(uri)
		require.Equal(t, string(f.text.Text), string(r.text.Text), "after %q", e)
		assert.Empty(t, parse.TreeDiff(r.tree, f.tree), "after %q", e)
		assert.Equal(t, published(t, fullOut, uri), published(t, rangedOut, uri), "after %q", e)
	}
	assert.Equal(t, text, string(full.server.Source(uri)))
}

// findCalled is the first Call named name under n.
func findCalled(n *parse.Node, name string) *parse.Node {
	if n.Kind == parse.Call && n.Text == name {
		return n
	}
	for _, c := range n.Children {
		if got := findCalled(c, name); got != nil {
			return got
		}
	}
	return nil
}

func TestDidChangeFileURIReparsesThroughSession(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib")
	require.NoError(t, os.MkdirAll(filepath.Join(lib, "My"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(lib, "My", "Mod.pm"),
		[]byte("package My::Mod;\nour @EXPORT = qw(twice);\nsub twice($) { }\n1;\n"), 0o644))
	t.Setenv("PERL5LIB", lib)
	path := filepath.Join(dir, "main.pl")
	uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
	roots, _ := incRoots(path)
	require.Contains(t, roots, lib, "the local lib is a root")

	h, out, _, reparses := syncHandler(t, true)
	text := "use My::Mod;\nmy @r = (twice 1, 2);\nmy $s = \"ünïcödé 😀\";\n"
	didOpen(t, h, uri, text)
	for v, e := range [][2]string{
		{"1, 2", "3, 4"},
		{"😀\";\n", "😀\";\nmy $t = 😀;\n"},
		{"ünïcödé", "ü"},
	} {
		c, next := replace(t, text, e[0], e[1])
		didChange(t, h, uri, v+2, c)
		text = next
		got := h.server.snapshot(uri).tree
		fresh := parse.NewSession().ParseSourceFrom(path, []byte(text), roots...)
		require.Empty(t, parse.TreeDiff(got, fresh), "after %q", e)
		call := findCalled(got, "twice")
		require.NotNil(t, call, "after %q", e)
		assert.True(t, call.Resolved && call.Prototype == "($)", "after %q: twice resolved with ($), got %+v", e, call)
		matchesFreshOpen(t, h, out, uri, text)
	}
	assert.Positive(t, reparses.Load(), "the owner re-parsed from its last tree")
	assert.Contains(t, parse.LoadedModules(h.server.snapshot(uri).tree), "My::Mod")
}

func TestDidChangeFullParseWhenFlagOff(t *testing.T) {
	const uri = "untitled:Untitled-3"
	h, out, init, _ := syncHandler(t, false)
	require.Equal(t, 1, init.Capabilities.TextDocumentSync)
	var computed atomic.Int32
	compute := h.server.compute
	h.server.compute = func(ctx context.Context, uri string, text []byte, prev *document, cache *parse.ModuleCache) (*document, error) {
		if cache != nil {
			t.Error("a parse re-parsed from a module cache with the flag off")
		}
		computed.Add(1)
		return compute(ctx, uri, text, prev, cache)
	}
	didOpen(t, h, uri, multiByte)
	text := multiByte
	for v, e := range [][2]string{{"1 + 2", "1 + 2 + 3"}, {"héllo", "hello"}} {
		_, text = replace(t, text, e[0], e[1])
		didChange(t, h, uri, v+2, lspTextDocumentContentChangeEvent{Text: text})
		matchesFreshOpen(t, h, out, uri, text)
	}
	assert.EqualValues(t, 3, computed.Load(), "every parse is a full parse")
	assert.Nil(t, h.server.owner(uri).cache, "no module cache")
}

func TestDidChangeRangeSurrogateAndCRLF(t *testing.T) {
	const uri = "untitled:Untitled-4"
	// 😀 is UTF-16 units 1 and 2 of line 0; lines 0 and 1 end in CRLF.
	const open = "a😀b\r\nxyz\r\nlast"
	at := func(sl, sc, el, ec int, text string) lspTextDocumentContentChangeEvent {
		return lspTextDocumentContentChangeEvent{Range: &lspRange{Start: lspPosition{Line: sl, Character: sc}, End: lspPosition{Line: el, Character: ec}}, Text: text}
	}
	for _, c := range []struct {
		name   string
		change lspTextDocumentContentChangeEvent
		want   string
	}{
		{"inside a surrogate pair", at(0, 2, 0, 2, "X"), "aX😀b\r\nxyz\r\nlast"},
		{"range ending inside a surrogate pair", at(0, 0, 0, 2, "Y"), "Y😀b\r\nxyz\r\nlast"},
		{"past end of a CRLF line", at(1, 99, 1, 120, "!"), "a😀b\r\nxyz!\r\nlast"},
		{"to the CR of a CRLF line", at(0, 3, 0, 4, ""), "a😀\r\nxyz\r\nlast"},
		{"joining CRLF lines", at(0, 4, 1, 0, " "), "a😀b xyz\r\nlast"},
		{"past EOF", at(9, 0, 12, 3, ";\n"), "a😀b\r\nxyz\r\nlast;\n"},
		{"past the last line's end", at(2, 4, 2, 50, "ing"), "a😀b\r\nxyz\r\nlasting"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h, out, _, _ := syncHandler(t, true)
			didOpen(t, h, uri, open)
			didChange(t, h, uri, 2, c.change)
			matchesFreshOpen(t, h, out, uri, c.want)
		})
	}
}

func TestDidChangeMultipleContentChangesInOrder(t *testing.T) {
	const uri = "untitled:Untitled-5"
	h, out, _, _ := syncHandler(t, true)
	didOpen(t, h, uri, "my $é = 1;\n")
	at := func(l, sc, ec int, text string) lspTextDocumentContentChangeEvent {
		return lspTextDocumentContentChangeEvent{Range: &lspRange{Start: lspPosition{Line: l, Character: sc}, End: lspPosition{Line: l, Character: ec}}, Text: text}
	}
	// Each position is in the text the change before it left, so applied
	// to the opened text they would leave another one.
	didChange(t, h, uri, 2,
		at(0, 0, 0, "our $😀 = 0;\n"), // our $😀 = 0;\nmy $é = 1;\n
		at(0, 4, 7, "$x"),            // our $x = 0;\nmy $é = 1;\n
		lspTextDocumentContentChangeEvent{Text: "my $y = 2;\n"},
		at(1, 0, 0, "print $y;\n"),
		at(0, 8, 9, "3"),
	)
	matchesFreshOpen(t, h, out, uri, "my $y = 3;\nprint $y;\n")
}

func TestDidChangeMidParseUsesMatchingTree(t *testing.T) {
	const uri = "untitled:Untitled-6"
	h, out, _, _ := syncHandler(t, true)
	const v1, v2, v3 = "my $a = 1;\nprint $a;\n", "my $a = 12;\nprint $a;\n", "my $a = 12;\nprint $a, 😀;\n"
	inFlight, cancelled := make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	var prevs []int
	reparse := h.server.compute
	h.server.compute = func(ctx context.Context, uri string, text []byte, prev *document, cache *parse.ModuleCache) (*document, error) {
		if prev != nil {
			// The tree a re-parse reuses is the one its text was parsed to.
			assert.Empty(t, parse.TreeDiff(prev.tree, parse.Parse(prev.text.Text)), "prev's tree is of prev's bytes")
			mu.Lock()
			prevs = append(prevs, prev.version)
			mu.Unlock()
		}
		if string(text) == v2 {
			close(inFlight)
			<-ctx.Done()
			close(cancelled)
			// It returns a tree, as a parser that never polls would.
		}
		return reparse(ctx, uri, text, prev, cache)
	}
	didOpen(t, h, uri, v1)

	c, _ := replace(t, v1, "1;", "12;")
	params, err := json.Marshal(lspDidChangeParams{TextDocument: lspVersionedTextDocumentIdentifier{URI: uri, Version: 2}, ContentChanges: []lspTextDocumentContentChangeEvent{c}})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didChange", Params: params}))
	await(t, inFlight, "v2's parse starts")

	c, _ = replace(t, v2, "$a;", "$a, 😀;")
	didChange(t, h, uri, 3, c)
	await(t, cancelled, "v3 cancels v2's parse")
	matchesFreshOpen(t, h, out, uri, v3)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 1}, prevs, "v2 and v3 both re-parse from v1, the last tree that landed")
}

func TestFullSyncChangeToEditPrefixSuffix(t *testing.T) {
	for _, c := range []struct {
		old, text string
		want      parse.Edit
	}{
		{"my $x;", "my $x;", parse.Edit{Start: 6, OldEnd: 6}},
		{"abc", "xyz", parse.Edit{Start: 0, OldEnd: 3, NewText: "xyz"}},
		{"", "", parse.Edit{}},
		{"", "abc", parse.Edit{Start: 0, OldEnd: 0, NewText: "abc"}},
		{"abc", "", parse.Edit{Start: 0, OldEnd: 3}},
		{"aaa", "aa", parse.Edit{Start: 2, OldEnd: 3}},
		{"aa", "aaa", parse.Edit{Start: 2, OldEnd: 2, NewText: "a"}},
		{"abab", "ab", parse.Edit{Start: 2, OldEnd: 4}},
		{"xaax", "xaaax", parse.Edit{Start: 3, OldEnd: 3, NewText: "a"}},
		{"é", "è", parse.Edit{Start: 1, OldEnd: 2, NewText: "\xa8"}},
		{"say 1;\n", "say 12;\n", parse.Edit{Start: 5, OldEnd: 5, NewText: "2"}},
	} {
		e := fullSyncEdit([]byte(c.old), []byte(c.text))
		assert.Equal(t, c.want, e, "%q to %q", c.old, c.text)
		require.True(t, 0 <= e.Start && e.Start <= e.OldEnd && e.OldEnd <= len(c.old), "%q to %q: %+v is not an edit of old", c.old, c.text, e)
		assert.Equal(t, c.text, c.old[:e.Start]+e.NewText+c.old[e.OldEnd:], "%q to %q: applying the edit", c.old, c.text)
	}
}
