"""Port tier 03_context to the mdtest topic format."""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '03_context'

covered = set()

covered |= topic(TIER, 'aggregates-in-context.md', 'An aggregate in two contexts', """
The tier's baseline: what an array becomes when something asks it for a
scalar, and what the comma operator does when the receiver is not a
list.

**Tier 03 context.** Introduces `join` and `list`. Depends on
02_variables.

The dependency is on the AGGREGATES, not on element access. `$a[0]` is
already a scalar and tells you nothing; `@a` in two contexts is the
whole subject. And context needs a thing to be in it -- `my $x = 1` is
the same program in either context, which is why this tier cannot come
before the one that supplies `@a`.

Two of the three cases below introduce no op the corpus can lint. That
is the tier's central fact rather than a gap in it: what separates the
contexts is the `s` versus `l` FLAG B::Concise prints on an op, and a
flag is not an op.
""", [
    ('01_scalar_of_array.t', '`my $n = @a` and `my $n = scalar(@a)` are the same program', """
The keyword that names this tier's subject compiles to no op of its own.
Both statements compile to `padav[@a] s` followed by `padsv_store` --
byte-identical optrees. What separates scalar context from list context
is the `s` versus `l` flag on `padav`.

So this case introduces no new op, and it is the case that proves the
tier cannot be linted by op NAME alone. A tier whose first case emitted
something new would imply the tier is a set of ops; it is a set of
contexts, of which one leaves no trace at all.
"""),
    ('02_interpolated_array.t', '`"@a"` is a `join`, decided at compile time', """
The source contains no function call and compiles to one. `"@a"` emits
`gvsv[*"]` -- fetching the list separator `$"` -- and then
`join[t5] sK/2`. No reader of the source would predict `join`, which is
the whole point: interpolation is a third context after scalar and list
in everything but name, and it is visible only in the optree.

The `gvsv` is tier 02's, claimed there with the package scalar; only
`join` is new here. That is what lets the case sit in 03 at all -- it
leans on an earlier tier's op rather than smuggling one in.

The separator is a single space by default, so three elements print as
`x y z`, unlike a bare `print @a` list which has nothing between its
arguments. That difference is the observable half of what the optree
shows.
"""),
    ('03_comma_in_scalar_context.t', 'The comma operator in scalar context discards its left operands', """
`my $last = (4,5,6)` is 6. The scalar half emits `pushmark v`,
`const[IV 6]`, `list sKP` -- the constants 4 and 5 are simply gone.
Two-thirds of the construct this case is about does not survive to the
optree, because discarding them is precisely what the comma operator
does here.

That makes this the tier's example of the lesson tier 01 records for
`my $x = 1+2` folding to `const[IV 3]`: ops LINT a declared tier and
cannot derive one, and a case demonstrating a construct may emit fewer
ops than the construct has parts.

The list half is the discriminator. The same parenthesised commas
assigned to an array keep all three values, so the pair shows that the
source text does not decide -- the assignment target does.
"""),
])

