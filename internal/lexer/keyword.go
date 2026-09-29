// ABOUTME: The keyword tables: which builtins leave an operator expected rather than a term,
// ABOUTME: and which leave a term expected with `//` still read as defined-or.

package lexer

import "strings"

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

// unidor is every named unary toke.c lexes with UNIDOR, which leaves
// XTERMORDORDOR rather than XTERM, so a `//` after it is defined-or:
//
//	$ grep -o 'UNIDOR(OP_[A-Z]*)' toke.c
//	UNIDOR(OP_GETC) UNIDOR(OP_POP) UNIDOR(OP_POS) UNIDOR(OP_READLINE)
//	UNIDOR(OP_BACKTICK) UNIDOR(OP_READLINK) UNIDOR(OP_SHIFT)
//	UNIDOR(OP_UNDEF) UNIDOR(OP_UMASK)
//
// OP_BACKTICK is KEY_readpipe.
var unidor = map[string]bool{
	"getc": true, "pop": true, "pos": true, "readline": true,
	"readpipe": true, "readlink": true, "shift": true, "undef": true,
	"umask": true,
}

// isUnidor reports whether `//` after this word is defined-or. The CORE::
// spelling is the same builtin.
func isUnidor(word string) bool {
	return unidor[strings.TrimPrefix(word, "CORE::")]
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
//
// print, printf and say are here for their FILEHANDLE slot, which takes a
// block: `print {$fh} "x"`. Their brace is ambiguous the same way, so it needs
// the same lookahead, and so does the parenthesised spelling -- perl accepts
// that one too, which is what makes it canon's re-readable output. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e 'print({$fh;} "x");'
//	print {$fh;} 'x';
//
// The parens are what canon emits, and without these three that emission did
// not re-lex: the `{` was classified from XTerm, the `;` ended a statement and
// the `}` was orphaned. Perl draws the line at the BUILTINS -- an arbitrary
// word's parenthesised block is a syntax error, `zzz({1;}print "b")` included
// -- so widening this table past them would claim a form perl rejects.
var blockTaking = map[string]bool{
	"map": true, "grep": true, "sort": true,
	"print": true, "printf": true, "say": true,
	"system": true, "exec": true,

	// Builtins only under their features (keyword_any, keyword_all), which
	// the lexer does not track. Listed so `any(` carries the brace
	// lookahead a bare `any {` already gets; for a user sub of that name the
	// same intuition decides, as it does after any word.
	"any": true, "all": true,
}

// takesBlock reports whether a `{` after this word might open a block.
func takesBlock(word string) bool {
	return blockTaking[word]
}

// ANY word may be followed by a block, which is why the brace after one is
// classified by intuitCurly rather than by a table. perl reads `WORD BLOCK`
// three ways and all three put the brace group and what follows in ONE
// statement. Measured on 5.42.0:
//
//	zzz {a};                   zzz { 'a' }       undeclared: indirect method
//	sub zzz {} zzz {a};        zzz({'a'})        declared: call, anon hash arg
//	sub zzz(&) {} zzz { 1 };   &zzz(sub { 1; })  prototyped: call, code ref
//
// Which reading applies is a symbol-table question, so the lexer cannot pick
// one -- but it does not have to. All three agree the brace opens a group
// belonging to this statement, and that is the only thing classification
// decides. What perl NEVER reads is a call SUBSCRIPTED by a hash, and that is
// what a word outside the table used to produce: `zzz { 1 } print "b"` lexed
// as `zzz(){1}` and the `print` after it had no operator before it.
//
// takesBlock stays narrow because it answers a DIFFERENT question -- which
// words have a block-shaped ARGUMENT SLOT the parser fills before the list
// (parseListOpBlock), and whose parenthesised spelling carries the lookahead
// past the `(` (listOpParen). Only map, grep and sort do.

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
//
// `ADJUST` is 5.38's class phaser and belongs with the five: perl reaches
// PREBLOCK for it too, and needs no `;` after the block. Measured on 5.42.0,
// two in a row with nothing between them both run, printing 3:
//
//	class Foo { field $x = 1; ADJUST { $x = 2 } ADJUST { $x++ } method m { $x } }
//
// Its absence was an order-dependent refusal rather than a missing form.
// `ADJUST { ... }` alone read as a bareword call subscripted by an anonymous
// hash, which needs a `;` -- so the block alone parsed, and the declaration
// that followed it did not.
var phaser = map[string]bool{
	"BEGIN": true, "END": true, "CHECK": true, "INIT": true,
	"UNITCHECK": true, "ADJUST": true,
}

// isPhaser reports whether a `{` after this word opens a phaser's block.
func isPhaser(word string) bool {
	return phaser[word]
}
