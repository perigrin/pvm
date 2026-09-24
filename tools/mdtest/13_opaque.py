"""Build tier 13's topic files from the extracted .t claims."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '13_opaque'
covered = set()

covered |= topic(TIER, 'heredocs.md', 'Heredocs', """
A heredoc's body begins on the NEXT line while the statement it sits in
continues on the SAME one, so the body token arrives after the semicolon
that ended the statement it belongs to.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

None of these three cases claims an op, because a heredoc emits none of
its own. Measured, `<<"EOT"` with an interpolation gives tier 01's
`multiconcat` and `<<'EOT'` and `<<~EOT` give a plain `const[PV]` --
indistinguishable from the same strings written inline. The token facts
are what say a heredoc was written at all.

All three refuse, all three name `trailing_tokens`, and that is one gap
rather than three. The LEXER produces the opener and the body as two
tokens in the right order with the right text -- every token fact here
passes. The parser has no grammar rule for a body token arriving where a
new statement should start. Naming the code in all three is what makes
"one gap" falsifiable: the day one of them begins refusing elsewhere,
these cases report the change instead of skipping on in silence.

This tier also settles a question GLOSSARY.md deferred: a heredoc body
token INCLUDES its terminator line. Perl cannot adjudicate -- there is no
token stream to inspect -- so the decision is made on what a consumer
needs, which is to know where the body ends.
""", [
    ('01_heredoc_interpolating.t', 'The interpolating heredoc', """
`<<"EOT"` interpolates. The declaration itself PARSES: measured,
`my $h = <<"EOT";` arrives as a declaration whose two terms are `$h` and
the opener, with no Unknown in it. The Unknown spans what comes NEXT --
the body token and the `print` after it -- and its code is
`trailing_tokens`, an expression that parsed with bytes remaining.

That is the heredoc's reordering stated as a parser failure. The output
proves the CONTENT and proves nothing about the spelling; a case
asserting only `parses` would report the same failure and leave the
boundary unlocated.
"""),
    ('02_heredoc_literal.t', 'The non-interpolating heredoc', """
`<<'EOT'` does not interpolate: `$name` in the body is five characters of
text, not a variable. Measured, this compiles to a plain
`const[PV "hello $name\\n"]` -- no multiconcat, because there is nothing
to interpolate.

This is the pair to the interpolating case, and the pair is the point.
The two spellings differ by one quote character and produce DIFFERENT
strings, which makes this the one heredoc property whose behaviour is
observable: the case above prints `hello world` and this prints
`hello $name` with `$name` in scope and holding `world`. Everything else
about a heredoc -- where it starts, where it ends, that it is a heredoc
at all -- is invisible below the token stream.

Refuses with the SAME CODE, and that is the claim rather than a copied
line: a refusal that changed cause between the two would mean the parser
sees the quoting, and it does not.
"""),
    ('03_heredoc_indented.t', 'The indentation-stripping heredoc', """
`<<~EOT` strips the terminator's indentation from every body line, so the
heredoc can be indented with the code around it.

The `~` belongs to the OPENER token, not to a separate operator: the
opener's text is `<<~EOT`. The stripping happens in the lexer, so by the
time an op exists the body is already `trimmed\\n` -- identical to what an
unindented `<<EOT` would have produced. The BODY TOKEN still carries the
leading spaces, which is what distinguishes this case from one that
simply did not indent.
"""),
])

covered |= topic(TIER, 'formats.md', 'Formats', """
A `format NAME =` declaration's body runs from the `=` to a line holding
a lone `.`, and it is the clearest case in the tier of a region the lexer
must delimit WITHOUT LEXING: `@<<<<<` inside a picture line is a
left-justified column, not an array sigil followed by two left shifts.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

`enterwrite` is claimed by this tier but it belongs to `write`. The
DECLARATION compiles to nothing: measured, a file holding only a format
and a `write` emits `enter`, `nextstate`, `enterwrite`, `leave`, and the
declaration accounts for none of them. The picture lines that are the
actual lexing problem are compiled into a format no op mentions.