covered |= topic(TIER, 'list-operators-in-context.md', 'List operators, each asked twice', """
Five operators that answer a different question depending on what
received the answer. Each appears twice below, once with the `s` flag
and once with `l`, because one half alone asks the parser for one answer
where the construct has two.

**Tier 03 context.** Introduces `reverse`, `sort`, `localtime`,
`mapstart`, `mapwhile`, `grepstart`, `grepwhile`. Depends on
02_variables.

None of these cases is separated from its own other half by an op NAME.
`my $x = reverse @a` emits `reverse[t4] sK/1` and `my @b = reverse @a`
emits `reverse[t6] lK/1` -- same op, same operand, different answer. The
discrimination is in the flag and in the OUTPUT, which is why the output
blocks carry this topic's weight.
""", [
    ('04_reverse.t', '`reverse`: reversed elements, or a reversed string', """
In list context `reverse` reverses the elements; in scalar context it
concatenates them and reverses the characters. Same op and same operand
twice: `my @b = reverse @a` emits `reverse[t6] lK/1`, `my $s = reverse @a`
emits `reverse[t4] sK/1`.

The scalar result is `321` and not `3 2 1`: reversing in scalar context
first joins the list with nothing between the elements, then reverses
the resulting string. A parser that treats the two as one operation on
"a value" has no way to say which of these it means.
"""),
    ('05_sort.t', '`sort`, and the `() =` count idiom in both contexts', """
`sort` in list context returns the sorted elements. The second statement
is the count idiom, `my $n = () = sort @a`, which is the only way to
observe the list-versus-scalar distinction on an `aassign`: the inner
assignment to an empty list runs in list context, and the outer
assignment then asks it how many elements it moved --
`aassign[t6] sKS`.

The third statement is that same idiom with an ARRAY receiving it, and
it is what makes the second a pair rather than a single measurement:
`my @c = (() = sort @a)` emits `aassign[t8] lKPS`. Same op, same
operand, same spelling of the idiom, and unrelated answers. In scalar
context the empty-list assignment reports how many elements it moved,
which is 3; in list context it yields what it assigned, which is the
empty list, so `@c` has 0 elements. Nothing in the source says which.
Without the third statement the `aassign` here is only ever seen with
`s`, and the claim that context is a flag would be made by a case that
never shows the other value of the flag.

No comparator block appears on purpose: `sort { $a <=> $b } @a` would
drag in a block and the comparison operator, which belong to later
tiers. Default sort is a string sort, which is why the digits come back
in the order they do.
"""),
    ('06_localtime.t', '`localtime`: a nine-element list, or one string', """
The sharpest discriminator in the tier, because its two contexts return
unrelated TYPES. `reverse` and `sort` return rearrangements of the same
data either way. `localtime` in list context is
(sec,min,hour,mday,mon,year,wday,yday,isdst), and in scalar context a
formatted string like `Sun Sep 21 14:03:11 2026`. Nothing about the call
site says which; the assignment target decides.

What this case pins is deliberately NOT the time. `localtime` depends on
the clock and the timezone, so pinning either result verbatim would make
it fail tomorrow and in another zone. Both assertions are facts about
SHAPE: the list has nine elements always, and the scalar string is one
element always. That is exactly the contextual difference, and it is the
part that does not move.
"""),
    ('09_map.t', '`map BLOCK` and `map EXPR`, a fork the optree cannot see', """
`map` takes either a BLOCK or an EXPRESSION, and the two forms are a
genuine parse fork -- `map BLOCK LIST` takes no comma after the block,
`map EXPR, LIST` requires one. Measured, `my @b = map { $_ } @a` and
`my @b = map($_, @a)` emit the same ops, in the same order, with the
same flags and targs:

    pushmark, pushmark, padav[@a] lM, mapstart lK, mapwhile lK, gvsv[*_]

So nothing an op-name lint reads distinguishes the halves, and a parser
that took the comma form for the block form emits an identical optree.
This case therefore asserts at the TOKEN STREAM: the comma IS the fork,
the source is written to contain exactly one (`qw()` builds the list
without any), and the token fact counts it. A parser that admitted a
comma after a block, or required one, changes that count.

What is not identical is not an op. The two optrees differ in the COP
SEQUENCE RANGE beside each lexical: `padav[@a:1,5]` under the block form
against `padav[@a:1,3]` under the expression form. A block is a SCOPE,
and measured, a bare `{ 1; }` standing where the map block stands
advances the counter by exactly the same 2. The trace is in pad
metadata, which the op-name lint does not read. "Byte-identical" is too
strong; "identical in every op" is the measured claim.

The fourth statement is the discriminating pair: `my $n = map { $_ } @a`
emits `mapstart sK` where the two before it emit `mapstart lK`, and in
scalar context the answer is the COUNT of what the list form returns.

The list is opened up by `$ENV{M}` because a CONSTANT argument erases
the construct. `$ENV{M}` is unset when the runner executes, so
`"$ENV{M}a"` is the one-character string `a` -- opaque at compile time,
fixed at run time, and the reason the first element comes back `a`
rather than the `x` that `qw` put there. No arithmetic appears in the
block on purpose: `map { $_ + 1 } @a` emits `add`, which is tier 04's. A
bare `$_` is the smallest body that is still a body.
"""),
    ('10_grep.t', '`grep BLOCK` and `grep EXPR`, where the arities disagree', """
The same fork as `map`, measured again for the other keyword and
holding: block form and comma form emit

    pushmark, pushmark, padav[@a] lM, grepstart lK, grepwhile lK, gvsv[*_]

The syntactic difference is a COMMA, which the optree does not record,
so the token stream is the only place the fork is observable. The source
holds exactly one comma and the token fact counts it. As with `map`, the
block form advances each lexical's COP SEQUENCE RANGE by 2, which is pad
metadata rather than an op.

WHERE grep IS SHARPER THAN map, and the reason both cases exist rather
than one: `map` in scalar context returns the count of what it would
have returned in list context, so its two halves report the same NUMBER
in different shapes. `grep` FILTERS, so the count is not the input's --
from three elements it keeps two, the scalar half reports 2 and the list
half produces the two surviving strings. Nothing folds; the
discrimination is between a count and the things counted, and they have
different arities from the input as well as from each other.

`$a[0]` is set from `$ENV{G}` for two reasons at once. A constant
argument erases the construct -- a wholly-constant grep folds and leaves
no `grepstart` to measure -- and grep needs an element that is FALSE, or
it filters nothing and the two contexts stop disagreeing about arity.
`$ENV{G}` is unset when the runner executes, so the element is undef,
which is both opaque at compile time and false at run time.

Written as a bare `$a[0] = $ENV{G}` rather than `"$ENV{G}"`, which is
what the `map` case uses. Measured, the interpolating form of a lone
variable emits `stringify`, and `stringify` is claimed by NO tier -- so
quotes that read as harmless would fail the corpus lint for an op the
READMEs do not yet describe. The plain assignment emits
`aelemfastlex_store` and `multideref`, both tier 02's. No comparison
appears in the block on purpose: `grep { $_ gt "a" } @a` emits `gt`,
which is tier 04's. A bare `$_` tests the element's own truth, which is
the smallest predicate there is.
"""),
])

