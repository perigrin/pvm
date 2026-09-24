"""Port tier 09_regex to the mdtest topic format."""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check  # noqa: E402

TIER = '09_regex'
covered = set()

covered |= topic(TIER, 'regex-matching.md', 'Matching and substitution', """
The tier's baseline: a pattern bound to a string, negated, interpolated,
and made to mutate its target.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals -- a match needs something to
match against, and the smallest such thing is a string literal.

Two measurements run through every case here. `=~` EMITS NO OP OF ITS
OWN, so the optree cannot tell you the operator was written; and a
constant pattern is compiled at COMPILE time while an interpolated one
is not, so one character of source is worth a whole op. Both are
recorded by the pairs below rather than by any single case.
""", [
    ('01_bare_match.t', 'A constant pattern compiles once', """
A pattern with no interpolation is compiled at COMPILE time: the match
emits `match` alone, with no `regcomp` beside it. This is the tier's
baseline, and measuring the constant form first is what makes the
interpolated case's extra op attributable.

The token claims are negative, and deliberately so. GLOSSARY.md's
`quote-like operator` is defined as a quote spelled with an OPERATOR
NAME -- q, qq, m, s, qr -- or with backticks; a bare `/abc/` is neither,
so there is no positive category to assert it under and this case does
not invent one. What it can say is what a mis-lex would produce: the
hazard for a bare-slash pattern is the slash read as DIVISION and the
pattern body read as Perl, which yields an `Operator("/")` and a
`Word("abc")`. Asserting their absence is the claim that survives.
"""),

    ('02_bind.t', 'The binding operator emits nothing', """
`=~` emits NO OP OF ITS OWN. The binding operator is absorbed into the
match op, which carries its target in its own flags -- the stream shows
`match(/"b"/)[$s:1,3]` and no `bind` anywhere.

So this case and the bare match above differ in their OPERAND, not in
their op set. That is why the binding needs a case of its own: the
optree cannot tell you the operator was written, only that the match
found a target other than `$_`, and the token stream is what records the
operator.
"""),

    ('03_negated_bind.t', 'The negated binding is two ops', """
`!~` is not a match op with a flag: it is `match` followed by `not`, a
second op stacked on top of the first. `not` is tier 04's op, claimed
there with the logical operators, and its appearance here is the negated
binding REUSING it rather than this tier introducing anything. The
measurement is still this tier's: reading the source would suggest a
negated match op, and there is none.

The pattern deliberately does not match, so the printed value is the
empty string rather than 0 -- Perl's false is `""` -- which is why the
brackets are there. Without them the pinned output would be a line with
nothing on it, and a mis-measured match would print the same.
"""),

    ('04_substitution.t', 'Substitution mutates in place', """
A substitution mutates its target in place and emits one `subst` op. No
assignment, no block, no dereference -- which is the measurement behind
this tier's claim that it could sit at 03.

The replacement `z` arrives as a `const` pushed before the `subst`, not
as a second argument to it, and there is no `sassign`: `$s` is named in
the subst op's own flags exactly as the match op names its target.
"""),

    ('05_interpolated.t', 'An interpolated pattern adds `regcomp`', """
An interpolated pattern cannot be compiled at compile time, so the
stream gains `regcomp` between the `padsv` that fetches the variable and
the `match` that uses the result. This is the ONLY form in the tier that
emits `regcomp`. Against the bare match above, one character of source
differs and a whole op appears; the tier's op set is their union.

Note also what the match op loses: `match()[$s:2,4]` carries no pattern
in its own dump, because it has none until `regcomp` runs.

The token claim is the sharpest in the tier. `$p` appears TWICE in the
source and must lex as ONE variable: the declaration's, with the
pattern's occurrence staying INSIDE the pattern token. A lexer that
recurses into a pattern -- reasonable-looking, since the pattern really
does interpolate -- produces two, and this is the assertion that catches
it. The optree cannot: perl interpolates at runtime either way and
`regcomp` is emitted regardless.
"""),
])

