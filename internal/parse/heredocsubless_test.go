// ABOUTME: Canon writes a heredoc's body where the source put it, and a sub-less special sub sub-less.
// ABOUTME: Both are spellings the tree records; Faithful forgives neither, and perl 5.44.0 judges both.

package parse_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
)

// faithfulUnforgiven fails t unless canon of src matches src token for
// token, with nothing forgiven.
func faithfulUnforgiven(t *testing.T, src string) {
	t.Helper()
	n := parse.Parse([]byte(src))
	ok, forgiven, why := parse.FaithfulTokens(sigTokens(src), parse.CanonTokens(n, []byte(src)))
	if !ok || len(forgiven) > 0 {
		t.Errorf("%q against its canon %q: %v forgiven by %v (%s), want true with nothing forgiven",
			src, parse.Canon(n, []byte(src)), ok, ruleNames(forgiven), why)
	}
}

// deparsesAlike fails t unless perl compiles canon of each source, and
// Deparse,-p reads it as it reads the source.
func deparsesAlike(t *testing.T, perl perltest.Perl, srcs []string) {
	t.Helper()
	dir := t.TempDir()
	deparse := func(name, text string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := perl.Command("-MO=Deparse,-p", path).Output()
		if err != nil {
			t.Errorf("perl -MO=Deparse,-p %q: %v", text, err)
		}
		return string(out)
	}
	for _, src := range srcs {
		canon := parse.Canon(parse.Parse([]byte(src)), []byte(src))
		if out, err := perl.Command("-c", "-e", canon).CombinedOutput(); err != nil {
			t.Errorf("perl -c %q: %v\n%s", canon, err, out)
		}
		if a, b := deparse("src.pl", src), deparse("canon.pl", canon); a != b {
			t.Errorf("Deparse,-p of %q is\n%s\nbut of its canon %q is\n%s", src, a, canon, b)
		}
	}
}

// heredocInArgList are heredocs in an argument list that runs on past the
// opener's line, reduced from op/debug.t, op/hook/require.t, op/attrs.t and
// op/hash-rt85026.t. perl reads a body from the line after its opener's and
// then resumes the list; measured on 5.44.0, the first is
//
//	fresh_perl_is("use B qw(SVf_IOK);\n", "Done\n", scalar({}), 'debugger');
//
// under Deparse,-p, the same as the one-line `f(<<'CODE', <<'EXPECT', {},
// 'debugger');` with both bodies after it.
var heredocInArgList = []string{
	"fresh_perl_is(<<'CODE', <<'EXPECT',\nuse B qw(SVf_IOK);\nCODE\nDone\nEXPECT\n    {}, 'debugger');\n",
	"fresh_perl_is(\n    'require Apack;',\n    <<'EOF_WANT',\nbefore Apack.pm at - line 1.\nEOF_WANT\n    { },\n    'post action');\n",
	"fresh_perl_is('our $d : Dummy', <<EOS, {},\nInvalid SCALAR attribute.\nEOS\n    \"same name\");\n",
	"is($x, 1, <<EOM);\nok\nEOM\n",
	"f(1, <<\"E\", sub {\nbody\nE\n  2; });\n",
}

// TestCanonKeepsHeredocBodyInArgList: canon writes each body after the
// argument that ended its opener's line in the source, so the file is
// faithful with nothing forgiven, and perl reads canon as it reads the source.
func TestCanonKeepsHeredocBodyInArgList(t *testing.T) {
	for _, src := range heredocInArgList {
		faithfulUnforgiven(t, src)
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), heredocInArgList)
	})
}