covered |= topic(TIER, 'context-from-inside.md', 'Asking what the context is', """
The other cases in this tier show context from the outside -- a
construct placed in one and then the other. These two ask perl directly.

**Tier 03 context.** Introduces `wantarray` and `caller`. Depends on
02_variables.

Both are placed at FILE SCOPE deliberately. A sub is where either
operator is USEFUL, not where it is legal, and a sub brings `entersub`
and `leavesub` -- tier 07's ops, four tiers forward, which the lint
would report as reaching ahead. The tier README previously claimed
`wantarray` was out of reach for exactly that reason; measured, that is
false, and the claim is withdrawn.

Observing the result is the part that needs care in both cases.
`defined $w ? ... : ...` emits `cond_expr` and `if (defined $w)` emits
`and`, and both are tier 06's. So both cases COUNT instead: a count is
an assertion about shape that needs no branch, and `scalar` is not an op
at all.
""", [
    ('07_wantarray.t', '`wantarray` at file scope is undef, and costs one op', """
`wantarray` is the only op in perl whose entire meaning is this tier: it
returns true in list context, false in scalar context, and undef in void
context.

Measured, at file scope it compiles to a bare `wantarray s` with no call
machinery at all, and returns undef because a file body is void context.

`my @seen = ($w)` holds exactly one element whether `$w` is undef or
not, so the printed `1` says the op ran and produced a value. That the
value is undef is what the measurement records:
`perl -e 'my $w = wantarray; print defined($w) ? "defined\\n" : "undef\\n"'`
prints `undef`.
"""),
    ('08_caller.t', '`caller` returns different AMOUNTS in the two contexts', """
Measured, this case's first two statements:

    $ perl -MO=Concise,-exec -e 'my $s = caller; my @l = caller;'
    3  <0> caller[t2] s
    7  <0> caller[t4] l

One op name, two context flags -- the `s`/`l` pair this tier is built
on. An op-NAME lint sees one `caller` and is satisfied by a case that
only ever put it in one context.

The behavioural half is stronger here than anywhere else in the tier. At
file scope there IS no caller, and the two contexts do not merely format
the same answer differently: `@l` is EMPTY, because list-context
`caller` with no frame returns the empty list, while `$s` is undef and a
scalar holding undef is still one element. A parser that imposed LIST
context on both prints `0 0`; one that imposed SCALAR on both prints
`1 1`. Both wrong readings are visible in one byte.

`caller` is placed in this tier rather than in tier 07, where a reader
expects a call-stack operator to live, because what makes its parse
distinctive is not the stack: it is that the SAME call site yields a
scalar or a list depending on what receives it. 74 of T1's 986 files
call it and the corpus named it nowhere.

THE TOKEN FACT HERE IS WEAK AND THE CASE SAYS SO. `one word whose text
is "print"` would pass any lexer that tokenises identifiers at all.
Measured, nothing better is available: `caller`, `scalar`, `$s`, `@l`
and `@one` each appear twice, and the fact grammar has only `one` and
`no`, so no count fits -- and every negative reachable here names a
spelling the source writes, which makes it false rather than merely
vacuous. What this case actually claims is its OUTPUT.
"""),
])

