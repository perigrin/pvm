// ABOUTME: The prototype scan: a balanced (...) after `sub NAME` is one opaque token, not variables.
// ABOUTME: Recognition is M1's; what a prototype DOES to a call is M4's.

package lexer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPrototypeIsOpaque: `sub f ($$)` has no variables in it.
//
// Measured before the scan existed:
//
//	"sub f ($$) { 1 }"      Operator(()  Variable($$))   <- swallowed the paren
//	"sub f (\\@\\@) { 1 }"  Operator(\)  Variable(@\)  Variable(@))
//	"sub f ($;$) { 1 }"     Variable($;)  Variable($)
//
// Every one of those is wrong, and two of them are wrong in a way that moves
// the token boundary past the closing paren -- which is how a mis-scanned
// prototype takes the rest of the declaration with it.
//
// perl scans it with scan_str (toke.c:5923), i.e. as a balanced string, and
// spec §5.5.2 says so. The contents are not lexed at all.
func TestPrototypeIsOpaque(t *testing.T) {
	for _, tc := range []struct{ src, proto string }{
		{"sub f ($$) { 1 }", "($$)"},
		{`sub f (\@\@) { 1 }`, `(\@\@)`},
		{"sub f ($;$) { 1 }", "($;$)"},
		{"sub f (&@) { 1 }", "(&@)"},
		{"sub f ($$$;$) { 1 }", "($$$;$)"},
		{"sub f (*) { 1 }", "(*)"},
		{"sub f (_) { 1 }", "(_)"},
		{"sub f () { 1 }", "()"},
	} {
		toks := Tokenize([]byte(tc.src))
		if !hasToken(toks, []byte(tc.src), Prototype, tc.proto) {
			t.Errorf("%q: no Prototype token %q: %s",
				tc.src, tc.proto, renderKinds(toks, []byte(tc.src)))
		}
		// And nothing inside it lexed as a variable, which is the failure
		// that moved the boundary.
		for _, tok := range toks {
			if tok.Kind == Variable {
				t.Errorf("%q: %q lexed as a Variable inside a prototype",
					tc.src, tc.src[tok.Start:tok.End])
			}
		}
	}
}

// TestPrototypeOnlyAfterSubName keeps the scan from eating ordinary parens.
//
// `f ($x)` is a call, not a declaration, and `my ($a, $b)` is a list. Only a
// `(` directly after `sub NAME` is a prototype.
func TestPrototypeOnlyAfterSubName(t *testing.T) {
	for _, src := range []string{
		"f ($x);",
		"my ($a, $b) = @_;",
		"print ($x);",
		"if ($x) { 1 }",
	} {
		toks := Tokenize([]byte(src))
		for _, tok := range toks {
			if tok.Kind == Prototype {
				t.Errorf("%q: %q lexed as a Prototype, but there is no sub declaration",
					src, src[tok.Start:tok.End])
			}
		}
	}
}

// TestSignatureIsNotAPrototype: under the signatures feature the same
// position holds a signature, whose contents ARE lexed.
//
// perl's own test is the feature flag, not the content: toke.c:5923 reads
//
//	if (*s == '(' && !is_sigsub)
//
// Measured:
//
//	$ perl -e 'sub f ($$) { 1 } print prototype(\&f)'          $$
//	$ perl -e 'use v5.36; sub g ($a, $b) {1} print prototype(\&g) // "undef"'
//	undef
//
// So `($a, $b)` under v5.36 is a signature with real variables in it.
func TestSignatureIsNotAPrototype(t *testing.T) {
	src := "use v5.36;\nsub g ($a, $b) { 1 }\n"
	toks := Tokenize([]byte(src))

	for _, tok := range toks {
		if tok.Kind == Prototype {
			t.Errorf("%q: %q lexed as a Prototype, but signatures are enabled",
				src, src[tok.Start:tok.End])
		}
	}
	if !hasToken(toks, []byte(src), Variable, "$a") {
		t.Errorf("a signature's parameters are variables: %s",
			renderKinds(toks, []byte(src)))
	}
}

// TestProtoCorpusLexes: every prototype declaration in comp/proto.t lexes
// without an error token.
//
// 46 declarations in that file alone, by the tighter grep; 11 T2 files
// contain one. TestT2CoreParses cannot reach 100% while they mis-lex.
func TestProtoCorpusLexes(t *testing.T) {
	root, files := corpusFiles(t)
	var target string
	for _, f := range files {
		if f == "comp/proto.t" {
			target = f
			break
		}
	}
	if target == "" {
		t.Skip("comp/proto.t not in the corpus")
	}

	src, err := os.ReadFile(filepath.Join(root, target))
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}
	toks := Tokenize(src)

	var protos, errs int
	for _, tok := range toks {
		switch tok.Kind {
		case Prototype:
			protos++
		case Error, UnknownRest:
			errs++
		}
	}
	if errs != 0 {
		t.Errorf("comp/proto.t: %d error tokens", errs)
	}
	// A count, not `> 0`. `protos > 0` passes with one prototype out of
	// thirty-two, which is a guard that cannot fail -- the defect class this
	// project keeps finding in its own tests.
	//
	// 32 is the lexer's count and it is the trustworthy one. Every grep over
	// this file disagrees with it and with the others:
	//
	//	grep -cE 'sub \w+ \([$@%\\&*;[]]*\)'   46   counts eval'd strings
	//	the same, minus lines with eval or a quote   26   misses multi-line
	//	                                                  and multi-per-line
	//
	// proto.t is a test OF prototypes, so it is full of `eval 'sub f($){}'`
	// -- prototype text inside a string literal, which is not a declaration.
	// The lexer is right to skip those; a line-oriented grep cannot.
	//
	// Pinned exactly so a regression in either direction is visible. Update
	// it when the corpus pin moves, and say why in the commit.
	const wantProtos = 32
	if protos != wantProtos {
		t.Errorf("comp/proto.t: %d Prototype tokens, want %d", protos, wantProtos)
	}
	t.Logf("comp/proto.t: %d prototypes, %d error tokens", protos, errs)
}
