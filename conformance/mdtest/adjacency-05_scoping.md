# Every scoping form, each beside another

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

## The whole tier in one body

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

```perl
use feature "state";
our $g = 1;
my $x = 2;
state $s = 3;
{
  local $g = $g + 10;
  my $x = $x * 2;
  state $s = $s + 100;
  print "$g $x $s\n";
}
print "$g $x $s\n";
```

```behavior
parses: yes
```

```output
11 4 103
1 2 3
```