Both cases refuse with `unimplemented_statement`, which is not the code
the heredocs carry. Those parse their statement and then meet a body
token they have no form for; this one never starts, because `format` is a
statement keyword the parser does not implement. Measured, the Unknown
spans `format STDOUT =` through the `write;` that follows it -- the
declaration and the statement after it swallowed together, which is what
an unimplemented keyword does to whatever it cannot find an end for. The
`format body` fact passes in both, so the picture lines are already
opaque to the lexer and what is missing is above it.
""", [
    ('06_format_write.t', 'A format declaration and the `write` that uses it', """
The picture here is a fixed line rather than a `@<<<` field, because a
field pads its column with spaces to a width the picture sets and the
runner compares bytes. A fixed line keeps this case about the LEXING and
leaves the field to the case below.

`write` prints the format's output and `print` prints after it, so the
two share one ordered stream on STDOUT.
"""),
    ('09_format_picture.t', 'A picture line that is not Perl', """
WHY A PLAIN PICTURE LINE CANNOT MAKE THIS CLAIM, which is the whole
reason this case exists beside the one above. `a fixed report line`
inside a format body is delimited identically by a lexer that treats the
body as opaque and by one that lexes it as Perl -- three Words either
way, and the region's extent is unchanged. Such a case measures
DELIMITING and says nothing about NOT LEXING, which is the other half of
this tier's thesis. `@<<<<<<< @>>>` separates them: opaque it is one
token, lexed as Perl it is `@` variables and shift operators.

Measured against our lexer, the body arrives as a single format body
token running from the first picture line through the lone `.`, with no
token inside it.

The argument line `$name,   $qty` IS ordinary Perl to perl, evaluated
when `write` runs. It is inside the opaque region all the same: the lexer
does not distinguish picture lines from argument lines, and it does not
need to.

The variables are `our` rather than `my` because a format body is
compiled in its own scope and cannot see a lexical declared beside it;
measured, `my $name` leaves the field empty.

The output has NO TRAILING WHITESPACE on either line, which the field
widths were chosen for: `widget` fills 6 of the 8 columns `@<<<<<<<` sets
and `7` is right-justified into `@>>>`, so the padding falls between the
two fields rather than after the last one.
"""),
])

covered |= topic(TIER, 'pod-and-data.md', 'Pod blocks and data sections', """
Two regions the lexer eats whole and the compiler never sees. A pod block
emits NOTHING -- not an op, not a `nextstate`; the line numbers in the
surrounding `nextstate`s move and that is the only trace. `__END__` and
`__DATA__` emit nothing either: reading the section emits `readline` and
`gv`, which describe the read and not the marker.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io. The dependency is here and it is specific:
the only way to observe that a data section exists is to read the `DATA`
filehandle it creates, and that is tier 10's `readline`.

Each construct appears twice, and the pairing is the tier's thesis split
in half. "Delimit without lexing the contents" is a claim about EXTENT
and a claim about CONTENTS, and a region whose contents are inert makes
only the first -- prose and two bare words are delimited the same way
whether or not they were lexed. The second case of each pair puts
hostile bytes inside the region so the two readings can be told apart.

The pod cases PASS; the data cases refuse with `not_a_term`, the third
distinct code in this tier. Both data cases' token facts pass, so the
lexer ends the program text at the marker and reads `<DATA>` as one
angle-bracket term; the refusal is the parser having no rule for a
trailing data section.

Both data sections are read in LIST context deliberately. The usual
spelling, `while (my $line = <DATA>)`, emits `defined` -- perl inserts it
around the readline so an empty last line does not end the loop -- and
`defined` belongs to no tier yet.

Every body here is written clear of any line beginning `--- `, which the
corpus format reads as a section marker with no awareness of Perl's
nesting. Tier 13 is the first tier able to hit that.
""", [
    ('07_pod_block.t', 'A pod block compiles to nothing', """
Measured, this source and the same source with the pod deleted emit the
IDENTICAL op sequence -- const, padsv_store, pushmark, padsv,
multiconcat, print -- differing only in the line numbers recorded on the
nextstates. So the pinned output cannot fail for the reason this case
exists: a lexer that read the pod as prose, as code, or as nothing at all
would print `1` either way.

That is why this case asserts the pod token. It is the whole assertion;
the rest is scaffolding proving the program still ran.