// sublessSpecialSubs are the sub-less declarations perl accepts: toke.c
// sends AUTOLOAD, DESTROY and the five phasers to yyl_sub at a statement
// boundary. Measured on 5.44.0, Deparse writes `DESTROY { 1 }`, `AUTOLOAD {
// 1 }`, `BEGIN { 1 }`, `END`, `INIT`, `CHECK` and `UNITCHECK` alike as `sub
// NAME {1;}`. The first is op/hash.t's, op/pos.t's and op/vec.t's.
var sublessSpecialSubs = []string{
	"{ package Class; DESTROY { ++$destroyed; } }\n",
	"DESTROY {\n    $messages .= 'destroyed ';\n}\n",
	"AUTOLOAD { 1; }\n",
	"BEGIN { 1; }\n",
	"END { 1; }\n",
	"INIT { 1; }\n",
	"CHECK { 1; }\n",
	"UNITCHECK { 1; }\n",
}

// TestCanonWritesSublessSpecialSub: canon writes a sub-less special sub
// without `sub`, the file is faithful with nothing forgiven, and the
// parse of canon keeps the spelling, so canon is a fixpoint.
func TestCanonWritesSublessSpecialSub(t *testing.T) {
	for _, src := range sublessSpecialSubs {
		faithfulUnforgiven(t, src)
		canon := parse.Canon(parse.Parse([]byte(src)), []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("canon of %q is %q, and canon of that is %q", src, canon, again)
		}
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), sublessSpecialSubs)
	})
}

// TestFaithfulHeredocBodyChangeIsADifference: a body is one significant
// token, compared like any other, wherever canon writes it.
func TestFaithfulHeredocBodyChangeIsADifference(t *testing.T) {
	rejects(t, "print <<EOT;\nhello\nEOT\n", "print <<EOT;\nhullo\nEOT\n")
	rejects(t, "f(<<EOT,\nhello\nEOT\n 1);\n", "f(<<EOT,\nhullo\nEOT\n 1);\n")
	// Moved past the next argument, it is a difference too.
	rejects(t, "f(<<EOT,\nhello\nEOT\n 1);\n", "f(<<EOT, 1);\nhello\nEOT\n")
}

// TestFaithfulHeredocBodyPunctuationIsInert: a `(` or `;` inside a body is
// part of the body's one token, so it neither pairs with a paren around it
// nor ends a statement, and extra-nesting still pairs the parens it should.
func TestFaithfulHeredocBodyPunctuationIsInert(t *testing.T) {
	for _, src := range []string{
		"f(<<A,\n((;\nA\n 1);\n",
		"print <<A;\n);(\nA\n",
		"ok((($y eq \"x\")), <<A);\n)(;\nA\n",
	} {
		for _, tok := range sigTokens(src) {
			if tok != "(" && tok != ")" && tok != ";" && strings.ContainsAny(tok, "();") && !strings.HasSuffix(tok, "A\n") {
				t.Errorf("%q: token %q holds body punctuation outside the body", src, tok)
			}
		}
		n := parse.Parse([]byte(src))
		if ok, forgiven, why := parse.FaithfulTokens(sigTokens(src), parse.CanonTokens(n, []byte(src))); !ok ||
			slices.ContainsFunc(ruleNames(forgiven), func(r string) bool { return r != "extra-nesting" }) {
			t.Errorf("%q against its canon %q: %v forgiven by %v (%s)", src, parse.Canon(n, []byte(src)), ok, ruleNames(forgiven), why)
		}
	}
	// The body's `)` does not stand in for a missing one.
	if ok, _, _ := parse.FaithfulTokens(sigTokens("f(<<A, 1);\n)\nA\n"), asCanon(sigTokens("f(<<A, 1;\n)\nA\n"))); ok {
		t.Errorf("a body's `)` stood in for the call's")
	}
}

