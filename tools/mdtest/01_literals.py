"""Build the 01_literals topic files from the extracted .t claims.

Sources, pins, token facts and refusal records are COPIED out of
corpus.json by gen.py. Only the prose below is written fresh.
"""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = """**Tier 01 literals.** Introduces `const`, `enter`, `leave`,
`multiconcat`, `nextstate`, `padrange`, `padsv`, `padsv_store`, `print`,
`pushmark`. Depends on nothing -- it is the only tier that can say that.
`my` and `print` appear throughout as FIXTURES rather than as subjects: a
literal has to be bound to something and observed somehow."""

covered = set()

covered |= topic('01_literals', 'numeric-radix.md', 'Numbers by radix', """
The spellings that say what BASE the digits are in, plus the decimal
integer they all deviate from, plus the separator that is a word
character everywhere else in perl.

""" + TIER + """

The radix is INVISIBLE after compilation. `0b1010`, `0xff`, `0377`,
`0o377` and `4_294_967_296` all arrive as a `const[IV ...]`,
indistinguishable from the same value written in decimal, so the optree
cannot say which spelling was written and output cannot either. Every
case here turns on its token fact for that reason -- except the leading
zero, where a lexer that does nothing special changes the VALUE and
output catches it.
""", [
    ('03_decimal_integer.t', 'A decimal integer', """
The baseline every other spelling in this tier deviates from, and the
tier's simplest construct.

The tier had twelve construct files and none of them was a plain
integer: `0.5` is the FRACTIONAL boundary and the rest are hexadecimal,
binary, octal, exponent and v-string -- every one a deviation from a
baseline no file stated. `= 42;` appeared only in the adjacency body,
which introduces nothing, so the tier's simplest construct was the one
it never asserted. `TestTierLiteralsCoversGlossary` named `decimal
integer` among this tier's boundaries and passed anyway, because it
joined the adjacency source in with the construct files'; excluding that
file made three boundaries fail at once, this being the third.

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $x = 42; print "$x\\n"'
    ... const[IV 42] ... padsv_store ... multiconcat ... print ...

`const[IV 42]`, an INTEGER const, where `0.5` gives `const[NV 0.5]`.
Same op, different SV type, and the optree is the only place that
difference is visible -- output prints `42` either way. The token fact
is what pins the spelling."""),

    ('01_binary.t', 'Binary: the `0b` prefix', """
A `0b` prefix makes the digits after it binary: `0b1010` is one token
denoting ten.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0b1010; print "$x\\n"'
    10

After compilation this is `const[IV 10]`, the same op `my $x = 10`
produces, so nothing downstream can say which spelling was written.
`perldata` lists the form under "Scalar value constructors", which is
why GLOSSARY.md carries it as a numeric literal."""),

    ('05_hexadecimal.t', 'Hexadecimal: the `0x` prefix', """
A `0x` prefix makes the digits after it hexadecimal, letters included:
`0xff` is one token denoting 255.

The trap is that `ff` is also a legal identifier. A lexer scanning the
digit `0`, stopping at the first non-digit and handing `xff` to the word
scanner produces `Number(0) Word(xff)` -- which the parser reads as a
number beside a bareword, not as an error. That is why the negative fact
names the word.

MEASURED perl 5.42.0, the three radix spellings and one decimal are the
SAME value:

    $ perl -e 'printf "%s %s %s\\n", 0xff, 0377, 0o377'
    255 255 255"""),

    ('08_octal_leading_zero.t', 'Octal by leading zero', """
A leading zero makes the digits after it octal: `0377` is one token
denoting 255, not three hundred and seventy-seven.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0377; print "$x\\n"'
    255

The radix is carried by a character a decimal literal may also begin
with, so this is the one radix form a lexer can get wrong by doing
NOTHING: scan the digits, read them as decimal, and `0377` becomes 377
with no token boundary out of place and nothing to report. Only the
VALUE differs, which is why this case pins output as well as tokens."""),

    ('09_octal_prefix.t', 'Octal by the `0o` prefix', """
An explicit `0o` prefix is octal too: `0o377` is one token denoting 255,
the same value `0377` denotes. Perl 5.34 added the spelling so octal
need not be signalled by a leading zero alone.

It is a separate case from the leading zero because it is a separate
LEXICAL form -- the two agree on the value and compile to the same
`const[IV 255]`, so the optree cannot tell them apart and only the token
stream can say which was written.

The trap is the hexadecimal one: `o377` is a legal identifier, so a
lexer that stops the number at the first non-digit produces `Number(0)
Word(o377)` and the parser sees a number beside a bareword rather than
an error."""),

    ('14_underscore_separators.t', 'Underscore separators', """
Underscores inside a numeric literal are separators, part of the token
and absent from the value: `4_294_967_296` is one token denoting
4294967296.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 4_294_967_296; print "$x\\n"'
    4294967296

The underscore is a WORD character everywhere else in perl, which is
what makes this a boundary rather than a detail. Two wrong answers are
reachable: a lexer that stops a number at the first non-digit produces
`Number(4) Word(_294_967_296)`, and a lexer that scans word characters
greedily after a digit produces one token whose text is right and whose
value is unparseable. The token facts below distinguish both from the
correct answer, which the printed value cannot -- the separators are
gone by the time anything prints."""),
])