covered |= topic(TIER, 'regex-delimiters.md', 'Delimiters and opacity', """
Choosing the delimiter and finding its match is the lexer's whole job in
this tier, so these are its central cases -- and every one of them is a
TOKEN claim, because delimiters erase themselves from the optree.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals.

Measured under 5.42.0, `m{abc}` and `/abc/` emit byte-identical op
streams, as do `s{a}{z}` and `s/a/z/`, and `m#abc#` and `/abc/`. The
optree cannot distinguish them and neither can the printed output, so a
delimiter case asserting only ops would measure nothing the bare forms
already measure and would go green against a lexer that silently
normalised every delimiter to `/`. This is the same argument
`conformance/README.md` makes for `5e-1`.
""", [
    ('06_alternative_delimiters.t', 'Bracketing and punctuation delimiters', """
`m{abc}` and `s{a}{z}` emit op streams BYTE-IDENTICAL to `/abc/` and
`s/a/z/` -- measured, down to the pattern text perl prints inside the
op: `match(/"abc"/) sKS`. So this case's whole measurement is its token
claims.

`s{a}{z}` is the case where the second pair's opening delimiter is free
to differ from the first's, which GLOSSARY.md records as true only when
the first pair is bracketing. `m!abc!` is the non-bracketing case, where
one character delimits both ends and nothing nests -- which is also why
its body holds no `!`.
"""),

    ('08_hash_delimiter.t', 'The comment character as a delimiter', """
`#` is PERL'S COMMENT CHARACTER, and it is also a legal delimiter. A
lexer reading left to right has already decided `m#abc#` is a comment
before it can know it was wrong. Every other delimiter in the tier is a
character with no other job at that position; `#` has one, and the two
readings differ by everything -- `m#abc#` is a match, while `m` followed
by a comment is a bareword and then nothing at all to end of line.

Measured under 5.42.0, the op stream is BYTE-IDENTICAL to the one for
`/abc/` and `s/a/z/`: `match(/"abc"/) sKS` and `subst(/"a"/)`. So
neither the optree nor the printed output can report a lexer that got
this wrong, and the token facts below are the only place the claim can
live.

Both spellings are here rather than the match alone, because the
substitution is where the hazard compounds: `s#a#z#` has THREE `#`
characters, and a lexer that recovers from the first by luck still has
two more to mis-read.
"""),

    ('09_nesting_delimiters.t', 'A bracketing delimiter nests', """
`m{a{b}c}` is a match on the pattern `a{b}c`: the inner braces are
CONTENT, not the terminator, and a lexer must count depth rather than
stop at the first `}`. This is the half of the tier's scope that a
delimiter FORM cannot express -- every other bracketing pattern here has
a body with no bracket in it, so all of them are satisfied by a lexer
that stops at the first closing character.

The asymmetry is what makes the rule a rule, and GLOSSARY.md records
both halves: bracketing delimiters nest, non-bracketing ones do not.
Measured, the non-bracketing counterpart is not merely different but
ILL-FORMED -- `"a!b!c" =~ m!a!b!c!` gives `Unknown regexp modifier "/b"`
and then a syntax error -- so it cannot appear in a case expected to
parse.

THE SUBSTITUTION IS WHAT PINS THE PATTERN'S EXTENT, and the `x` and `y`
around the target are why. A lexer that stopped at the first `}` would
take the pattern as `a{b` -- which still matches, so a bare match would
report 1 either way -- but the replacement would then cover three
characters instead of five and leave `}c` behind. `xoky` says the whole
five characters went, bounded on both sides by text the pattern must not
have touched.
"""),
])

