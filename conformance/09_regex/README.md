# 09_regex

Match, substitution, binding, and alternative delimiters.

## Why this tier sits here

A regex needs a string to match against and a variable to hold it. That is
the whole requirement. `$s =~ /abc/` is a literal (tier 01) bound to a pad
slot (tier 02) and then matched -- nothing in it reaches for an operator, a
block, a subroutine or a reference.

Check 2 asks whether this tier could move earlier. It could, and by a long
way. Nothing between tier 03 and tier 08 is needed: the binding operator is
part of the match op's own syntax rather than a general operator, and a
substitution mutates its target in place without a block, a call or a
dereference. Measured, `my $s = "abc"; $s =~ s/a/z/; print $s;` compiles to
tier 01's ops plus one `subst`. The tier could sit at 03 without breaking
the dependency check.

It sits at 09 because that is one topological sort of a partial order, not
because eight tiers precede it. What tier 10 needs from it is the
observation that a match is the first construct whose OPERAND IS NOT PERL --
the delimited pattern is a second language embedded in the source, and the
lexer must delimit it without lexing it. Tiers 13 and 14 are built entirely
on that problem; this tier is where it first appears in an easy form, with
one pair of delimiters and no re-entry.

## What this tier claims, and what it does not

**Its claims are about DELIMITERS and OPACITY, not about pattern
syntax.** A pattern is one token -- `m{a{b}c}` is a single quote-like
operator, nested braces and all -- so every fact this tier can state is
about the boundary of that token rather than its contents. The files
test delimiters exhaustively, because choosing the delimiter and finding
its match is the lexer's whole job here, and they assert opacity through
negatives: the delimiter did not leak out, and the contents were not
lexed as code.

**Pattern internals are out of scope by decision, not by oversight.**
`conformance/README.md` records the measurement and the reasoning. A
parser that treats a pattern body as an opaque string passes this tier,
and is correct to.

**`14_recursive` is the deliberate exception.** `(?{ })` puts a Perl
statement inside a pattern and `s///e` puts one in a replacement, so
there the contents are code and a lexer must re-enter. That tier carries
those cases; this one stops at the delimiter.

## DEPENDS ON

    01_literals

The spec names `09_regex` as one of the two tiers whose DEPENDS ON is not
N-1, and it is right: nothing here needs `08_references`. A pattern is not a
reference, a captured group is not a reference, and `qr//` -- which does
produce a reference-ish scalar -- is claimed here only in its non-recursive
form (see below). Pairing this tier with `08_references` in an adjacency
file would assert nothing, which is exactly the failure mode the spec warns
about.

`01_literals` rather than `nothing`, because a match needs something to match
against and the smallest such thing is a string literal. `02_variables` is
the near miss: every file here binds through a pad variable, so `padsv` is
present throughout. But `my $s = "abc"` is already a tier 01 fixture -- tier
01's own files declare and print variables -- so the honest floor is the
literal, and the pad slot comes with it. Declaring 02 would assert a
dependency on element access, `delete`/`exists` and `$::`, none of which
appear here.

## INTRODUCES

    match pos qr regcomp split subst trans

## Why those ops, and not the ones the source implies

Measured with `perl -MO=Concise,-exec` under 5.42.0, taking the union across
the forms this tier covers and subtracting what earlier tiers claim. Six
things that reading the source would not tell you:

- **`=~` emits no op of its own.** There is no `bind` in the stream. The
  binding operator is absorbed into the `match` or `subst` op, which carries
  its target in its own flags -- `match()[$s:1,7]`. A file testing `=~` and
  a file testing a bare `/abc/` against `$_` differ in their operand, not in
  their op.

- **`!~` is `not` wrapped around the same op.** `$s !~ /zzz/` is `match`
  followed by `not`; `$s !~ s/a/z/` is `subst` followed by `not`. `not` is
  tier 04's op, claimed there with the logical operators, and its
  appearance here is the negated binding reusing it rather than a new
  construct. The measurement is still this tier's: negation of a match is
  not a match op with a flag, it is a second op stacked on top.

- **Alternative delimiters erase themselves entirely.** `m{abc}` and
  `/abc/` emit byte-identical op streams, as do `s{a}{z}` and `s/a/z/` and
  `m!a!` and `m#abc#`. The optree cannot distinguish them, so delimiter
  coverage in this tier is a TOKEN claim, not an op claim -- the same
  argument `conformance/README.md` makes for `5e-1`. Every delimiter file
  here must carry an `expect tokens` section or it measures nothing the
  plain form does not already measure.

  `m#abc#` is the sharpest of them and gets a file of its own. `#` is
  Perl's COMMENT CHARACTER, so a lexer reading left to right has already
  decided to discard the rest of the line before it can know it was
  wrong -- and because the op stream is identical to `/abc/`'s, neither
  the optree nor the printed output would report the mistake.

