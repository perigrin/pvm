// ABOUTME: The position-dependent scanners: variables, words, numbers, and the ambiguous punctuation.
// ABOUTME: What / < & and x mean depends on the expect state, which is why they live together.

package lexer

import (
	"strings"
	"unicode/utf8"
)

// scanVariable lexes a sigil and its name.
//
// The `$#` forms are the trap this exists for, and t/base/lex.t line 10 is
// the canonical case. Measured with B::Concise:
//
//	$ perl -MO=Concise,-exec -e '$x = $#[0];'
//	3  <#> aelemfast[*#] s
//
//	$ perl -e '@# = (42); print $#[0]'
//	42
//
// So `$#[0]` is element 0 of the array `@#`: the sigil is `$`, the NAME is
// `#`, and `[0]` is an ordinary subscript. It is NOT "$# followed by a
// subscript" -- there is no last-index term in it at all. The spec's gloss
// and this issue's own text both said otherwise, and both were wrong.
//
// The contrast: `$#x` IS last-index-of-@x, one lexeme. Same two opening
// bytes; what follows decides.
func scanVariable(l *lexer) bool {
	start := l.pos
	c := l.src[l.pos]
	if c != '$' && c != '@' && c != '%' {
		return false
	}

	// `%` and `@` are only sigils in term position; in operator position
	// they are modulus and... also `@` is never an operator, but `%` is, and
	// dispatching on position here is what keeps `$a % $b` from lexing as a
	// hash.
	if c == '%' && !l.expect.wantsTerm() && !l.containerTypeSigil() {
		return false
	}

	// `*@` is the glob named `@` -- see globStar. Emitted as the one-byte
	// name, the way `*-` and `*+` arrive, rather than as a sigil that would
	// take the `;` after it as an array's name.
	if c == '@' && l.globStar && l.toks[len(l.toks)-1].End == start {
		l.pos++
		l.emit(Operator, start)
		return true
	}

	l.pos++
	if l.pos >= len(l.src) || l.bareSignatureSigil(start) {
		l.emit(Variable, start)
		return true
	}

	// `$#` is either last-index (`$#name`, `$#{...}`, `$#$ref`, `$#+`) or the
	// variable named `#`. toke.c's test is the byte after it:
	//
	//	if (   s[1] == '#'
	//	    && (   isIDFIRST_lazy_if_safe(s+2, PL_bufend, UTF)
	//	        || memCHRs("{$:+-@", s[2])))
	//
	// `+`, `-` and `@` are one-byte punctuation names -- `$#+` is the last
	// index of @+, measured `2` after `"ab" =~ /(a)(b)/` on 5.42.0 -- and
	// `:` begins a package-qualified one, `$#::x`. Anything else, `[`
	// included, means the name is `#`.
	if c == '$' && l.src[l.pos] == '#' {
		l.pos++
		if l.pos < len(l.src) {
			switch b := l.src[l.pos]; {
			case b == '+' || b == '-' || b == '@':
				l.pos++
			case b == '*' && l.expect == XPostDeref:
				// `$r->$#*`, the postfix last index: the star closes the
				// dereference as it does in `->@*`.
				l.pos++
			case isWordByte(b) || b == '{' || b == '$' || b == ':':
				l.scanVarName()
			}
		}
		l.emit(Variable, start)
		return true
	}

	// A POSTFIX SLICE: `$r->@[0,1]`, `$r->%{'a'}`. The sigil is the whole
	// token and the bracket after it is an ordinary opener, so it reaches
	// the bracket stack.
	//
	// Without this the sigil scanner took `@[` as a punctuation variable
	// named `[`. No opener was ever pushed, and the matching `]` then popped
	// whatever was underneath -- corrupting the stack for the rest of the
	// file, which is why a brace much later in postderef.t was misread.
	//
	// `@*` and `%*` are NOT this case: there the star IS the whole
	// dereference and one Variable token is right. Measured:
	//
	//	$ perl -MO=Deparse -e 'my $r=[1,2,3]; my @s = $r->@[0,1]; my @t = $r->@*;'
	//	my(@s) = @$r[0, 1];
	//	my(@t) = @$r;
	if l.expect == XPostDeref && (l.src[l.pos] == '[' || l.src[l.pos] == '{') {
		l.emit(Variable, start)
		return true
	}

	// A sigil applied to an EXPRESSION is its own token, and the expression
	// after it is lexed normally so the parser can read it.
	//
	// `${EXPR}` and `$$x` are the two shapes. Consuming them here produced
	// one Variable token spanning the whole form, and the parser then made a
	// childless leaf of it -- no interior, so nothing inside was reachable.
	if l.startsDerefExpression() {
		l.emit(DerefSigil, start)
		return true
	}

	// Whitespace and comments may stand between a sigil and its NAME, as
	// they may before a deref's brace: scan_ident runs skipspace first.
	// Measured on 5.42.0, `my $ bits = 1` declares $bits, and a `$` then a
	// comment then `b` on the next line declares $b. The token keeps the bytes between, so the source
	// round-trips. Without a name after them, the space is not reached
	// across -- scanVarName's one-byte punctuation rule stands.
	// A braced NAME after spaces the same way: `$ {^XY}` is `${^XY}` and
	// `$ { foo }` is `$foo`, measured on 5.42.0. startsDerefExpression has
	// already declined it as an expression.
	if brace := skipBlanks(l.src, l.pos); brace > l.pos && brace < len(l.src) &&
		l.src[brace] == '{' && l.bracedNameFollowsAt(brace) {
		l.pos = brace
		l.scanVarName()
		l.emit(Variable, start)
		return true
	}

	if name := l.nameAfterSpace(l.pos); name > l.pos {
		l.pos = name
		l.scanIdentRunes()
		l.emit(Variable, start)
		return true
	}

	l.scanVarName()
	l.emit(Variable, start)
	return true
}

// skipBlanks returns the index of the first byte at or after i that is not
// a space or a tab.
func skipBlanks(src []byte, i int) int {
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	return i
}