covered |= topic(TIER, 'regex-quote-likes.md', 'Quote-likes that are not matches', """
Two constructs wearing regex syntax that the optree says are something
else: `qr//` compiles a pattern without matching anything, and `tr///`
is not a regex at all.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals.

Both sit in this tier for the same reason: the LEXER cannot tell them
from a match. `qr` and `tr` are operator names in the same table as `m`
and `s`, and `tr` needs the entry in that table saying a second region
follows. What they emit afterwards is a different question, and in both
cases the answer is one op with no `match` beside it.
""", [
    ('07_qr.t', 'A compiled pattern, uncoupled from matching', """
Plain `qr/abc/` emits a single `qr` op and nothing else: no `regcomp`,
no `match`, no block, no recursion. It is closer to a literal than to
anything tier 14 does.

Tier 14 owns `qr//` WITH EMBEDDED CODE, which is a re-entry problem.
Claiming the plain form there would make tier 14 depend on this tier for
the non-recursive half of one construct, so it stays here and tier 14
claims only what embedding adds.

Printing the object is safe because its stringification is deterministic
-- and it is `(?^:abc)`, NOT `(?^u:abc)`. The `u` appears only under a
unicode_strings-like feature bundle; this case runs without one.
"""),

    ('11_transliteration.t', '`tr///` wears `s///`\'s syntax and is not a regex', """
`tr///` is the only quote-like besides `s///` that takes TWO delimited
regions -- and neither of its regions is a pattern. That asymmetry is
the trap.

WHY THE TWO-REGION SHAPE IS THE LEXICAL CLAIM. A lexer that knows only
`s` takes a second region reads `tr/./Z/` as `tr/./` followed by the
stray tokens `Z` and `/`, which is a different token stream and a
different program. Nothing about `tr`'s first region announces that a
second one follows; the operator NAME is the only signal, so a lexer
must carry a table of which quote-like names take two regions, and `tr`
(with its synonym `y`) is the entry most often missing from it.

AND `tr` IS NOT `s` WHERE IT COUNTS. Measured under 5.42.0 on the same
subject, `$_ = "a.c"; tr/./Z/` prints `aZc` while `$_ = "a.c"; s/./Z/`
prints `Z.c`. `tr`'s `.` is the CHARACTER dot; `s`'s `.` is the
metacharacter matching anything. So a parser that desugared `tr` into
`s` -- an easy thing to do when the syntax is this similar -- produces a
program that compiles, runs, and prints the wrong answer. Both spellings
run on one subject so the results sit side by side, and a desugared `tr`
prints `Z.c 1 Z.c` instead of `aZc 1 Z.c`.

THE RETURN VALUE IS THE SECOND BEHAVIOURAL PIN. `tr` returns the COUNT
of characters it transliterated, where a bare match returns a boolean.
Measured, `"a.c" =~ tr/./Z/` is 1. The count and the boolean agree at 1
here, which is why the count is pinned beside the two strings rather
than trusted to carry the claim alone.

The op is `trans`, measured as `trans[$t:1,2]
sP/TRANS=ONLY_UTF8_INVARIANTS` -- one op, no `match`, no `regcomp`, no
`subst`, which is the optree agreeing that `tr` is not a regex construct
at all despite wearing the syntax of one.

THE REPLACEMENT CHARACTER IS `Z` AND NOT `X`, which looks arbitrary and
is not. `no word whose text is "Z"` is the fact that falsifies a lexer
leaking the second region -- and with `X` as the replacement the fact is
false before `tr` is reached, because `$ENV{X}` puts a `Word("X")` in
the source's first line. A negative token fact is only a claim about the
construct if the character it names appears nowhere else.

The subject reads `$ENV{X}`, unset when the runner executes it, because
a `tr` on a constant is a folding candidate and a folded `tr` measures
nothing.
"""),
])