covered |= topic(TIER, 'range.md', 'Three operators spelled `..`', """
`..` is not one operator in three contexts; it is three operators
wearing one spelling, and CONTEXT chooses. In list context it builds a
list -- by arithmetic over numbers, by perl's magic increment over
strings -- and in scalar context it is a stateful flip-flop that
remembers whether it is on.

**Tier 03 context.** Introduces `range`, `flip` and `flop`. Depends on
02_variables.

Before these cases the corpus claimed none of the three. The only `..`
anywhere was `for my $i (1 .. 2)` in tier 07's adjacency file, which is
loop scaffolding rather than a claim.

ONE `..` EMITS ALL THREE OPS, which defeats the obvious discriminator:
`range`, `flip` and `flop` all appear for a plain list range, because
the optimiser builds the flip-flop machinery even where the context
means it can never be used. So the op set does NOT separate the three
readings and the weight falls on OUTPUT, which separates them cleanly.

Every endpoint below is a runtime value for a PLACEMENT reason, not a
parsing one. Measured, `my @r = (1 .. 3)` emits no range, flip or flop
op while `my @r = (1 .. scalar(@a))` emits all three; the same holds for
string endpoints. Constant folding is an OPTIMISER effect reaching only
the op stream -- `(1 .. 3)` lexes as `Operator("..")` and parses to a
range either way -- but a folded range emits nothing for the lint to
check against this tier's INTRODUCES block. `scalar(@a)` serves rather
than tier 04's `$ENV{X} // <default>` because `dor` is out of this
tier's op budget, and because a scalar-context array is this tier's own
subject.
""", [
    ('11_range_list.t', 'The numeric list range', """
`..` in list context builds the list between its endpoints. The
endpoint is `scalar(@a)` so that the tier's `range` declaration has
something to check.

The negative fact is the lexical half: `..` must lex as ONE operator.
A lexer that read it as two `.` concatenation operators has a different
program, and the source contains a `.` only inside the `..`, which is
what makes the negative reachable rather than vacuous.
"""),
    ('12_range_string.t', 'The string range counts by magic increment', """
`"az" .. "bb"` is three elements and none of them is a number. It looks
like the numeric range and is a separate claim, because the SUCCESSOR
function is different: a numeric range adds one, a string range applies
the magic increment that also drives `$s++` on a string.

`az` to `ba` is the carry -- the last character wraps from `z` to `a`
and the one before it advances. A parser that read this as an arithmetic
range over numified endpoints would get `0 .. 0` and print a single
zero, since `"az"` numifies to 0.

The COUNT is the second half of the claim. `scalar(@r)` is 3, this
tier's own scalar-context idiom applied to the result -- a range in list
context producing a list whose length is then taken in scalar context,
which is the tier's subject twice over.
"""),
    ('13_flipflop.t', '`..` in scalar context is stateful', """
The third reading, and the one that makes the operator worth three
cases. The other two build lists; this one builds nothing and carries
state between evaluations.

INDICES 2 AND 3 ARE THE CLAIM. At both of them `$on[$_]` is 0 and
`$off[$_]` is 0 -- both operands false -- and the flip-flop selects them
anyway, because it turned on at index 1 and does not turn off until
index 4. No truthiness test can produce that: a parser that read `..`
here as a list range, or as a boolean `or`, or as anything stateless,
selects only indices 1 and 4.

That is why the operands are ARRAY ELEMENTS rather than the `$_ .. $_` a
first draft used. Measured, `grep { $_ .. $_ }` over `(0,1,0,0,1,0)`
prints `1 1` -- and so does `grep { $_ }` over the same list. The
stateful reading and plain truthiness agree, so the case would have
passed against a parser that had never heard of a flip-flop. Two
separate operand arrays are what make the state observable.

The `grep` form is deliberate: the loop spelling of the same claim needs
`enteriter` and `leaveloop`, which are tiers 05 and 06, while
`grepstart` and `grepwhile` are this tier's own.
"""),
])

