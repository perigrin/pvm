// ABOUTME: The keyword table: which builtins leave an operator expected rather than a term.
// ABOUTME: One bit per keyword, because the expect state has exactly two outcomes.

package lexer

// niladic is every perl builtin that takes no argument, so a value has just
// been produced and an OPERATOR comes next.
//
// One bit rather than a taxonomy. The parser needs to know a named unary from
// a list operator -- they bind differently -- but the lexer only needs to
// know which way the expect state moves, and that has two outcomes. A table
// with more categories than outcomes grows a third nothing reads.
//
// Measured, not recalled. Every one of the 251 keywords in perl's
// regen/keywords.pl was compiled as
//
//	perl -e 'use v5.38; no warnings; my $x = KEYWORD / 2 ;'
//
// and exactly these 21 succeeded. Success means perl read the slash as
// division, which it does only where no argument is expected; the other 230
// fail with "Search pattern not terminated" because the slash opened a match.
//
// `use v5.38` matters to the result. Under it `break`, `evalbytes` and `fc`
// take arguments and drop out; without it `say`, `state`, `class`, `method`,
// `field`, `try` and `defer` appear to qualify, but only because an unenabled
// keyword is an ordinary bareword and a bareword followed by a slash is
// division:
//
//	$ perl -MO=Deparse -e 'my $x = zzznotakeyword / 2 ;'
//	my $x = 'zzznot' / 2;
//
// Those are feature-gated, not niladic, and tabling them would break the 5
// corpus files of §0.13 rank 6 that use them as plain identifiers.
var niladic = map[string]bool{
	"time": true, "times": true, "wantarray": true, "wait": true,
	"fork": true, "getppid": true, "getlogin": true,

	// The getent/setent/endent families: all three walk a system database
	// with no argument.
	"getgrent": true, "gethostent": true, "getnetent": true,
	"getprotoent": true, "getpwent": true, "getservent": true,
	"setgrent": true, "setpwent": true,
	"endgrent": true, "endhostent": true, "endnetent": true,
	"endprotoent": true, "endpwent": true, "endservent": true,
}

// isNiladic reports whether a value follows this word rather than an operator.
//
// An unknown word is NOT niladic, and that is the hedge M0 chose rather than
// an oversight. perl's own answer for an undeclared bareword is division --
// it is a string literal -- but perl decides with the symbol table, and three
// different parses come out of the same bytes depending on what a BEGIN block
// declared:
//
//	                          my $x = zzz / 2 ;
//	undeclared                'zzz' / 2          division
//	sub zzz { }               zzz(/ 2 ;/)        a pattern
//	sub zzz () { }            0.5                division, folded
//
// Guessing operator turns every `split /re/` into division and mis-lexes the
// rest of the line. Guessing term costs a division after a bareword that was
// really a value, and the token boundaries stay recoverable. Term is the
// cheaper mistake, which is why it stays the default for everything the table
// does not name.
func isNiladic(word string) bool {
	return niladic[word]
}

// blockTaking is every builtin whose first argument may be a BLOCK rather
// than an expression, so a `{` after it is ambiguous and needs intuitCurly.
//
// Three, not a taxonomy, and measured the same way as the table above -- each
// was compiled with a block first argument and with an expression first
// argument, and only these three accept both:
//
//	$ perl -MO=Deparse -e 'my @b = map { $_ => 1 } @a;'
//	my(@b) = map({$_, 1;} @a);
//	$ perl -MO=Deparse -e 'my @b = map +{ x => $_ }, @a;'
//	my(@b) = map({'x', $_}, @a);
//
// `do` and `eval` also take a brace, but theirs is never a hash -- `do {}` is
// always a block -- so they need no heuristic and are not here. `sort` takes
// a comparator block whose `{` is decided the same way, which is why it is.
var blockTaking = map[string]bool{
	"map": true, "grep": true, "sort": true,
}

// takesBlock reports whether a `{` after this word might open a block.
func takesBlock(word string) bool {
	return blockTaking[word]
}

// phasers are the compile-time and run-time blocks. A `{` after one is always
// a block and never a hash, so unlike map/grep/sort they need no lookahead --
// the word alone settles it, the way `sub NAME` does.
//
// perl treats them as declarations, which is exactly why: toke.c's
// yyl_just_a_word reaches PREBLOCK for a phaser the same way it does for a
// sub name (6636). Without this the brace of `BEGIN {` was read from XTerm as
// an anonymous hash, its `}` left XOperator instead of XState, and a bare
// block after it was read as a hashref too.
//
// `class` and `package` are not here: their block is reached through
// sawPackageWord, which also has a NAME to pass first.
var phaser = map[string]bool{
	"BEGIN": true, "END": true, "CHECK": true, "INIT": true,
	"UNITCHECK": true,
}

// isPhaser reports whether a `{` after this word opens a phaser's block.
func isPhaser(word string) bool {
	return phaser[word]
}
