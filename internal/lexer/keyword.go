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