covered |= topic('01_literals', 'numeric-point.md',
                 'The decimal point, the exponent and the sign', """
Where a numeric literal STARTS and STOPS. Four of these five cases are
about a `.` that is also the concatenation operator, and the fifth is
about a `-` that is never part of the number at all.

""" + TIER + """

MEASURED perl 5.42.0, and this is the reason the token layer exists:

    $ perl -e 'printf "%.17g %.17g %.17g\\n", .5, 0.5, 5e-1'
    0.5 0.5 0.5

Three spellings, the SAME VALUE to 17 significant digits. No behavioural
probe can see which was written, so each case here carries a token
assertion and the output pin is only a guard against a lexer that got
the value wrong as well.
""", [
    ('02_decimal.t', 'A decimal point between digits', """
`0.5` is one token, not `0` `.` `5`.

This is the case that already works, and it is here for two reasons. It
is the baseline the two cases after it deviate from -- `.5` and `5e-1`
are the same construct with the integer part removed and an exponent
sign added -- so a regression here would explain both without being
separately diagnosed. And a corpus whose every case refuses cannot
demonstrate that passing is reachable.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0.5; print "$x\\n"'
    0.5"""),

    ('06_leading_decimal.t', 'A leading decimal point', """
`.5` is one token, not the concatenation operator `.` followed by `5`.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = .5; print "$x\\n"'
    0.5

This refused from 38c95d23 until the lexer learned the rule: our lexer
produced `Operator(.) Number(5)`, because `scanNumber` required a
leading DIGIT. Fixed under issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe,
which made a `.` before a digit start a number where a TERM is expected
and leave it as concatenation where an operator is. The token fact below
was written while it still refused, and it is what the fix had to
satisfy. Nothing in it changed."""),

    ('13_trailing_decimal.t', 'A trailing decimal point', """
A trailing point with no digits after it is still part of the numeric
literal: `1.` is one token, not `1` followed by the concatenation
operator.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 1.; print "$x\\n"'
    1

The mirror image of the leading point, and the pair is why both are
here: `.5` puts the dot where a lexer expects an operator and `1.` puts
it where a lexer expects more digits. A scanner that requires a digit on
BOTH sides of the point gets each of them wrong in a different way, and
a corpus holding only one of the two would diagnose that as a single
fault.

The output cannot distinguish the spellings -- `1.` prints `1`, exactly
as `1` does -- so the token assertion is what this case turns on."""),

    ('07_negative.t', 'A leading minus is not part of the literal', """
`-1` is two tokens, a negation operator and the literal `1`. This is the
asymmetry the signed exponent is the other half of: the sign of an
EXPONENT is part of the token; a sign in front of the number never is.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = -1; print "$x\\n"'
    -1

    $ perl -MO=Concise -e 'my $x = -1;' 2>&1 | grep const
    const[IV -1] s/FOLD

The second measurement is why this case exists. The optree holds ONE
FOLDED CONSTANT: the optimiser has already applied the negation and
erased the operator, so nothing downstream of compilation can tell `-1`
from a hypothetical single negative literal. Behaviour cannot see it
either -- both would print `-1`. The token stream is the only place the
two tokens are still two, which is this tier's argument for the token
layer stated where it is cheapest to check.

Our lexer gets this right today: `Operator(-) Number(1)`. The case is
here to keep it right, not to report that it is wrong."""),

    ('11_signed_exponent.t', 'The sign of an exponent is part of the literal', """
`5e-1` is one token, not `5e` minus `1`.

MEASURED perl 5.42.0:

    $ perl -e 'print 5e-1, "\\n"'
    0.5

    $ perl -e 'print 5e, "\\n"'
    Bareword found where operator expected (Missing operator before "e"?)
    syntax error at -e line 1, near "5e"

The second measurement is what makes the split WRONG rather than merely
different: `5e` is not a number, so a lexer emitting `Number("5e")` has
produced a token perl would reject.

REFUSES as of 38c95d23. Our lexer produces `Number(5e) Operator(-)
Number(1)`, and the same split affects `5e+1`, `5E-1` and `1.5e-3`.
Unsigned `5e1` lexes correctly, which is why it went unnoticed. The
refusal is LEXICAL, so the parser returns no Unknown node and there is
no refusal code to name.

No issue: this case is the record. The construct was found by writing
it, so there is nowhere earlier for it to have been filed, and
duplicating the token stream into a tracker would give it a second place
to go stale."""),
])

