# use, require and import

`use Foo LIST` locates a file, loads it, and then calls
`Foo->import(LIST)`. `require` is the same construct with the import
step removed, and `import` is an ordinary method call. Those three
statements are this whole area.

**Tier 12 packages.** Introduces `require`. Depends on 07_subroutines.

`use` EMITS NOTHING -- it is a `BEGIN` block that has finished before
there is an optree to dump. `require` is the one construct here that
reaches run time and the one op this tier introduces: `require strict;`
emits `const[PV "strict.pm"] s/BARE` and then `require`, `require $m`
emits `padsv` and the same `require`. One op, several sources. `import`
emits `pushmark`, two `const`, `method_named` and `entersub` -- tiers
01, 07 and 11 between them, with nothing left over. There is no import
op; the reason `use` is heavier than `require` is a call the op stream
cannot distinguish from any other call.

Because both halves of `use` are compile-time, the cases observe them
through `%INC` and the symbol table rather than through anything the
loaded module emits. The `04`/`05` pair is where the tier's claim
actually lives: neither case alone separates loading from importing.

## `use strict;` lexes, and that is all it establishes

A pragma is a `BEGIN` block, finished before there is a runtime optree.
The op stream for this case is the stream for `my $x = "ok"; print
"$x\n"` alone, differing only in the feature bits printed on
`nextstate`, which are a field rather than an op.

WHAT THIS CASE DOES NOT REACH, stated plainly because the pin does not
say it. Measured: delete both `use` lines and this program still prints
`ok`. So the pin establishes that the lines LEX and the statement is
delimited, and it establishes nothing about the load or the import. A
pragma cannot do better -- its effect is lexical and compile-time, and
`$^H` and `${^WARNING_BITS}` consulted at run time report the CALLER's
scope, measured OFF inside the very file that turned them on.

It is kept because `use` with a pragma is the overwhelmingly common
spelling and a lexer that mis-delimited it would fail here. The case
that reaches the rest is the next one.

```perl
use strict;
use warnings;
my $x = "ok";
print "$x\n";
```

```behavior
parses: yes
```

```output
ok
```

## `use POSIX;` loads the file AND calls import

This is the case that makes `use` FALSIFIABLE, through a module with an
exporter and the symbol table. `use POSIX;` puts `floor` into `main::`,
and deleting the `use` line changes both pinned lines -- so the output
measures the statement rather than its absence of syntax errors.

POSIX is core, its import list is not versioned in any way this case
observes, and nothing here depends on what it prints, because it prints
nothing.

MEASURED perl 5.42.0:

    $ perl -e 'use POSIX; print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\n"'
    yesyes

```perl
use POSIX;
print "POSIX loaded: ", ($INC{"POSIX.pm"} ? "yes" : "no"), "\n";
print "floor imported: ", ($main::{"floor"} ? "yes" : "no"), "\n";
```

```behavior
parses: yes
```

```output
POSIX loaded: yes
floor imported: yes
```

## `use POSIX ();` loads the file and suppresses import

The empty parentheses are not an empty argument list but the absence of
one. Same module and the same two observations as the case above, one
statement apart:

    use POSIX;      loaded yes    imported yes
    use POSIX ();   loaded yes    imported no

The first column is `require`'s half of `use`, the second is `import`'s.
The two cases' op streams are byte-identical to each other and to the
two `print` statements alone, which is why the symbol table is where the
difference is visible at all. That is the tier's defining difficulty
stated as a measurement: `use` is real, and no optree records it.

A pragma could not stand in here, because a pragma's whole effect IS its
import and there would be nothing left to observe.

MEASURED perl 5.42.0:

    $ perl -e 'use POSIX (); print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\n"'
    yesno

```perl
use POSIX ();
print "POSIX loaded: ", ($INC{"POSIX.pm"} ? "yes" : "no"), "\n";
print "floor imported: ", ($main::{"floor"} ? "yes" : "no"), "\n";
```

```behavior
parses: yes
```

```output
POSIX loaded: yes
floor imported: no
```

## `require strict;` -- the tier's one runtime op

The bareword-to-filename rewrite is the compiler's: by the time the op
runs, `strict` has become the string `"strict.pm"`, emitted as
`const[PV "strict.pm"] s/BARE` and consumed by `require`. So `require`
takes a filename, never a package name, and the bareword spelling is
sugar the op stream cannot see.

`strict` is chosen because it is core, prints nothing, and its version
does not reach the output. The load is observed through `%INC` rather
than through anything the module itself emits.

MEASURED perl 5.42.0:

    $ perl -e 'require strict; print +($INC{"strict.pm"} ? "yes" : "no"), "\n"'
    yes

```perl
require strict;
print "loaded ", ($INC{"strict.pm"} ? "yes" : "no"), "\n";
```

```behavior
parses: yes
```

```output
loaded yes
```

## `require $m;` takes the filename from a variable

This is the case that shows the bareword form was sugar. `require
strict` emits `const[PV "strict.pm"] s/BARE` then `require`; this emits
`padsv` then `require`. One op, two sources, and only the operand
differs -- which is also why `require Foo::Bar` and `require
"Foo/Bar.pm"` are the same statement and `require $m` with `$m` holding
`"Foo::Bar"` is not.

MEASURED perl 5.42.0:

    $ perl -e 'my $m = "warnings.pm"; require $m; print +($INC{$m} ? "yes" : "no"), "\n"'
    yes

```perl
my $m = "warnings.pm";
require $m;
print "loaded ", ($INC{$m} ? "yes" : "no"), "\n";
```

```behavior
parses: yes
```

```output
loaded yes
```

## `import` is an ordinary method call

`Marker->import("tag")` emits `pushmark`, `const[PV "Marker"] sM/BARE`,
`const[PV "tag"] sM`, `method_named[PV "import"]`, `entersub` -- tiers
01, 07 and 11 between them and nothing left over. The claim is that
`import` is not special, and the way to demonstrate it is to call it
directly and get the same stream `use` would have produced.

Called at run time rather than from `BEGIN` so the output order is the
source order. Under `BEGIN` the same call runs during compilation and
prints first, which is true of any BEGIN block and is tier 05's
business, not this tier's.

MEASURED perl 5.42.0:

    $ perl -e 'package Marker; sub import { print "import $_[0] $_[1]\n" } package main; Marker->import("tag"); print "after\n"'
    import Marker tag
    after

```perl
package Marker;
sub import { print "import $_[0] $_[1]\n" }
package main;
Marker->import("tag");
print "after\n";
```

```behavior
parses: yes
```

```output
import Marker tag
after
```
