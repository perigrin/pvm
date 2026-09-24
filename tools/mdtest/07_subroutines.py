"""Tier 07 subroutines: declaration and call forms, arguments, extent."""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '07_subroutines'

INTRO = """**Tier 07 subroutines.** Introduces `anoncode`, `argcheck`,
`argdefelem`, `argelem`, `entersub`, `leavesub`, `lock`, `return`, `warn`.
Depends on 06_control."""

covered = set()

covered |= topic(TIER, 'subroutines.md', 'Declaring and calling', """
A subroutine is declared, and then reached. This topic is the two halves
of that: where the CV comes from -- a named declaration, an anonymous
expression -- and the four spellings that call it.

%s

`entersub` COVERS EVERY CALL FORM, and the differences are flags:
`f(1)`, bare `f`, `&f` and `&f(2)` all emit it, the ampersand forms
carrying `/AMPER` and the code-reference call `/STRICT`. A lint that
reads op NAMES sees one construct where the language has four, which is
why this tier needs one case per form rather than one op claim.
""" % INTRO, [
    ('01_named_sub.t', 'A named subroutine, declared and called', """
`sub f {...}` installs a CV under a package name and `f(1)` calls it
through that name.

THE DECLARATION EMITS NOTHING into the main program. It is compile-time:
the CV is built and installed, and the main optree that runs afterwards
contains only the call -- `pushmark`, the `gv` naming the sub, and
`entersub`. Measured 5.42.0, `sub f { $_[0] + 1 } print f(1), "\\n"`
gives `gv[IV \\&main::f]` rather than `gv[*f]`, because the call resolved
to the CV at compile time: the sub was already declared when the call was
compiled. A call compiled BEFORE the declaration gets `gv[*f]` and looks
the name up at run time. Same op either way.

The body, where this tier's subject actually lives, is a separate optree
that `-MO=Concise,-exec` never prints.
"""),

    ('02_anon_sub.t', 'An anonymous subroutine is a value', """
`sub { ... }` builds a closure in the enclosing scope and yields it;
`$c->(4)` calls it through the variable holding it.

This is the ONE PLACE a subroutine construct emits an op into the main
program, and the reason `anoncode` is in this tier's INTRODUCES list at
all. A named declaration vanishes at compile time; an anonymous one
cannot, because the closure has to be built where it is written --
`anoncode` is that build, and it runs every time control reaches the
expression. Measured 5.42.0, `my $c = sub { $_[0] * 2 }` compiles
`anoncode[CV CODE]` then `padsv_store`, and the call is `entersub[t3]`
with the CV coming off the pad instead of out of a `gv`.

`anoncode[CV CODE]` is opaque on purpose: the body it wraps is a CV that
B::Concise prints only when asked for it by name, and an anonymous sub
has no name to ask with.
"""),

    ('03_call_forms.t', 'Four spellings of one call', """
`f(...)`, bare `f`, `&f` and `&f(...)` all compile to `entersub`, and
only `/AMPER` in the flags column distinguishes the ampersand pair.

THE SEMANTIC DIFFERENCE IS REAL AND THIS FILE ONCE NAMED IT WITHOUT
ASSERTING IT. `&f` with no parens passes the CALLER's `@_` through
untouched rather than an empty list. The earlier source used
`sub answer { 42 }`, which ignores its arguments, so all four spellings
printed 42 and a parser that never implemented the forwarding passed:
the header said the distinction was real and the pin said nothing.

Two changes make it observable and both are necessary -- the callee
REPORTS its arguments, and the calls happen INSIDE a sub whose own `@_`
is non-empty. At file scope there is no caller `@_` to forward and the
four spellings would still agree. Measured 5.42.0 with `outer(1, 2)`:
`answer()` prints `[]`, bare `answer` prints `[]`, `&answer` prints
`[1 2]`, `&answer()` prints `[]`.

THE THIRD LINE IS THE CLAIM: only the parenless ampersand forwards. The
FOURTH is why `&f()` is here rather than treated as a noisier `&f` -- an
EXPLICIT empty list passes nothing, so the ampersand alone does not cause
forwarding, the ABSENCE of an argument list does. A parser reading `&f`
and `&f()` as the same call would print `[1 2]` twice.

Measured, bare `f` compiles as `gv[*f]` where the parenthesised form
gave `gv[IV \\&main::f]`: the parser had to decide whether `f` was a call
at all. Still one `entersub`.
"""),

    ('11_code_ref_call.t', 'Calling through a code reference', """
`$code->(...)`: the callee is a VALUE rather than a name, and the
argument list is parenthesised because there is no other way to spell it.

WHY THIS FORM HAS NO EXTENT QUESTION is the reason it sits beside the two
cases that do. An extent decided by a DECLARATION needs a declaration the
parser has already seen; here there is no name to have seen, so the
callee is whatever `$code` holds at run time, the parens are mandatory,
and the argument list ends where they close. The same construct that
makes the callee unknowable until run time makes its argument extent
knowable at parse time.

Measured 5.42.0, the call is `padsv` then `entersub[t3]` where the named
forms carry `gv` -- a callee read out of a pad slot rather than resolved
by name. Still `entersub`.

The empty call is the second statement rather than a case of its own:
`$code->()` with no arguments is where a parser treating `->(` as an
operator taking an operand would refuse, and it costs one line.

The token facts pin the arrow NEGATIVELY, the only form available, since
`->` appears twice and the vocabulary is `one` or `no`. They assert the
arrow is ONE token and not two -- neither a minus nor a greater-than is
anywhere in this source, so a lexer that split `->` would fail both at
once. The positive half is carried by the case parsing at all.
"""),
])