// nameAfterSpace returns where an identifier starts after the whitespace and
// comments from i, or i when no whitespace or comment is there or no name
// follows them.
func (l *lexer) nameAfterSpace(i int) int {
	j := i
	for j < len(l.src) {
		switch c := l.src[j]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			j++
		case c == '#':
			for j < len(l.src) && l.src[j] != '\n' {
				j++
			}
		default:
			if j == i {
				return i
			}
			r, _ := utf8.DecodeRune(l.src[j:])
			if identStart(r, l.utf8Pragma) {
				return j
			}
			return i
		}
	}
	return i
}

// startsDerefExpression reports whether the sigil at l.pos-1 is applied to an
// expression rather than to a name. l.pos is just past the sigil.
//
// Two forms qualify, and each has a near neighbour that does NOT:
//
//	${ $h->{k} }   expression      vs  ${name}  a NAME in braces
//	$$x            deref of $x     vs  $$       the process id
//
// The brace case is decided by what follows it: an identifier that runs to
// the closing brace is a name, anything else is an expression. The sigil case
// is decided by what follows the SECOND sigil: a name makes it a dereference,
// nothing makes it the punctuation variable `$$`.
func (l *lexer) startsDerefExpression() bool {
	if l.pos >= len(l.src) {
		return false
	}

	// Whitespace may stand between the sigil and the brace. `$ {*$glob}{Keys}`
	// is a real form and appears in the corpus; checking the byte immediately
	// after the sigil missed it, and the sigil then fell back to the
	// swallowing path it was meant to replace. Measured on perl 5.42.0:
	//
	//	$ perl -e 'use Symbol qw(gensym); my $g = gensym();
	//	           $ {*$g}{K} = 5; print $ {*$g}{K};'
	//	5
	brace := l.pos
	for brace < len(l.src) && (l.src[brace] == ' ' || l.src[brace] == '\t') {
		brace++
	}
	if brace < len(l.src) && l.src[brace] == '{' {
		return !l.bracedNameFollowsAt(brace)
	}

	// `$$x` and `@$x` are dereferences; a bare `$$` is the pid, and
	// `$$ref[0]` is still a dereference of `$ref`. A second sigil with a
	// NAME after it is the whole rule -- with nothing after it there is
	// nothing to dereference.
	// Spaces may stand before the second sigil too: measured on 5.42.0,
	// `$ $name1` is `$$name1` and `@ $r` is `@$r`.
	if brace >= len(l.src) || l.src[brace] != '$' {
		return false
	}
	next := brace + 1
	return next < len(l.src) &&
		(isWordByte(l.src[next]) || l.src[next] == '{' || l.src[next] == '$')
}

// bracedNameFollowsAt reports whether `{` at brace opens a plain NAME, as in
// `${foo}` -- an identifier, optionally spaced, then the closing brace.
//
// Anything else is an expression: `${$x}`, `${ $h->{k} }`, `@{[ 1, 2 ]}`.
func (l *lexer) bracedNameFollowsAt(brace int) bool {
	i := brace + 1
	for i < len(l.src) && (l.src[i] == ' ' || l.src[i] == '\t') {
		i++
	}

	// `${^TEST}`, `${^TAINT}`, `${^UNICODE}` -- the caret control variables
	// are NAMES, and the caret is part of the name rather than an operator.
	// Splitting one into a sigil and an expression made `^TEST` a term the
	// parser had to guess at, which is what TestLexDotTGoldenStream caught.
	if i < len(l.src) && l.src[i] == '^' {
		i++
	}

	nameStart := i
	for i < len(l.src) && isWordByte(l.src[i]) {
		i++
	}
	if i == nameStart {
		return false
	}
	for i < len(l.src) && (l.src[i] == ' ' || l.src[i] == '\t') {
		i++
	}
	return i < len(l.src) && l.src[i] == '}'
}

// scanVarName consumes an identifier, a punctuation variable, or a braced
// name after a sigil.
func (l *lexer) scanVarName() {
	if l.pos >= len(l.src) {
		return
	}
	switch c := l.src[l.pos]; {
	case c == '\'':
		// `$'` is the postmatch variable: the apostrophe is the whole name
		// -- unless an identifier start follows, when it is the old package
		// separator naming main. toke.c's parse_ident takes it at the first
		// position as at any other; measured, `$main::b = 5; print $'b`
		// prints 5. So comp/package.t:17's `$'b` is `$main::b`.
		//
		// ponytail: ungated, as the interior separator in scanIdentRunes is.
		// perl gates both on feature apostrophe_as_package_separator, off
		// from the 5.41 bundle; there `$'b` is `$'` and a bareword, a syntax
		// error in every spelling but `$'x3`-style repeats. Gate both when
		// the lexer tracks feature bundles: 01a0ec10-0f98-7801-afbf-5c02bdc9c3c9.
		l.pos++
		if l.pos < len(l.src) {
			r, _ := utf8.DecodeRune(l.src[l.pos:])
			if identStart(r, l.utf8Pragma) {
				l.scanIdentRunes()
			}
		}
	case c == ':' && l.leadingPackageSeparator():
		// `$::x` is `$main::x`. scanIdentRunes requires a leading identifier
		// byte and `:` is not one, so this lexed as `Variable("$:")` -- a
		// punctuation variable named `:` -- followed by stray tokens. A
		// plausible wrong answer rather than an error, which is why nothing
		// caught it.
		//
		// Measured: `our $x = 5; print $::x` prints 5.
		//
		// `$:` ALONE is a real punctuation variable, the format line-break
		// set, so the two colons and a name after them are all required --
		// see leadingPackageSeparator.
		l.pos += 2
		l.scanIdentRunes()
	case c >= '0' && c <= '9':
		// A digit name is the whole run of digits: scan_ident's parse_ident
		// runs with STOP_AT_FIRST_NON_DIGIT. Measured on 5.42.0, `@119797`
		// is one array and `$10` the tenth capture.
		for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
			l.pos++
		}
	case l.scanIdentRunes():
		// Consumed by the identifier scanner, which handles both package
		// separators and the utf8-widened class.
	case c == '{':
		// A braced NAME: `${name}`, `${^TAINT}`. Reached only when
		// startsDerefExpression has already declined, so the braces here are
		// punctuation around an identifier and the match can be shallow.
		depth := 0
		for l.pos < len(l.src) {
			if l.src[l.pos] == '{' {
				depth++
			} else if l.src[l.pos] == '}' {
				depth--
				l.pos++
				if depth == 0 {
					return
				}
				continue
			}
			l.pos++
		}
	case c == '^' && l.caretNameFollows():
		// A caret control variable: `$^O`, `$^W`, `$^_`. The caret and the
		// character after it are ONE name, which is the same rule
		// bracedNameFollowsAt already applies to `${^TAINT}` -- the braced
		// spelling was repaired when TestLexDotTGoldenStream caught
		// `${^TEST}` splitting, and this bare one was left behind.
		//
		// Without this the punctuation-variable case below took one byte,
		// so `$^O` lexed as `Variable($^)` and a separate `Word(O)`. Most
		// spellings still yielded a parseable tree -- `my $x = $^O;`
		// canonicalised to `$^;O()` with Unknown=0 -- so no node count
		// reported it.
		//
		// Measured: `perl -MO=Deparse -e 'my $x = $^O;'` -> `my $x = $^O;`
		l.pos += 2
	case c == '$':
		// `$$` followed by a name is a dereference -- `$$ref`, `@$ref` --
		// and otherwise it is the process id, a whole name on its own.
		// toke.c's scan_ident takes the dereference only when an identifier
		// start, a digit, another `$`, a `{` or `::` follows:
		//
		//	if (*s == '$' && s[1]
		//	    && (   isIDFIRST_lazy_if_safe(s+1, PL_bufend, is_utf8)
		//	        || isDIGIT_A((U8)s[1]) || s[1] == '$' || s[1] == '{'
		//	        || memBEGINs(s+1, ..., "::")) )
		//
		// Recursing unconditionally took the NEXT byte as a punctuation name,
		// so `$$,`, `$$;`, `$$)` and even `$$ ` each lexed as one variable
		// that swallowed a comma, a terminator, a closer or a space.
		// Measured, `my @a = ($$, 1)` has two elements.
		l.pos++
		if l.pos < len(l.src) {
			n := l.src[l.pos]
			if isWordByte(n) || n == '$' || n == '{' ||
				n == ':' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ':' {
				l.scanVarName()
			}
		}
	default:
		// A punctuation variable: $_, $0, $!, $@, $/ and the rest. One byte.
		l.pos++
	}
}

