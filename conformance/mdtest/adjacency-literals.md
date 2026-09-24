# Every literal, each beside another

One body holding every spelling this tier's cases introduce, each
adjacent to another.

**Tier 01 literals.** Introduces nothing of its own; it is the mixture
that is the subject. Depends on nothing, so there is no earlier tier to
pair with -- the adjacency is entirely within 01.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one
construct each, which is what makes them diagnosable: when the leading
point refused, the construct that refused was the only one present. That
same property is why a corpus of such cases cannot reach an ADJACENCY
bug -- a parser that handles every construct alone and mishandles a pair
goes green over the pair.

MEASURED perl 5.42.0, and this is not hypothetical:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses; `ADJUST` followed by anything does not. No
one-construct-per-case corpus can ever see that, because every case is
one construct by definition.

`TestTierLiteralsAdjacency` reads the tier's construct sources for the
literal each binds and requires the spelling to appear here, so this
body cannot fall behind the tier by a construct. It cannot check the
ADJACENCY itself; see the note in that test about `padrange` absorbing
`pushmark`, which is why more adjacent constructs emit FEWER ops.

## The whole tier in one body

Every numeric spelling the tier introduces -- binary, decimal,
hexadecimal, leading point, negative, octal by leading zero, octal by
prefix, signed exponent, trailing point, underscore separators, and both
v-string forms -- plus an integer, a single-quoted string, an
interpolating string and a `q` list.

This body refused from 7711154e until issue
01a0c13f-97f5-7f98-b32d-07245ec6ddfe, and it refused for a BORROWED
reason: the `.5` on line 4 was the leading-point gap reaching here, not
a second bug. The note recorded at the time said an adjacency case
cannot pass while any construct it holds refuses, and that composing
only the working constructs would make it green and make it stop
covering the tier. Fixing the lexer cleared both in one change, which is
that prediction coming true.

The tier's other two refusals -- the exponent split and the v-strings --
are LEXICAL and produce no Unknown at all, so they leave no code here to
name. See `TestTierLiteralsRefusalsCited` for why a case must not name a
refusal it does not have.

`qw(a b c)` prints as `abc` rather than `a b c`: in a print LIST the
three words are separate arguments and `$,` is unset, so nothing
separates them. Binding them to an array instead would print `a b c` --
but it would also emit `aassign`, `padav` and `join`, which are
02_variables' array machinery and unclaimed here. The lint caught that,
and this is the version that keeps `qw` in the tier that owns it.

```perl
my $bin = 0b1010;
my $dec = 0.5;
my $hex = 0xff;
my $lead = .5;
my $neg = -1;
my $oct = 0377;
my $octp = 0o377;
my $exp = 5e-1;
my $trail = 1.;
my $usep = 4_294_967_296;
my $vb = 65.66.67;
my $vv = v65.66.67;
my $int = 42;
my $sq = 'plain';
my $dq = "$int-$dec";
my $esc = "a\nb";
my $qop = q(one);
print "$bin $dec $hex $lead $neg $oct $octp $exp $trail $usep $vb $vv $int $sq $dq ", qw(a b c), " $qop ", $esc, "\n";
```

```behavior
parses: yes
```

```output
10 0.5 255 0.5 -1 255 255 0.5 1 4294967296 ABC ABC 42 plain 42-0.5 abc one a
b
```
