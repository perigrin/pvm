#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another -- and paired with 06_control, the tier this one depends on.
#
# TIER 07 subroutines
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of dc1bea2c. Issue 01a0c432-fbd5. Refusal trailing_tokens.
#
# THIS FILE STARTED REFUSING WHEN THE CALL FORMS JOINED IT, which is the
# corpus working rather than a regression. It parsed clean while it held
# only parenthesised calls; the two parenless call sites below are what
# our parser declines, at the same `trailing_tokens` site as
# `08_parenless_extent.t` and `09_prototype_extent.t` -- it reads the
# callee as a complete term and then finds a number with no operator
# between them. Measured at dc1bea2c: three Unknown nodes, all
# `trailing_tokens`, where each extent file alone produces one.
#
# The marker comes off when the parenless form lands, and the two extent
# files' markers come off with it.
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
# THE CALL FORMS, added by the call-form slice (issue 01a0c432-fbd5).
# The slice's own files are one form each; this is where they sit beside
# one another, which is the pair a parser handling each alone can still
# get wrong. Present below: `answer()` parenthesised, bare `answer`,
# `&answer` with the ampersand, `$anon->(2, 3)` through a code
# reference, and `f 1, 2` parenless with a greedy extent. Five forms,
# all compiling to `entersub`, which is exactly why they need separate
# source rather than separate op claims.
#
# THE PROTOTYPE NEEDS A BLOCK, and the reason is the sharpest single
# measurement in this file. `sub g ($)` is a PROTOTYPE only where the
# signatures feature is OFF. This file says `use v5.36`, which turns
# signatures on, and under it perl reads the same three characters as a
# SIGNATURE and enforces arity instead -- measured 5.42.0:
#
#   $ perl -e 'use v5.36; sub g ($) { "g" } print g 1, 2;'
#   Too many arguments for subroutine 'main::g' (got 2; expected 1)
#
#   $ perl -e 'sub g ($) { "g[$_[0]]" } print g 1, 2; print "\n"'
#   g[1]2
#
# The same three characters, two different features, and only the
# feature state decides which. The `no feature "signatures"` /
# `use feature "signatures"` pair around `sub g` is what lets both
# readings live in one file: `pick` is declared above it and keeps its
# signature, `g` is declared inside it and gets a prototype.
#
# A BLOCK would have been the tidier spelling and is measurably wrong
# here. `{ no feature "signatures"; sub g ($) {...} }` compiles the bare
# braces as a loop -- `enterloop`, `stub`, `leaveloop` -- and `stub` is
# an op 11_oo introduces, so the dependency lint correctly refuses this
# file for reaching four tiers forward to declare a sub. The file-scope
# toggle emits no ops at all.
#
# `@greedy` and `@cut` are the BUILTIN extent question (issue 01a0c730),
# standing beside the user-sub one on the last two lines. `warn "a", "b"`
# is greedy and yields ONE value; `warn("a"), "b"` is cut by the paren
# and yields TWO -- the `12` on this file's third line. The two routes to
# the same question are what makes them worth having adjacent: a user
# sub's extent is decided by its DECLARATION and a builtin's by a PAREN
# at the call site, and a parser can implement either and not the other.
# `12_builtin_extent.t` is the one-construct half and records why `die`,
# which asks the identical question, is not here: it terminates, and a
# file that dies has no output to pin.
#
# `local $SIG{__WARN__} = sub { }` is what makes the pair observable at
# all. `warn` writes to STDERR, which `--- expect output` does not read;
# an installed handler is called INSTEAD of that write, so the text goes
# nowhere and `warn`'s RETURN VALUE is the only thing left to count.
#
# MEASURED perl 5.42.0, the call-form lines:
#
#   42 42 42
#   12
#   f[1-2]
#   g[1]2
#
# THOSE LAST TWO LINES ARE THE ADJACENCY THAT MATTERS. `print f 1, 2`
# and `print g 1, 2` are the same call-site shape, and they print
# different things: f is greedy and takes both arguments, while g's
# prototype cuts the extent to one and the `2` falls through to the
# enclosing `print`. A parser that handled each file alone and got the
# pair wrong would go green over two separate corpus files and fail
# here, which is the whole reason the adjacency file exists.
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
sub answer { 42 }
sub f { return "f[" . join("-", @_) . "]" }
no feature "signatures";
sub g ($) { return "g[" . $_[0] . "]" }
use feature "signatures";
my $anon = sub { pick($_[0]) . "/" . &pick($_[1], "tiny") };
my $seen = 0;
bump($seen);
local $SIG{__WARN__} = sub { };
my @greedy = (warn "a", "b");
my @cut = (warn("a"), "b");
print pick(1), " ", pick(50), " ", $anon->(2, 3), " ", $seen, "\n";
print &answer, " ", answer(), " ", answer, "\n";
print scalar @greedy, scalar @cut, "\n";
print f 1, 2;
print "\n";
print g 1, 2;
print "\n";

--- expect output
small big-2 small/tiny 1
42 42 42
12
f[1-2]
g[1]2

--- expect parses