covered |= topic(TIER, 'regex-pattern-positions.md', 'Patterns in unexpected positions', """
A pattern in an ARGUMENT slot and a regex operator on the LEFT of an
assignment. Both are positions a parser modelled on ordinary expressions
gets wrong, and in both the failure is silent.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals.

Each case is arranged so the OUTPUT reports the mis-parse, because the
optree will not: `split`'s two spellings emit byte-identical ops while
printing different answers, and `pos` is the same op in rvalue and
lvalue position, differing only in perl's printed flags. Both subjects
read `$ENV{X}`, unset when the runner executes them, because a `split`
or a match on a constant is a folding candidate and a folded one
measures nothing.
""", [
    ('10_split_pattern.t', "`split`'s first argument is a pattern", """
`split /,/, $s` puts a regex literal in an argument slot, where a parser
reading `/` as division produces a tree perl never builds.

THE FAILURE IS SILENT. `split /,/, $s` and `split ",", $s` behave
identically on every ordinary input, so a parser that mistook the
pattern for a division, recovered, and produced a string would print
exactly what a correct one prints.

Except on ONE input, and it is the case's whole measurement. A single
SPACE as split's first argument is perl's awk-compatibility special
case: the STRING `" "` means "split on runs of whitespace and discard
leading whitespace", while the PATTERN `/ /` means what it says, one
space. Measured under 5.42.0 on `"  a b "`, `split / /` gives `||a|b`
and `split " "` gives `a|b`. So the two spellings are a PARSE apart, and
perl itself distinguishes them at parse time rather than at runtime.

AND THE OPTREE CANNOT SEE IT. Measured, both spellings emit the same op
with the same pattern text printed inside it -- `split(/" "/ => @p:2,3)
[t4] vK/LVINTRO,ASSIGN,LEX,IMPLIM` -- byte-identical, for two programs
that print different things. That is this tier's standing argument in
its sharpest form, because here the erasure hides a difference the
output can still see.

The token facts are the only place the distinction can be asserted, and
they are COUNTED rather than merely forbidden. `no operator whose text
is "/"` forbids the division reading: a lexer that scanned `/ /` as two
divide operators around a space would emit two `Operator("/")` tokens
where the pattern belongs. `one string literal whose text is "\\" \\""`
is the counting half and the sharper of the two -- this source holds
exactly ONE quoted string spelled `" "`, the second split's argument,
and the first split's `/ /` must NOT be one. A lexer that read a
slash-delimited pattern as a string literal would produce TWO tokens
matching that text and fail the count. Measured, ours emits `Quote("/
/")` for the pattern and `Quote("\\" \\"")` for the string, and only the
second is a string literal by the glossary's rule.

Neither fact can be written as `one word whose text is "split"`: the
source calls `split` twice, on purpose, since the whole measurement is
the two spellings side by side.
"""),

    ('12_pos.t', '`pos` is a named operator in lvalue position', """
`pos` is LVALUE-CAPABLE and tied to the regex engine: it reads and
WRITES the match position `//g` leaves behind on a scalar.

THE LVALUE FORM IS WHY THIS IS A PARSE AND NOT A CALL. `pos($s) = 0`
puts a named operator on the LEFT of an assignment, which almost nothing
in Perl does -- measured, the optree puts the `pos` op underneath
`sassign`'s left arm, `<1> pos[t3] sKRM*/1` then `<2> sassign vKS/2`. A
parser that modelled `pos` as an ordinary named unary returning a value
would refuse that assignment or silently discard it, and the second case
is the dangerous one: the program still runs.

THE OUTPUT REPORTS THE DISCARD, which is the case's whole design. `//g`
on a scalar advances a stored position, so two successive `/b/g` matches
against `"abcabc"` find the two `b` characters in turn. Measured,
without the reset the two positions are `2 5`; with it they are `2 2`.
So a parser that dropped `pos($s) = 0` prints `2 5` where this case pins
`2 2`, and the discarded statement is the only thing that could account
for it.

The op is the same in both positions -- rvalue `my $a = pos($s)` and
lvalue `pos($s) = 0` differ in the flags perl prints, not in the op --
so the op claim is earned by either spelling and the LVALUE claim is
behavioural.

THE TOKEN FACTS ARE ABOUT THE MODIFIER, not about `pos`. `pos` lexes as
an ordinary `Word` and this source holds three of them, so the facts
grammar's `one`/`no` cannot pin it. What the facts CAN pin is the thing
`pos` is useless without: `/b/g` carries a MODIFIER, and the modifier is
part of the quote token. `no word whose text is "g"` forbids the
modifier escaping the quote -- a lexer that stopped at the closing
delimiter would leave `g` behind as a stray `Word("g")`, a program that
still parses and still runs, printing `2 2` for the wrong reason. `no
word whose text is "b"` forbids the pattern BODY escaping it, which is
what a lexer reading the slashes as division would produce. Together
they say the three characters `b/g` were consumed by the quote and by
nothing else.

A positive fact is not available for the glossary's reason, not by
omission: the source spells `m/b/g` twice, deliberately, since the
measurement IS the two matches, so `one quote-like operator whose text
is "m/b/g"` would be false at a count of two.
"""),
])

