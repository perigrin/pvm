---
title: EXPR foreach LIST swaps its body and its list at Unknown=0, and only the spelling that refuses is correct
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T22:12:56.600355997Z
updated: 2026-09-26T22:12:56.600355997Z
---

## Measured finding

`EXPR foreach LIST;` — the statement-modifier form — builds a tree with the BODY
and the LIST SWAPPED, and scores **Unknown = 0** while doing it. The ratchet
cannot see this. Canon can, and canon is how it was found.

Measured at `2257e08e`:

```
                                        unk  canon
$s += $_ foreach 1..3;                   0   foreach ($s += $_) 1 .. 3
f() foreach @a;                          0   foreach (f()) @a
f($x) foreach @a;                        0   foreach (f($x)) @a
$s = 1 foreach @a;                       0   foreach ($s = 1) @a
$x++ foreach @a;                         0   foreach ($x++) @a
push @r, $_ foreach @a;                  0   foreach (push(@r , $_)) @a
test_op($a,$b) foreach qw(x);            0   foreach (test_op($a , $b)) qw(x)
```

Seven of seven. The modified expression has become the loop's LIST and the list
has become its BODY, and the terminating `;` is dropped. perl reads it the other
way round:

```
$ perl -MO=Deparse -e 'my @a=(1,2); print $_ foreach @a;'
print $_ foreach (@a);
```

The `print` spelling is the one exception and it is instructive — it scores
Unknown = 1 and canons CORRECTLY as `print $_ foreach @a;`. So the only
spelling that refuses is the only spelling that is right, and every spelling
that scores clean is wrong. A count over this corpus rewards the corruption.

`for` as a modifier behaves the same as `foreach`; `if`, `unless`, `while` and
`until` all score Unknown = 1 and canon correctly, so the defect is specific to
the two loop-list modifiers that take a LIST rather than a condition.

### How it was reached

Bucketing every dirty perl.git `t/` file's FIRST Unknown span by the construct
the refusal stumbled over left a residual group of 10 files and 265 nodes whose
first Unknown is a LONE `}` byte with nothing else refused nearby and a
well-formed block around it — the only group in the corpus where the "bare
closer is a cascade" hypothesis actually holds. `op/args.t` is the clean example:
two Unknowns in the whole file, both a single `}`, at offsets 1761 and 2076.
Walking into those blocks finds a statement-modifier `foreach` that over-consumed
and left the closer behind.

Corpus reach: 251 modifier-shaped `EXPR for/foreach LIST;` lines across
perl.git `t/`. Affected files whose FIRST failure is this shape:
`comp/parser.t`, `op/args.t`, `op/assignwarn.t`, `op/die_keeperr.t`,
`perf/optree.t`, `re/pat_advanced.t`, `io/msg.t`, `op/array.t`,
`op/tiehandle.t`, `re/pat_re_eval.t`.

## Acceptance criteria

- [ ] `$s += $_ foreach 1..3;` canon's as `$s += $_ foreach 1 .. 3;` — the expression is the BODY, the list is the LIST, and the `;` survives `go test ./internal/parse -run TestStatementModifierForeachKeepsItsBody -count=1`
- [ ] the same holds for the `for` spelling, for a parenless call (`f() foreach @a`), for an assignment, for a postfix `++`, and for a list operator (`push @r, $_ foreach @a`) — the five shapes measured wrong above `go test ./internal/parse -run TestStatementModifierForeachKeepsItsBody -count=1`
- [ ] `print $_ foreach @a;` scores 0 Unknown and KEEPS its correct canon — today it is correct at Unknown=1, so the fix must not trade a right tree for a clean count `go test ./internal/parse -run TestStatementModifierForeachKeepsItsBody -count=1`
- [ ] REGRESSION GUARD, passes today: `if`, `unless`, `while` and `until` modifiers already canon correctly and must be untouched by a change to the loop-list modifiers `go test ./internal/parse -run TestModifier -count=1`
- [ ] REGRESSION GUARD, passes today: the BLOCK forms `for (LIST) { ... }` and `foreach my $x (LIST) { ... }` are a different production and must not move `go test ./internal/parse -run TestForLoopForms -count=1`
- [ ] the perl.git `t/` ratchet is regenerated in the same commit and every mover verified individually; expect the count to RISE on files where a swallowed `}` currently hides a whole block, and explain each riser `go test ./internal/parse -run TestPerlGitTRatchet -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