- **The nesting rule needs a pattern that nests, and the forms do not
  show it.** A delimiter FORM -- `m{}`, `s{}{}` -- says a bracketing
  delimiter was used; it does not say the brackets NEST. Every bracketing
  pattern the tier shipped with had a body containing no bracket, so all
  of them are satisfied by a lexer that stops at the first `}`.
  `09_nesting_delimiters.t` is the file that tells them apart: measured,
  `m{a{b}c}` matches the five-character string `a{b}c`, and `s{a{b}c}{ok}`
  turns `xa{b}cy` into `xoky`, which pins the pattern's full extent on
  both sides. The non-bracketing counterpart cannot be written at all --
  `"a!b!c" =~ m!a!b!c!` is a syntax error, not a different match -- which
  is the asymmetry `GLOSSARY.md` records.

- **`regcomp` appears only when the pattern interpolates.** `/abc/` compiles
  its pattern once, at compile time, and emits `match` alone. `/$p/` must
  build the pattern at runtime and emits `padsv`, then `regcomp`, then
  `match`. This is the third lesson of the corpus in its sharpest form: the
  two files differ in one character of source and in a whole op. Both belong
  in the tier, and the tier's set is their union.

- **`qr//` is claimed here, not by tier 14.** Tier 14 owns `qr//` WITH
  EMBEDDED CODE, which is a re-entry problem. Plain `qr/abc/` emits a single
  `qr` op and nothing else -- no `regcomp`, no `match`, no block, no
  recursion -- so it is closer to a literal than to anything tier 14 does.
  Claiming it there would make tier 14 depend on this tier for the
  non-recursive half of one construct. It stays here, and tier 14 claims
  only what embedding adds.

- **No `gvsv`, though `$1` needs one.** `print $1` after a match emits
  `gvsv[*1]`, because capture variables are package globals. That is tier
  02's op -- it is what `$::` and any other global access emits -- and
  capture ACCESS is a variable construct that happens to be populated by a
  regex. This tier's files may read `$1` as a fixture; the op belongs to 02.

- **`split` takes a PATTERN where an expression would go, and the optree
  cannot see which was written.** This was scoped out once and is now in,
  and the reason it came back is that it is the tier's own argument in its
  strongest form. Measured under 5.42.0, `split / /, $s` and `split " ",
  $s` emit byte-identical ops -- `split(/" "/ => @p:2,3)` for both, down
  to the pattern text perl prints inside the op -- while printing
  DIFFERENT ANSWERS, because a lone space as a STRING is perl's awk
  special case and the same space as a PATTERN is not. Everywhere else in
  this tier the erasure hides a distinction nothing observable depends on;
  here it hides one the output reports. `10_split_pattern.t` pins both
  spellings side by side.

- **`tr///` wears `s///`'s syntax and is not a regex at all.** It is the
  only quote-like besides `s///` taking two delimited regions, so a lexer
  needs `tr` (and its synonym `y`) in whatever table tells it a second
  region follows -- the operator NAME is the only signal, since nothing
  about the first region announces a second. But neither region is a
  pattern: measured, `tr/./X/` on `"a.c"` gives `aXc` where `s/./X/` gives
  `X.c`, because `tr`'s `.` is the character and `s`'s is the
  metacharacter. A parser that desugared one into the other compiles,
  runs, and prints the wrong string. The op is `trans` and it stands alone
  -- no `match`, no `regcomp` -- which is the optree agreeing that `tr` is
  not a regex construct while the lexer cannot tell.

- **`pos` is a named operator in LVALUE position.** `pos($s) = 0` puts the
  `pos` op under `sassign`'s left arm, which almost nothing in Perl does,
  and a parser modelling `pos` as an ordinary named unary either refuses
  the assignment or discards it silently. The discard is the dangerous
  case and `12_pos.t` makes the output report it: with the reset two
  successive `//g` matches both land at 2, without it the second lands at
  5. The op is the same in both positions -- rvalue and lvalue differ in
  perl's printed flags, not in the op -- so the op claim is earned by
  either spelling and the lvalue claim is behavioural.

`padsv`, `padsv_store`, `const`, `print` and `pushmark` appear throughout and
are tier 01's, already claimed.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
