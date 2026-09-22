# 13_opaque

Heredocs, `format`, `qx`, `<*>`, POD, `__END__`/`__DATA__`, `pack`
templates, `qq` delimiters.

## Why this tier sits here

Every earlier tier introduces a construct the compiler can be asked about.
This one introduces constructs the compiler never sees. A heredoc is a
string by the time ops exist; POD and `__END__` are eaten by the lexer and
leave nothing behind at all. Measured under 5.42.0, a program with a pod
block between two statements compiles to exactly the ops of the program
without it.

So this tier is not about what Perl means. It is about where one token
ends and the next begins, in the six places where the ordinary rule --
read until the delimiter -- does not apply. A heredoc's content begins on
the next LINE while its statement continues on the same one. A format body
is picture lines, not Perl, so `@<<<<<` is a column and not an array sigil
followed by two left shifts. A data section runs to end of file. Each is a
LEXING problem, and none of them is a compilation one.

Could this tier move earlier? Two answers, and only the second is the real
one.

The cheap answer is that its files need earlier machinery: `__DATA__` is
only observable by reading the handle it creates, which is tier 10's
`readline`, and a heredoc has to be bound to something and printed. True,
and not the point -- that argument would place the tier at 11, not at 13.

The real answer is that a lexing tier can only be POSED once the ordinary
language is established. A heredoc is hard because the lexer must suspend
the statement it is in the middle of, consume lines belonging to a
construct that started earlier, and resume. There is no statement to
suspend until the tiers that define statements have run. And the
dependency runs the other way too: a parser that cannot lex a heredoc
cannot be tested on ANY file containing one, so every construct these
files touch has to already be measured on its own, or a refusal here would
be undiagnosable. That is a late position by construction.

Tier 14 takes the harder half of the same problem: re-entering Perl INSIDE
a delimiter. Here the lexer must delimit without lexing the contents;
there it must lex the contents too.

## DEPENDS ON

    10_io

The dependency is `__DATA__`, and it is specific. The marker itself
compiles to nothing -- it is a lexer instruction, and the optree of a
program with a data section is identical to one without. The only way to
observe that the section exists is to read the `DATA` filehandle it
creates, and reading a filehandle is `readline`, which tier 10 introduces.
Without tier 10 this tier could write a `__DATA__` section and could not
demonstrate it had one.

Nothing else here needs 10. A heredoc needs a scalar and a print, which
are tiers 05 and 01. `qx` and `<*>` need an array and a print. The
declared prerequisite is the one that is load-bearing, not the largest one
the files happen to touch.

## INTRODUCES

    backtick enterwrite glob pack unpack

## Why those ops, and not the ones the source implies

Nine constructs, five ops. The gap is the tier.

**Five of the nine emit nothing that is theirs.**

- **Heredocs emit `const` or `multiconcat`** -- tier 01's, both of them.
  Measured, `<<"EOT"` with an interpolation gives `multiconcat`, `<<'EOT'`
  and `<<~EOT` give a plain `const[PV]`. Indistinguishable from the same
  strings written inline. The heredoc is invisible, and NOT claiming
  `const` here is the point rather than an omission.
- **POD emits nothing.** Not an op, not a `nextstate`, nothing. The line
  numbers in the surrounding `nextstate`s move, and that is the only trace.
- **`__END__` and `__DATA__` emit nothing.** Reading the section emits
  `readline` and `gv`, which belong to tiers 10 and 02 and describe the
  read, not the marker.
- **A `format` DECLARATION emits nothing.** `enterwrite` is claimed here,
  but it is emitted by `write`, not by `format`. The declaration --
  including the picture lines that are the actual lexing problem -- is
  compiled into a format that no op mentions.

- **`qq` emits NOTHING OF ITS OWN.** Measured, `qq{plain}` is tier 01's
  `const[PV "plain"]` and `qq{v=$x}` is tier 01's `multiconcat` -- the
  same two ops `"plain"` and `"v=$x"` give. So this tier REACHES
  `multiconcat` rather than introducing it, and `qq` adds nothing to the
  list above. The optree records what a string IS and never how it was
  spelled, which leaves the delimiter observable only in the token
  stream. `14_qq_delimiters.t` is the file that carries all of its weight
  in token facts because of it.

