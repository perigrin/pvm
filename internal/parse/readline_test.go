// ABOUTME: `readline $fh` is the same op as `<$fh>`, and the subject reported only the latter.
// ABOUTME: A leaf's angle-delimited TEXT cannot see a builtin call, which arrives as a Call node.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// TestReadlineBuiltinIsASite: the function and the syntax are one op.
//
//	$ perl -MO=Concise,-exec -e 'open my $fh,"<","/dev/null"; my $l = readline $fh;'
//	c  <1> readline[t5] sKS/1
//
// isReadlineTerm (sites.go:299) matches a LEAF whose text is angle-delimited,
// which is the lexer's own decision and correct as far as it goes. It says
// nothing about the builtin, which arrives as a Call node and was reported
// nowhere -- 20 of the 25 WRONG readline sites over the corpus.
func TestReadlineBuiltinIsASite(t *testing.T) {
	for _, src := range []string{
		`my $line = readline $fh;`,
		`readline($yunk);`,
		`readline $f{g};`,
		`readline *{$f{g}};`,
		`readline undef;`,
		`while (readline $fh) { 1; }`,
		`my $x = readline $ary[0];`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindReadline) {
			t.Errorf("%s\n  reports no readline site; perl emits the readline op", src)
		}
	}

	// The angle form must keep working -- it is the case that already did.
	for _, src := range []string{
		`my $line = <$fh>;`,
		`my $line = <FH>;`,
		`my $line = <>;`,
		`my $line = <STDIN>;`,
		`my $line = <My::Handle>;`,

		// A UTF-8 handle name. t/uni/readline.t:64 reads `<hòฟ>` and perl
		// compiles it to readline, measured -- an ASCII-only identifier
		// class calls it a glob and loses the site, which is the defect
		// 0db016fd fixed in the lexer's quote guard.
		"use utf8; my $line = <hòฟ>;",

		// perl's rule here is looser than a package name: a single colon
		// is accepted. Measured, `<a:b>` is readline.
		`my $line = <a:b>;`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindReadline) {
			t.Errorf("%s\n  reports no readline site", src)
		}
	}

	// A GLOB is a different op and must not be claimed. Measured:
	//
	//	$ perl -MO=Concise,-exec -e 'my @f = <*.c>;'
	//	6  <@> glob[t3] lK/1
	//
	// The lexer already separates these, so this asserts the separation
	// survives rather than introducing it.
	for _, src := range []string{
		`my @f = <*.c>;`,
		`my @f = <$h{x}>;`,
		`my @f = <a b>;`,
	} {
		if hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindReadline) {
			t.Errorf("%s\n  reports a readline site; perl emits glob", src)
		}
	}

	// A sub whose NAME merely begins with readline is not the builtin.
	for _, src := range []string{
		`my $x = readline_thing();`,
		`my $x = $obj->readline;`,
	} {
		if hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindReadline) {
			t.Errorf("%s\n  reports a readline site; this is not the builtin", src)
		}
	}

	// Skipping a method NAME must not skip its ARGUMENTS. A first attempt
	// dropped the whole right operand of `->` and lost the reference in
	// `threads->create(\&f, $i)` -- measured, op/threads.t went from clean
	// to three WRONG srefgen sites and re/pat.t gained two.
	for _, src := range []string{
		`threads->create(\&do_sort, $i);`,
		`$obj->readline(\@args);`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindReference) {
			t.Errorf("%s\n  reports no reference site; a method's arguments are still walked", src)
		}
	}
}
