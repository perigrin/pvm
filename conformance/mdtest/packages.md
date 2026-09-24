# The package statement

`package NAME;` and `package NAME { ... }` -- the two spellings of the
construct that moves the compiler's notion of the current package.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

NEITHER SPELLING EMITS AN OP. `package Foo;` contributes no runtime op
at all; its only trace is the package name inside the `nextstate` that
follows, which is a field of an op tier 01 already claims. The block
form emits `enterloop`, `stub` and `leaveloop` -- a bare block, which is
tier 05's construct, plus tier 11's placeholder. Nothing in either
stream says `package`.

So both cases pin the construct by a FULLY QUALIFIED CALL instead.
Delete the `package` line and the sub lands in `main::`, the call is to
an undefined subroutine, and perl dies. That is what a tier of no-op
constructs has in place of an op check: every pinned line is falsifiable
by deleting a statement.

## `package Foo;` switches package for the rest of the file

`package Greet;` switches the current package for the rest of the
enclosing block or file. The whole effect is on the compiler, so the
call `Greet::hello()` is the only evidence the statement took effect --
a parser that dropped it would put `hello` in `main::` and this program
would die on an undefined subroutine.

MEASURED perl 5.42.0:

    $ perl -e 'package Greet; sub hello { "hello" } package main; print Greet::hello(), "\n"'
    hello

```perl
package Greet;
sub hello { return "hello" }
package main;
print Greet::hello(), "\n";
```

```behavior
parses: yes
```

```output
hello
```

## `package Foo { ... }` scopes the switch to a block

The same construct as the statement form with a scope attached, and the
spelling `class Foo { ... }` also uses -- which is why tier 11 could not
be ordered after this tier. Its three ops belong to tiers 05 and 11;
nothing in the stream says `package`.

MEASURED perl 5.42.0:

    $ perl -e 'package Louder { sub shout { "HELLO" } } print Louder::shout(), "\n"'
    HELLO

```perl
package Louder {
    sub shout { return "HELLO" }
}
print Louder::shout(), "\n";
```

```behavior
parses: yes
```

```output
HELLO
```
