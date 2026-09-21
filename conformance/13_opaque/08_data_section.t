#!perl
# `__DATA__` ends the program text and opens a filehandle onto what follows.
# The marker compiles to nothing; only the read emits ops.
#
# TIER 13 opaque
# INTRODUCES the data section
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file.
#
# One Unknown. Both token facts pass, so the lexer ends the program text at
# `__DATA__` and reads `<DATA>` as one angle-bracket term; the refusal is
# the parser's. Note 05_glob_angle.t PASSES with the same token category in
# the same position, so the gap is not the angle brackets -- it is the
# trailing data section, which no earlier tier can produce.
#
# This is the one construct in the tier with an observable consequence, and
# the consequence is not the marker's. `my @lines = <DATA>` emits `gv`,
# `readline`, `padav` and `aassign` -- tiers 02 and 10 -- and would emit
# exactly those for a handle opened with `open`. Nothing in the op stream
# says `__DATA__` was what created the handle.
#
# The section is read in LIST context deliberately. The usual spelling,
# `while (my $line = <DATA>)`, emits `defined` -- perl inserts it around the
# readline so an empty last line does not end the loop -- and `defined`
# belongs to no tier yet. Slurping avoids introducing an op that has
# nothing to do with this tier's subject. That the two spellings differ at
# all is the kind of thing the op lint is for.
#
# `__END__` is the same construct under another name: measured, both end
# the program text and both open DATA in package main. One file covers
# both; the adjacency file uses the other spelling.
#
# THE THIRD LINE OF OUTPUT IS EMPTY, AND IT IS THE CORPUS FORMAT'S OWN
# BLANK LINE. A section's body runs to the next `--- ` marker, including
# the blank separator line before it, and no other section notices --
# trailing whitespace in a Perl program is whitespace. Inside a data
# section it is DATA. So the `--- source` here holds `one\ntwo\n\n` and
# perl prints three lines.
#
# Pinned as measured rather than worked around. A file that shifted the
# data to dodge the blank line would hide a property every future data
# section file inherits: a data section is the one construct whose content
# extends to the section boundary, so the boundary is part of what it
# holds. Writing it down is cheaper than each author rediscovering it as a
# CORPUS BUG report.
#
# MEASURED perl 5.42.0, against a file whose data section ends with a blank
# line, which is what the runner writes:
#
#   $ perl data.pl | od -c
#   0000000   o   n   e  \n   t   w   o  \n  \n
#   0000011

--- source
my @lines = <DATA>;
print @lines;
__DATA__
one
two

--- expect parses

--- expect output
one
two


--- expect tokens
one readline operator whose text is "<DATA>"
one data section whose text is "__DATA__\none\ntwo\n\n"
