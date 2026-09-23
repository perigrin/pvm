#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 03 context
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which makes them
# diagnosable: when `05_sort.t` refuses, the construct that refused is the
# only one present. That property is also why such files cannot reach an
# adjacency bug -- a parser that handles every construct alone and
# mis-handles a pair goes green over the pair.
#
# Here the constructs sit on consecutive statements: an array in scalar
# context, the same array interpolated, the comma operator in scalar
# context, `sort`, `reverse`, `localtime`, `wantarray`, `caller`, `map`
# and `grep`. They are
# adjacent as statements rather than nested, because nesting them would
# need an operator to join them and this tier is before the operator tier.
#
# EVERY DISCRIMINATING CONSTRUCT APPEARS IN BOTH CONTEXTS, which is what
# distinguishes this file from a list of the tier's constructs. This tier's
# subject is not a set of constructs, it is a set of PAIRS: the two halves
# of each emit the same op and differ only by the `s` versus `l` flag
# B::Concise prints. A file holding one half of each pair asks the parser
# for one answer where the construct has two, and a parser that collapsed
# the two contexts into a single answer would go green over it.
#
# So `@a` appears in scalar context and in list context, `reverse` twice,
# the empty-list count idiom twice, `localtime` twice, `caller` twice,
# and `map` and `grep` twice each. Measured, that is `padav s`/`padav l`,
# `reverse sK/1`/`reverse lK/1`, `aassign sKS`/`aassign lKPS`,
# `localtime s`/`localtime l`, `caller s`/`caller l`,
# `mapstart sK`/`mapstart lK` and `grepstart sK`/`grepstart lK` -- seven
# pairs, in one body, with the tier's remaining constructs in scope.
#
# This tier depends on 02_variables, and the adjacency with the earlier
# tier is the first line: `@a` is tier 02's array, and every statement
# after it puts that same array in a different context. Pairing this tier
# with the one it depends on is the array being re-used, not a separate
# statement about arrays.
#
# `sort` is a default string sort, so `(3,1,2)` comes back `1 2 3`. The
# reverse of that array in scalar context is `213` -- the elements joined
# with nothing between them, then the string reversed -- not `2 1 3`,
# which is what the list-context `@revd` on the next line gives.
#
# `$moved` is 3 and `@kept` is empty from the SAME idiom: in scalar
# context the empty-list assignment reports how many elements it moved,
# and in list context it yields what it assigned, which is nothing.
#
# `$fields` is 9 because `localtime` in list context returns nine
# elements; the count idiom is what puts it in list context while still
# yielding a scalar to print. `$stamp` is the scalar-context half, a
# formatted string, and it is COUNTED rather than printed -- pinning it
# would make this file depend on the clock and the timezone. `@seen` holds
# two elements whatever the time is, which is the shape claim `$stamp`
# can make and the string cannot.
#
# `$who` and `@frame` are `caller` in the two contexts, and at file
# scope the two do not merely format one answer differently -- they
# return different AMOUNTS of it. `$who` is undef and `@frame` is the
# EMPTY list, which is the `0` in this file's output after the `3`.
# That pair is `08_caller.t`'s subject; here it stands beside the rest.
#
# `map` and `grep` close the file, each in both contexts. `@mapped` is
# the three elements and `$mapcount` is the number 3; `@kept2` is the
# three elements and `$grepcount` is 3 as well, because every element of
# `(3, 1, 2)` is true and nothing is filtered out. That grep keeps
# everything HERE is not a weakening: `10_grep.t` is where grep filters,
# using an element bound to `$ENV{G}`, and this file's job is adjacency
# -- that both keywords appear beside the other constructs, each with
# `grepstart`/`mapstart` measured in both `s` and `l`. Bringing `$ENV`
# in here to make the filter bite would add an opacity this file does not
# otherwise need.
#
# Neither `map` nor `grep` FOLDS despite `@a` being wholly constant,
# which the comma operator two statements up does. Measured, a block is
# not constant-foldable however constant its input, so `mapstart` and
# `grepstart` both survive.
#
# `..`'S THREE READINGS close the file, and they are the tier's sharpest
# context pair because the three are not one operator in three contexts
# -- they are three operators wearing one spelling, chosen by context.
# `@ranged` is the numeric list range, `@lettered` is the string range
# whose successor is perl's magic increment (`az ba bb`, carrying), and
# `@window` is the scalar-context FLIP-FLOP.
#
# `1 2 3 4` is the flip-flop's claim and the reason its operands are two
# separate arrays. At indices 2 and 3 BOTH operands are false, and the
# elements are selected anyway, because the flip-flop turned on at index
# 1 and does not turn off until index 4. A parser reading `..` there as
# a list range, or as a boolean, selects only 1 and 4. `13_flipflop.t`
# records that a first draft using `$_ .. $_` was indistinguishable from
# plain truthiness and had to be replaced.
#
# Fewer ops appear than the constructs suggest. `padrange` fuses
# consecutive `my` declarations, and the comma operator in scalar context
# erases its own left operands -- so the op set is a UNION across the
# tier's files rather than a property of this one. It happens that this
# file does emit all seven, but that is a fact about this program, not a
# rule the format requires.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/03_context/00_adjacency.t
#   3 3 1 2 3 1 2 6 1 2 3 3 0 213 2 1 3 9 3 0 3 1 2 3 3 1 2 3 1 2 3 az ba bb 1 2 3 4

--- source
my @a = (3, 1, 2);
my $count = @a;
my @copy = @a;
my $interp = "@a";
my $last = (4, 5, 6);
my @sorted = sort @a;
my $moved = () = sort @a;
my @kept = (() = sort @a);
my $rev = reverse @a;
my @revd = reverse @a;
my $fields = () = localtime;
my $stamp = localtime;
my $want = wantarray;
my $who = caller;
my @frame = caller;
my @mapped = map { $_ } @a;
my $mapcount = map { $_ } @a;
my @kept2 = grep { $_ } @a;
my $grepcount = grep { $_ } @a;
my @ranged = (1 .. scalar(@a));
my @seed = ("az");
my @lettered = ($seed[0] .. "bb");
my @on = (0, 1, 0, 0, 0, 0);
my @off = (0, 0, 0, 0, 1, 0);
my @idx = (0, 1, 2, 3, 4, 5);
my @window = grep { $on[$_] .. $off[$_] } @idx;
my @seen = ($want, $stamp, $who);
print "$count @copy $interp $last @sorted $moved ", scalar(@kept),
    " $rev @revd $fields ", scalar(@seen), " ", scalar(@frame),
    " @mapped $mapcount @kept2 $grepcount",
    " @ranged @lettered @window\n";

--- expect output
3 3 1 2 3 1 2 6 1 2 3 3 0 213 2 1 3 9 3 0 3 1 2 3 3 1 2 3 1 2 3 az ba bb 1 2 3 4

--- expect parses
