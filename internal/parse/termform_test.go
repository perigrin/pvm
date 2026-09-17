// ABOUTME: Term forms the T1 sweep found missing: declarations in expressions, anon subs, blocks, globs, filetests.
// ABOUTME: Each was the first cause of failure in dozens of corpus files, measured rather than guessed.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclarationAsTerm: `my` is a named unary at level 19 (§4.6), so it is a
// term as much as a statement form.
//
// `open my $fh, '<', $path` is the idiomatic three-argument open and appears
// in nearly every corpus file that touches a filehandle -- it was the first
// failure in 50 T1 files.
func TestDeclarationAsTerm(t *testing.T) {
	for _, src := range []string{
		"open my $fh, '<', $p;",
		"open my $fh, $p or die;",
		"f(my $x);",
		"push @a, my $y;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestDeclaratorAsIdentifier is the other half, and the regression the first
// draft of the above caused.
//
// `my`, `our`, `state` and `field` are ordinary barewords in most positions.
// Treating every occurrence as a declaration made `$h{field}` parse its KEY
// as one. A declarator takes a target only when a variable or `(` follows.
func TestDeclaratorAsIdentifier(t *testing.T) {
	for _, src := range []string{
		"my $x = $h{field};",
		"my $x = $h{state};",
		"my $x = $h{my};",
		"f(our => 1);",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: a declarator used as an identifier must parse: %v",
				src, kinds(root))
		}
	}
}

// TestAnonSubAsTerm: `sub { ... }` with no name.
func TestAnonSubAsTerm(t *testing.T) {
	for _, src := range []string{
		"my $c = sub { 1 };",
		"f(sub { 1 });",
		"my $c = sub { my $x = shift; $x };",
		// `subtest 'name' => sub { 1 };` is NOT here: `subtest` is an
		// unresolved bareword, which by design consumes no arguments, so the
		// `'name' => sub {...}` after it is orphaned. That is the
		// unresolved-call-with-arguments shape, which needs its own work --
		// 42 T1 files first fail on it.
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestBlockOperators: `eval BLOCK` and `do BLOCK` take a block, not an
// expression, so the named-unary path cannot read them -- it parses the `{`
// as an anonymous hash.
//
// `eval EXPR` and `do EXPR` are different operators with the same spelling,
// and the `{` is what separates them. That is also how perl decides.
func TestBlockOperators(t *testing.T) {
	for _, src := range []string{
		"eval { 1 };",
		"eval { 1 } or die;",
		"my $r = eval { 1 };",
		"do { 1 };",
		// The EXPR forms still work.
		"eval '1';",
		"my $r = eval $code;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestGlobTerm: `*foo` is a glob in term position, not multiplication.
func TestGlobTerm(t *testing.T) {
	for _, src := range []string{
		"my $x = *foo;",
		"my $x = \\*foo;",
		"my $x = *{$name};",
		// And `*` is still multiplication in operator position.
		"my $x = $a * $b;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestFileTestOperators: `-e`, `-d` and the rest are UNIOP (toke.c:6255), so
// they bind like a named unary.
//
// The lexer emits `-` and `e` as separate tokens, and the prefix table would
// otherwise read the `-` as unary minus on a bareword. The check has to come
// BEFORE the prefix branch, which is why `ok(-e $f)` failed while
// `my $b = -e $f` worked -- the second reached a different path.
func TestFileTestOperators(t *testing.T) {
	for _, src := range []string{
		"my $b = -e $f;",
		"ok(-e $f);",
		"if (-d $dir) { 1 }",
		"chdir 't' if -d 't';",
		// And `-` is still subtraction after a term.
		"my $x = $a - $b;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}