// caretNameFollows reports whether the `^` at the cursor begins a caret
// control variable rather than standing as the punctuation variable `$^`.
//
// A character must follow, and it must be one perl folds into the name.
// Measured on perl 5.42.0 by compiling `my $x = $^C;` for every C:
//
//	$^A .. $^Z   compile     every uppercase letter
//	$^_  $^^     compile
//	$^[          compiles    the caret form of a control character
//	$^a .. $^z   Bareword found where operator expected
//	$^0 .. $^9   Number found where operator expected
//
// So a LOWERCASE letter is not part of the name -- perl reads `$^` and then a
// bareword. Digits behave the same. The rule is therefore the uppercase
// range plus `_`, `^` and `[`, not "any identifier byte": widening it to the
// identifier class would swallow the word after a bare `$^`.
//
// `$^` ALONE is real -- the format top-of-page name -- which is why the
// character after the caret is required. This is the same shape as
// leadingPackageSeparator below, where `$:` alone is also a real variable.
func (l *lexer) caretNameFollows() bool {
	if l.pos+1 >= len(l.src) {
		return false
	}
	switch c := l.src[l.pos+1]; {
	case c >= 'A' && c <= 'Z', c == '_', c == '^', c == '[':
		return true
	}
	return false
}

// leadingPackageSeparator reports whether `::` at the cursor introduces a
// package-qualified name, as the `$::x` shorthand for `$main::x`.
//
// Both colons AND a name are required. `$:` is the format line-break
// variable and `$::` with nothing after it is not a name perl accepts, so
// taking either would trade one wrong answer for another.
func (l *lexer) leadingPackageSeparator() bool {
	if l.pos+1 >= len(l.src) || l.src[l.pos+1] != ':' {
		return false
	}
	// TWO COLONS ARE ENOUGH, and nothing after them needs checking.
	//
	// This used to require a word byte at pos+2, which admitted `$::x` and
	// refused the BARE main stash -- `$::{n}`, `%::`, `@::`. There the package
	// name is empty and the name ends AT the second colon, so there is no word
	// byte to find, and requiring one let `$:` (a real punctuation variable,
	// the format line-break set) take the first colon and stranded the second
	// as an Operator: `$::{n}` lexed as `Variable($:) Operator(:) ...`.
	//
	// Measured 5.42.0 -- all three are the main stash:
	//
	//	perl -MO=Deparse -e 'our $n = 5; my $r = $::{n};'
	//	  ->  my $r = $main::{'n'};
	//	perl -MO=Deparse -e 'my @k = keys %::;'
	//	  ->  my(@k) = keys %main::;
	//	perl -e 'print exists $::{STDOUT} ? "yes" : "no"'   ->  yes
	//
	// `$main::{n}` and `$Pkg::{n}` were already right, so only the empty
	// package name was affected -- one spelling of a rule repaired and its
	// sibling left, the same shape as `${^TAINT}` being right while `$^O` was
	// wrong.
	//
	// A first fix listed the openers `{ [ :` and still left `keys %::;`
	// refusing, because a `;` ends the name as surely as a subscript does.
	// Inverting that to "anything but a word byte" then made both branches
	// unconditional, which is the measurement's real answer: a scalar named
	// `$:` cannot be followed by a second colon and still be `$:`, so the
	// second colon alone settles it.
	//
	// `$:` ALONE still stays one token, because this guard is never reached
	// without that second colon.
	return true
}

