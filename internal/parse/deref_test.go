// ABOUTME: A sigil applied to an expression is a prefix operator, not a leaf.
// ABOUTME: `${$h->{k}}` used to be ONE token, so everything inside it was outside the tree.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDereferenceHasInterior: the expression inside a dereference is in the
// tree as children.
//
// The lexer used to brace-match `${` to its closer and emit one Variable
// token, so the parser saw a leaf:
//
//	"${$h->{k}};"   [source_file statement term]
//
// Every name, call and constructor inside was outside the tree. Nothing
// failed -- the node round-trips, and a walk over nodes simply found nothing
// there -- which is worse than an Unknown: an Unknown refuses, while silence
// claims there is nothing to see. `internal/parse/sites.go` had to hedge all
// five oracle markers for one of these on exactly that reasoning.
//
// perl reads a dereference as an operator applied to an expression:
//
//	$ perl -MO=Deparse -e 'my %h; my $r = ${$h{k}};'
//	my $r = ${$h{'k'};};
func TestDereferenceHasInterior(t *testing.T) {
	for _, tc := range []struct {
		src   string
		inner parse.Kind
	}{
		// A subscript chain inside the braces.
		{"my $r = ${$h->{k}};", parse.Index},
		// An anonymous array, which perl reports with its own op.
		{"my @a = @{[ 1, 2 ]};", parse.AnonArray},
		// The brace-less forms: a sigil applied to a scalar.
		{"my $v = $$x;", parse.Term},
		{"my @v = @$x;", parse.Term},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", tc.src, kinds(root))
			continue
		}
		deref := firstOfKind(root, parse.Unary)
		if deref == nil {
			t.Errorf("%q: a dereference is a Unary, got %v", tc.src, kinds(root))
			continue
		}
		if len(deref.Children) == 0 {
			t.Errorf("%q: the dereference has no interior -- the whole point "+
				"is that the expression inside is in the tree: %v",
				tc.src, kinds(root))
			continue
		}
		if !containsKind(deref, tc.inner) {
			t.Errorf("%q: the interior must hold a %v: %v",
				tc.src, tc.inner, kinds(deref))
		}
	}

	// The braces GROUP; they do not construct. Reading them as an anonymous
	// hash would make the tree say a hash is being BUILT around the very
	// expression being dereferenced -- a wrong tree that round-trips, the
	// same shape as the `*{EXPR}` glob defect.
	root := parse.Parse([]byte("my $r = ${$h->{k}};"))
	if firstOfKind(root, parse.AnonHash) != nil {
		t.Errorf("the braces of ${EXPR} group, they do not construct: %v",
			kinds(root))
	}
}

// TestBracedNameIsNotADereference: `${name}` is a NAME in braces.
//
// The braces there are punctuation, which is what makes `"${foo}bar"`
// interpolate `$foo` followed by a literal `bar`. Splitting one into a sigil
// and an expression would make `foo` a bareword term the parser has to guess
// at.
//
// The caret control variables are the sharp case. `${^TAINT}` is a NAME --
// the caret belongs to it -- and treating the brace as an expression made
// `^TAINT` a term, caught by TestLexDotTGoldenStream rather than by anything
// aimed at it. Hence this.
func TestBracedNameIsNotADereference(t *testing.T) {
	for _, src := range []string{
		"my $v = ${foo};",
		"my $v = ${ foo };",
		"my $v = ${^TAINT};",
		"my $v = ${^UNICODE};",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Unary) != nil {
			t.Errorf("%q: a braced NAME is one variable, not a dereference "+
				"of an expression: %v", src, kinds(root))
		}
	}
}

// TestPunctuationVariablesStillLex: `$$` is the process id.
//
// The rule that makes `$$x` a dereference is "a sigil, then a sigil, then a
// name". With nothing after the second sigil there is no name, so `$$` stays
// the punctuation variable it has always been. Measured:
//
//	$ perl -e 'print "pid $$\n"'
//	pid 12345
func TestPunctuationVariablesStillLex(t *testing.T) {
	for _, src := range []string{
		"my $pid = $$;",
		"print $$;",
		// The neighbours, which have never been dereferences either.
		"my $e = $@;",
		"my $s = $_;",
		"my $b = $!;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Unary) != nil {
			t.Errorf("%q: a punctuation variable is not a dereference: %v",
				src, kinds(root))
		}
	}

	// And the contrast, which is the whole reason the rule needs a name
	// after the second sigil.
	root := parse.Parse([]byte("my $v = $$x;"))
	if firstOfKind(root, parse.Unary) == nil {
		t.Errorf("`$$x` IS a dereference: %v", kinds(root))
	}
}

// TestDereferenceRoundTrips: every form reproduces its source exactly.
//
// Round-trip proves no byte was lost, never that the parse was right
// (chapter 7 §7.2(c)) -- the four wrong trees of 5b415101 all round-tripped.
// It is still the invariant a token split is most likely to break, because
// one form now covers several tokens where it covered one, and a gap between
// them would be invisible to every other test here.
func TestDereferenceRoundTrips(t *testing.T) {
	for _, src := range []string{
		"my $r = ${$h->{k}};",
		"my @a = @{[ 1, 2 ]};",
		"my $v = $$x;",
		"my @v = @$x;",
		"my $v = ${foo};",
		"my $v = ${^TAINT};",
		"my $pid = $$;",
		"my $x = ${[{a=>214}]}[0];",
		// The spaced form, legal and in the corpus. Measured on perl 5.42.0:
		//
		//	$ perl -e 'my $CX = "x"; $ {$CX} = 17;'
		//	${$CX;} = 17;
		"$ {$CX} = 17;",
		// A glob applied to a scalar, which reaches the parser only now that
		// the deref split exposes it.
		"my $io = *$glob{IO};",
	} {
		root := parse.Parse([]byte(src))
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("round trip lost bytes:\n want %q\n  got %q", src, got)
		}
		if strings.Contains(src, "${") && containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