**Four emit something, and none of the ops is the construct.**

- **`qx{...}` emits `backtick`**, over a `const[PV "echo hi"]`. Backticks
  and `qx` produce the same op, which is why `GLOSSARY.md` puts them in
  one category.
- **`<*.nonexistent-xyz>` emits `glob`.** Note what this does NOT prove:
  `<DATA>` emits `readline` from the same syntax. The angle brackets are
  one token to a lexer and two different ops to perl, and the difference is
  whether the name inside is a filehandle -- a fact the lexer does not
  have. Asserting `glob` here asserts about the pattern case only.
- **`pack` and `unpack` emit `pack` and `unpack`, and the CONSTRUCT is
  the template, which emits nothing.** `"C3"` is a count and a type code
  to perl's packing engine and three characters of string to the lexer,
  which is why the pair belongs in this tier rather than beside the
  operators: the template is an opaque region spelled as a string
  literal, exactly as a format's picture lines are an opaque region
  spelled as lines of text. The ops are claimed here because this is
  where they are first emitted, not because either op is the subject.

  **Both fold, and the fold is the trap.** Measured,
  `print unpack("A3", pack("A3","abc"))` emits NO `pack` op at all: the
  pack ran at compile time and left a `const[PV "abc"] s/FOLD` for the
  `unpack` to read. A file written over constant arguments would claim
  `pack` in the block above and emit none, so the lint would have
  nothing to check. `12_pack_template.t` and `13_unpack_template.t` take
  their argument from `$ENV{X} // <default>`, which is unset when the
  runner executes and therefore yields the default as a RUNTIME value.

So the `--- expect tokens` sections carry this tier. That is why four
categories were added to `GLOSSARY.md` for it -- `readline operator`, `pod
block`, `data section`, `format body` -- and why the heredoc files assert
both the opener and the body rather than settling for the output. Where a
construct emits no op, a token fact is the only assertion left, and a file
asserting only `parses` and `output` would go green against a lexer that
read the pod block as six statements of arithmetic.

This tier also settles a question `GLOSSARY.md` deferred to it: whether a
heredoc body token includes its terminator line. It does. Perl cannot
adjudicate -- there is no token stream to inspect, which is this tier's
subject -- so the decision is made on what a consumer needs, which is to
know where the body ends.

## What the tier measured

Six of the fifteen files refuse as of 9750d03b, and the shape of the
refusals is the result. The three added since -- `12_pack_template.t`,
`13_unpack_template.t` and `14_qq_delimiters.t` -- all pass, which is the
expected shape: none of them asks the parser for a construct it lacks.
Their opaque regions are a string literal's contents, and a string
literal is tier 01's.

**Three codes, not one gap.** Every refusing file names the refusal site
it waits on rather than describing it, and the sites are distinct:
`unimplemented_statement` for the two format declarations, which are a
statement keyword the parser does not implement; `trailing_tokens` for the
three heredocs and the adjacency file, whose statement parses and which
then meet a body token arriving after the semicolon it belongs to; and
`not_a_term` for the two data sections, a token that cannot begin a term.
Three sites for six constructs means closing one will not close the
others, which uncited skips could not have said.

**Every token fact in the tier passes, including in the files that
refuse.** The lexer already delimits all six constructs: heredoc openers
and bodies in three spellings, format picture lines, pod blocks, data
sections, and both uses of the angle-bracket term. The Unknowns come from
the parser -- there is no grammar rule for `format NAME = BODY`, for a
statement containing a heredoc, or for a trailing data section.

That split is only visible because the files assert tokens. A file
asserting `parses` alone would report the same five refusals and say
nothing about where the boundary lies, and the natural reading of "the
heredoc test fails" is that heredoc lexing is broken. It is not.

