// ABOUTME: Holds the token cache to a fresh Tokenize* after every edit, in all three start states (spec §6.3.2).
// ABOUTME: Also the resync rule (Kind, shifted span, Flags and State all equal), its locality on edit5k.pl, and its edge cases.

package lexer

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// cacheEdit is one edit: bytes [at, at+del) of the source replaced by ins.
type cacheEdit struct {
	at, del int
	ins     string
}

// apply returns src with e applied, and the edit's end in the result.
func (e cacheEdit) apply(src []byte) ([]byte, int) {
	out := append(append(bytes.Clone(src[:e.at]), e.ins...), src[e.at+e.del:]...)
	return out, e.at + len(e.ins)
}

// editCache applies e to src and to c, and returns the edited source and the
// number of tokens c re-lexed.
func editCache(c *TokenCache, src []byte, e cacheEdit) ([]byte, int) {
	out, newEnd := e.apply(src)
	return out, c.Edit(out, e.at, e.at+e.del, newEnd)
}

// checkCache asserts c holds tokenize(src) and that resuming from each of its
// checkpoints yields that stream's suffix, so a later edit can start there.
func checkCache(t *testing.T, c *TokenCache, src []byte, tokenize func([]byte) []Token) {
	t.Helper()
	full := tokenize(src)
	if !reflect.DeepEqual(c.Tokens, full) && len(c.Tokens)+len(full) > 0 {
		t.Fatalf("cache differs from a fresh tokenization of %q: %s", src, firstTokenDiff(src, c.Tokens, full))
	}
	if len(c.Checkpoints) == 0 || c.Checkpoints[0].Byte != 0 {
		t.Fatalf("the cache's first checkpoint is not byte 0")
	}
	for i, cp := range c.Checkpoints {
		if i > 0 && cp.Byte <= c.Checkpoints[i-1].Byte {
			t.Fatalf("checkpoints out of order at %d", cp.Byte)
		}
		j := suffixAt(full, cp.Byte)
		if j < 0 || cp.Byte > len(src) {
			t.Fatalf("checkpoint at %d is not a token boundary of %q", cp.Byte, src)
		}
		if got, want := Resume(src, cp), full[j:]; !reflect.DeepEqual(got, want) && len(got)+len(want) > 0 {
			t.Fatalf("resuming the cache's checkpoint at %d of %q: %s", cp.Byte, src, firstTokenDiff(src, got, want))
		}
	}
	// No failed lookahead in a fresh lex reads past a checkpoint the cache
	// kept: the tokens before each depend on no byte after it.
	peek := peekEnds(src, c.Checkpoints[0].State)
	for _, cp := range c.Checkpoints {
		if peek[cp.Byte] > cp.Byte {
			t.Fatalf("checkpoint at %d of %q: a lookahead before it read to %d", cp.Byte, src, peek[cp.Byte])
		}
	}
}

// peekEnds lexes src from start and maps each token boundary to the lexer's
// peekEnd there.
func peekEnds(src []byte, start State) map[int]int {
	l := start.Clone().l
	l.src = src
	m := map[int]int{0: 0}
	for l.pos < len(l.src) {
		l.step(scanOne)
		m[l.pos] = l.peekEnd
	}
	return m
}

// incrementalEdits are spec §7.7.3's eight edits, the seeds of
// internal/parse/testdata/fuzz/FuzzIncremental.
var incrementalEdits = map[string]struct {
	src string
	e   cacheEdit
}{
	"insert quote":              {editSub, cacheEdit{20, 0, "\""}},
	"insert heredoc":            {editSub, cacheEdit{23, 0, "  print <<EOF;\n"}},
	"delete heredoc terminator": {"sub f {\n  print <<EOF;\nbody\nEOF\n  return 1;\n}\nprint f();\n", cacheEdit{28, 4, ""}},
	"insert pod":                {editSub, cacheEdit{23, 0, "=pod\n"}},
	"insert __END__":            {editSub, cacheEdit{23, 0, "__END__\n"}},
	"prototype":                 {"sub f() { 1 }\nmy @a = (1, 2);\nprint f @a;\n", cacheEdit{6, 0, "\\@"}},
	"insert brace":              {editSub, cacheEdit{23, 0, "{\n"}},
	"split token":               {editSub, cacheEdit{15, 0, "|"}},
}