// scanWord lexes a bareword or keyword.
//
// Whether a word is a keyword, a function name or a bareword string cannot be
// decided here -- it needs the symbol table and the parser -- so the lexer
// reports Word and leaves the classification alone.
func scanWord(l *lexer) bool {
	start := l.pos
	// The repeat operator touching its count: `"ab"x4` is `x` then `4`.
	// toke.c splits it before keyword lookup, in operator position only:
	//
	//	case 'x':
	//	    if (isDIGIT(s[1]) && PL_expect == XOPERATOR) {
	//	        s++;
	//	        Mop(OP_REPEAT);
	//	    }
	//
	// Before keyword lookup, so no declaration changes it: measured on
	// 5.42.0, `sub x4 {9} my $x = "ab" x4;` still repeats. The count is left
	// for scanNumber.
	if l.expect == XOperator && l.src[start] == 'x' &&
		start+1 < len(l.src) && l.src[start+1] >= '0' && l.src[start+1] <= '9' {
		l.pos++
		l.emit(Word, start)
		return true
	}
	if !l.scanIdentRunes() {
		return false
	}
	l.keywordBeforeApostrophe(start)
	l.takeRepeatAssign(start)
	// `use utf8` widens the class for everything after it, so the pragma has
	// to be noticed as it is lexed rather than in a prepass.
	l.notePragma(start)
	l.emit(Word, start)
	return true
}

// keywordBeforeApostrophe ends a word at an apostrophe that follows a
// keyword: the apostrophe opens a string there. toke.c's yyl_keylookup scans
// the first word WITHOUT package parts and looks it up; only a non-keyword
// goes on to yyl_just_a_word, where `'` joins a package name. So `print'x'`
// prints and `eval'f()'` evals, while `foo'bar` is `foo::bar`. A `::` comes
// first in toke.c too, so a word already qualified with one is left alone.
func (l *lexer) keywordBeforeApostrophe(start int) {
	word := string(l.src[start:l.pos])
	i := strings.IndexByte(word, '\'')
	if i <= 0 || strings.Contains(word[:i], "::") || !perlKeywords[word[:i]] {
		return
	}
	l.pos = start + i
}

// takeRepeatAssign extends a just-scanned `x` into the `x=` operator.
//
// `x=` is the thirteenth compound assignment and the ONLY one perl spells
// with a letter; the other twelve are punctuation, so the operator scanner
// that forms `+=` and `.=` from punctuation runs never reaches this one.
// Measured before this existed, `$t x= 2` lexed as `Word(x) Operator(=)` --
// a Word where an operator belongs, which the parser reported as
// trailing_tokens.
//
// The kind stays WORD, which the caller's emit supplies. That is the kind
// this lexer gives every word-shaped operator (`x`, `cmp`, `and`), and the
// glossary's `word-shaped operator` category is Word plus
// `parse.IsWordShapedOperator` -- which already answers yes for `x=`, reading
// it out of the precedence table at level 9. A Word also leaves a TERM
// expected, which is what an operator wants next.
//
// OPERATOR POSITION ONLY, and the fat comma is why. Measured 5.42.0:
//
//	$ perl -MO=Deparse -e 'my %h = (x=>1);'          my(%h) = ('x', 1);
//	$ perl -MO=Deparse -e 'my $t; my @a=($t x=> 2);' syntax error near "$t x"
//
// In term position `x=>1` is the autoquoted bareword `x` and a `=>`; in
// operator position perl has already formed `x=` and the `>` is then junk,
// which is what the error reports. So the expect state decides, exactly as
// it decides `/` and `<`, and no `=>` exclusion is needed on top of it.
//
// ONE `=`, never two. Measured, `x==` is a syntax error and `x=~ 3` applies
// `x=` to `~3`:
//
//	$ perl -e 'my $s="ab"; $s x=~ 3;'  ->  $s x= 18446744073709551612
//
// Both readings are perl having formed `x=` and lexed the rest separately,
// so the scanner stops after the first `=`.
//
// EXACTLY the word `x`. `$t xx= 2` is the identifier `xx`, and `$tx= 2` is
// the variable `$tx` -- scanVariable claims that one before this is reached.
func (l *lexer) takeRepeatAssign(start int) {
	if l.expect.wantsTerm() {
		return
	}
	if l.pos != start+1 || l.src[start] != 'x' {
		return
	}
	if l.pos >= len(l.src) || l.src[l.pos] != '=' {
		return
	}
	l.pos++
}

// scanVString lexes the `v`-prefixed spelling of a v-string: `v65.66.67`.
//
// It runs BEFORE scanWord, because `v65` is a legal identifier and the word
// scanner claimed it -- leaving `Word(v65) Operator(.) Number(66.67)`, a
// concatenation of a bareword with a number where perl has one string. The
// parser cannot see that: it receives a valid expression and returns no
// Unknown node.
//
// ONE DOT is enough when the `v` is there, and that asymmetry is the whole
// rule. Measured 5.42.0, the same single dot lands in two categories and
// the prefix is the only difference:
//
//	$ perl -e 'my $v = v5.36; print length($v)'   2
//	$ perl -e 'print 5.36'                        5.36
//
// Two characters, so `v5.36` is a string of ordinals; bare `5.36` is the
// float it looks like, which is why `scanNumber` still needs two dots and
// this needs one.
//
// A dot is STILL required. Perl reads bare `v5` as a v-string too --
// measured, `length(v5)` is 1 -- but `v5` with no dot is also an ordinary
// identifier here, and claiming it would take every `v`-plus-digits
// bareword with it. That is a third gap, not this one.
//
// LOWERCASE ONLY, and an earlier revision accepted `V` as well. Measured
// on 5.42.0, capital V is never a v-string:
//
//	$ perl -e 'my $v = V5.36; print "[$v]"'
//	[V536]
//	$ perl -Mstrict -e 'my $v = V5.36; print $v'
//	Bareword "V5" not allowed while "strict subs" in use
//	$ perl -e 'use V5.36; print "ok"'
//	Can't locate V5.pm in @INC
//
// `V5.36` is the bareword `V5` concatenated with `.36`, and `use V5.36`
// LOADS A MODULE. Accepting it here made `noteSignatures` read a module
// load as a feature-gating version and turn signatures on -- the silent
// change of meaning this scanner exists to get right.
func scanVString(l *lexer) bool {
	if l.src[l.pos] != 'v' {
		return false
	}
	if !l.expect.wantsTerm() {
		return false
	}
	start := l.pos
	l.pos++
	if l.pos >= len(l.src) || !isDigit(l.src[l.pos]) {
		l.pos = start
		return false
	}
	dots := l.scanNumberRun()
	// A dot with no digit after it is not the v-string's: `v10.v257` is
	// `v10 . v257`, measured on 5.42.0 to deparse as "\n\x{101}".
	if l.src[l.pos-1] == '.' {
		l.pos--
		dots--
	}
	if dots < 1 {
		l.pos = start
		return false
	}
	l.emit(Quote, start)
	return true
}

