// ABOUTME: One subscript is one shape, arrow or not: `$h{k}` and `$h->{k}` both Index.
// ABOUTME: A subscript after `->` is not an anonymous constructor, which is what it used to parse as.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSubscriptShapeIsOneShape: §4.14 specifies ONE node for a subscript,
// with the arrow as a flag -- `Index{Base, Idx, Arrow}`. The parser produced
// two unrelated shapes instead:
//
//	$h{k}     [index term call]
//	$h->{k}   [binary term anon_hash call]
//
// The second is wrong twice. It is a different shape for the same operation,
// so every consumer rule has to match both (chapter 6 §6.1.6, "why not (a)":
// the variation tax is observable at 23 kinds, not hypothetical at 200). And
// `anon_hash` is a false claim: `{k}` after an arrow is a subscript, not a
// hash constructor. A consumer reading that tree sees a hash being built.
//
// The same held for `$a->[0]` as `anon_array`.
func TestSubscriptShapeIsOneShape(t *testing.T) {
	for _, pair := range []struct{ plain, arrow string }{
		{"$h{k};", "$h->{k};"},
		{"$a[0];", "$a->[0];"},
		{"$h{a}{b};", "$h->{a}{b};"},
	} {
		plainKinds := strings.Join(kinds(parse.Parse([]byte(pair.plain))), " ")
		arrowKinds := strings.Join(kinds(parse.Parse([]byte(pair.arrow))), " ")
		if plainKinds != arrowKinds {
			t.Errorf("%q and %q must have the same shape\n plain: %s\n arrow: %s",
				pair.plain, pair.arrow, plainKinds, arrowKinds)
		}
	}
}

// TestArrowSubscriptIsNotAConstructor states the second half directly, so a
// regression that made both sides equally wrong would still fail.
func TestArrowSubscriptIsNotAConstructor(t *testing.T) {
	for _, src := range []string{
		"$h->{k};",
		"$a->[0];",
		"$r->{a}[0];",
		"f()->{k};",
		// The chains where the SECOND bracket carries no arrow, which is the
		// case intuit_curly is usually credited with. It is not: a `]` or a
		// `}` that closed a subscript leaves an operator expected, so the
		// brace stack already reads the next `{` as another subscript. The
		// lookahead is only needed where both readings are grammatical, and
		// after a subscript they are not.
		"$a[0]{k};",
		"$x->[0]{k};",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.AnonHash) || containsKind(root, parse.AnonArray) {
			t.Errorf("%q: a subscript after `->` is not a constructor: %v",
				src, kinds(root))
		}
		if !containsKind(root, parse.Index) {
			t.Errorf("%q: must produce an Index: %v", src, kinds(root))
		}
	}
}

// TestConstructorsStillParse is the other side: `->` is not the only thing
// that precedes a brace, and a fix that turned every `{...}` into a subscript
// would break every anonymous hash in the corpus.
func TestConstructorsStillParse(t *testing.T) {
	for _, src := range []string{
		"my $h = {a => 1};",
		"my $a = [1, 2];",
		"f({a => 1});",
		"my $x = {};",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if !containsKind(root, parse.AnonHash) && !containsKind(root, parse.AnonArray) {
			t.Errorf("%q: a constructor must stay a constructor: %v", src, kinds(root))
		}
	}
}

// TestArrowCallStillWorks: `->` before a `(` is a code dereference, and
// before a NAME is a method call. Neither is a subscript, and the fix must
// not capture them.
func TestArrowCallStillWorks(t *testing.T) {
	for _, src := range []string{
		"$code->();",
		"$obj->method;",
		"$obj->method(1);",
		"$obj->$name;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
