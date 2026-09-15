// ABOUTME: The fuzz target and its seed corpus: spec §7.6.2's four invariants on ANY input.
// ABOUTME: Invariant 4 is structural rather than fuzzed, because go test -fuzz hangs on a loop instead of failing.

package lexer

import (
	"strings"
	"testing"
)

// fuzzSeeds are the §7.6.2 seeds plus the constructs this milestone added.
// Real Perl mutates into interesting near-Perl; random bytes mostly do not.
var fuzzSeeds = []string{
	"my $x = 42;",
	"print <<'EOF';\nbody\nEOF\n",
	"s{a}{b}ge;",
	"q{nested {braces} here}",
	"$x = $#[0];",
	"my @a = map { $_*2 } @b;",
	"sub f(\\@) {}",
	"<<~EOT;\n  indented\n  EOT\n",
	// Added here: each one is a construct an issue in this milestone landed,
	// so a mutation of it explores a code path that exists.
	"s\\a\\b\\r;",
	"q xfoox",
	"use utf8;\n$Føø::Bær = 1;",
	"$main'a = $'b;",
	"=pod\ntext\n=cut\n",
	"__END__\nnot perl ) ( $$$\n",
	"format STDOUT =\n@<<<<<\n$a\n.\n",
	"$foo =~ s/x/substr(<<EOF, 0, 0)/e;\nbody\nEOF\n",
	"print <<A, <<B;\nfirst\nA\nsecond\nB\n",
	"1 << 3",
	"$x <STDIN>",
}

// FuzzLexer is the target. Run it with
//
//	go test ./internal/lexer/ -run '^$' -fuzz '^FuzzLexer$' -fuzztime 1000000x
//
// or `make fuzz-lexer`. `-run` alone does NOT fuzz -- it executes the seed
// corpus only -- which is why the acceptance criterion names the make target
// rather than a -run invocation. An earlier draft of this issue had a -run
// command claiming 1M execs, and it would have passed without fuzzing at all.
func FuzzLexer(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		checkInvariants(t, src)
	})
}

// TestFuzzSeeds runs the invariants over the seed corpus in the ordinary test
// suite, so they are checked on every `go test` rather than only when someone
// remembers to fuzz. Crashers the fuzzer finds land in testdata/fuzz/ and are
// replayed here too.
func TestFuzzSeeds(t *testing.T) {
	for _, src := range fuzzSeeds {
		checkInvariants(t, src)
	}
	// Inputs that are not Perl at all: an LSP sees half-typed buffers, and
	// the invariants hold on any bytes whatsoever.
	for _, src := range []string{
		"",
		"\x00\x01\x02",
		strings.Repeat("{", 100),
		strings.Repeat("<<", 50),
		"q",
		"s/",
		"<<",
		"=pod",
		"'",
		"\\",
	} {
		checkInvariants(t, src)
	}
}

// checkInvariants is §7.6.2 invariants 1-3. Invariant 4, forward progress, is
// enforced structurally inside the lexer (see step) because `go test -fuzz`
// does not detect an infinite loop -- it hangs, reporting nothing at all.
func checkInvariants(t *testing.T, src string) {
	t.Helper()

	// INVARIANT 1: never panic. A panic fails the test automatically, so the
	// call itself is the assertion.
	toks := Tokenize([]byte(src))

	// INVARIANT 2: positions are monotonic and in bounds.
	prev := 0
	for i, tok := range toks {
		if tok.Start < prev {
			t.Fatalf("%q: token %d starts at %d, before the previous ended at %d",
				src, i, tok.Start, prev)
		}
		if tok.End < tok.Start {
			t.Fatalf("%q: token %d has End %d before Start %d", src, i, tok.End, tok.Start)
		}
		if tok.End > len(src) {
			t.Fatalf("%q: token %d ends at %d, past the input length %d",
				src, i, tok.End, len(src))
		}
		prev = tok.End
	}

	// INVARIANT 3: lossless. Spec §7.6.2 notes this "catches more real bugs
	// than the other three combined", and the corpus survey bore that out --
	// it found a heredoc dropping 141 bytes at EOF that nothing else saw.
	var sb strings.Builder
	for _, tok := range toks {
		sb.WriteString(src[tok.Start:tok.End])
	}
	if sb.String() != src {
		t.Fatalf("lossy tokenization:\n got %q\nwant %q", sb.String(), src)
	}
}