// TestFaithfulTwoHeredocsKeepOrder: two openers on one line read their
// bodies in order; canon keeps it, and swapped bodies are a difference.
func TestFaithfulTwoHeredocsKeepOrder(t *testing.T) {
	for _, src := range []string{
		"f(<<A, <<B);\nfirst\nA\nsecond\nB\n",
		"f(<<A, <<B,\nfirst\nA\nsecond\nB\n  1);\n",
		// One body inside the statement and one after it: the second
		// opener's line is not the first's.
		"f(<<A,\nfirst\nA\n  <<B);\nsecond\nB\n",
	} {
		faithfulUnforgiven(t, src)
	}
	rejects(t, "f(<<A, <<B);\nfirst\nA\nsecond\nB\n", "f(<<B, <<A);\nsecond\nB\nfirst\nA\n")
	rejects(t, "f(<<A, <<B,\nfirst\nA\nsecond\nB\n  1);\n", "f(<<A, <<B,\nsecond\nA\nfirst\nB\n  1);\n")
}

// indentedAndInterpolating follow the plain rule. Measured on 5.44.0,
// `my $x = 1; f(1,\n<<~EOT, 2);\n    a $x\n    EOT\n` is `f(1, "a $x\n", 2)`
// under Deparse,-p.
var indentedAndInterpolating = []string{
	"my $x = 1; f(1,\n<<~EOT, 2);\n    a $x\n    EOT\n",
	"my $x = 1; f(<<\"EOT\",\na $x (\nEOT\n  2);\n",
	"my $x = 1; f(<<~\"EOT\",\n    a ${x};\n    EOT\n  2);\n",
}

// TestFaithfulIndentedAndInterpolatingHeredocs: an indented or interpolating
// body is one token, as a plain one is; the `$x` inside it never leaks out as
// a token of its own, and canon keeps it in place.
func TestFaithfulIndentedAndInterpolatingHeredocs(t *testing.T) {
	for _, src := range indentedAndInterpolating {
		faithfulUnforgiven(t, src)
		if toks := sigTokens(src); slices.Index(toks, "$x") != 1 || slices.Contains(toks[2:], "$x") || slices.Contains(toks, "x") {
			t.Errorf("%q: the body's variable leaks out as a token: %q", src, toks)
		}
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), indentedAndInterpolating)
	})
}

// TestFaithfulSublessFormRejectsOrdinaryNames: perl takes only the special
// names sub-less. Measured on 5.44.0, Deparse leaves `foo { 1 }`, `new { 1 }`
// and `CLONE { 1 }` as an expression statement, `foo {1} ;`, so canon never
// writes them as a declaration.
func TestFaithfulSublessFormRejectsOrdinaryNames(t *testing.T) {
	for _, src := range []string{"foo { 1 };\n", "new { 1 };\n", "CLONE { 1 };\n"} {
		n := parse.Parse([]byte(src))
		if canon := parse.Canon(n, []byte(src)); strings.Contains(canon, "sub") {
			t.Errorf("canon of %q is %q, a declaration", src, canon)
		}
		var decl func(*parse.Node) bool
		decl = func(n *parse.Node) bool {
			return n.Kind == parse.Declaration || slices.ContainsFunc(n.Children, decl)
		}
		if decl(n) {
			t.Errorf("%q parses with a Declaration", src)
		}
		rejects(t, src, "sub "+strings.TrimSuffix(src, " 1 };\n")+" 1; }\n")
	}
	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		out, err := perl.Command("-MO=Deparse", "-e", "foo { 1 }").CombinedOutput()
		if err != nil || strings.Contains(string(out), "sub foo") {
			t.Errorf("perl -MO=Deparse -e 'foo { 1 }': %v\n%s", err, out)
		}
	})
}

// TestFaithfulSublessSpecialSubBodyStillCompared: the sub-less spelling is
// written, not forgiven, so a different body, a different name, or an added
// `sub` is a difference.
func TestFaithfulSublessSpecialSubBodyStillCompared(t *testing.T) {
	rejects(t, "DESTROY { 1; }\n", "DESTROY { 2; }\n")
	rejects(t, "DESTROY { 1; }\n", "AUTOLOAD { 1; }\n")
	rejects(t, "DESTROY { 1; }\n", "sub DESTROY { 1; }\n")
	rejects(t, "sub DESTROY { 1; }\n", "DESTROY { 1; }\n")
}