The block runs from a `=` in column 1 followed by an identifier through
the `=cut` line, and the TERMINATOR IS PART OF THE TOKEN, for the same
reason a heredoc body's is: every byte belongs to the block.
"""),
    ('10_pod_not_perl.t', 'A pod block holding text that is not a program', """
The block holds an unclosed `(`, a `$this` that is not a declared
variable and two stray semicolons. Measured, perl compiles and runs the
program without noticing any of it: the pod is eaten by the lexer before
the parser exists. Measured against our lexer, the same bytes arrive as
ONE pod token with nothing lexed inside.

`=head1` rather than `=pod`, because the block's opener is any `=` in
column 1 followed by an identifier and the case above already covers
`=pod`.

This case PASSES, in a tier where eight of the fifteen files refuse, and
that is the useful shape: the pod block is the construct this tier gets
entirely right, opaque contents and all, so the tier's refusals cannot be
read as "the lexer cannot handle opaque regions".
"""),
    ('08_data_section.t', 'A data section read back through `DATA`', """
`__DATA__` ends the program text and opens a filehandle onto what
follows. This is the one construct in the tier with an observable
consequence, and the consequence is not the marker's: `my @lines =
<DATA>` emits `gv`, `readline`, `padav` and `aassign` -- tiers 02 and 10
-- and would emit exactly those for a handle opened with `open`. Nothing
in the op stream says `__DATA__` created the handle.

Measured, the two statements before the marker parse cleanly and the
Unknown spans `__DATA__` to end of file. The glob case passes with the
same token category in the same position, so the gap is not the angle
brackets -- it is the trailing data section, which no earlier tier can
produce.

THE THIRD LINE OF OUTPUT IS EMPTY, AND IT IS THE CORPUS FORMAT'S OWN
BLANK LINE. A section's body runs to the next `--- ` marker including the
blank separator before it, and no other section notices -- trailing
whitespace in a Perl program is whitespace. Inside a data section it is
DATA. Pinned as measured rather than dodged, because every future
data-section case inherits it.

`__END__` is the same construct under another name: measured, both end
the program text and both open DATA in package main. The adjacency case
uses the other spelling.
"""),
    ('11_data_not_perl.t', 'A data section holding bytes that are not a program', """
The section holds `sub not_compiled { $x <=> }`, which is a syntax error
as Perl -- `<=>` with no right operand -- and `q{ unbalanced`, whose
quote never closes. Measured, perl compiles the program and prints both
lines back: neither is a program, and neither was ever offered to the
parser. Measured against our lexer, the whole section arrives as ONE data
section token with nothing lexed inside.

The unbalanced `q{` is the sharper case for the lexer/parser split: a
lexer that had read the section would still be inside an unterminated
quote at end of file.

The third line of output is the corpus format's blank separator again,
recorded here because every data-section case inherits it and a case that
dodged it would hide the property.
"""),
])

covered |= topic(TIER, 'quote-like-terms.md', 'Opaque regions inside one term', """
Four constructs whose opaque region fits inside a single term: the
command quote, the angle-bracket glob, the `pack`/`unpack` template, and
`qq`'s arbitrary delimiters. Everything above this heading suspends
ordinary lexing across LINES; these are delimited within an expression,
and all five cases PASS.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

Three of the five cases emit an op and none of the ops is the construct.
`qx{...}` emits `backtick` over a `const[PV]`, the same op backticks
emit, which is why GLOSSARY.md puts the two spellings in one category.
`pack` and `unpack` emit `pack` and `unpack`, and the CONSTRUCT is the
template, which emits nothing. `qq` emits nothing of its own at all:
measured, `qq{plain}` is tier 01's `const[PV "plain"]` and `qq{v=$x}` is
tier 01's `multiconcat` -- so this tier REACHES `multiconcat` rather than
introducing it, and the delimiter is observable only in the token stream.

Two of the five touch the world, so both were written for a result that
does not depend on the machine: `echo hi` writes the same three bytes on
every system with a POSIX shell, and `*.nonexistent-xyz` matches nothing
in any directory the runner might sit in. A case calling `date` or
globbing `*.t` would be unreproducible and must not be written.
""", [
    ('04_backtick_command.t', 'The command quote', """
`qx{...}` runs a shell command and returns its output, emitting
`backtick` over the command string. Measured, `echo hi` writes the three
bytes `hi\\n` including the trailing newline.

`qx` is the one construct in this tier whose own op is visible, and even
that op is shared with the backtick spelling.
"""),
    ('05_glob_angle.t', 'The angle-bracket glob', """
`<*.pattern>` in term position is a GLOB, not a pair of comparisons, and
it compiles to `glob` rather than to `readline`.

The token fact is what makes this a glob case rather than a count case.
`<*.nonexistent-xyz>` and `<DATA>` are the SAME token category -- our
lexer cannot tell a handle name from a pattern without knowing what the
name means, which is a parsing question -- so this asserts the
angle-bracket term and the data-section cases assert the other use of it.
perl separates them at the optree, where this is `glob` and that is
`readline`.

Note `scalar` emits no op: perl imposes scalar context on `@none` at
compile time and the array is simply evaluated there.
"""),
    ('12_pack_template.t', 'The `pack` template', """
`C3` is a count and a type code to `pack` and three ordinary characters
to the lexer, which reads it as a string literal and never looks inside.

THE FOLDING TRAP, and it is why the argument comes from the environment.
Measured, `print unpack("A3", pack("A3","abc"))` emits NO `pack` op at
all: the pack ran at compile time and left a `const[PV "abc"] s/FOLD` for
the `unpack` to read. A case written over constant arguments would claim
`pack` and emit none, so the lint would have nothing to check and a
compiler that had never heard of `pack` would pass it. `$ENV{X}` is unset
when the runner executes, so `//` yields the default as a RUNTIME value.

THE WRONG PARSE THIS RULES OUT: a lexer that reads `C3` -- or `A3`, or
`x2` -- as anything but three characters of string. A template is the one
argument shape in this tier that LOOKS like it wants lexing, and the
temptation to give it its own scanner is exactly the temptation a
format's picture lines present. The token fact says it is one string
literal, whole, with `C` and `3` never separated.

`C3` rather than `A3`, because `A` pads with SPACES: `pack("A5","abc")`
is `abc  `, and pinning that would put trailing whitespace in the pinned
output for the repository's hook to strip. `C` takes ordinals and
produces printable letters, and 65, 66, 67 are `A`, `B`, `C` in every
encoding perl builds against.

`length` would be the natural way to observe a packed string and NO TIER
CLAIMS IT, so the output is the packed bytes themselves.
"""),
    ('13_unpack_template.t', 'The `unpack` template', """
`unpack` reads the same opaque template `pack` writes, and returns a LIST
where `pack` returns one string. Measured, the two differ in the optree
beyond their names: `pack[t5] sK/2` against `unpack lK/2` -- LIST
context, `l` where pack has `s`. A compiler that implemented `pack` and
stopped would pass the case above and fail here.

`unpack` folds over constant arguments exactly as `pack` does, so the
source uses the same `$ENV{X} // default` idiom for the same reason.

THE WRONG PARSE THIS RULES OUT: `unpack` read as a scalar-returning call.
The pinned output is three numbers, which only exists if the result was a
list of three that `@c` absorbed -- so this is one of the few cases in a
tier of token facts where the BEHAVIOUR carries a real claim.

`C3` is the template above, deliberately: the round trip is the
assertion. That case packs 65, 66, 67 into `ABC` and this unpacks `ABC`
back into 65, 66, 67. Either alone could be satisfied by a compiler that
had the template wrong in a compensating way; the pair cannot.

`"@c"` interpolates the array with `$"` between elements, which is a
space by default -- the `gvsv[*"]` and `join` in the optree, tiers 02 and
03. `print @c` would print `656667` with no separator and would not show
the list had three elements.
"""),
    ('14_qq_delimiters.t', "`qq`'s arbitrary delimiters", """
WHAT IS PARSING-DISTINCTIVE ABOUT `qq` IS ARBITRARY DELIMITERS. `qq{}`,
`qq()`, `qq[]` and `qq!!` are one operator with four terminators, and a
lexer that hardcodes `"` or scans for a fixed closer gets three of them
wrong. GLOSSARY.md already records that bracketing delimiters NEST, and
this is where the corpus measures it.

THE WRONG PARSE THIS RULES OUT, stated as the token a broken lexer would
produce: `Quote("qq{a{b}")`. A lexer that scans forward to the FIRST `}`
ends the string there, leaving `c=$x}` as loose tokens and a stray
CloseBracket. That is the single most likely qq bug, and it is why the
nesting case is here rather than a fourth plain delimiter. Measured under
5.42.0, `print length(qq{a{b}c}), "\\n"` prints `5` -- `a`, `{`, `b`,
`}`, `c` -- so the inner `}` is not a terminator. The negative fact names
the wrong token directly, and `qq{a{b}` appears nowhere else here, so it
is a real claim rather than one defeated by its own source.

`$ENV{X} // 1` is the corpus idiom for a runtime value, and here it also
gives the interpolating case something to interpolate that is not a
constant. `$ENV{X}` lexes as Variable, Operator, Word, CloseBracket --
NOT as one token -- so it cannot satisfy or defeat either fact.

Measured against our lexer, the four spellings arrive as four Quote
tokens, the first of them spanning both inner braces.
"""),
])

covered |= topic(TIER, 'adjacency-13_opaque.md', 'Every opaque region, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another.

**Tier 13 opaque.** Introduces nothing of its own; it is the mixture that
is the subject. Depends on 10_io, and the dependency is live here: the
`__END__` section is read through the DATA filehandle, which is tier 10's
`readline`.

ADJACENCY MATTERS MORE HERE THAN ANYWHERE EARLIER, because every
construct in this tier CHANGES THE LEXER'S MODE and then has to hand it
back. A format body, a pod block and a heredoc body each suspend ordinary
lexing, consume lines by a rule of their own, and resume. A case holding
one of them tests that the mode was entered and left once. This one tests
that leaving one mode leaves the lexer able to enter the next.

THE ORDER IS NOT ARBITRARY. The format body sits FIRST, before any
statement, because its picture lines are the region least like Perl -- if
resuming after them is wrong, everything downstream is wrong for a reason
the single-construct cases cannot name. The heredoc body and the pod
block follow, so the heredoc's line-oriented terminator is immediately
followed by another line-oriented terminator of a different kind. The
data section is last because it has to be: it ends the program text, so
nothing can follow it.

A HAZARD THIS CASE HAD TO AVOID, and it is a real limitation of the
corpus format rather than a quirk of this body: `--- ` at the start of a
line is recognised as a SECTION MARKER, with no awareness of Perl's own
nesting. A heredoc body, a pod block or a data section whose content held
such a line would be cut in two. The bodies here are written clear of it.
Tier 13 is the first tier able to hit this, because it is the first with
multi-line opaque regions.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
THREE CODES, ONE NAMED. Measured, the three Unknowns carry
`unimplemented_statement` (the format declaration), `trailing_tokens`
(the heredoc body) and `not_a_term` (the data section) -- the same three
the format, heredoc and data-section cases carry alone, and nothing else.
`hasCode` is membership, so naming any one is a true claim; the middle
one is named because it is the only one whose SPAN says something the
single-construct cases cannot.

THAT SPAN IS THE NEGATIVE THIS CASE EXISTS FOR. Alone, the heredoc's
`trailing_tokens` Unknown covers the body and the statement after it.
Here it covers the body, the pod block, AND the statement after those --
measured, from `heredoc line` through `my @none = <*.nonexistent-xyz>;`.
The pod block sits inside an Unknown and contributes none of its own,
which is the adjacency claim stated as a measurement: the lexer left the
heredoc's opaque region and entered the pod block's cleanly, and the
parser's failure to resume did not compound into a second failure at the
join.

Three Unknowns for three distinct constructs and nothing more. `qx`, the
glob and the pod block pass here as they pass alone. So no Unknown
appears at a JOIN between two constructs, which is the negative worth
having: the lexer leaves each opaque region cleanly and the failures do
not compound.

Every token fact below passes. The lexer handles all six constructs
adjacent to one another; the refusals are the parser's, and the token
section is what locates them there.

The spelling is `__END__` rather than `__DATA__` to cover the other half
of the data-section construct -- measured, both end the program text and
both open DATA in package main.

`write` prints between the two `print`s rather than at the end, so format
output and print output share one ordered stream on STDOUT. The output
ends with a BLANK LINE, and it is the corpus format's own.
"""),
])

check(TIER, covered)