covered |= topic(TIER, 'adjacency-09_regex.md', 'Every regex construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another, and every delimiter form the tier teaches.

**Tier 09 regex.** Introduces nothing of its own; it is the mixture that
is the subject. Depends on 01_literals.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why a corpus of such cases cannot
reach an ADJACENCY bug -- and in a tier whose operand is not Perl, the
adjacency bug is the one to expect. A lexer that delimits a pattern by
scanning for the next `/` is correct on every case here taken alone and
wrong the moment an `s{a}{z}` sits between two matches.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
The constructs sit consecutively: a constant match, a negated match
through a bracketing delimiter, a match through a non-bracketing one, a
match through the comment character, a match whose pattern NESTS its own
delimiter, an interpolated match, a `qr//`, four substitutions delimited
by brackets, slashes, hashes and nested brackets, a `tr///`, a `split`
on a pattern, and a `//g` match read through `pos` -- then one print
that reads every result.

THE LAST THREE ARE WHERE THE ADJACENCY CLAIM EARNS ITS KEEP a second
time. `tr/./Z/` is a second two-region quote-like, and a lexer whose
table says only `s` takes two regions mis-terminates it and then
mis-terminates everything after it -- a failure a one-construct-per-case
`tr` test cannot reach, because there is nothing after it to break.
`split / /` puts a pattern in an ARGUMENT SLOT immediately following
that, which is the position a lexer recovering from a botched `tr` is
most likely to read as division. The pattern is a single space, the
least distinguishable body a slash-delimited pattern can have.

EVERY DELIMITER FORM THE TIER TEACHES APPEARS HERE, which is the part
that needed fixing. The body shipped holding `m{}`, `s{}{}` and `qr//`
alone -- so the one bug it exists to reach was unreachable for every
form most likely to produce it.

Three of the added forms are the ones that cannot NEST. `m!abc!`,
`s/c/y/` and `s#d#w#` each close with the character they opened with, so
a lexer has no bracket depth to count and must simply stop at the next
occurrence. The fourth is the opposite case and is here for the
contrast: `m{a{b}c}` and `s{a{b}c}{ok}` carry their own opening brace
inside the pattern, where a lexer MUST count depth. That pair of rules
is what GLOSSARY.md records, and neither `m{zzz}` nor `s{a}{z}`
exercises either half.

The hash forms are the sharpest pair. `#` is Perl's comment character,
so a lexer that has not yet recognised the `m` or `s` has already
discarded the rest of the line. Putting them BETWEEN other delimiter
forms rather than alone is the adjacency claim in its strongest version:
recovering from `m#abc#` is not enough if the `s#d#w#` three statements
later is then read as a comment.

The tier's dependency on 01_literals is present rather than decorative:
every pattern here is matched against a string literal bound to a pad
slot, and the final print interpolates every result into one
double-quoted string. Pairing with 08_references instead would assert
nothing.

The substitutions run LAST on purpose. They mutate `$s`, which the five
matches above them read; running any earlier would make those results
depend on statement order in a way that hides a mis-parse behind a
plausible-looking output. Each of the three that touch `$s` mutates a
DIFFERENT character -- `a`, `c` and `d` -- so the final `zbyw` records
that all three ran, where two substitutions of one character would leave
the second's failure invisible. The nested substitution needs its own
target, `$n`, because its pattern is five characters `$s` does not hold.
"""),
])

check(TIER, covered)