covered |= topic('01_literals', 'vstrings.md', 'V-strings', """
Two dots make a string. A v-string is the boundary where a thing that
looks entirely like a number is not one, and GLOSSARY.md records the
decision under "numeric literal": a v-string is not in that category.

""" + TIER + """

MEASURED perl 5.42.0:

    $ perl -e 'my $n = 65.66; print "$n\\n"'
    65.66

ONE DOT IS A NUMBER AND TWO DOTS ARE A STRING. That is the whole
boundary, and it is decided by COUNTING DOTS after the token has already
started -- a lexer cannot know which it is scanning until it reaches the
second point or the end. `perldata` documents the form under "Version
Strings", away from numbers.

Both cases below refuse, and both refusals are LEXICAL: the parser
receives a well-formed expression, returns ZERO Unknown nodes, and
leaves no refusal code to name. The token assertions are the only place
the disagreement is visible. Neither has an issue, because each was
found by writing the case.
""", [
    ('15_vstring_bare.t', 'Bare: three dot-separated parts', """
`65.66.67` is one token denoting the three-character string `ABC`, with
no `v` anywhere.

MEASURED perl 5.42.0:

    $ perl -e 'my $v = 65.66.67; print "$v\\n"'
    ABC

REFUSES as of 7711154e. Our lexer produces `Number("65.66.67")` -- one
token, which is right, in the wrong CATEGORY, which is not. The parser
sees a number where perl sees a string and the value never reaches
output because we do not run the program."""),

    ('16_vstring_v.t', 'The `v` prefix', """
A leading `v` makes a v-string of what follows, one dot or many:
`v65.66.67` is ONE token denoting `ABC`.

MEASURED perl 5.42.0:

    $ perl -e 'my $v = v65.66.67; print "$v\\n"'
    ABC

    $ perl -e 'my $v = v5.42.0; print join(".", map ord, split //, $v), "\\n"'
    5.42.0

The second measurement is the form this repository writes constantly --
`v5.42.0` is how every `use` line spells a version -- and it is the same
construct.

REFUSES as of 7711154e. Our lexer produces `Word("v65") Operator(".")
Number("66.67")`: `v65` is a legal identifier, so the word scanner
claims it and the rest of the v-string is read as a concatenation of a
bareword with a number. The PARSER cannot see this. It receives Word
Operator Number, reads a valid expression, and returns ZERO Unknown
nodes -- the same shape the signed exponent records, and the same reason
the token layer exists."""),
])

