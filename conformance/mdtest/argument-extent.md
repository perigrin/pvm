# Argument extent

How far a parenless argument list runs. Three cases where the answer is
decided somewhere other than the call site, and one that establishes the
precondition for asking at all.

**Tier 07 subroutines.** Introduces `anoncode`, `argcheck`,
`argdefelem`, `argelem`, `entersub`, `leavesub`, `lock`, `return`, `warn`.
Depends on 06_control.

A USER SUB'S EXTENT IS DECIDED BY ITS DECLARATION AND A BUILTIN'S BY A
PAREN AT THE CALL SITE. Two routes to one question, and a parser can
implement either and not the other -- which is why both halves are here.

`lock` is in this topic for proximity and nothing else, and its own case
says so. It declares no sub, calls no sub, and has no argument protocol;
what it does have is a named unary's operand, which is the shape the rest
of this topic is about.

## A parenless call is greedy

A parenless call to a declared callee with no prototype consumes the
whole remaining list, even nested inside another list operator that looks
like it owns those arguments.

THIS CASE AND THE NEXT ARE A PAIR and neither means much alone. The call
sites are byte-identical but for the callee's name -- `print f 1, 2;`
here and `print g 1, 2;` there -- and only the declaration above differs.
The outputs differ, so the extent of a parenless argument list is decided
by something that is not at the call site at all.

Measured 5.42.0, BOTH constants sit inside the inner pushmark: `pushmark`,
`pushmark`, `const[IV 1]`, `const[IV 2]`, `gv[IV \&main::f]`, `entersub`,
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

```perl
sub f { return "f[@_]" }
print f 1, 2;
print "\n";
```

```behavior
parses: yes
refuses: 01a0c432-fbd5
refusal: trailing_tokens
```

```output
f[1 2]
```

```tokens
one operator whose text is ","
one numeric literal whose text is "1"
one numeric literal whose text is "2"
```

## A `($)` prototype cuts the extent to one

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
argument and not `g`'s. Note `gv[IV \"$"]` rather than
`gv[IV \&main::g]`: with a prototype in force perl has resolved the call
differently again. Still one `entersub`, which is why the op lint cannot
tell these two cases apart and the outputs have to.

WHY THIS REFUSES: the same `trailing_tokens` as its pair, at the same
place and for the same reason. The prototype is not what our parser
stumbles on -- it has not got as far as caring.

The token facts assert that `($)` is NOT three punctuation tokens. A
lexer reading it as `(`, `$`, `)` would produce a stream in which this
looks like a parenthesised call, which is the other call form entirely.

```perl
sub g ($) { return "g[@_]" }
print g 1, 2;
print "\n";
```

```behavior
parses: yes
refuses: 01a0c432-fbd5
refusal: trailing_tokens
```

```output
g[1]2
```

```tokens
no operator whose text is "("
no operator whose text is ")"
one operator whose text is ","
one numeric literal whose text is "2"
```

## An undeclared callee is a syntax error

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

```perl
f 1, 2;
sub f { return "f[@_]" }
```

```behavior
parses: no
```

```tokens
one operator whose text is ","
one numeric literal whose text is "1"
one numeric literal whose text is "2"
```

## A builtin's extent is cut by a paren

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

```perl
local $SIG{__WARN__} = sub { };
my @greedy = (warn "a", "b");
my @cut = (warn("a"), "b");
print scalar @greedy, scalar @cut, "\n";
```

```behavior
parses: yes
```

```output
12
```

```tokens
one word whose text is "__WARN__"
one word whose text is "local"
```

## `lock` is a named unary with its own op

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

```perl
my $n = $ENV{N} // 7;
lock($n);
print "locked ", $n, "\n";
```

```behavior
parses: yes
```

```output
locked 7
```

```tokens
one word whose text is "lock"
```
