#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 01 literals
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `06_leading_decimal.t` refuses, the construct that
# refused is the only one present. That same property is why a corpus of
# such files cannot reach an ADJACENCY bug -- a parser that handles every
# construct alone and mis-handles a pair goes green over the pair.
#
# MEASURED perl 5.42.0, and this is not hypothetical:
#
#   class Foo { ADJUST { 1 } }                   0 Unknowns
#   class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both
#
# `ADJUST` alone parses; `ADJUST` followed by anything does not. No
# one-construct-per-file corpus can ever see that, because every file is
# one construct by definition.
#
# So each tier carries one file where its constructs sit next to each
# other. Here that is EVERY numeric spelling the tier's construct files
# introduce -- binary, decimal, hexadecimal, leading point, negative,
# octal by leading zero, octal by prefix, signed exponent, trailing point,
# underscore separators, and both v-string forms -- plus an integer, a
# single-quoted string, an interpolating string and a `qw` list.
#
# `TestTierLiteralsAdjacency` reads the tier's construct files for the
# literal each binds and requires the spelling to appear here, so this
# list cannot fall behind the tier by a file. It cannot check the
# ADJACENCY itself; see the note in that test about `padrange` absorbing
# `pushmark`, which is why more adjacent constructs emit FEWER ops.
#
# This tier depends on nothing, so there is no earlier tier to pair with;
# the adjacency is entirely within 01.
#
# This file refused from 7711154e until issue
# 01a0c13f-97f5-7f98-b32d-07245ec6ddfe, and it refused for a BORROWED
# reason: the `.5` on line 4 was `06_leading_decimal.t`'s gap reaching
# here, not a second bug. The note recorded at the time said the adjacency
# file cannot pass while any construct it holds refuses, and that
# composing only the working constructs would make it green and make it
# stop covering the tier. Fixing the lexer cleared both files in one
# change, which is that prediction coming true.
#
# The tier's other two refusals -- the exponent split and the v-strings --
# are LEXICAL and produce no Unknown at all, so they leave no code here to
# name. See `TestTierLiteralsRefusalsCited` for why a file must not name a
# refusal it does not have.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/01_literals/00_adjacency.t
#   10 0.5 255 0.5 -1 255 255 0.5 1 4294967296 ABC ABC 42 plain 42-0.5 abc one a
#   b
#
# `qw(a b c)` prints as `abc` rather than `a b c`: in a print LIST the
# three words are separate arguments and $, is unset, so nothing separates
# them. Binding them to an array instead would print `a b c` -- but it
# would also emit aassign, padav and join, which are tier 02's array
# machinery and unclaimed here. The lint caught that, and this is the
# version that keeps `qw` in the tier that owns it.
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it is what carries the output's own trailing newline,
# and a blank line at END of file is what `end-of-file-fixer` strips. A
# section after it puts the blank line in the middle of the file, where
# the hook has no quarrel with it.

--- source
my $bin = 0b1010;
my $dec = 0.5;
my $hex = 0xff;
my $lead = .5;
my $neg = -1;
my $oct = 0377;
my $octp = 0o377;
my $exp = 5e-1;
my $trail = 1.;
my $usep = 4_294_967_296;
my $vb = 65.66.67;
my $vv = v65.66.67;
my $int = 42;
my $sq = 'plain';
my $dq = "$int-$dec";
my $esc = "a\nb";
my $qop = q(one);
print "$bin $dec $hex $lead $neg $oct $octp $exp $trail $usep $vb $vv $int $sq $dq ", qw(a b c), " $qop ", $esc, "\n";

--- expect output
10 0.5 255 0.5 -1 255 255 0.5 1 4294967296 ABC ABC 42 plain 42-0.5 abc one a
b

--- expect parses