// isDigit reports whether c is an ASCII decimal digit.
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// startsLeadingDecimal reports whether the cursor sits on a `.` that
// begins a numeric literal written with no digit before the point.
//
// Two conditions, and POSITION is the load-bearing one. A digit must
// follow, so `.` alone and `..` are untouched -- but a digit following is
// not sufficient, because `$a .5` is concatenation. Measured 5.42.0, the
// same four bytes parse two ways with nothing but context between them:
//
//	my $r = $a .5;     ->  my $r = $a . '5';    concatenation
//	print $a .5        ->  print $a 0.5;       filehandle and a number
//
// The second prints nothing, because `$a` lands in print's filehandle
// slot. So the rule is the expect state, the same mechanism §3.2 uses for
// a leading `%`, `<`, `&` and `/`: where a TERM is expected nothing else
// can start with a dot, and where an OPERATOR is expected a dot is
// concatenation regardless of what follows.
// ONE version-string exception survives, measured rather than defensive
// and found by a test rather than anticipated. It was TWO until
// `scanVString` began claiming a one-dot `v5.36` whole; the paragraphs
// below record which one went and why, because a count in a header that
// disagrees with the code beneath it is how a comment starts lying.
//
// `use v5.36` USED TO reach here with a term expected and a digit
// following, because `v5` lexed as an ordinary Word -- so without the
// first guard the `.36` became one Number and `pendingVersionMajor` never
// saw the Operator it reassembles the version from. The feature bundle
// then did not turn on, which silently changes what the rest of the file
// means: `use v5.36` enables signatures, and `sub g ($a, $b)` is a
// PROTOTYPE without it.
//
// It no longer reaches here. `scanVString` claims ONE dot as of 01a0dc26,
// so `v5.36` is one Quote before the number scanner sees a dot at all,
// and the guard below has no caller left that needs it: measured, both
// `internal/lexer` and `internal/parse` stay green without it. It is kept
// for one commit rather than deleted in the same change that stranded it,
// the way `continuesVersionString`'s Number branch was.
//
// `require(v5.5.630)` needs the second. `pendingVersionMajor` clears on
// the Number it pairs with, so by the third part it is already zero and
// the first guard no longer applies. What rules it out is that the
// previous token is a Number with no space before the dot: a v-string's
// parts are adjacent, and a leading decimal never follows a number
// directly. `(1,.5)` is a real leading decimal and its previous token is
// the comma, not the 1.
func startsLeadingDecimal(l *lexer) bool {
	return l.src[l.pos] == '.' &&
		l.expect.wantsTerm() &&
		!continuesVersionString(l) &&
		l.pos+1 < len(l.src) &&
		isDigit(l.src[l.pos+1])
}

// continuesVersionString reports whether the dot at the cursor separates
// two parts of a v-string such as `v5.36` or `v5.5.630`.
//
// ADJACENCY is the whole rule: a v-string's parts touch, so the token
// before the dot ends exactly where the dot begins. `(1,.5)` is a real
// leading decimal and the token before its dot is the comma; `5 . .5`
// has a space. Neither is adjacent to a digit-bearing token.
//
// The predecessor is a WORD, and only a word. An earlier revision also
// accepted a `Number`, because the two dots of `v5.5.630` had different
// predecessors when that run arrived as three tokens. `scanVString` now
// claims a two-dot run whole, so nothing reaches a second dot here:
// measured, `foo.5.6` lexes as `Word(foo) Quote(.5.6)` and
// `v5.5.630` as one `Quote`.
//
// The Number branch was deleted rather than left with a comment
// promising a test, which is what it had become -- its only witness was
// the `v5.5.630` row in `scanrows_test.go`, and that row now exercises
// `scanVString` instead. Deleting it changes no suite: `internal/lexer`
// and `internal/parse` are green without it.
//
// Checking `pendingVersionMajor` instead would still be wrong, for the
// original reason: it would miss every v-string outside a `use`, because
// `noteSignatures` returns early unless a `use` is pending, so
// `require(v5.5.630)` never sets it.
func continuesVersionString(l *lexer) bool {
	if len(l.toks) == 0 {
		return false
	}
	prev := l.toks[len(l.toks)-1]
	if prev.End != l.pos {
		return false
	}
	if prev.Kind != Word {
		return false
	}
	_, ok := versionPrefix(string(l.src[prev.Start:prev.End]))
	return ok
}

// scanNumber lexes an integer or float literal, or the v-string that a
// second decimal point turns the same digits into. The exotic forms --
// `0x_1234`, `0x0p0` -- are §0.13 rank 9 and belong to a later issue; this
// covers what the M0 corpus reaches.
//
// DOT COUNTING is what separates the two kinds, and it cannot be decided
// before the run has been scanned: `65.66` is a float and `65.66.67` is the
// three-character string `ABC`, and they are the same bytes up to the second
// point. Measured 5.42.0:
//
//	$ perl -e 'my $n = 65.66; print $n'          65.66
//	$ perl -e 'my $v = 65.66.67; print $v'       ABC
//	$ perl -e 'my $x = 5.42.0; print $x+0'       0
//
// The last is the category claim: perl holds a string of ordinals, so `+0`
// is 0 rather than 5.42. GLOSSARY.md records the decision under "numeric
// literal" -- a v-string is not in that category.
func scanNumber(l *lexer) bool {
	c := l.src[l.pos]
	if !isDigit(c) && !startsLeadingDecimal(l) {
		return false
	}
	start := l.pos
	dots := l.scanNumberRun()
	// One dot is a float; two make a v-string, which this lexer spells as a
	// Quote because that is its kind for a string. `internal/conformance/
	// categories.go` reads a Quote with no operator name as the glossary's
	// "string literal", and `v` is not a quote-operator name.
	if dots >= 2 {
		l.emit(Quote, start)
		return true
	}
	l.emit(Number, start)
	return true
}

