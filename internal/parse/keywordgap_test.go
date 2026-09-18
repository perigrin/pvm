// ABOUTME: delete, exists, goto and do FILE: four keywords absent from the table, so four Unknowns.
// ABOUTME: Each is a named unary by parse shape; what it accepts as an argument is a later milestone's.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeleteParses: `delete` takes one argument and binds tightly.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'my %h; my $x = delete $h{a} || 1;'
//	my $x = delete $h{'a'} || 1;
//
// The `||` stays OUTSIDE, so `delete` is a named unary rather than a list
// operator. perl restricts its argument further -- "delete argument is not a
// HASH or ARRAY element or slice" -- but that is an arity rule the parse
// shape does not carry, and §4.8.3 puts argument checking after M1.
func TestDeleteParses(t *testing.T) {
	for _, src := range []string{
		`delete $h{k};`,
		`delete $h->{k};`,
		`delete $a[0];`,
		`my $x = delete $h{k};`,
		`delete @h{qw(a b)};`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestExistsParses: `exists` is membership, and a distinct name from
// `defined`.
//
// Measured: `my %h = (u => undef); exists $h{u}` is true where
// `defined $h{u}` is false. B::SoN gives each its own IR kind for that
// reason (§4.14.2); here the name on the Call carries it.
func TestExistsParses(t *testing.T) {
	for _, src := range []string{
		`exists $h{k};`,
		`exists $h->{k};`,
		`if (exists $h{k}) { 1 }`,
		`my $b = exists $h{k};`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestGotoForms: a label, an expression, or `&NAME`.
//
// §4.14.2 groups `goto` under `LoopEx{Op,Label}` with last/next/redo, which
// already parse. Measured: `perl -MO=Deparse -e 'sub f {} goto &f;'` gives
// `goto \&f;` -- one argument, whatever its shape.
func TestGotoForms(t *testing.T) {
	for _, src := range []string{
		`goto &foo;`,
		`goto LABEL;`,
		`goto $where;`,
		`goto &$code;`,
		`goto HERE if $x;`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestDoFileAndDoBlock: `do EXPR` and `do BLOCK` are different operators
// with one spelling, and the `{` is what separates them -- the same test
// perl makes, and the one `term.go:134` already applies.
func TestDoFileAndDoBlock(t *testing.T) {
	for _, src := range []string{
		`do 'f.pl';`,
		`do $file;`,
		`my $r = do './config.pl';`,
		// The block form must keep its existing path.
		`do { 1 };`,
		`my $r = do { 1 };`,
		`do { 1 } while $x;`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestKeywordGapsConsumeExactlyTheirArguments is the negative scenario that
// matters most: a named unary that swallows the rest of the statement turns
// one gap into a cascade.
//
// A second statement after each keyword must survive as its own statement.
func TestKeywordGapsConsumeExactlyTheirArguments(t *testing.T) {
	for _, src := range []string{
		"delete $h{k};\nmy $after = 1;\n",
		"exists $h{k};\nmy $after = 1;\n",
		"goto &foo;\nmy $after = 1;\n",
		"do 'f.pl';\nmy $after = 1;\n",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		var statements int
		for _, c := range root.Children {
			if c.Kind == parse.Statement {
				statements++
			}
		}
		if statements != 2 {
			t.Errorf("%q: want 2 statements, got %d -- the keyword swallowed "+
				"what followed it: %v", src, statements, kinds(root))
		}
	}
}

// TestKeywordsAreStillIdentifiers is the other negative, and it is the one
// this project has already been caught by once.
//
// `internal/parse/termform_test.go` records that treating every declarator
// as a declaration made `$h{field}` parse its KEY as one. A keyword added
// without the identifier case repeats that: `delete` is a perfectly good
// hash key, and `=>` quotes the word to its left (§4.5.4).
func TestKeywordsAreStillIdentifiers(t *testing.T) {
	for _, src := range []string{
		`my $x = $h{delete};`,
		`my $x = $h{exists};`,
		`my $x = $h{goto};`,
		`my $x = $h{do};`,
		`f(delete => 1);`,
		`my %h = (exists => 1, goto => 2);`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: a keyword used as an identifier must parse: %v",
				src, kinds(root))
		}
	}
}
