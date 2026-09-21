#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another -- and paired with 06_control, the tier this one depends on.
#
# TIER 07 subroutines
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `06_signature.t` refuses, the construct that refused is
# the only one present. That same property is why a corpus of such files
# cannot reach an ADJACENCY bug -- a parser that handles every construct
# alone and mis-handles a pair goes green over the pair.
#
# Here the constructs are a named sub, a signature with a default, `@_`
# read positionally, `@_` WRITTEN through, three call forms, an anonymous
# sub, and `return` both early and trailing. The aliasing write is the one
# construct here whose effect is visible from outside the sub at all:
# `bump($seen)` returns nothing anyone looks at, and `$seen` is 1 in the
# printed line only because `$_[0]++` reached the caller's variable. Put
# next to a signatured sub on purpose: `pick` binds its arguments by
# signature and `bump` by `@_`, so both of the tier's argument protocols
# are in one body, which is the pair a parser handling each alone can
# still get wrong.
#
# MEASURED perl 5.42.0, on why `bump` has no signature: `@_` inside a
# signatured sub is populated, but reading it warns --
#
#   $ perl -e 'use v5.36; sub f ($a) { scalar(@_) } f(9)'
#   Use of @_ in scalar with signatured subroutine is experimental
#
# -- and a warning on stderr is noise this corpus has no section to pin.
# The two protocols therefore sit in two subs rather than one.
#
# The pairing with 06_control is not decoration: the `for` loop with
# `next` inside the body is the earlier tier's construct sitting directly
# against this tier's `return`, which is the pair the DEPENDS ON line
# names. A `return` reached from inside a loop has to
# unwind the loop as well as the sub, and nothing else in the corpus puts
# those two exits next to each other.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/07_subroutines/00_adjacency.t
#   small big-2 small/tiny 1
#
# This file is also the tier's sharpest statement of what the op lint can
# and cannot see. The source below contains a signature, a parameter
# default, a loop, a `next`, three `return`s, an aliasing write through
# `$_[0]` and an ampersand call, and the main program's op stream
# contains exactly two ops this tier introduces -- `anoncode` and
# `entersub` -- and not one op from 06_control either, because the loop
# is inside the sub too. Everything else compiled into a CV that
# `-MO=Concise,-exec` does not print. The tier README's last section is
# where that is written down; this file is the measurement behind it.
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it is what carries the output's own trailing newline,
# and a blank line at END of file is what `end-of-file-fixer` strips.

--- source
use v5.36;
sub pick ($n, $label = "small") {
    return $label if $n < 10;
    for my $i (1 .. 2) {
        next if $i == 1;
        return "big-" . $i;
    }
    return "none";
}
sub bump { $_[0]++ }
my $anon = sub { pick($_[0]) . "/" . &pick($_[1], "tiny") };
my $seen = 0;
bump($seen);
print pick(1), " ", pick(50), " ", $anon->(2, 3), " ", $seen, "\n";

--- expect output
small big-2 small/tiny 1

--- expect parses