covered |= topic(TIER, 'arguments.md', 'Arguments', """
How a sub receives its parameters. Two protocols: `@_`, which is tier
02's array populated by the call, and a signature, which is the one
construct in this tier the optimiser does NOT erase.

%s

The two protocols cannot share a sub without noise. Measured 5.42.0,
`@_` inside a signatured sub is populated but reading it warns --
`Use of @_ in scalar with signatured subroutine is experimental` -- and
a warning on stderr is not something this corpus pins.
""" % INTRO, [
    ('04_args_array.t', 'Arguments arrive in `@_`', """
A sub reads its arguments as array elements: `$_[0]`, `$_[1]`, and
`scalar @_` for the count.

THE CONSTRUCT IS SPELLED ENTIRELY IN TIER 02'S SYNTAX. `@_` is an array
and `$_[0]` is element access; nothing about reading arguments needs an
op this tier introduces. What makes it tier 07's is that `@_` is only
populated by a call, which is `entersub` -- the array exists, but outside
a sub it holds nothing a caller put there. So the body's ops are
`aelemfast`, `rv2av` and friends, all tier 02's, and this tier claims
none of them.

Measured 5.42.0, the body of `sub n { scalar @_ }` is `gv[*_]`,
`rv2av`, `av2arylen`, `leavesub` -- and that `leavesub` is invisible to a
lint reading the main program alone.
"""),

    ('06_signature.t', 'A signature names the parameters', """
`sub f ($a, $b = 3)` binds by position and supplies a default for the
ones omitted.

A SIGNATURE IS NOT ERASED, which is worth stating because it is the
opposite of what the rest of this tier warns about. Measured 5.42.0,
`sub f ($a, $b = 3) { $a + $b }` compiles to `argcheck(1,1,-)`, an
`argelem` per parameter, and an `argdefelem` guarding the default
expression -- distinct ops carrying the declared arity, not ordinary pad
assignments that happen to read `@_`. Compare `sub f { my ($a,$b) = @_ }`,
which is `padrange` and `aassign` and nothing else: the two spellings
really are different optrees.

All of which is inside the sub, where a main-program lint cannot see it.

`use v5.36` itself adds no ops. Measured, it changes the `nextstate` hint
flags -- `v:us,*,&,{,$,fea=6` instead of `v:{` -- and nothing more, so
the feature pragma this case needs costs the op stream nothing.
"""),

    ('07_args_alias.t', '`@_` aliases the caller\'s variables', """
`$_[0]++` inside a sub increments the CALLER's variable, because the
elements of `@_` are not copies of the arguments -- they ARE the
arguments, the same SVs under different names.

THIS IS THE CASE THE `@_` READING CASE IS NOT. Reading `$_[0]` is
something every array in the language supports and therefore measures
nothing about `@_` in particular. Writing through it is the whole
difference: an ordinary array's element is its own storage and a write to
it is local, and `@_`'s is not. The contrast is in the source rather than
in prose about it -- `bump_alias` writes through `$_[0]` and the caller
sees the change; `bump_copy` takes the conventional `my ($n) = @_` copy
and writes to that, and the caller does not. Same increment, same op, two
different variables, which is why `$untouched` stays 1.

MEASURED 5.42.0, THE OPS ARE THE SAME, and that is the finding.
`bump_alias` is `aelemfast[*_]`, `postinc`, `leavesub`; `bump_copy` is
`padrange`, `aassign`, `padsv`, `postinc`, `leavesub`. One `postinc`
each. The aliasing is not an op and not a flag on one: it is a property
of what `aelemfast[*_]` fetches, established when `entersub` filled `@_`
with the caller's SVs rather than with copies of them. No op claim can
express it and no optree comparison can find it -- only running the
program can, which is why this case leans on its pinned output entirely.
"""),

    ('05_return.t', 'Explicit return, early and trailing', """
An explicit `return` leaves the sub with a value, and an early `return`
inside a branch leaves it before the last statement runs.

THE SOURCE CONSTRUCT AND THE OP ARE DIFFERENT THINGS. A `return` that IS
the last thing the body does compiles to nothing: the value is already on
the stack and `leavesub` takes it, so perl deletes the op. Only a return
that jumps -- guarded by a branch, with statements after it -- survives.
Measured 5.42.0, `sub f { return $_[0] + 1 }` is `aelemfast_lex_or_rv2av`,
`const`, `add`, `leavesub` and no `return`; `sub g { return 1 if $_[0]; 0 }`
compiles `and` guarding a `pushmark`/`return` pair.

That is why this tier declares 06_control as its prerequisite. Without a
branch there is no early exit, without an early exit there is no `return`
op, and the construct the tier is named for would be unobservable in any
optree the corpus could produce.

Both of those live in the sub's own optree, which a main-program
measurement does not reach. The BEHAVIOUR is what this case asserts and
it distinguishes the two: `classify(0)` prints `zero` only if the early
return fired, and `positive` only if it did not.
"""),
])