const editSub = "sub f {\n  my $foo = 1;\n  return $foo;\n}\nprint f();\n"

// fixtureEdits are a few edits for each resume fixture: at the start, the middle and
// the end, inserting, deleting and replacing.
func fixtureEdits(src string) []cacheEdit {
	n := len(src)
	return []cacheEdit{
		{0, 0, "x"}, {0, min(3, n), ""}, {n / 2, 0, ";"}, {n / 2, min(2, n-n/2), "{"},
		{n / 3, 0, "\""}, {n / 3, 0, "<<A;\n"}, {n / 4, min(5, n-n/4), "/ 2;"}, {n, 0, "\nfoo();"},
	}
}

// resyncEdits are edits where reusing old tokens would be wrong although an
// old checkpoint looks usable.
var resyncEdits = map[string]struct {
	src string
	e   cacheEdit
}{
	// The old checkpoint after `a;` shifts to after the inserted `x;`, but
	// what follows it is now `1`, not `b`.
	"checkpoint inside the edit": {"a;b;c;\n", cacheEdit{2, 1, "x;1"}},
	// A `<` in term position looked ahead past the `;` for a `>` and, finding
	// none, was an Operator: the `>` the edit adds makes it a Readline, so
	// the `;` after it is no checkpoint (review of #510).
	"readline lookahead at EOF":   {"<;", cacheEdit{2, 0, ">"}},
	"readline lookahead past a ;": {"my @l = <STDIN;\nfoo();\n", cacheEdit{len("my @l = <STDIN;"), 0, ">"}},
	// The same for a prototype's search for its `)` (fuzz find).
	"prototype lookahead past a ;": {"sub A(0};0", cacheEdit{len("sub A(0};0"), 0, ")"}},
	"heredoc lookahead past a ;":   {"print <<'; foo();\nbar();\n", cacheEdit{len("print <<'; foo();\nbar();\n"), 0, "'"}},
}

func TestTokenCacheEditMatchesFull(t *testing.T) {
	for _, s := range resumeStarts {
		for name, c := range resyncEdits {
			t.Run(s.name+"/"+name, func(t *testing.T) {
				b := []byte(c.src)
				cache := newTokenCache(b, s.start, 0)
				out, _ := editCache(cache, b, c.e)
				checkCache(t, cache, out, s.tokenize)
			})
		}
		for name, c := range incrementalEdits {
			t.Run(s.name+"/"+name, func(t *testing.T) {
				b := []byte(c.src)
				for _, stride := range []int{0, checkpointStride} {
					cache := newTokenCache(b, s.start, stride)
					out, _ := editCache(cache, b, c.e)
					checkCache(t, cache, out, s.tokenize)
				}
			})
		}
		for name, src := range resumeFixtures {
			t.Run(s.name+"/"+name, func(t *testing.T) {
				b := []byte(src)
				for _, e := range fixtureEdits(src) {
					cache := newTokenCache(b, s.start, 0)
					out, _ := editCache(cache, b, e)
					checkCache(t, cache, out, s.tokenize)
				}
			})
		}
	}
}

// latencyEdit is the latency benchmarks' 1-char edit (internal/psc
// editOffset): a digit before the `0` of `$_ > 0` in the middle `sub scale`.
func latencyEdit(t *testing.T) ([]byte, cacheEdit) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
	if err != nil {
		t.Fatal(err)
	}
	var subs []int
	for i := 0; ; {
		j := bytes.Index(src[i:], []byte("\nsub scale {"))
		if j < 0 {
			break
		}
		subs, i = append(subs, i+j), i+j+1
	}
	if len(subs) == 0 {
		t.Fatal("edit5k.pl has no `sub scale`")
	}
	sub := subs[len(subs)/2]
	return src, cacheEdit{sub + bytes.Index(src[sub:], []byte("$_ > 0")) + len("$_ > "), 0, "4"}
}