covered |= topic(TIER, 'adjacency-03_context.md', 'Every construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to tier 02's array, which is what this tier
depends on.

**Tier 03 context.** Introduces nothing of its own; the mixture is the
subject. Depends on 02_variables.

The tier's other cases are one construct each, which makes them
diagnosable: when the `sort` case fails, the construct that failed is
the only one present. That property is also why such cases cannot reach
an adjacency bug -- a parser that handles every construct alone and
mis-handles a pair goes green over the pair.

The constructs sit on consecutive STATEMENTS rather than nested, because
nesting them would need an operator to join them and this tier comes
before the operator tier.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
EVERY DISCRIMINATING CONSTRUCT APPEARS IN BOTH CONTEXTS, which is what
distinguishes this from a list of the tier's constructs. The tier's
subject is not a set of constructs, it is a set of PAIRS whose halves
emit the same op and differ only by the `s` versus `l` flag. A body
holding one half of each asks the parser for one answer where the
construct has two, and a parser that collapsed the two contexts into a
single answer would go green over it.

So `@a` appears in scalar and in list context, `reverse` twice, the
empty-list count idiom twice, `localtime` twice, `caller` twice, and
`map` and `grep` twice each. Measured, that is `padav s`/`padav l`,
`reverse sK/1`/`reverse lK/1`, `aassign sKS`/`aassign lKPS`,
`localtime s`/`localtime l`, `caller s`/`caller l`,
`mapstart sK`/`mapstart lK` and `grepstart sK`/`grepstart lK` -- seven
pairs in one body, with the tier's remaining constructs in scope.

The adjacency with the earlier tier is the first line: `@a` is tier 02's
array, and every statement after it puts that same array in a different
context. Pairing this tier with the one it depends on is the array being
re-used, not a separate statement about arrays.

READING THE OUTPUT. `sort` is a default string sort, so `(3,1,2)` comes
back `1 2 3`. The reverse of that array in scalar context is `213` --
the elements joined with nothing between them, then the string reversed
-- not `2 1 3`, which is the list-context `@revd` on the next line.
`$moved` is 3 and `@kept` is empty from the SAME idiom. `$fields` is 9
because list-context `localtime` returns nine elements; `$stamp` is the
scalar half, a formatted string, and it is COUNTED rather than printed
-- pinning it would make this case depend on the clock and the timezone,
while `@seen` holds two elements whatever the time is. `$who` and
`@frame` are `caller` in the two contexts, and at file scope they return
different AMOUNTS: `$who` is undef and `@frame` is the EMPTY list, which
is the `0` after the `3`.

`@mapped` is the three elements and `$mapcount` is 3; `@kept2` is the
three elements and `$grepcount` is 3 as well, because every element of
`(3, 1, 2)` is true and nothing is filtered. That grep keeps everything
HERE is not a weakening -- the `grep` case is where grep filters, using
an element bound to `$ENV{G}`, and this body's job is adjacency.
Bringing `$ENV` in here would add an opacity it does not otherwise need.
Neither `map` nor `grep` FOLDS despite `@a` being wholly constant, which
the comma operator two statements up does: measured, a block is not
constant-foldable however constant its input.

`..`'s three readings close the body. `@ranged` is the numeric list
range, `@lettered` is the string range whose successor is the magic
increment (`az ba bb`, carrying), and `@window` is the scalar-context
FLIP-FLOP whose `1 2 3 4` selects indices 2 and 3 where both operands
are false.

Fewer ops appear than the constructs suggest. `padrange` fuses
consecutive `my` declarations, and the comma operator in scalar context
erases its own left operands -- so the op set is a UNION across the
tier's cases rather than a property of this one. It happens that this
body does emit all seven pairs, but that is a fact about this program,
not a rule the format requires.
"""),
])

check(TIER, covered)
