# Every construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another.

**Tier 10 io.** Introduces nothing of its own; the mixture is the
subject. Depends on 03_context.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why such a corpus cannot reach an
ADJACENCY bug: a parser handling every construct alone and mishandling
a pair goes green over the pair.

## The whole tier in one body

The pairs that matter are the ones a one-construct case cannot make.

A scalar readline followed by a LIST readline on the SAME handle is the
whole of this tier's context dependency in two adjacent statements --
the second reads what the first left, so the printed `2` is only
correct if both contexts were resolved and resolved DIFFERENTLY. The
two readline cases each open a fresh handle and so can never disagree
about position.

The second pairing is `eof` inside a print LIST that also holds an
array, which puts the tier's own op in an argument position rather than
alone in a statement.

THE THIRD PAIRING IS `select` WITH `say`, and it is the one that needs
two constructs most. `say "round trip"` names NO HANDLE, and its bytes
land in `$buf` rather than on stdout, because the `select($out)` above
it changed where "no handle" points. Neither case alone can make that
claim: the `say` case passes its handle explicitly and the `select`
case redirects a `print`. Here the two constructs are the same
assertion -- a parser that dropped either one puts `round trip` on
stdout and leaves the buffer empty.

The four-argument `select` follows, the other operator sharing that
name: measured, `select($out)` emits `select` and
`select(undef,undef,undef,0)` emits `sselect`, a different op reached
by a different argument count. Both spellings are here because the
tier's declared set holds both and an adjacency case must reach every
op the tier introduces.

The pairing with 03_context is the scalar-versus-list readline itself:
context is not a separate construct to place beside this one, it is the
thing selecting which readline happens.

The ops this reaches beyond the tier's own, all claimed earlier and
none new: `gv`, `padav` and `aassign` (02), `cond_expr` and `goto` from
the ternary (06), `undef` from the syscall arguments (04), `srefgen`
from `\my $buf` (08).

`use feature "say"` is required and is not decoration: without it `say`
is not this tier's op at all but a method call. Measured, the pragma
changes no op in this body beyond enabling `say` itself.

```perl
use feature "say";
open(my $in, "<", \"first\nsecond\nthird\n");
my $head = <$in>;
my @rest = <$in>;
print STDOUT $head;
print STDOUT scalar(@rest), " left, eof ", (eof($in) ? "yes" : "no"), "\n";
close($in);
open(my $out, ">", \my $buf);
my $prev = select($out);
say "round trip";
select($prev);
close($out);
my $ready = select(undef, undef, undef, 0);
our $AUTOLOAD;
sub AUTOLOAD { my $n = $AUTOLOAD; $n =~ s/.*:://; return "auto:$n" }
*sq = sub { $_[0] * $_[0] };
print $buf, "ready $ready sq ", sq(3), " glob ", ref(\*STDOUT), " ", missing(), "\n";
```

```behavior
parses: yes
```

```output
first
2 left, eof yes
round trip
ready 0 sq 9 glob GLOB auto:missing
```