// scanNumberRun advances the cursor over one numeric run and reports how
// many decimal points it contained.
//
// Shared with scanVString, which differs only in the `v` it consumes first.
func (l *lexer) scanNumberRun() int {
	dots := 0
	// A RADIX PREFIX rules the exponent sign out, because `e` is a hex
	// digit there rather than an exponent marker. Measured 5.42.0:
	// `print 0x1e-1` is 29, so that `-` is subtraction.
	hex := l.pos+1 < len(l.src) && l.src[l.pos] == '0' &&
		(l.src[l.pos+1] == 'x' || l.src[l.pos+1] == 'X')
	// Octal too, spelled `0o17` or `017`: `01.1p0` is an octal float.
	radix := hex || l.pos+1 < len(l.src) && l.src[l.pos] == '0' &&
		(byteIn(l.src[l.pos+1], "bBoO") || isDigit(l.src[l.pos+1]))
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '.':
			// A DOUBLE dot is the range operator, not part of the literal.
			// `1..5` is three tokens; consuming the dots gave one
			// Number("1..5") and the range was gone before the parser saw
			// it.
			//
			// `$a..$b` was never affected, so only literal endpoints broke
			// -- and the cost is measured: `my @r = (1..5)` infers @r as
			// Str, because `internal/infer/infer.go:645-659` matches the
			// anonymous `.` before ever seeing `..` (§4.14.2).
			//
			// One dot still belongs to the number: `1.5` is a float, and
			// perl takes `1_000.5` and `1.5e10` whole.
			if l.pos+1 < len(l.src) && l.src[l.pos+1] == '.' {
				return dots
			}
			dots++
		case isWordByte(c):
			// The SIGN OF AN EXPONENT is part of the literal, and this is
			// the only place a sign ever is. Measured 5.42.0:
			//
			//	$ perl -e 'print 5e-1'   0.5
			//	$ perl -e 'print 5e'     Bareword found where operator
			//	                         expected (Missing operator
			//	                         before "e"?)
			//
			// So `Number("5e")` is a token perl would reject, which is
			// what makes the old split wrong rather than merely different.
			// The sign must be ADJACENT to the `e`: `5-1` is arithmetic,
			// and so is the second `-` of `5e-1-1`.
			if !hex && (c == 'e' || c == 'E') &&
				l.pos+2 < len(l.src) &&
				(l.src[l.pos+1] == '-' || l.src[l.pos+1] == '+') &&
				isDigit(l.src[l.pos+2]) {
				l.pos += 2
				continue
			}
			// A DECIMAL takes no other letter: its `e` only before a sign,
			// a digit or `_` (toke.c:13142-13144), so `5x3` is `5 x 3` and
			// `1if $b` is `1 if $b`, measured on 5.42.0. A radix literal's
			// letters are its digits and prefix, and stay.
			if !radix && !isDigit(c) && c != '_' &&
				!((c == 'e' || c == 'E') && l.pos+1 < len(l.src) &&
					byteIn(l.src[l.pos+1], "+-0123456789_")) {
				return dots
			}
		default:
			return dots
		}
		l.pos++
	}
	return dots
}

// scanAngle lexes `<FH>` in term position, and reports false in operator
// position so the caller emits a comparison operator instead.
//
// Spec §3.2: the same `<` is a readline where a value is expected and a
// comparison where one just ended.
func scanAngle(l *lexer) bool {
	if l.src[l.pos] != '<' || !l.expect.wantsTerm() {
		return false
	}
	// `<<>>` is the 5.22 double diamond, which reads every argument as a
	// filename and never as a command. Checked first because its inner `<>`
	// would otherwise close the scan at the second `<`.
	if l.pos+3 < len(l.src) && l.src[l.pos+1] == '<' &&
		l.src[l.pos+2] == '>' && l.src[l.pos+3] == '>' {
		start := l.pos
		l.pos += 4
		l.emit(Readline, start)
		return true
	}

	// A readline's contents are a filehandle name, a scalar, or nothing --
	// and a GLOB's are a shell pattern, which is neither. Measured:
	//
	//	$ perl -MO=Deparse -e 'my @g = <*.c>; my @h = <~/x>;'
	//	my(@g) = glob('*.c');
	//	my(@h) = glob('~/x');
	//
	// Both are the same token to the lexer; which one it is depends on the
	// contents, and the parser can read those. What matters here is only
	// where the form ENDS.
	//
	// A newline ends the candidate: `$a < $b` on one line and `$c > $d` on
	// the next must not join into one token across them. Everything else is
	// permitted, because perl accepts `<a b>` -- a glob pattern with a space
	// in it -- as readily as `<*.c>`.
	j := l.pos + 1
	for j < len(l.src) && l.src[j] != '>' && l.src[j] != '\n' {
		j++
	}
	if j >= len(l.src) || l.src[j] != '>' {
		return false
	}
	start := l.pos
	l.pos = j + 1
	l.emit(Readline, start)
	return true
}

// fileTests are the 27 letters that make `-X` a file-test operator.
//
// perl returns UNIOP for them (toke.c:6255-6261 FTST), so they bind exactly
// like a named unary. The set was measured rather than transcribed: a letter
// is a filetest iff `my $v = -L $f;` compiles to an `ft*` op.
//
//	$ perl -MO=Concise -e 'my $f="/etc"; my $v = -o $f;'   fteowned
//	$ perl -MO=Concise -e 'my $f="/etc"; my $v = -n $f;'   method_named "n"
//
// The 25 letters Deparse renders back as `-L $f`, plus `o` and `O`, which are
// filetests that Deparse spells `-O` for both. `a h i j m n q v y` and the
// remaining capitals are NOT in the family, and `-n $f` is the shape they all
// take: negate applied to a method call.
var fileTests = map[byte]bool{
	'e': true, 'f': true, 'd': true, 'r': true, 'w': true, 'x': true,
	's': true, 'z': true, 'l': true, 'p': true, 'S': true, 'b': true,
	'c': true, 't': true, 'u': true, 'g': true, 'k': true, 'T': true,
	'B': true, 'A': true, 'M': true, 'C': true, 'o': true, 'R': true,
	'W': true, 'X': true, 'O': true,
}