func TestTokenCacheResyncIsLocal(t *testing.T) {
	src, e := latencyEdit(t)
	cache := NewTokenCache(src, StartState())
	out, relexed := editCache(cache, src, e)
	checkCache(t, cache, out, Tokenize)
	if total := len(cache.Tokens); relexed*20 >= total {
		t.Fatalf("a 1-char edit re-lexed %d of %d tokens, want fewer than 5%%", relexed, total)
	} else {
		t.Logf("a 1-char edit re-lexed %d of %d tokens", relexed, total)
	}
}

// TestTokenCacheResyncRequiresEqualFlagsAndState tampers with the old token or
// the old checkpoint state at the point an edit would resync, and requires the
// cache to look past it to the next one. A resync that took the tampered point
// would re-lex no more than the untampered edit does.
func TestTokenCacheResyncRequiresEqualFlagsAndState(t *testing.T) {
	src := []byte("my $x = 1;\nif ($a) { f(); }\nmy $y = 2;\nmy $z = 3;\n")
	e := cacheEdit{strings.Index(string(src), "$a"), 2, "$ab"}
	brace := strings.Index(string(src), "}")
	baseline := func() *TokenCache { return newTokenCache(src, StartState(), 0) }
	_, want := editCache(baseline(), src, e)

	tamper := map[string]func(c *TokenCache){
		"Flags": func(c *TokenCache) {
			i := suffixAt(c.Tokens, brace)
			if !c.Tokens[i].OpensBlock {
				t.Fatal("the block's `}` does not carry OpensBlock")
			}
			c.Tokens[i].OpensBlock = false
		},
		"State": func(c *TokenCache) {
			for i := range c.Checkpoints {
				if c.Checkpoints[i].Byte == brace+1 {
					c.Checkpoints[i].State.l.utf8Pragma = true
					return
				}
			}
			t.Fatal("no checkpoint after the block's `}`")
		},
	}
	for name, f := range tamper {
		t.Run(name, func(t *testing.T) {
			c := baseline()
			f(c)
			out, got := editCache(c, src, e)
			if got <= want {
				t.Fatalf("re-lexed %d tokens with the resync point's %s tampered, %d without: it resynced there", got, name, want)
			}
			checkCache(t, c, out, Tokenize)
		})
	}
}

func TestTokenCacheEditAtOffsetZeroAndEOF(t *testing.T) {
	for _, s := range resumeStarts {
		for _, src := range []string{"my $x = 1; foo();\nbar();\n", "print <<A;\nbody\nA\nz();\n", ""} {
			n := len(src)
			for _, e := range []cacheEdit{
				{0, 0, "my $w = 0;\n"}, {0, min(1, n), ""}, {n, 0, "baz();\n"}, {n, 0, "\""},
				{0, n, ""}, {0, n, "my $q = 1;\n"},
			} {
				b := []byte(src)
				cache := newTokenCache(b, s.start, 0)
				out, _ := editCache(cache, b, e)
				checkCache(t, cache, out, s.tokenize)
				if e.at == 0 && e.del == n && e.ins == "" && len(cache.Tokens) != 0 {
					t.Errorf("%s: deleting all of %q leaves %d tokens", s.name, src, len(cache.Tokens))
				}
			}
		}
		// Inserting into an empty buffer, then again at its end.
		cache := newTokenCache(nil, s.start, 0)
		out, _ := editCache(cache, nil, cacheEdit{0, 0, "my $x = 1;\n"})
		checkCache(t, cache, out, s.tokenize)
		out, _ = editCache(cache, out, cacheEdit{len(out), 0, "foo();\n"})
		checkCache(t, cache, out, s.tokenize)
	}
}

