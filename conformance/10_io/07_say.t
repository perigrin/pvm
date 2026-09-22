#!perl
# `say` is a FILEHANDLE OP, not a print variant, and it is FEATURE-GATED:
# the same bytes parse two different ways depending on a pragma written
# earlier in the file.
#
# TIER 10 io
# INTRODUCES say
# USES my, open, close, print, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# WITHOUT THE FEATURE THE SPELLING IS NOT A SYNTAX ERROR. It is a METHOD
# CALL, and that is the finding this file exists for. Measured, the same
# three tokens with and without the pragma:
#
#   $ perl -MO=Concise,-exec -e 'my $x=1; say $x'
#   6  <0> pushmark s
#   7  <0> padsv[$x:1,2] sM
#   8  <.> method_named[PV "say"]
#   9  <1> entersub[t2] vKRS/TARG
#
#   $ perl -MO=Concise,-exec -e 'use feature "say"; my $x=1; say $x'
#   6  <0> pushmark s
#   7  <0> padsv[$x:2,3] s
#   8  <@> say vK
#
# Identical source, TWO VALID PARSES. Without the feature perl reads
# `say $x` as indirect-object method dispatch -- call the method `say` on
# the invocant `$x` -- and it compiles clean. It fails at RUNTIME, not at
# compile time:
#
#   $ perl -e 'my $x=1; say $x'
#   Can't locate object method "say" via package "1" ...
#   exit code: 255
#
#   $ perl -c -e 'my $x=1; say $x'
#   -e syntax OK
#
# That is the same shape as `11_oo/10_isa_infix.t`'s finding, met from
# the other side. There the featureless spelling is a SYNTAX ERROR, so a
# parser that got the gate wrong at least fails loudly. Here it is a
# clean compile of the WRONG PROGRAM, which is strictly worse: a lexer
# that ignores the pragma produces a parse tree nothing refutes until the
# program runs.
#
# WHY THE FEATURELESS FORM IS NOT A SOURCE SECTION HERE, and this is a
# tier constraint rather than a choice. It emits `method_named` and
# `entersub`. `entersub` is tier 07's and 07 is earlier, so that one is a
# legal use -- but `method_named` is tier 11's, and 11 is LATER than 10.
# The dependency lint forbids a file from emitting an op no earlier tier
# claims, so the featureless parse cannot be run from this tier at all.
# It is recorded above as a measurement and nowhere asserted, which is
# the honest shape: the corpus can state the fact without being able to
# pin it here. A file that pins it belongs in tier 11, beside the `isa`
# file whose finding it mirrors.
#
# WHY `say` IS TIER 10 AND NOT TIER 01 BESIDE `print`. It takes a
# filehandle through the SAME `rv2gv` every other handle in this tier
# goes through:
#
#   say   $out $line   padsv[$out] rv2gv sKR/1 padsv[$line] say   vKS
#   print $out "x"     padsv[$out] rv2gv sKR/1 const        print vKS
#
# The handle machinery is identical; only the op name and the appended
# newline differ. `say` is this tier's construct because the thing it
# does that `print` does not -- add a record separator -- is a property
# of writing to a HANDLE, and the thing it shares with `print` is this
# tier's `rv2gv`.
#
# THE OUTPUT IS THE FALSIFYING HALF AND THE `print` IS WHAT MAKES IT ONE.
# Both statements write to the same in-memory handle, in the same
# argument position, one line apart. Measured:
#
#   $ perl conformance/10_io/07_say.t
#   [said
#   printed]
#
# The newline between `said` and `printed` is the whole assertion: it is
# in no string in this source, so it can only have come from `say`. A
# parser that compiled `say` as `print` emits `[saidprinted]`; one that
# appended a newline to both emits a third line. The brackets are what
# make the trailing byte visible -- without them `[said\nprinted]` and
# `[said\nprinted\n]` would differ only in bytes the format cannot show.
#
# `$ENV{X} // "said"` keeps the argument a RUNTIME value, so nothing
# folds: a constant argument to `say` would let an optimiser reach the
# same output by a different route. `X` is unset when the runner
# executes the file.
#
# `use feature "say"` rather than `use v5.36`, which would also enable
# it. Measured, the version bundle sets `strict` as well -- the ops carry
# `/STRICT` and the nextstate flags read `fea=6` against `fea=15` -- so
# it turns on more than this file is about. The narrow pragma names the
# gate that decides the parse and nothing else.
#
# THE TOKEN FACTS ARE WHAT CATCH THE REWRITE, because behaviour cannot.
# A parser that read the featureless spelling -- rewriting `say $out
# $line` into `$out->say($line)`, which is what perl itself does without
# the pragma -- produces an ARROW, and this source has none. That
# negative is checkable because `->` appears nowhere else in the file.
#
# THE PARENTHESIS NEGATIVE IS NOT AVAILABLE HERE and the reason is worth
# recording, because it is the obvious second guard and it is FALSE.
# `11_oo/10_isa_infix.t` can assert `no operator whose text is "("`
# because nothing in its source is parenthesised. This file cannot:
# `open(...)` and `close($out)` are both this tier's constructs in their
# ordinary spelling, so `(` is present three times and asserting its
# absence would be a claim the source refutes. Rewriting the opens to
# their paren-free form to buy the guard would make the file's own
# subject unidiomatic for the sake of a second assertion the arrow
# already makes.
#
# The positive count is the other half: `one word whose text is "say"`
# says the construct survived the lex as ONE word rather than being split
# or swallowed. It counts ONE and not two even though `say` appears twice
# in the source, because the other occurrence is inside `use feature
# "say"` where it is a STRING and not a word -- the glossary's `word`
# category is a bare identifier, and a quoted one is a string literal.
# That distinction is exactly what this file's subject turns on, so the
# count asserting it is not an accident of the vocabulary.
#
# `expect output` is written before `expect parses` rather than last: the
# blank line after it carries the output's trailing newline, and a blank
# line at END of file is what `end-of-file-fixer` strips.

--- source
use feature "say";
my $line = $ENV{X} // "said";
open(my $out, ">", \my $buf);
say $out $line;
print $out "printed";
close($out);
print "[$buf]\n";

--- expect output
[said
printed]

--- expect parses

--- expect tokens
one word whose text is "say"
no operator whose text is "->"
