#!perl
# `-e` is a NAMED UNARY whose name is punctuation, and the whole family
# was absent from the corpus.
#
# TIER 04 operators
# INTRODUCES the file test operator
# USES nothing from a later tier
# STATUS refuses as of this file. Issue 01a0cf64-8436-7f82-bcb5-587f0eba266f.
#
# perlop's level 10, alongside `defined` and `ref`. `-e -d -f -s -z -r -w
# -x -M -A -C` and the rest are one operator family with a one-character
# name, and not one of them appeared anywhere in the corpus.
#
# WHAT A PARSER GETS WRONG IS THE MINUS. `-e $f` is a file test; `-$e`
# is arithmetic negation; `-bareword` is the string `"-bareword"`, which
# tier 01's `07_negative.t` already measures as two tokens. So a single
# `-` has at least three readings and the one that applies is decided by
# what follows it -- a letter that names a test, a sigil, or a word.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print -e "/etc/hostname" ? "yes" : "no"'
#   yes
#
# The op is `ftis`, which is one op for the whole `-e`/`-f`/`-r` group --
# perl distinguishes them by a private flag rather than by op name, so
# the op stream cannot tell `-e` from `-r` and only the source can.
#
# `/etc/hostname` is a path that exists on this machine and on the CI
# image, and the file pins `yes` rather than a path. A test against a
# path the file creates would need tier 10's `open`; a test against a
# path guaranteed absent would pin `no` and prove the same thing, but
# `yes` is the answer that fails if the operator is not implemented at
# all -- an unparsed `-e` yields undef, which is false.
#
# THE TOKEN FACT IS A NEGATIVE AND IT CURRENTLY FAILS, which is the
# point. Measured against our own lexer, `-e $f` arrives as
# `Operator("-") Word("e")` -- two tokens, the minus split from the
# name. Perl lexes it as one file-test operator; we do not.
#
# The refusal names NO CODE, because the failure is LEXICAL and the
# parser produces no Unknown at all -- it reads the split tokens as a
# negation of a bareword and builds a tree. That is the same shape tier
# 01 records for its exponent and v-string files, and the reason the
# token layer exists: three other checks all pass over this file.
#
# So the fact asserts what a correct lexer produces: no bare `-`
# operator, because the whole `-e` is the name. A lexer that split it
# leaves a `-` that could be read as negation or as the start of a
# bareword, which is the three-way ambiguity this file is about.
#
# The ternary is tier 06's `cond_expr` and out of budget here, so the
# file uses the boolean directly: perl's true prints as `1` and its
# false as the empty string, which `29_logical_not.t` measures.

--- source
my $f = $ENV{X} // "/etc/hostname";
print "[", (-e $f), "]\n";

--- expect output
[1]

--- expect parses

--- expect tokens
no operator whose text is "-"
