// ABOUTME: The lexer's structural invariants: lossless spans, monotonic positions, forward progress.
// ABOUTME: These hold for every input, valid Perl or not, and are what the round-trip metric rests on.

package lexer

import (
	"strings"
	"testing"
)

// TestTokenizeTrivia is the smallest thing that proves trivia are TOKENS
// rather than skipped. Every byte of input belongs to exactly one token, so
// whitespace-only input is one trivia token spanning the whole source -- not
// an empty token list.
//
// This is the decision that makes lossless round-trip reachable at all. A
// lexer that discards whitespace can never satisfy invariant 3, and
// retrofitting it later is a rewrite rather than an addition.
func TestTokenizeTrivia(t *testing.T) {
	const src = "  \t\n  "
	toks := Tokenize([]byte(src))

	if len(toks) != 1 {
		t.Fatalf("whitespace-only input gave %d tokens, want exactly 1: %v", len(toks), toks)
	}
	if toks[0].Kind != Whitespace {
		t.Errorf("Kind = %v, want Whitespace", toks[0].Kind)
	}
	if toks[0].Start != 0 || toks[0].End != len(src) {
		t.Errorf("span = [%d,%d), want [0,%d)", toks[0].Start, toks[0].End, len(src))
	}
}

// TestErrorKindsSpanSource pins the rule that an error is a token kind, not a
// return value.
//
// Tokenize has no error result, so the only way to say "I could not lex this"
// while keeping every byte accounted for is a token that covers those bytes.
// Spec §2.15 item 18 names two: Error for a byte that cannot start any token,
// and UnknownRest for a construct whose end could not be found.
//
// Later issues assert on these -- the identifier class wants a byte outside
// Latin-1 word characters to be an Error, and the lex.t gate wants NO Error
// tokens in the whole file -- so they are declared here or they are declared
// twice.
func TestErrorKindsSpanSource(t *testing.T) {
	// \x95 is not a word character in Latin-1 and cannot start any Perl
	// token. Verified against perl 5.42: "Unrecognized character \x95".
	const src = "\x95"
	toks := Tokenize([]byte(src))

	if len(toks) != 1 {
		t.Fatalf("got %d tokens, want 1: %v", len(toks), toks)
	}
	if toks[0].Kind != Error {
		t.Errorf("Kind = %v, want Error", toks[0].Kind)
	}
	if toks[0].Start != 0 || toks[0].End != 1 {
		t.Errorf("span = [%d,%d), want [0,1)", toks[0].Start, toks[0].End)
	}

	// Both kinds must exist and be distinct, or a later issue cannot tell
	// "this byte is wrong" from "this construct never ended".
	if Error == UnknownRest {
		t.Error("Error and UnknownRest are the same kind; they report different failures")
	}
}

// TestForwardProgressGuardFires is the one test in this package that cannot
// be written against the public API, and it is deliberate.
//
// Spec §7.6.2 invariant 4: every lexer step consumes at least one byte. A
// Perl lexer WILL infinite-loop during development -- an unterminated
// heredoc or quote-like operator that fails to advance the cursor is the
// classic case -- and `go test -fuzz` does not detect infinite loops. It
// hangs.
//
// There is no Perl lexing yet, so no INPUT can stall the cursor: a test that
// feeds sources and checks termination proves nothing today and would pass
// with the guard deleted. The test that actually fails when the guard is
// removed installs a scan state that does not advance and asserts the guard
// fires.
func TestForwardProgressGuardFires(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("a scan step that consumed no bytes did not trip the guard: " +
				"invariant 4 is unenforced, and a future non-advancing state will hang the fuzzer")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "forward progress") {
			t.Errorf("guard panicked with %v; the message must name forward progress "+
				"so the next person knows which invariant broke", r)
		}
	}()

	l := &lexer{src: []byte("anything")}
	// A scan that returns without moving the cursor. This is the shape every
	// unterminated-construct bug takes.
	l.step(func(*lexer) {})
}

// TestPositionMonotonicity is spec §7.6.2 invariant 2. Positions run forward,
// never overlap, and never leave the source. An LSP maps every token to a
// range in a buffer; a token whose End precedes its Start, or exceeds the
// file, is a crash in the editor rather than a wrong answer here.
func TestPositionMonotonicity(t *testing.T) {
	for _, src := range []string{
		"",
		" ",
		"\x95",
		"  \t\n  ",
		"\x95\x95",
		" \x95 ",
	} {
		toks := Tokenize([]byte(src))
		prev := 0
		for i, tok := range toks {
			if tok.Start < prev {
				t.Errorf("%q: token %d starts at %d, before the previous token ended at %d",
					src, i, tok.Start, prev)
			}
			if tok.End < tok.Start {
				t.Errorf("%q: token %d has End %d before Start %d", src, i, tok.End, tok.Start)
			}
			if tok.End > len(src) {
				t.Errorf("%q: token %d ends at %d, past the input length %d",
					src, i, tok.End, len(src))
			}
			prev = tok.End
		}
	}
}

// TestLossless is spec §7.6.2 invariant 3, and the note in the spec is worth
// repeating: it catches more real bugs than the other three combined.
//
// It is also weaker than it looks, and the chain review caught that -- this
// property holds for ANY partition of the bytes, including one token spanning
// the whole file. It proves nothing about DELIMITATION. That is what the
// lex.t golden token stream is for; this test only proves no byte was lost or
// duplicated.
func TestLossless(t *testing.T) {
	for _, src := range []string{
		"",
		" ",
		"\x95",
		"  \t\n  ",
		"\x95 \x95",
		"\t\t\n\n  \x95",
	} {
		var sb strings.Builder
		for _, tok := range Tokenize([]byte(src)) {
			sb.WriteString(src[tok.Start:tok.End])
		}
		if sb.String() != src {
			t.Errorf("lossy tokenization:\n got %q\nwant %q", sb.String(), src)
		}
	}
}
