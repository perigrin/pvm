"""Port tier 05_scoping to mdtest topics.

Three topics: the three lexical declarations, the two constructs that
scope by TIME rather than by name, and the adjacency body.
"""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '05_scoping'

covered = set()

covered |= topic(TIER, 'declarations.md', 'Lexical declarations', """
`my`, `our` and `state`: three ways to bind a NAME for a region of
source, told apart by what the name refers to and when its
initialisation runs.

**Tier 05 scoping.** Introduces `enterloop`, `leaveloop`, `once`.
Depends on 04_operators.

THREE OPS FOR FIVE CONSTRUCTS IS NOT TIDY, AND IT IS WHAT PERL EMITS.
Measured under 5.42.0 with `perl -MO=Concise,-exec`, `my` has no op of
its own: `my $x = 1` compiles to `const` then `padsv_store[$x]
vKS/LVINTRO`, both tier 01's, claimed there. `our` emits `gvsv` and
`sassign`, both tier 02's. What separates them is a FLAG -- `gvsv[*x]
s/OURINTR` against plain `gvsv[*x] s` -- and the ops list does not
record flags. Only `state` adds an op of its own, `once`.

So the op stream cannot tell these cases from earlier tiers' cases, and
the OUTPUT is what distinguishes them. Every case below prints from two
live declarations of the same name, because a parser that conflates two
of these forms still produces a plausible tree and is caught only by the
value it prints.
""", [
    ('01_my.t', '`my` shadows, and the outer binding survives', """
This is where `my` stops being scaffolding. Tier 01 writes `my $x = 1`
in every case, but only to hold a literal still; nothing there asks what
the `my` does. Here the question is the scope, and the answer is visible
only because two declarations of the SAME NAME are live at once and
print different values.

The op stream cannot tell this case from a tier 01 one. Both emit
`const` then `padsv_store[$x] vKS/LVINTRO`, and the two `$x` differ only
by pad slot -- `[$x:1,5]` against `[$x:3,4]` in the concise output. That
is the honest position of this whole tier: what distinguishes it is the
declared subject, not the op set.

The block brings `enterloop`/`leaveloop` with it, which is this tier's
own op pair and the subject of the bare block case.
"""),
    ('02_our.t', '`our` names a package variable, and the shadow proves it', """
`our $x = 1` looks like `my $x = 1` and is nothing like it: `my` makes a
pad slot, `our` makes a lexically-scoped NAME for a symbol-table entry
that already exists. The probe is the shadow -- inside the block `$x` is
the `my`, and `$main::x` is the `our`'s referent, unshadowed and still
1. A parser that read `our` as `my` would print `2 2`, or nothing at
all.

It emits no op of its own; `gvsv` and `sassign` are tier 02's, and the
difference from a bare assignment is the `s/OURINTR` flag the ops list
cannot see. The output is what distinguishes the constructs here, which
is why the case exists even though the op stream has nothing new to
show.

No `use strict`, so the qualified `$main::x` needs no declaration of its
own; the `our` on line 1 is what makes the unqualified `$x` inside the
block refer to the same scalar.
"""),
    ('03_state.t', '`state` scopes like `my`, and initialises once', """
`state $n = 1` compiles to `once(other->...)` wrapped around an
otherwise ordinary `padsv_store[$n] vKS/LVINTRO,STATE`. The pad op is
tier 01's; `once` is the whole of what `state` adds, and it is this
tier's only op a single construct owns outright.

WHAT THIS CASE CANNOT SHOW, and why that is not a gap. The point of
`state` is that the value persists across calls, which needs a sub to
call twice. `sub` is tier 07 and `entersub` is not in this tier's
budget; a bare block is the only re-enterable-LOOKING construct
available and it is not re-enterable -- it runs exactly once, so `once`
firing once is indistinguishable from an ordinary initialisation. The
behavioural probe for persistence belongs to tier 07. What is asserted
here is the part that IS observable: the declaration compiles, and its
scoping is lexical.

The `use feature "state"` line is a tier 12 construct in a tier 05 case,
affordable because it emits NO RUNTIME OP -- measured, not assumed.
Under `-MO=Concise,-exec` the pragma leaves nothing in the op stream at
all; it only flips a bit in the `nextstate` hints field
(`v:%,{,fea=15`), and the lint reads op names, not hints. `use v5.36`
was measured too: same ops, a different hints value. The explicit
feature import is preferred because it names the one thing this case
needs rather than dragging in a bundle.
"""),
])