covered |= topic(TIER, 'argument-extent.md', 'Argument extent', """
How far a parenless argument list runs. Three cases where the answer is
decided somewhere other than the call site, and one that establishes the
precondition for asking at all.

%s

A USER SUB'S EXTENT IS DECIDED BY ITS DECLARATION AND A BUILTIN'S BY A
PAREN AT THE CALL SITE. Two routes to one question, and a parser can
implement either and not the other -- which is why both halves are here.

`lock` is in this topic for proximity and nothing else, and its own case
says so. It declares no sub, calls no sub, and has no argument protocol;
what it does have is a named unary's operand, which is the shape the rest
of this topic is about.
""" % INTRO, [
    ('08_parenless_extent.t', 'A parenless call is greedy', """
A parenless call to a declared callee with no prototype consumes the
whole remaining list, even nested inside another list operator that looks
like it owns those arguments.

THIS CASE AND THE NEXT ARE A PAIR and neither means much alone. The call
sites are byte-identical but for the callee's name -- `print f 1, 2;`
here and `print g 1, 2;` there -- and only the declaration above differs.
The outputs differ, so the extent of a parenless argument list is decided
by something that is not at the call site at all.

Measured 5.42.0, BOTH constants sit inside the inner pushmark: `pushmark`,
`pushmark`, `const[IV 1]`, `const[IV 2]`, `gv[IV \\&main::f]`, `entersub`,
`print`. So both are `f`'s. In the prototype case `const[IV 2]` falls
AFTER the `entersub` and belongs to `print`. That is the whole
difference, and it is visible in the optree as well as in the output.

`"f[@_]"` rather than a `join` because the body must not introduce a `(`
or a second `,`: the token facts assert on both, and a body that spelled
its own would make those assertions about the wrong punctuation. They are
asserted on the two argument literals rather than on `@_`, because
`"f[@_]"` is ONE Quote token -- the array never reaches the token stream,
and a fact about it would be satisfied vacuously.

WHY THIS REFUSES. Our parser reads `f` as a complete term and then finds
`1` with no operator between them, which is `trailing_tokens`: it has no
notion that a bareword followed by a list may be a call. The token facts
are what a fix has to keep true -- there is no `(` anywhere and exactly
one `,`, so nothing may quietly rewrite this into the parenthesised form
while claiming to have handled the parenless one.
"""),

    ('09_prototype_extent.t', 'A `($)` prototype cuts the extent to one', """
The same parenless call site under a callee with a `($)` prototype: the
extent is ONE argument, and the second falls through to the enclosing
`print`.

THIS IS HALF OF A MEASUREMENT. Set the two call sites side by side:

    sub f     { return "f[@_]" }   print f 1, 2;   # f[1 2]
    sub g ($) { return "g[@_]" }   print g 1, 2;   # g[1]2

The call sites differ only in the callee's name. Everything deciding how
far `1, 2` runs is in the DECLARATION -- its prototype, and indeed
whether there is a declaration at all, since an undeclared callee makes
the whole statement a syntax error. A parser cannot answer the extent
question from the call site's tokens, which is why the corpus measures it
rather than assuming it.

Measured 5.42.0, `const[IV 2]` is AFTER the `entersub`, so it is print's
argument and not `g`'s. Note `gv[IV \\"$"]` rather than
`gv[IV \\&main::g]`: with a prototype in force perl has resolved the call
differently again. Still one `entersub`, which is why the op lint cannot
tell these two cases apart and the outputs have to.

WHY THIS REFUSES: the same `trailing_tokens` as its pair, at the same
place and for the same reason. The prototype is not what our parser
stumbles on -- it has not got as far as caring.

The token facts assert that `($)` is NOT three punctuation tokens. A
lexer reading it as `(`, `$`, `)` would produce a stream in which this
looks like a parenthesised call, which is the other call form entirely.
"""),

    ('10_undeclared_callee.t', 'An undeclared callee is a syntax error', """
A parenless call to a callee perl has NOT YET SEEN is not an ambiguous
argument extent. It is a syntax error, and there is no call to have an
extent at all.

THIS IS THE PRECONDITION OF THE TWO EXTENT CASES ABOVE. Both of their
answers presuppose that perl read `f 1, 2` as a call; this is the
measurement that says when it does not. Measured 5.42.0:
`Number found where operator expected (Do you need to predeclare "f"?)`,
then `syntax error`, then `had compilation errors.`

THE `sub f` ON THE SECOND LINE IS THE POINT, not an oversight. The
obvious reading is that the definition below fixes the call above, and it
does not: perl parses top to bottom and the call site is read before the
declaration is reached. Delete that line and the error is identical,
which is why it is here -- a source that merely omitted the sub would
measure "undefined subroutine", a weaker and different fact, and one that
happens at RUNTIME rather than at parse time.

Measured the same with and without `use strict`, so this is the PARSER's
answer rather than strictness. The greedy case carries no `use strict`
either, which keeps the pair comparable.

THERE IS NO OUTPUT BLOCK. perl never runs a program it will not compile,
so there are no bytes to pin, and an EMPTY block would be the claim that
the program RUNS and prints nothing -- a different and false statement.

The token facts are what remains assertable, and they are the two the
extent pair assert, for the same reason: nothing may quietly read this as
the parenthesised form while claiming to have handled the parenless one.
"""),

    ('12_builtin_extent.t', 'A builtin\'s extent is cut by a paren', """
`warn "a", "b"` takes both arguments; `warn("a"), "b"` takes one and
leaves the other to the enclosing list.

THIS IS THE BUILTIN HALF of what the two user-sub cases measure. There
the extent was decided by the DECLARATION -- `f` greedy, `g ($)` cut to
one -- and the call sites were byte-identical. Here there is no
declaration to consult: `warn` is a builtin and its extent is decided at
the CALL SITE, by a paren or its absence.

Measured 5.42.0, `my @greedy = (warn "a", "b")` is 1 element and
`my @cut = (warn("a"), "b")` is 2. `warn` returns 1, so the greedy form
yields ONE value -- both strings went to `warn` -- and the cut form
yields TWO, the `1` and the `"b"` the paren pushed out of `warn`'s reach.
Printed as counts, `12`.

THE OPTREE SAYS THE SAME THING AND THE LINT CANNOT READ IT. Measured,
`const[PV "b"]` sits INSIDE warn's arguments in one and AFTER the `warn`
in the other, exactly the shape the parenless pair records for `f` and
`g`. `opsOf` collects op NAMES in execution order and the lint compares
SETS, so it sees `const`, `const`, `warn` twice over and cannot say which
side of the operator either constant fell on. The counts carry the claim.

HOW A WARNING BECAME OBSERVABLE is the whole reason this measures `warn`
and not `die`. The runner compares STDOUT byte for byte and `warn` writes
to STDERR, which nothing here reads. `local $SIG{__WARN__} = sub { }`
installs a handler, which perl calls INSTEAD of writing to STDERR, so the
text goes nowhere and `warn`'s RETURN VALUE is the only thing left to
observe. The handler is EMPTY on purpose: one that printed would put the
warning text on STDOUT and the pin would then depend on perl's message
format -- `a b at FILE line N.` -- which carries the file's own path and
line number. The count does not.

`die` ASKS THE IDENTICAL QUESTION AND IS NOT HERE. 218 of T1's 986 files
use it, so its absence is a gap rather than a decision about relevance.
Measured, there is no spelling that makes it observable: `$SIG{__DIE__}`
does not prevent termination --
`local $SIG{__DIE__} = sub { print "D" }; die "x"; print "never"` prints
`D` and exits 255, and a program that dies has no STDOUT to pin -- and a
`die` under a runtime-false guard is a DEAD branch, which says nothing
about how far its arguments ran. The spelling that works is
`eval { die ... }`, and `eval` is not yet in the corpus. When a tier
claims it, `die`'s extent case can be written against this one as its
pair.

THE COUNTS THIS CASE WOULD MOST LIKE TO ASSERT ARE NOT EXPRESSIBLE. There
are TWO words spelled `warn`, TWO `(` and TWO strings spelled `"b"` --
one of each on either side of the question -- and the grammar admits only
`one` and `no`, so every count naming the construct directly would be
FALSE of a correct lex. So the facts are written on what IS unique, and
both are about the machinery that makes the measurement possible rather
than about the extent itself: `__WARN__` pins that the handler is
installed EXACTLY ONCE (two would mean the second silently replaced the
first; none would mean the warnings went to STDERR and this measured
nothing while still printing `12`), and `local` pins the dynamic scope
that makes the handler apply to the two statements after it.
"""),

    ('13_lock.t', '`lock` is a named unary with its own op', """
`lock` takes a VARIABLE, emits its own op, and -- absent actual
contention -- does nothing observable at all.

WHY IT IS IN THIS TIER, SAID PLAINLY: `lock` is not about subroutines. It
declares no sub, calls no sub, and has no argument protocol, and its op
budget is so small that almost any tier could hold it -- `const padsv
padsv_store print pushmark` are all tier 01's, so nothing about the ops
places it. Given a free choice it would sit with the unary operators in
04_operators. It is recorded here rather than argued into a theme it does
not fit.

WHAT IT ACTUALLY MEASURES is the parse. `lock $x` is a named unary whose
operand is a variable and whose value is never used, in statement
position -- the shape a parser is most likely to mis-handle by treating
the keyword as an ordinary bareword function call. Measured, it is not
one: `lock` is in the op stream as an op of its own. A parser that
compiled it to a subroutine call would emit `entersub` and a `gv` naming
a sub that does not exist, and the program would die at run time.

THREADS, AND A CLAIM THIS CASE DOES NOT MAKE. It is tempting to say
`lock` works "without threads". That is not what was measured. The pinned
interpreter is a THREADED build -- measured, `useithreads=define`,
`x86_64-linux-thread-multi` -- so what is established is that `lock`
compiles and runs whether or not `threads.pm` has been loaded, and
nothing about an unthreaded perl. Uncontended, the statement is a no-op
with respect to output, which is precisely why it pins `$n` rather than
anything about locking: there IS no observable effect, so the only honest
assertion is that the statement compiled, ran, and left the program
otherwise unchanged.

`$ENV{N}` is unset when the runner executes, so `$n` is 7 by the `//`
default -- the corpus idiom for a runtime value, used here for this
tier's own reason as well as the general one: a `lock` on a folded
constant would be a weaker measurement of the operand's arity.

THE TOKEN FACT is the discrimination against the other spelling. There is
no `unlock` in perl -- a lock is released when its scope exits -- and a
parser inventing one as the obvious counterpart would introduce a word
this case says is absent.
"""),
])

