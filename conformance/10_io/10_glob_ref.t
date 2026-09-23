#!perl
# `\*STDOUT` is a reference to a GLOB, which is its own reference type --
# not SCALAR, not CODE, not the filehandle it names.
#
# TIER 10 io
# INTRODUCES nothing of its own
# USES ref and srefgen from 08_references
#
# `09_glob_assign.t` writes to the symbol table; this reads from it. The
# pair is what makes the glob a first-class thing rather than a spelling
# of assignment: a glob can be referenced, passed and stored like any
# other value, and `ref` names it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print ref(\*STDOUT), "\n"'
#   GLOB
#
# Neither `SCALAR` nor `GLOB(0x...)`. `ref` returns the type NAME, and
# `GLOB` is a type this corpus never named before -- tier 08 covers
# `SCALAR`, `ARRAY`, `HASH` and `CODE` and stops there.
#
# THE BACKSLASH IS THE PART A PARSER CAN GET WRONG. `\*STDOUT` is a
# reference-to-glob; `*STDOUT` alone is the glob itself, and `\*` is not
# a compound operator -- it is tier 08's `\` applied to a term that
# happens to start with `*`. A lexer that read `\*` as one token, or that
# read `*STDOUT` as multiplication by a bareword, produces something
# `ref` would not call `GLOB`.
#
# `STDOUT` rather than a glob this file creates, because a bareword
# filehandle is the one glob guaranteed to exist without the file
# installing anything -- and it is this tier's own subject, which is why
# the file sits here rather than beside `ref` in tier 08.

--- source
print ref(\*STDOUT), "\n";

--- expect output
GLOB

--- expect parses

--- expect tokens
one word whose text is "ref"