covered |= topic(TIER, 'dynamic-scope.md', 'Dynamic scope and the bare block', """
The two constructs here are the ones that scope by TIME. `local` binds a
value for the duration of a block, and the bare block is the thing whose
exit performs the restore.

**Tier 05 scoping.** Introduces `enterloop`, `leaveloop`, `once`.
Depends on 04_operators.

`enterloop`/`leaveloop` is this tier's own op pair and the surprising
one: a bare block is a LOOP THAT RUNS ONCE. The two cases are together
because the pair is incidental in one and load-bearing in the other --
`leaveloop` is where `local`'s restore happens.
""", [
    ('04_local.t', '`local` restores on block exit', """
This is the one form in the tier that is not lexical. `my`, `our` and
`state` all bind a NAME for a region of source; `local` binds a VALUE
for a region of TIME. Nothing about the name changes -- `$x` is the same
package scalar throughout -- and what the block changes is what that
scalar holds while the block is running.

The probe is the second print: the block shows the localised 2, and
after it `$x` is 1 again, restored by the block exit with no statement
putting it back. That restoration is the entire construct.

Like `our`, it emits no op of its own: `gvsv` and `sassign` are tier
02's, and the difference from `our` is a flag, `gvsv[*x] s/LVINTRO`
against `s/OURINTR`, which the ops list does not record. The `enterloop`
and `leaveloop` are the bare block's, this tier's own, and here they are
load-bearing rather than incidental -- `leaveloop` is where the restore
happens.

No `use strict`, deliberately. Under strict, `local $x` on an undeclared
package variable is a compile error ("Global symbol "$x" requires
explicit package name") and would need an `our $x` first, making the
case a test of two constructs. The corpus does not run under strict
unless a case asks for it, so the plain assignment on line 1 is what
brings `$main::x` into existence.
"""),
    ('05_bare_block.t', 'A bare block is a scope, and a loop that runs once', """
A bare `{ ... }` at statement position is a SCOPE: declarations inside
it are gone on the far side, and it executes exactly once. It compiles
to `enterloop`/`leaveloop` -- perl builds it as a loop that runs once,
so that `last`, `next` and `redo` have something to act on inside it.
The braces look like grouping and the optree says loop.

The contrast is measured, because the same braces produce different ops
depending on what precedes them. Under 5.42.0:

    $ perl -MO=Concise,-exec -e '{ print "a\\n"; }'
    ... enter nextstate enterloop nextstate pushmark const print leaveloop leave

    $ perl -MO=Concise,-exec -e 'my $c = 1; if ($c) { my $y = 2; print "$y\\n"; }'
    ... and enter nextstate const padsv_store ... print leave leave

The `if` block gets `enter`/`leave`, tier 01's ops, because nothing can
jump out of it; only the bare form gets the loop pair. With a constant
condition -- `if (1) { ... }` -- the block ops vanish entirely, folded
away, which is the optimiser-erases-constructs warning in miniature. So
`enterloop`/`leaveloop` belong to this tier and not to 06_control, even
though 06 is where loops live.

The loop-ness is not probed behaviourally. Showing it would take a
`last` inside the block, and `last` compiles to a `last` op that no tier
at or before 05 claims -- measured; the op stream carries a literal
`last` between `print` and `leaveloop`. That probe belongs with the loop
control tier. What this case asserts is the part inside the budget: the
scope closes, and the block is entered once and not repeated.

The second block is what shows "once": a construct perl builds as a loop
prints `again` a single time. Its ops are the same pair -- `{ 1; }`
keeps its `enterloop`/`leaveloop` under 5.42.0, measured rather than
assumed, so an optimiser that erased a do-nothing block would be visible
as a missing op rather than as identical output.
"""),
])

covered |= topic(TIER, 'adjacency-05_scoping.md', 'Every scoping form, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another, and each initialised from the value it shadows so the shadowing
order is observable rather than merely compiled.

**Tier 05 scoping.** Introduces nothing of its own; it is the mixture
that is the subject. Depends on 04_operators.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why a corpus of such cases cannot
reach an ADJACENCY bug -- a parser that handles every declaration alone
and mishandles a pair goes green over the pair.

This tier's adjacency has a hazard the literal tiers do not: all five
constructs declare the SAME KIND OF THING, so a parser can conflate two
of them and still produce a plausible tree. `our $g` and `local $g`
differ only by a flag in the optree; `state $s` and `my $s` differ only
by a flag and a wrapper. So each declaration is paired with a probe that
only the correct reading survives.

DEPENDS ON names 04_operators, and the arithmetic is where this body
crosses that edge: `+` and `*` are tier 04's `add` and `multiply`, and
putting them in the initialisers is what makes this case exercise the
dependency rather than only 05.

The union rule is visible here in the direction that surprises. This
body has MORE declarations than any construct case and emits a smaller
distinct set than their union in one respect -- `padrange` never
appears, because the declarations are separated by other statements
rather than consecutive. `padrange` ABSORBS `pushmark` when three `my`
declarations do sit together, so a tier's declared ops are a UNION
ACROSS ITS CASES and never a property of any one of them.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
Five scoping forms on consecutive statements, three of them inside the
fifth, each carrying a probe that a conflating parser fails:

- `local $g = $g + 10` -- the right-hand `$g` is the OUTER value, 1,
  read before the save; a parser that localised first would compute from
  an undefined value and warn.
- `my $x = $x * 2` -- the right-hand `$x` is the OUTER `my $x`, 2,
  because the new pad slot is not visible until the statement ends. 4,
  not undef, is what says the declaration's scope starts late.
- `state $s = $s + 100` -- the same rule, giving 103.

The `11 4 103` line inside the block and `1 2 3` outside it are one
assertion each about which binding won, in both directions.

`use feature "state"` is a tier 12 construct here, affordable because it
emits no runtime op at all -- measured; see the `state` case.
"""),
])

check(TIER, covered)
