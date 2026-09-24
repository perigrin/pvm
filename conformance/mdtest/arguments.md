# Arguments

How a sub receives its parameters. Two protocols: `@_`, which is tier
02's array populated by the call, and a signature, which is the one
construct in this tier the optimiser does NOT erase.

**Tier 07 subroutines.** Introduces `anoncode`, `argcheck`,
`argdefelem`, `argelem`, `entersub`, `leavesub`, `lock`, `return`, `warn`.
Depends on 06_control.

The two protocols cannot share a sub without noise. Measured 5.42.0,
`@_` inside a signatured sub is populated but reading it warns --
`Use of @_ in scalar with signatured subroutine is experimental` -- and
a warning on stderr is not something this corpus pins.

## Arguments arrive in `@_`

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

```perl
sub describe {
    return $_[0] . "-" . $_[1] . "-" . scalar(@_);
}
print describe("a", "b", "c"), "\n";
```

```behavior
parses: yes
```

```output
a-b-3
```

## A signature names the parameters

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

```perl
use v5.36;
sub add_up ($a, $b = 3) { $a + $b }
print add_up(1), " ", add_up(1, 10), "\n";
```

```behavior
parses: yes
```

```output
4 11
```

## `@_` aliases the caller's variables

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

```perl
sub bump_alias {
    $_[0]++;
}
sub bump_copy {
    my ($n) = @_;
    $n++;
}
my ($aliased, $untouched) = (1, 1);
bump_alias($aliased);
bump_copy($untouched);
print "$aliased $untouched\n";
```

```behavior
parses: yes
```

```output
2 1
```

## Explicit return, early and trailing

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

```perl
sub classify {
    return "zero" if $_[0] == 0;
    return "negative" if $_[0] < 0;
    return "positive";
}
print classify(0), " ", classify(-5), " ", classify(7), "\n";
```

```behavior
parses: yes
```

```output
zero negative positive
```