// IsFileTest reports whether text is a whole file-test operator, `-e` and the
// rest of the 27.
//
// Exported because the parser needs the same answer and this is where the
// table lives: two copies of a 27-entry set is two things to keep in step.
func IsFileTest(text string) bool {
	return len(text) == 2 && text[0] == '-' && fileTests[text[1]]
}

// IsTrivia reports whether a kind carries no syntax: whitespace, a comment, a
// pod block, or a data section.
//
// Exported for the same reason as IsFileTest, and with a worse history. This
// set was restated in FOUR places -- `parse.isTrivia`, `fidelity.go`,
// `canon_test.go` and `use_test.go` -- and when `DataSection` joined it, three
// were updated and the fourth was not. A TIER_2 gate found the divergence;
// nothing in the suite did, because `firstWordOf`'s walk cannot reach a data
// section today and the drift was latent.
//
// The test copies live in `package parse_test`, which is why they could not
// call `parse.isTrivia` and restated it instead. That package boundary is the
// mechanism, so the predicate belongs HERE, beside the Kind it tests, where
// every caller can reach it.
//
// `HeredocBody` is deliberately absent. It was measured as trivia under
// 01a0c13f and rejected: canon must emit a body's bytes, and dropping them
// produced `my $h = <<"EOT";print($h);`, which does not compile. A body is a
// child of the statement that opened it, not trivia.
func IsTrivia(k Kind) bool {
	switch k {
	case Whitespace, Comment, Pod, DataSection:
		return true
	}
	return false
}

// scanFileTest lexes `-e`, `-d`, `-M` and the rest of the family as ONE
// operator whose text includes the minus.
//
// perlop's named-unary level holds the whole family, so this is one rule for
// 27 spellings. Before it existed the lexer emitted `Operator(-) Word(e)` and
// left the minus undecided -- and a bare `-` in front of a word has not said
// which of perl's three readings it meant.
//
// THE EXPECT STATE IS NOT THE GUARD, which is where this differs from
// `takeRepeatAssign`. There is no subtraction reading to protect: measured
// 5.42.0, a filetest where a term has already been read is a SYNTAX ERROR,
// not a minus.
//
//	$ perl -e 'my $x = 1 -e "/etc";'   syntax error at -e line 1, near "1 -e "
//	$ perl -e 'my ($a,$b); $a-e$b;'    syntax error at -e line 1, near "$a-e"
//
// perl formed `-e` in both and then had nowhere to put it. The one spelling
// that compiles does so only because `print` takes an indirect filehandle,
// and the `-e` is STILL one operator there:
//
//	$ perl -MO=Concise -e 'my ($a,$b); print $a -e $b;'
//	  rv2gv <- $a        $a is the filehandle
//	  ftis  <- $b        and `-e $b` the filetest
//
// So the scanner claims `-X` wherever it appears and never consults
// `l.expect`. Three boundaries decide it instead, each measured:
//
// EXACTLY ONE LETTER, not the head of a longer word. `-e1` and `-ee` deparse
// as `-$f->e1` and `-$f->ee`, method calls on a negated scalar, and `-exists`
// is the builtin `exists` negated. So the letter must not be followed by an
// identifier byte.
//
// ADJACENT, with no space. This is the sharpest one, because all the bytes of
// a filetest are present:
//
//	$ perl -MO=Deparse -e 'my $f="/etc"; print - e $f;'
//	print -$f->e;
//
// NOT IN FRONT OF A FAT COMMA. `-bareword` is the string `"-bareword"`, and
// that reading survives when the bareword is a filetest letter:
//
//	$ perl -MO=Deparse -e 'my $x = { -e => 1 };'
//	my $x = {'-e', 1};
//
// which is why the `=>` lookahead is here and not in the parser. It is the
// one place a following token, rather than the two bytes themselves, decides.
//
// An absent operand needs no rule: `$_="/etc"; print -e;` deparses as
// `print -e $_`, so the operator is whole with nothing after it.
func scanFileTest(l *lexer) bool {
	if l.src[l.pos] != '-' || l.pos+1 >= len(l.src) {
		return false
	}
	if !fileTests[l.src[l.pos+1]] {
		return false
	}
	// One letter only: a longer word is a method name, not a filetest.
	after := l.pos + 2
	if after < len(l.src) && isWordByte(l.src[after]) {
		return false
	}
	// `-e => 1` autoquotes the whole thing as a string key.
	if sep := skipSpaceFrom(l.src, after); sep+1 < len(l.src) &&
		l.src[sep] == '=' && l.src[sep+1] == '>' {
		return false
	}
	start := l.pos
	l.pos = after
	l.emit(Operator, start)
	return true
}

// scanAmp lexes `&` as a function sigil in term position.
//
// §0.13 rank 7, 3 corpus files. Spec §3.2's `&` row: operator position yields
// `&&` or bitwise-and, term position yields `&name`.
func scanAmp(l *lexer) bool {
	if l.src[l.pos] != '&' || !l.expect.wantsTerm() {
		return false
	}
	// `&&` is an operator even where a term is expected -- it is the
	// low-precedence form. Only a single `&` before a name is a sigil.
	if l.pos+1 < len(l.src) && l.src[l.pos+1] == '&' {
		return false
	}
	start := l.pos
	l.pos++
	l.emit(FuncSigil, start)
	return true
}

// scanComment lexes `#` to end of line. Trivia, and therefore a token.
func scanComment(l *lexer) bool {
	if l.src[l.pos] != '#' {
		return false
	}
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
	l.emit(Comment, start)
	return true
}

// operators is checked longest-first so `<=>` is not read as `<=` then `>`.
var operators = []string{
	"<=>", "**=", "||=", "&&=", "//=", "...", "<<=", ">>=", "^^=",
	"=~", "!~", "->", "++", "--", "**", "==", "!=", "<=", ">=",
	"&&", "||", "^^", "//", "..", "::", "+=", "-=", "*=", "/=", ".=",
	"%=", "^=", "|=", "&=", "=>", "<<", ">>",
	"+", "-", "*", "/", "%", ".", ",", "=", "<", ">", "!", "?", ":",
	"&", "|", "^", "~", "(", ")", "[", "]", "{", "}", "\\",
}

