# Declaring and calling

A subroutine is declared, and then reached. This topic is the two halves
of that: where the CV comes from -- a named declaration, an anonymous
expression -- and the four spellings that call it.

**Tier 07 subroutines.** Introduces `anoncode`, `argcheck`,
`argdefelem`, `argelem`, `entersub`, `leavesub`, `lock`, `return`, `warn`.
Depends on 06_control.

`entersub` COVERS EVERY CALL FORM, and the differences are flags:
`f(1)`, bare `f`, `&f` and `&f(2)` all emit it, the ampersand forms
carrying `/AMPER` and the code-reference call `/STRICT`. A lint that
reads op NAMES sees one construct where the language has four, which is
why this tier needs one case per form rather than one op claim.

## A named subroutine, declared and called

`sub f {...}` installs a CV under a package name and `f(1)` calls it
through that name.

THE DECLARATION EMITS NOTHING into the main program. It is compile-time:
the CV is built and installed, and the main optree that runs afterwards
contains only the call -- `pushmark`, the `gv` naming the sub, and
`entersub`. Measured 5.42.0, `sub f { $_[0] + 1 } print f(1), "\n"`
gives `gv[IV \&main::f]` rather than `gv[*f]`, because the call resolved
to the CV at compile time: the sub was already declared when the call was
compiled. A call compiled BEFORE the declaration gets `gv[*f]` and looks
the name up at run time. Same op either way.

The body, where this tier's subject actually lives, is a separate optree
that `-MO=Concise,-exec` never prints.

```perl
sub double { $_[0] * 2 }
print double(21), "\n";
```

```behavior
parses: yes
```

```output
42
```

## An anonymous subroutine is a value

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

```perl
my $triple = sub { $_[0] * 3 };
print $triple->(14), "\n";
```

```behavior
parses: yes
```

```output
42
```

## Four spellings of one call

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
gave `gv[IV \&main::f]`: the parser had to decide whether `f` was a call
at all. Still one `entersub`.

```perl
sub answer { "[@_]" }
sub outer {
    print answer(), "\n";
    print answer, "\n";
    print &answer, "\n";
    print &answer(), "\n";
}
outer(1, 2);
```

```behavior
parses: yes
```

```output
[]
[]
[1 2]
[]
```

## Calling through a code reference

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

```perl
my $code = sub { return "c[@_]" };
print $code->(1, 2), "\n";
print $code->(), "\n";
```

```behavior
parses: yes
```

```output
c[1 2]
c[]
```

```tokens
no operator whose text is "-"
no operator whose text is ">"
```