**The adjacency file adds a negative.** Three Unknowns for three
constructs, none at a join between two, so leaving one opaque region does
not damage the lexer's handling of the next. Measured, its
`trailing_tokens` Unknown spans the heredoc body, the pod block, and the
statement after both -- so the pod block sits INSIDE an Unknown and
contributes none of its own, which is the negative stated as a
measurement rather than as prose.

**The tier's title has two halves and only one was originally asserted.**
"Delimit without lexing the contents" is a claim about extent AND a claim
about contents, and a region whose contents are inert makes only the
first. `a fixed report line` inside a format body is delimited identically
by a lexer that treats the body as opaque and by one that lexes it as
Perl -- three Words either way, and the region's extent is unchanged --
so it cannot tell the two readings apart. The same held for the pod
block's prose and the data section's `one`/`two`.

Three files close that: `09_format_picture.t` puts `@<<<<<<< @>>>` in a
format body, where the opaque reading is one picture line and the Perl
reading is array variables and shift operators;
`10_pod_not_perl.t` puts an unclosed `(` and stray semicolons in a pod
block; `11_data_not_perl.t` puts a `<=>` with no right operand and an
unterminated `q{` in a data section. Measured, each arrives as ONE token
with nothing lexed inside it, and perl compiles and runs all three. The
heredoc files already had hostile content -- `$name` in the body, and
leading indentation in the `<<~` case -- so they needed nothing.

**A format limitation this tier is the first able to reach.** A section
body runs to the next line starting `--- `, with no awareness of Perl's
nesting, and a section body also includes the blank separator line before
that marker. Neither matters for ordinary Perl -- trailing whitespace is
whitespace, and no earlier tier can put a line at column 1 inside a
multi-line opaque region. Both matter here:

- **The blank separator becomes DATA.** `08_data_section.t` pins three
  lines of output for a two-line data section, and the third is the corpus
  format's own blank line. Pinned as measured rather than dodged, because
  every future data-section file inherits it.
- **A `--- ` line inside a heredoc, pod block or data section cuts the file
  in two.** Measured against `ParseFile`: a heredoc body holding the line
  `--- not a marker` ends the source section there and the file is rejected
  with `unknown section "not a marker"`. So the limitation is real and the
  corpus cannot express such a program -- but it FAILS LOUDLY rather than
  silently mis-reading, because an unrecognised section name is already an
  error. A body whose stray line happened to spell a real section name
  (`--- expect output`, say) would be the silent case, and nothing
  currently prevents it. The files here are written clear of it.

## HARD MARKERS

    heredoc	<<
    format	\nformat·
    qx	qx
    glob-angle	<*

Four of the corpus's twelve `hardMarkers` place here, the largest share of any
tier, and that concentration is the point: this tier's subject is the
constructs whose CONTENTS the lexer must not read, which is the same
property that made them hard. The list came from
`internal/parse/easy_test.go`, which used it to carve an "easy" subset out
of T1; in a graded corpus each entry is a tier placement rather than a
filter, so that file was deleted and the placements moved to the tiers.

- `heredoc`, §2's sublexing row, 61 files at 11.5% clean. Three files back
  it -- `01_heredoc_interpolating.t`, `02_heredoc_literal.t`,
  `03_heredoc_indented.t` -- because the three differ in exactly the way
  the lexer must notice.
- `format`, §5, 2 files at 0.0%. `06_format_write.t` and
  `09_format_picture.t`. The probe carries a LEADING newline, written
  `\n`, because `format` must be at the start of a line to be a format
  declaration rather than a word.
- `qx`, a scanner row, 5 files at 20.0%. `04_backtick_command.t`.
- `glob-angle`, a scanner row, 2 files at 0.0%. `05_glob_angle.t`. Its
  probe `<*` is the one shape that cannot be a readline or a comparison.

The probe column is the SOURCE probe verbatim, `·` standing for a
significant trailing space and `\n` for a newline, because a Markdown
block cannot carry either safely -- an editor that strips trailing
whitespace would widen `format` from `\nformat·` to `\nformat`, and a
wider probe matches more, which is the direction that lets a tier pass
without the construct. `TestEveryHardMarkerPlaced` in
`internal/conformance` requires a file here that each probe finds.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
