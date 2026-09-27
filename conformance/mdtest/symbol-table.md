# Globs and the symbol table

`*name` writes a name, `\*name` reads one, and `AUTOLOAD` is what
happens when the lookup finds nothing.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

THESE SIT IN 10_io RATHER THAN 08_references BECAUSE OF WHAT THEY EMIT.
Measured, `*alias = sub {...}` emits `rv2gv`, which this tier claims for
its filehandles -- a glob and a filehandle are the same thing to perl,
which is the whole reason `open(my $fh, ...)` and `*STDOUT` live in one
namespace. Tier 08 owns references and could not have them. The three
cases reach one symbol table from three sides: one installs a name, one
references a name, and one fails to find a name.

The typeglob did not appear in this corpus at all before these cases.
Every `*` in it was either multiplication or a `sprintf` width
specifier -- `"%0*d"`, `"%*d"`, `"%-*d"` -- and ZERO were the sigil,
so nothing here could tell a lexer that treats `*` as always
multiplication from one that gets it right.

(An earlier version of this passage gave a count: three occurrences
across 1,141 lines. That was true when written and is not now -- the
corpus has grown and the number counted multiplications too. The claim
that survives is the KIND, not the tally, so it is stated that way.)

## `*` is a sigil and an operator, resolved by position

`*name` introduces a fourth namespace using the same character as
multiplication. The fork is a one-character lexical one and it is
resolved by POSITION -- a sigil where a term is expected, an operator
where one just ended, which is the expect-state mechanism the spec
describes for `%`, `<`, `&` and `/`.

BOTH READINGS APPEAR IN ONE STATEMENT PAIR. The sigil installs `sq`
into the symbol table; the two multiplications inside and after it are
ordinary arithmetic. A lexer that resolved `*` by looking only at the
character produces either a syntax error or a multiplication where the
installation belongs, and `9 6` is unreachable either way.

The installed sub is called by NAME on the next line, which is what
makes the installation observable: a parser could accept the assignment
and do nothing, and `sq($n)` would then be a call to an undefined sub.

THE TOKEN FACT CANNOT BE ABOUT THE SIGIL, and that is itself the
measurement. Our lexer emits `Operator("*")` for all three occurrences
-- the glob sigil and both multiplications -- so no count of `*`
separates them and a fact naming it would be a claim about the source's
punctuation. The glossary has no typeglob category to assert instead.
What the case can claim is the shape around it: `sub` appears exactly
once, in the anonymous constructor.

```perl
my $n = $ENV{X} // 3;
*sq = sub { $_[0] * $_[0] };
print sq($n), " ", $n * 2, "\n";
```

```behavior
parses: yes
```

```output
9 6
```

```tokens
one word whose text is "sub"
```

## `\*STDOUT` is a reference to a GLOB, its own type

Not SCALAR, not CODE, not the filehandle it names. Measured, `ref(\*STDOUT)`
is `GLOB` -- a type this corpus never named before, since tier 08
covers `SCALAR`, `ARRAY`, `HASH` and `CODE` and stops there.

THE BACKSLASH IS THE PART A PARSER CAN GET WRONG. `\*STDOUT` is a
reference-to-glob; `*STDOUT` alone is the glob itself, and `\*` is not a
compound operator -- it is tier 08's `\` applied to a term that happens
to start with `*`. A lexer that read `\*` as one token, or that read
`*STDOUT` as multiplication by a bareword, produces something `ref`
would not call `GLOB`.

`STDOUT` rather than a glob this case creates, because a bareword
filehandle is the one glob guaranteed to exist without installing
anything -- and it is this tier's own subject.

```perl
print ref(\*STDOUT), "\n";
```

```behavior
parses: yes
```

```output
GLOB
```

```tokens
one word whose text is "ref"
```

## `AUTOLOAD` catches a call to a sub that does not exist

A sub name special to the LANGUAGE rather than to the program: perl
calls it when a named sub cannot be found. The same symbol table seen
from the other side -- here the lookup fails and perl falls back.

`missing` is never defined. The call finds nothing, perl dispatches to
`AUTOLOAD`, and `$AUTOLOAD` holds the fully qualified name that was
sought -- `main::missing`, which the substitution trims to `missing`.

WHAT A PARSER GETS WRONG HERE IS THE CALL ITSELF. `missing()` has no
declaration anywhere in the source, so a parser that resolves calls at
parse time has nothing to resolve; one that treats an unknown bareword
followed by parens as a call gets it right and defers the question to
runtime, which is what perl does. `07_subroutines/10_undeclared_callee.t`
makes the same claim for an ordinary sub; the difference here is that
the sub genuinely does not exist and the program still works.

THE TOKEN FACT COUNTS ONE `AUTOLOAD` IN A SOURCE THAT SPELLS IT THREE
TIMES, and that is the claim rather than an oversight. Two of the three
are `$AUTOLOAD`, a Variable, and only the sub name is a bare Word. The
fact separates the special sub name from the special variable that
carries its argument -- different things wearing one spelling. A lexer
that let the sigil fall off, or that read the bare name as a variable,
fails it.

The `s///` is tier 09's op, reached rather than introduced: `$AUTOLOAD`
arrives package-qualified and the unqualified name is what makes the
output readable as a claim.

```perl
our $AUTOLOAD;
sub AUTOLOAD { my $n = $AUTOLOAD; $n =~ s/.*:://; return "auto:$n" }
print missing(), "\n";
```

```behavior
parses: yes
```

```output
auto:missing
```

```tokens
one word whose text is "AUTOLOAD"
```

## `foreach $pkg (LIST)` aliases a PACKAGE variable and restores it

The fourth side of the symbol table, and the one that is not a glob
spelling. `foreach` with a BARE variable rather than a `my` declaration
aliases the package slot itself for the body of the loop, which is why
`enteriter`'s operand here is `rv2gv` over a `gv[*i]` -- tier 10's op,
claimed for its filehandles -- where `foreach my $x` gets a pad slot and
`padsv`. Measured on 5.42.0:

    foreach my $x (@l)   ->  enteriter ... padsv[$x] LVINTRO
    foreach $i (@l)      ->  enteriter ... rv2gv <- gv[*i]

The alias is UNDONE at the exit, which is what the third `print` pins:
`$i` holds `"before"` again after the loop has assigned it twice, so the
loop localises the global rather than assigning to it. Nothing about that
is visible in the loop body, and a reader that took the bare spelling for
an ordinary assignment would print `"b"` on the last line.

THE ORDER OF THE HEAD IS WHAT THIS CASE PINS FOR A PARSER. The variable
comes before the parens and the list inside them -- `foreach $i (@l)`, not
`foreach ($i) @l` -- and the two spellings of the head differ only in
whether a declarator precedes the variable. Our parser read the bare
variable into the LIST slot and the list into a bare term after it, at
Unknown = 0, and emitted `for ($i) 2 {3;}`: a tree that is not a parse of
its source, which only a canon could see (`01a0e071`).

The token facts are the two spellings this case is about, and there is no
`my` fact even though the source spells one: `my @l` declares the LIST,
not the loop variable, so a fact counting `my` would assert the opposite
of the point.

```perl
our $i = "before";
my @l = ($ENV{A} // "a", "b");
foreach $i (@l) { print $i }
print "\n";
print "$i\n";
```

```behavior
parses: yes
```

```output
ab
before
```

```tokens
one word whose text is "foreach"
one word whose text is "our"
```
