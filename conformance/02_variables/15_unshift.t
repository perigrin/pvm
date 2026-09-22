#!perl
# `unshift` obeys the same array-first rule as `push`, at the other end.
#
# TIER 02 variables
# INTRODUCES nothing 14_push.t does not
# USES nothing from a later tier
#
# This file exists because the rule and the OP are separable claims. A
# parser can hard-code `push`'s aggregate first slot as a special case
# for the one keyword it was tested on, and `unshift` is the second
# keyword that rule has to cover. T1 uses it in 15 files.
#
# MEASURED perl 5.42.0. The flags are `push`'s exactly, over a different
# op:
#
#   $ perl -MO=Concise,-exec -e 'my @a = (20, 30); my @head = (10);
#         unshift @a, @head;'
#   ... <0> padav[@a:1,3] lRM
#       <0> padav[@head:2,3] l
#       <@> unshift[t5] vK/2
#
# `lRM` on the container, `l` on the flattened list, and `unshift` where
# `push` had `push`. So the tier claims a second op and no new rule --
# which is the honest shape of this file, and why its INTRODUCES line
# says so rather than inventing a construct name for it.
#
# The ELEMENT it prints is what separates the two behaviourally:
#
#   $ perl conformance/02_variables/15_unshift.t
#   1
#   3
#   10
#
# `@head` still holds ONE element -- the flattened slot is read, not
# consumed, which is `14_push.t`'s first line mirrored. `@a` holds three,
# the count a flattening parser also reaches, so the count proves
# nothing on its own. `$a[0]` is 10, the value that arrived from
# `@head`, and that is the line that separates this operator from
# `push`: under `push` the same three-element result would have 20
# there. A parser that got the operator right and the END wrong passes a
# count check and fails this.
#
# The token facts count, as `14_push.t`'s do, and the second is the
# falsifying half: exactly ONE word spelled `unshift` and NO word spelled
# `shift`. A lexer that read `unshift` as `un` followed by `shift` -- or
# that longest-matched the keyword table wrongly -- would produce a
# `shift` here and fail. `shift` is tier 11's op, so it is also the
# spelling this file must not accidentally contain.

--- source
my @a = (20, 30);
my @head = (10);
unshift @a, @head;
print scalar(@head), "\n";
print scalar(@a), "\n";
print $a[0], "\n";

--- expect output
1
3
10

--- expect parses

--- expect tokens
one word whose text is "unshift"
no word whose text is "shift"