// bitwiseStringOps are the `bitwise` feature's string operators and their
// assignments, longest first. See scanOperator for why they are not in the
// operators table.
var bitwiseStringOps = []string{"&.=", "|.=", "^.=", "&.", "|.", "^.", "~."}

// scanOperatorName lexes the symbol a typed-Perl operator declaration is
// named by, `sub + :infix(ADD) (Num $x, Num $y) Num;` (RFC 0001, "Operator
// declarations"), as the Word a sub's name is. Only in a `.pmt`: in perl,
// measured on 5.42.0, `sub + { 1 }` is "Illegal declaration of anonymous
// subroutine". A bracket, comma or colon after `sub` still opens a
// signature, a body or an attribute.
func scanOperatorName(l *lexer) bool {
	if !l.typed || !l.sawSubWord {
		return false
	}
	for _, ops := range [][]string{bitwiseStringOps, operators} {
		for _, op := range ops {
			if !strings.HasPrefix(string(l.src[l.pos:min(l.pos+len(op), len(l.src))]), op) {
				continue
			}
			if strings.ContainsAny(op, "()[]{},:") {
				return false
			}
			start := l.pos
			l.pos += len(op)
			l.emit(Word, start)
			return true
		}
	}
	return false
}

// scanOperator lexes punctuation.
func scanOperator(l *lexer) bool {
	// The string-bitwise operators exist only under the `bitwise` feature,
	// and without it the `.` belongs to what follows: measured on 5.42.0,
	// `$a |.5` is `$a | 0.5` unfeatured and `$a |. 5` featured. The lexer
	// does not track that feature, so the operator is taken whenever the `.`
	// is NOT immediately followed by a digit -- which reads every spaced
	// spelling right, and `$a |.5` under the feature is the ceiling.
	for _, op := range bitwiseStringOps {
		end := l.pos + len(op)
		if end > len(l.src) || string(l.src[l.pos:end]) != op {
			continue
		}
		if op[len(op)-1] == '.' && end < len(l.src) && l.src[end] >= '0' && l.src[end] <= '9' {
			break
		}
		start := l.pos
		l.pos = end
		l.emit(Operator, start)
		return true
	}
	// `~~` is ONE operator, smartmatch, only where an operator is expected;
	// toke.c:
	//
	//	if (FEATURE_SMARTMATCH_IS_ENABLED &&
	//	    s[1] == '~' && (PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR))
	//	    ... NCEop(OP_SMARTMATCH);
	//
	// In term position it is two bitwise nots, and `~~$x` -- forcing scalar
	// context -- is an idiom that must keep lexing that way. It is not in the
	// operators table for that reason: the table is position-blind. The
	// smartmatch feature is on by default and the lexer does not track it;
	// `no feature 'smartmatch'` followed by `$x ~~ $y` is the ceiling.
	if l.src[l.pos] == '~' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '~' &&
		(l.expect == XOperator || l.expect == XTermOrDorDor) {
		start := l.pos
		l.pos += 2
		l.emit(Operator, start)
		return true
	}
	for _, op := range operators {
		if l.pos+len(op) > len(l.src) {
			continue
		}
		if string(l.src[l.pos:l.pos+len(op)]) == op {
			start := l.pos
			l.pos += len(op)
			// A closing bracket ends a term, so an operator follows it;
			// every other operator opens a term. Without the distinction,
			// `$x[0] <FH>` reads its angle brackets as a readline.
			kind := Operator
			// Except `*]`, the glob named `]` (see globStar): a name, and
			// taken as a closer it would pop whatever bracket is open.
			globName := op == "]" && l.globStar && l.toks[len(l.toks)-1].End == start
			if (op == ")" || op == "]" || op == "}") && !globName {
				kind = CloseBracket
			}
			l.emit(kind, start)
			return true
		}
	}
	return false
}

// scanSemicolon ends a statement, returning the machine to XSTATE.
func scanSemicolon(l *lexer) bool {
	if l.src[l.pos] != ';' {
		return false
	}
	start := l.pos
	l.pos++
	l.emit(Semicolon, start)
	return true
}

// scanBarePattern lexes `/.../` in TERM position.
//
// scanQuoteLike handles the keyword forms (`m//`, `s///`); this one has no
// keyword, so the expect state is the ONLY thing distinguishing a pattern
// from division. Spec §3.2's `/` row. Measured:
//
//	$ perl -MO=Deparse -e 'my @a=(1); my $n = @a / 2;'
//	my $n = @a / 2;
//
// t/base/lex.t:20 is the harder case, `eval '$foo{1} / 1;'` -- division,
// because a hash subscript closes a term.
//
// Bare `?...?` is deliberately absent: it was removed as a match operator in
// perl 5.22 and is a syntax error in 5.42, verified.
func scanBarePattern(l *lexer) bool {
	if l.src[l.pos] != '/' || !l.expect.wantsTerm() {
		return false
	}
	// XTERMORDORDOR wants a term, except that `//` is defined-or there:
	//
	//	if ((PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR) && s[1] == '/')
	if l.expect == XTermOrDorDor && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
		return false
	}
	start := l.pos
	l.pos++
	if !l.scanDelimitedBody('/', 0) {
		l.emit(UnknownRest, start)
		return true
	}
	patEnd := l.pos - 1
	l.scanModifiers()
	l.emit(Quote, start)
	l.queueBlockHeredocs(start+1, patEnd, codeBlockOpeners)
	return true
}

// scanPostDerefStar lexes `&*` and `**` after `->` as the whole postfix
// dereference they are, one Variable each, as `->@*` is: `$r->&*` is `&$r`
// and `$r->**` is `*$r`, measured on 5.42.0. Elsewhere `&` is the function
// sigil and `**` is exponentiation.
func scanPostDerefStar(l *lexer) bool {
	if l.expect != XPostDeref || l.pos+1 >= len(l.src) || l.src[l.pos+1] != '*' {
		return false
	}
	if c := l.src[l.pos]; c != '&' && c != '*' {
		return false
	}
	start := l.pos
	l.pos += 2
	l.emit(Variable, start)
	return true
}