// TestTokenCacheNoResyncTerminates inserts a backtick, which edit5k.pl holds
// none of, so the string it opens runs to EOF and nothing resyncs. The re-lex
// is then one pass to EOF, not one per candidate resync point.
func TestTokenCacheNoResyncTerminates(t *testing.T) {
	src, _ := latencyEdit(t)
	if bytes.IndexByte(src, '`') >= 0 {
		t.Fatal("edit5k.pl holds a backtick: the inserted one would close")
	}
	at := bytes.Index(src, []byte("use strict;"))
	cache := NewTokenCache(src, StartState())
	out, relexed := editCache(cache, src, cacheEdit{at, 0, "`"})
	checkCache(t, cache, out, Tokenize)
	if fresh := len(cache.Tokens) - suffixAt(cache.Tokens, at); relexed < fresh {
		t.Fatalf("re-lexed %d tokens of the %d from the edit to EOF: it resynced", relexed, fresh)
	}
	if relexed > len(cache.Tokens) {
		t.Fatalf("re-lexed %d tokens for a %d-token file", relexed, len(cache.Tokens))
	}
}

func TestTokenCacheShiftNoUnderflow(t *testing.T) {
	for _, src := range []string{
		"my $x = 1;\nmy $y = 2;\nfoo();\n",
		"use strict;\nsub f { 1 }\nprint f();\n",
	} {
		b := []byte(src)
		first := strings.IndexAny(src, ";}") + 1
		for _, stride := range []int{0, checkpointStride} {
			cache := newTokenCache(b, StartState(), stride)
			out, _ := editCache(cache, b, cacheEdit{0, first, ""})
			checkCache(t, cache, out, Tokenize)
			for _, tok := range cache.Tokens {
				if tok.Start < 0 || tok.Start > tok.End || tok.End > len(out) {
					t.Fatalf("%q: token %d..%d out of 0..%d", src, tok.Start, tok.End, len(out))
				}
			}
			for _, cp := range cache.Checkpoints {
				if cp.Byte < 0 || cp.Byte > len(out) {
					t.Fatalf("%q: checkpoint at %d out of 0..%d", src, cp.Byte, len(out))
				}
			}
		}
	}
}

func TestTokenCacheCheckpointsInvalidatedByNewHeredoc(t *testing.T) {
	src := []byte("my $a = 1;\nmy $b = 2;\nmy $c = 3;\nEOF\nmy $d = 4;\nmy $e = 5;\n")
	cache := newTokenCache(src, StartState(), 0)
	inside := len(cache.Checkpoints)
	out, _ := editCache(cache, src, cacheEdit{len("my $a = 1;\n"), 0, "print <<EOF;\n"})
	checkCache(t, cache, out, Tokenize)
	var body Token
	for _, tok := range cache.Tokens {
		if tok.Kind == HeredocBody {
			body = tok
		}
	}
	if body.Kind != HeredocBody {
		t.Fatal("the inserted <<EOF has no body")
	}
	after := 0
	for _, cp := range cache.Checkpoints {
		if body.Start <= cp.Byte && cp.Byte < body.End {
			t.Errorf("a checkpoint at %d lies inside the heredoc body %d..%d", cp.Byte, body.Start, body.End)
		}
		if cp.Byte >= body.End {
			after++
		}
	}
	if after == 0 || after >= inside {
		t.Errorf("%d checkpoints after the body, of %d before the edit", after, inside)
	}
}

// lookaheadSequences are edit pairs whose first edit resyncs at an old
// checkpoint a lookahead in the new stream reads past, and whose second then
// edits after that checkpoint (re-review of #510).
var lookaheadSequences = []editSequence{
	{"a;  \nb;\n", 0, []cacheEdit{{0, 0, "<"}, {4, 0, ">"}}},
	stdinSequence(),
}

type editSequence struct {
	src    string
	stride int
	seq    []cacheEdit
}

// stdinSequence types `<` before STDIN past the first 512-byte stride, then
// `>` after the `foo()` that follows its `;`.
func stdinSequence() editSequence {
	src := strings.Repeat("1;\n", 170) + "my @l = STDIN; foo();\nbar();\n"
	foo := strings.Index(src, "foo()") + len("foo()")
	return editSequence{src, checkpointStride, []cacheEdit{{strings.Index(src, "STDIN"), 0, "<"}, {foo + 1, 0, ">"}}}
}