covered |= topic('01_literals', 'strings.md', 'Strings and quoting', """
What the DELIMITER decides. Three cases, and the op stream can see none
of the distinctions any of them make.

""" + TIER + """

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $s = q{plain}; print $s'
    ... const[PV "plain"] ... padsv_store ... print ...
    $ perl -MO=Concise,-exec -e 'my $s = "plain"; print $s'
    ... const[PV "plain"] ... padsv_store ... print ...

Byte-identical. A single-quoted string, a double-quoted string with
nothing to interpolate, and `q{...}` all arrive as one `const`. That is
why GLOSSARY.md keeps `string literal` and `quote-like operator` as
separate categories, and why these claims are lexical: a case asserting
`one string literal whose text is "hi"` must not be satisfied by
`qw(hi)`, which is not a string at all.

What output CAN still see is escape processing, which is what separates
the first two cases:

    $ perl -e 'print length("a\\nb")'       3
    $ perl -e 'print length(q{a\\nb})'      4

So one prints two lines and the other prints one, and a lexer that
treated the two delimiters alike fails exactly one of them, whichever
way it was wrong.
""", [
    ('04_double_quoted.t', 'Double quotes process escapes', """
`"a\\nb"` is THREE characters, and the same bytes in single quotes are
four.

INTERPOLATION is deliberately NOT exercised here. `"v=$x"` needs a
runtime operand to be worth anything, and the corpus idiom for one --
`$ENV{X} // <default>` -- uses `dor`, which is 04_operators' op and
three tiers away. Interpolation on an ARRAY is already claimed by
03_context, which is where the construct earns its keep. What this tier
owns is the DELIMITER, and the delimiter is what the escape
distinguishes."""),

    ('12_single_quoted.t', 'Single quotes do not', """
A single-quoted string does not interpolate, and its backslash escapes
are not the double-quoted set: `'a\\nb'` is FOUR characters.

The tier was chartered for "numbers, strings, quoting and qw" and
shipped twelve numeric files and no string file at all. The three string
boundaries its glossary check names were satisfied only by the adjacency
body, which introduces nothing -- so the tier asserted the numeric half
of its vocabulary and borrowed the rest from a case that exists to
compose, not to introduce.

A lexer that applied double-quoted escape processing here would print
`a`, a newline and `b`. `length` is claimed by no tier, so the count is
not measured in the case: the OUTPUT is the four characters themselves,
and a lexer that collapsed `\\n` would print three characters across two
lines instead.

The token fact spells the backslash DOUBLED, `'a\\\\nb'`, because the
fact's text is compared against the source bytes and the source holds a
literal backslash. The trailing `"\\n"` is a second quote token, which is
why the fact counts `one` of a specific text rather than one quote: a
bare count would be two and the claim would be about the case's
punctuation rather than about its construct."""),

    ('10_quote_operators.t', '`q` with three delimiters', """
`q` and `qq` are QUOTE-LIKE OPERATORS, not string literals: a different
glossary category making a different claim, and the delimiter is chosen
rather than fixed. That separation is only worth having if some case
asserts each side of it, and until this one no corpus file asserted this
side.

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $s = q(a b); print $s'
    ... const[PV "a b"] ... padsv_store ... print ...

One `const`, exactly as `'a b'` gives. So `q(a b)` and `'a b'` are the
same op and DIFFERENT TOKENS, which is precisely the case where only a
token fact can carry the claim.

The DELIMITERS are what this case pins, and each is a separate chance to
be wrong: `q(...)`, `q{...}` and `q!...!` are all one construct wearing
three delimiters, and a lexer that hard-codes one of them fails the
others. `qq` is not spelled separately because what separates it from
`q` is escape processing, which the two quoted-string cases above
already pin for the delimiter-less spellings -- repeating it here would
measure the same thing twice.

`qw` is NOT here. Measured, `my @w = qw(a b c)` emits `aassign` and
`padav`, both 02_variables' ops, because a word list needs an array to
land in. It belongs to the tier that owns arrays, and a case here would
have to borrow two ops from a later tier to hold it."""),
])

covered |= topic('01_literals', 'adjacency-literals.md',
                 'Every literal, each beside another', """
One body holding every spelling this tier's cases introduce, each
adjacent to another.

**Tier 01 literals.** Introduces nothing of its own; it is the mixture
that is the subject. Depends on nothing, so there is no earlier tier to
pair with -- the adjacency is entirely within 01.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one
construct each, which is what makes them diagnosable: when the leading
point refused, the construct that refused was the only one present. That
same property is why a corpus of such cases cannot reach an ADJACENCY
bug -- a parser that handles every construct alone and mishandles a pair
goes green over the pair.

MEASURED perl 5.42.0, and this is not hypothetical:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses; `ADJUST` followed by anything does not. No
one-construct-per-case corpus can ever see that, because every case is
one construct by definition.

`TestTierLiteralsAdjacency` reads the tier's construct sources for the
literal each binds and requires the spelling to appear here, so this
body cannot fall behind the tier by a construct. It cannot check the
ADJACENCY itself; see the note in that test about `padrange` absorbing
`pushmark`, which is why more adjacent constructs emit FEWER ops.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
Every numeric spelling the tier introduces -- binary, decimal,
hexadecimal, leading point, negative, octal by leading zero, octal by
prefix, signed exponent, trailing point, underscore separators, and both
v-string forms -- plus an integer, a single-quoted string, an
interpolating string and a `q` list.

This body refused from 7711154e until issue
01a0c13f-97f5-7f98-b32d-07245ec6ddfe, and it refused for a BORROWED
reason: the `.5` on line 4 was the leading-point gap reaching here, not
a second bug. The note recorded at the time said an adjacency case
cannot pass while any construct it holds refuses, and that composing
only the working constructs would make it green and make it stop
covering the tier. Fixing the lexer cleared both in one change, which is
that prediction coming true.

The tier's other two refusals -- the exponent split and the v-strings --
are LEXICAL and produce no Unknown at all, so they leave no code here to
name. See `TestTierLiteralsRefusalsCited` for why a case must not name a
refusal it does not have.

`qw(a b c)` prints as `abc` rather than `a b c`: in a print LIST the
three words are separate arguments and `$,` is unset, so nothing
separates them. Binding them to an array instead would print `a b c` --
but it would also emit `aassign`, `padav` and `join`, which are
02_variables' array machinery and unclaimed here. The lint caught that,
and this is the version that keeps `qw` in the tier that owns it."""),
])

check('01_literals', covered)