covered |= topic(TIER, 'adjacency-07_subroutines.md',
                 'Every construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to 06_control, the tier this one depends on.

%s

The tier's other cases are one construct each, which is what makes them
diagnosable: when the signature case refuses, the construct that refused
is the only one present. That same property is why a corpus of such cases
cannot reach an ADJACENCY bug -- a parser that handles every construct
alone and mis-handles a pair goes green over the pair.
""" % INTRO, [
    ('00_adjacency.t', 'The whole tier in one body', """
Present here: a named sub, a signature with a default, `@_` read
positionally, `@_` WRITTEN through, five call forms, an anonymous sub,
`return` both early and trailing, `lock`, and the builtin extent pair.

THE ALIASING WRITE IS THE ONE CONSTRUCT WHOSE EFFECT IS VISIBLE FROM
OUTSIDE THE SUB. `bump($seen)` returns nothing anyone looks at, and
`$seen` is 1 in the printed line only because `$_[0]++` reached the
caller's variable. It sits next to a signatured sub on purpose: `pick`
binds its arguments by signature and `bump` by `@_`, so both of the
tier's argument protocols are in one body, which is the pair a parser
handling each alone can still get wrong. They are two subs rather than
one because, measured 5.42.0, reading `@_` inside a signatured sub warns
-- `Use of @_ in scalar with signatured subroutine is experimental` --
and a warning on stderr is noise this corpus has no section to pin.

THE PAIRING WITH 06_CONTROL IS NOT DECORATION. The `for` loop with `next`
inside the body is the earlier tier's construct sitting directly against
this tier's `return`, which is the pair the DEPENDS ON line names. A
`return` reached from inside a loop has to unwind the loop as well as the
sub, and nothing else in the corpus puts those two exits next to each
other.

THE PROTOTYPE NEEDS A FEATURE TOGGLE, and the reason is the sharpest
single measurement here. `sub g ($)` is a PROTOTYPE only where the
signatures feature is OFF. This body says `use v5.36`, which turns
signatures on, and under it perl reads the same three characters as a
SIGNATURE and enforces arity instead -- measured 5.42.0,
`use v5.36; sub g ($) { "g" } print g 1, 2;` gives
`Too many arguments for subroutine 'main::g' (got 2; expected 1)`, while
without the pragma `sub g ($) { "g[$_[0]]" } print g 1, 2` prints
`g[1]2`. The same three characters, two different features, and only the
feature state decides which. The `no feature "signatures"` /
`use feature "signatures"` pair around `sub g` is what lets both readings
live in one body: `pick` is declared above it and keeps its signature,
`g` is declared inside it and gets a prototype.

A BLOCK WOULD HAVE BEEN TIDIER AND IS MEASURABLY WRONG.
`{ no feature "signatures"; sub g ($) {...} }` compiles the bare braces
as a loop -- `enterloop`, `stub`, `leaveloop` -- and `stub` is an op
11_oo introduces, so the dependency lint correctly refuses a file
reaching four tiers forward to declare a sub. The file-scope toggle emits
no ops at all.

THE LAST TWO PRINTED LINES ARE THE ADJACENCY THAT MATTERS. `print f 1, 2`
and `print g 1, 2` are the same call-site shape and print different
things: `f` is greedy and takes both arguments, while `g`'s prototype
cuts the extent to one and the `2` falls through to the enclosing
`print`. A parser that handled each case alone and got the pair wrong
would go green over two separate cases and fail here, which is the whole
reason an adjacency case exists.

`@greedy` and `@cut` stand the BUILTIN extent question beside the
user-sub one. `warn "a", "b"` is greedy and yields ONE value;
`warn("a"), "b"` is cut by the paren and yields TWO -- the `12` on the
third printed line. A user sub's extent is decided by its DECLARATION and
a builtin's by a PAREN at the call site, and a parser can implement
either and not the other. `local $SIG{__WARN__} = sub { }` is what makes
the pair observable at all: `warn` writes to STDERR, which the pinned
output does not read, and an installed handler is called INSTEAD of that
write, leaving `warn`'s RETURN VALUE as the only thing to count.

WHY THIS REFUSES, and why that is the corpus working rather than a
regression: it parsed clean while it held only parenthesised calls. The
two parenless call sites are what our parser declines, at the same
`trailing_tokens` site as the two extent cases -- it reads the callee as
a complete term and then finds a number with no operator between them.
Measured at dc1bea2c: three Unknown nodes, all `trailing_tokens`, where
each extent case alone produces one. The marker comes off when the
parenless form lands, and the extent cases' markers come off with it.

THIS IS ALSO THE TIER'S SHARPEST STATEMENT OF WHAT THE OP LINT CAN AND
CANNOT SEE. The source below contains a signature, a parameter default, a
loop, a `next`, three `return`s, an aliasing write through `$_[0]` and an
ampersand call, and the MAIN program's op stream contains exactly two ops
this tier introduces -- `anoncode` and `entersub` -- and not one op from
06_control either, because the loop is inside the sub too. Everything
else compiled into a CV that `-MO=Concise,-exec` does not print.
Measured with the sub bodies included, the same source adds seventeen
ops: `aelemfast and argcheck argdefelem argelem enteriter eq iter join
leaveloop leavesub lt multiconcat next postinc return rv2av unstack`.
That difference is the size of the blind spot, in one file.
"""),
])

check(TIER, covered)