func TestTokenCacheEditSequenceMatchesFull(t *testing.T) {
	for i, c := range lookaheadSequences {
		for _, s := range resumeStarts {
			cur := []byte(c.src)
			cache := newTokenCache(cur, s.start, c.stride)
			for j, e := range c.seq {
				cur, _ = editCache(cache, cur, e)
				t.Run(s.name+"/lookahead"+itoa(i)+"/edit"+itoa(j), func(t *testing.T) { checkCache(t, cache, cur, s.tokenize) })
			}
		}
	}
	src := []byte("sub f {\n  print <<EOF;\nbody;\nEOF\n  return 1;\n}\nmy $x = f();\nprint $x / 2;\n")
	term := bytes.Index(src, []byte("EOF\n  return"))
	seq := []cacheEdit{
		{term, 4, ""},                  // delete the terminator: the body runs to EOF
		{term, 0, "EOF\n"},             // put it back
		{0, 0, "my $y = 1;\n"},         // shift everything
		{term + 11, 0, "EOF\n"},        // a terminator one line early
		{term + 11, 4, ""},             // and gone again
		{len("my $y = 1;\n"), 0, "\""}, // an unterminated string
		{len("my $y = 1;\n"), 1, ""},   // closed by deleting it
		{len("my $y = 1;\n") + 4, 0, "{"},
	}
	for _, s := range resumeStarts {
		for _, stride := range []int{0, 16, checkpointStride} {
			cur := src
			cache := newTokenCache(cur, s.start, stride)
			for i, e := range seq {
				cur, _ = editCache(cache, cur, e)
				t.Run(s.name+"/edit"+itoa(i), func(t *testing.T) { checkCache(t, cache, cur, s.tokenize) })
			}
		}
	}
}

// FuzzTokenCacheEdit is the cache property on any input and any two edits,
// for each of the three start states: after the first edit, after a second
// at an independent offset, and after the edit that undoes the second, the
// cache equals a fresh Tokenize* of the bytes and every checkpoint it holds
// resumes to that stream's suffix with no lookahead reading past it.
//
//	go test ./internal/lexer/ -run '^$' -fuzz '^FuzzTokenCacheEdit$' -fuzztime 2000000x
func FuzzTokenCacheEdit(f *testing.F) {
	for _, c := range incrementalEdits {
		f.Add(c.src, c.e.at, c.e.del, c.e.ins, 0, 0, "")
	}
	for _, c := range resyncEdits {
		f.Add(c.src, c.e.at, c.e.del, c.e.ins, 0, 0, "")
	}
	for _, c := range lookaheadSequences {
		f.Add(c.src, c.seq[0].at, c.seq[0].del, c.seq[0].ins, c.seq[1].at, c.seq[1].del, c.seq[1].ins)
	}
	for _, src := range resumeFixtures {
		for _, e := range fixtureEdits(src) {
			f.Add(src, e.at, e.del, e.ins, len(src)-e.at, 0, e.ins)
		}
	}
	for _, s := range fuzzSeeds {
		f.Add(s, len(s)/2, 1, ";", len(s), 0, "x")
	}
	f.Fuzz(func(t *testing.T, src string, at, del int, ins string, at2, del2 int, ins2 string) {
		b := []byte(src)
		e := clampEdit(len(b), at, del, ins)
		mid, _ := e.apply(b)
		e2 := clampEdit(len(mid), at2, del2, ins2)
		undo := cacheEdit{e2.at, len(e2.ins), string(mid[e2.at : e2.at+e2.del])}
		// A stride keeps the checkpoints, and so checkCache's resumes, to
		// some 64 rather than one per statement.
		stride := (len(b) + len(ins)) / 64
		for _, s := range resumeStarts {
			cache := newTokenCache(b, s.start, stride)
			cur := b
			for _, e := range []cacheEdit{e, e2, undo} {
				cur, _ = editCache(cache, cur, e)
				checkCache(t, cache, cur, s.tokenize)
			}
		}
	})
}

// clampEdit is the edit at, del, ins made valid for an n-byte source.
func clampEdit(n, at, del int, ins string) cacheEdit {
	at = min(max(at, 0), n)
	return cacheEdit{at, min(max(del, 0), n-at), ins}
}
