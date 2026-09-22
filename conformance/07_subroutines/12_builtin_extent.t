#!perl
# A LIST OPERATOR's parenless argument list is GREEDY, and a paren cuts
# it. `warn "a", "b"` takes both; `warn("a"), "b"` takes one and leaves
# the other to the enclosing list.
#
# TIER 07 subroutines
# INTRODUCES the builtin list operator's argument extent
# USES my, print, scalar, local, %SIG, anonymous sub
# MEASURED perl 5.42.0
#
# THIS IS THE BUILTIN HALF of what `08_parenless_extent.t` and
# `09_prototype_extent.t` measured for USER subs, and the pair should be
# read together. There the extent was decided by the DECLARATION -- `f`
# greedy, `g ($)` cut to one -- and the call sites were byte-identical.
# Here there is no declaration to consult: `warn` is a builtin and its
# extent is decided at the CALL SITE, by a paren or its absence. Two
# routes to one question, and a parser can have either and not the other.
#
# MEASURED perl 5.42.0:
#
#   my @greedy = (warn "a", "b");   1 element
#   my @cut    = (warn("a"), "b");  2 elements
#
# `warn` returns 1, so the greedy form yields ONE value -- both strings
# went to `warn` -- and the cut form yields TWO, the `1` and the `"b"`
# the paren pushed out of `warn`'s reach. Printed as counts, `12`.
#
# THE OPTREE SAYS THE SAME THING AND THE LINT CANNOT READ IT. Measured:
#
#   9  <$> const[PV "a"] s
#   a  <$> const[PV "b"] s     <- INSIDE warn's arguments
#   b  <@> warn[t4] sKP/1
#   i  <$> const[PV "a"] s
#   j  <@> warn[t7] sK/1
#   k  <$> const[PV "b"] s     <- AFTER warn, so the list's
#
# `const[PV "b"]` before the `warn` in one and after it in the other,
# which is exactly the shape `08_parenless_extent.t` records for `f` and
# `g`. `opsOf` collects op NAMES in execution order and the lint compares
# SETS, so it sees `const`, `const`, `warn` twice over and cannot say
# which side of the operator either constant fell on. The counts are what
# carry the claim.
#
# HOW A WARNING BECAME OBSERVABLE, which is the whole reason this file
# measures `warn` and not `die`. The runner compares STDOUT byte for
# byte, and `warn` writes to STDERR -- a stream nothing here reads, as
# `10_io/README.md` records. `local $SIG{__WARN__} = sub { }` installs a
# handler, which perl calls INSTEAD of writing to STDERR, so the warning
# text goes nowhere and the only thing left to observe is `warn`'s return
# value. That is the value this file counts.
#
# The handler is EMPTY on purpose. A handler that printed would put the
# warning text on STDOUT and the pin would then depend on perl's message
# format -- "a b at FILE line N." -- which carries the file's own path
# and line number. The count does not.
#
# `die` CANNOT BE WRITTEN THIS WAY AND IS NOT IN THIS FILE. It asks the
# identical extent question and 218 of T1's 986 files use it, so its
# absence is a gap rather than a decision about relevance. Measured,
# there is no spelling that makes it observable here:
#
#   - `$SIG{__DIE__}` does NOT prevent termination. Measured,
#     `local $SIG{__DIE__} = sub { print "D" }; die "x"; print "never"`
#     prints `D` and exits 255. The handler runs; the program still
#     dies, and a file that dies has no STDOUT to pin.
#
#   - A `die` under a runtime-false guard compiles and never fires, so
#     the op is in the stream and the file prints normally -- but the
#     branch is DEAD, and a dead branch says nothing about how far its
#     arguments ran. It would pin the presence of `die` and not its
#     extent, which is the thing worth measuring.
#
# The spelling that works is `eval { die ... }`, and `eval` is not yet in
# the corpus. When a tier claims it, `die`'s extent file can be written
# against this one as its pair. Recorded here rather than in an issue,
# because this file already holds the measurement that says why.
#
# THE TOKEN FACTS ARE THE OTHER HALF, and they are the same two
# `08_parenless_extent.t` makes, for the same reason: the extent is a
# question about PARENTHESES, so a parser must not be free to add or
# drop one while claiming to have handled the construct.
#
# The counts this file would most like to assert are not expressible.
# There are TWO words spelled `warn`, TWO `(` and TWO strings spelled
# `"b"` -- one of each on either side of the question -- and the grammar
# admits only `one` and `no`, so every count that names the construct
# directly would be FALSE of a correct lex. `GLOSSARY.md` records that
# limit; `08_parenless_extent.t` can assert `one operator whose text is
# ","` only because its source holds a single call.
#
# So the facts are written on what IS unique, and both are about the
# machinery that makes the measurement possible rather than about the
# extent itself:
#
#   - `one word whose text is "__WARN__"` pins that the handler is
#     installed EXACTLY ONCE. Two would mean the second silently
#     replaced the first and the counts below would be measuring a
#     different program; none would mean the warnings went to STDERR and
#     the file measured nothing while still printing `12`.
#
#   - `one word whose text is "local"` pins the dynamic scope that makes
#     the handler apply to the two statements after it. A lexer that
#     read `local` as part of the variable beside it, or dropped it as a
#     modifier, changes the count.
#
# Neither is the extent claim. The extent claim is the OUTPUT, `12`, and
# these two are what stop that output being produced by a program other
# than the one written here.

--- source
local $SIG{__WARN__} = sub { };
my @greedy = (warn "a", "b");
my @cut = (warn("a"), "b");
print scalar @greedy, scalar @cut, "\n";

--- expect output
12

--- expect parses

--- expect tokens
one word whose text is "__WARN__"
one word whose text is "local"
