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

// TestNoFeatureSignaturesTurnsItOff: `no feature` puts the feature back off,
// so a `($)` after it is a PROTOTYPE again.
//
// The sibling of TestSignatureIsNotAPrototype above, reached from the other
// direction. perl parses the pragma pair and the bare declaration
// identically, because the feature state is the whole test and `no` changes
// it:
//
//	$ perl -MO=Deparse -e 'use v5.36; no feature "signatures"; sub g ($) { return "a" . $_[0] }'
//	sub g ($) { return 'a' . $_[0]; }
//	$ perl -MO=Deparse -e 'sub g ($) { return "a" . $_[0] }'
//	sub g ($) { return 'a' . $_[0]; }
//
// And the prototype's effect on the call is back with it -- measured 5.42.0,
// under either spelling of the `no`:
//
//	perl -e 'use v5.36; no feature "signatures"; sub g ($) { return "g[$_[0]]" } print g 1, 2'
//	g[1]2
//	perl -e 'use v5.36; no feature ":5.36";      sub g ($) { return "g[$_[0]]" } print g 1, 2'
//	g[1]2
//
// BOTH `no feature` spellings are here because perl expands the bundle:
// deparsing `no feature ":5.36"` prints `no feature 'current_sub', ...,
// 'signatures', ...`, so the bundle names signatures explicitly and turning
// it off is not our inference.
//
// A version assertion is NOT one of them. `no v5.36` and `no 5.036` lex to
// the same Word+Quote shape as `use v5.36`, but they are `no VERSION` -- a
// demand that perl be OLDER than that -- and touch no feature at all.
// Measured 5.42.0: `use v5.36; no v5.36;` dies "Perls since v5.36.0 too
// modern--this is v5.42.0", and `no 5.036` deparses to a bare
// `BEGIN { no 5.036; }` beside an untouched `use feature ':5.36'`. So the
// version arms of noteSignatures get no false path, and the two cases below
// pin that: after a version assertion the feature is still ON and `($a, $b)`
// is still a signature.
func TestNoFeatureSignaturesTurnsItOff(t *testing.T) {
	for _, tc := range []struct {
		src   string
		proto string
	}{
		{"use v5.36;\nno feature \"signatures\";\nsub g ($) { return $_[0] }\n", "($)"},
		{"use v5.36;\nno feature \":5.36\";\nsub g ($) { return $_[0] }\n", "($)"},
		// A non-empty prototype whose contents LOOK like a parameter list is
		// the case that matters: `($a, $b)` is a valid prototype spelling
		// too, so only the feature state tells the two apart.
		{"use v5.36;\nno feature \"signatures\";\nsub g ($a, $b) { 1 }\n", "($a, $b)"},
		{"use v5.36;\nno feature \"signatures\";\nsub g ($$) { return $_[0] }\n", "($$)"},
	} {
		toks := Tokenize([]byte(tc.src))
		if !hasToken(toks, []byte(tc.src), Prototype, tc.proto) {
			t.Errorf("%q: no Prototype token %q, so `no feature` did not turn signatures off: %s",
				tc.src, tc.proto, renderKinds(toks, []byte(tc.src)))
		}
	}

	// A version assertion is not a feature pragma. Signatures stay ON.
	//
	// `use feature ":5.36"` is in the same list because it turns them on: the
	// bundle names signatures, and the false path's arrival is what taught
	// this arm to read a bundle at all. Its `:5.34` neighbour is NOT here,
	// because that bundle has no signatures in it -- measured 5.42.0,
	// `use feature ":5.10"` deparses to `use feature 'say', 'state',
	// 'switch'` and nothing else.
	for _, src := range []string{
		"use v5.36;\nno v5.36;\nsub g ($a, $b) { 1 }\n",
		"use v5.36;\nno 5.036;\nsub g ($a, $b) { 1 }\n",
		"use feature \":5.36\";\nsub g ($a, $b) { 1 }\n",
		"use feature \":5.40\";\nsub g ($a, $b) { 1 }\n",
	} {
		toks := Tokenize([]byte(src))
		for _, tok := range toks {
			if tok.Kind == Prototype {
				t.Errorf("%q: %q lexed as a Prototype, but signatures should still be on here",
					src, src[tok.Start:tok.End])
			}
		}
	}

	// And the bundle reader does not fire on things that are not bundles. A
	// feature name is `\w+` and a bundle is `:V.NN`; `:5.34` and `:5.10` are
	// bundles that do not name signatures, and a colon anywhere but the
	// front means the string is neither. `"a:5.36"` is here because a loose
	// colon search made it a bundle, and perl accepts no such name.
	for _, tc := range []struct {
		text string
		want bool
	}{
		{`":5.36"`, true}, {`':5.36'`, true}, {`":5.40"`, true},
		{`":5.34"`, false}, {`":5.10"`, false}, {`":all"`, false},
		{`"signatures"`, false}, {`"Foo::Bar"`, false}, {`"a:5.36"`, false},
		{`"5.36"`, false}, {`":"`, false}, {`""`, false},
	} {
		if got := featureBundleHasSignatures(tc.text); got != tc.want {
			t.Errorf("featureBundleHasSignatures(%s) = %v, want %v", tc.text, got, tc.want)
		}
	}

	// A `no feature` that names some OTHER feature leaves signatures alone.
	for _, src := range []string{
		"use v5.36;\nno feature \"say\";\nsub g ($a, $b) { 1 }\n",
		"use v5.36;\nno feature \":5.10\";\nsub g ($a, $b) { 1 }\n",
	} {
		toks := Tokenize([]byte(src))
		for _, tok := range toks {
			if tok.Kind == Prototype {
				t.Errorf("%q: %q lexed as a Prototype, but `no VERSION` is a version assertion and leaves signatures on",
					src, src[tok.Start:tok.End])
			}
		}
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
