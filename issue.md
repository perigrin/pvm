---
title: EXPR foreach LIST swaps its body and its list at Unknown=0, and only the spelling that refuses is correct
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T22:12:56.600355997Z
updated: 2026-10-11T02:47:56.581941105Z
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

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `$s += $_ foreach 1..3;` canons as `$s += $_ foreach 1 .. 3;` at Unknown=0, the body first and the `;` kept, as perl Deparses it. Fixed by 34737c46.

Three claims are reworded because measurement showed them wrong:

- `print $_ foreach @a;` was "correct at Unknown=1" only because canon echoes an Unknown's bytes. It now parses at Unknown=0 and canons as `print($_) foreach @a;`.
- `if`, `unless`, `while` and `until` did NOT already canon correctly. Measured at 652f4792, `$y = 1 if $x;` came back as `if ($y = 1) $x` at Unknown=0: one canon case produced all six, and 34737c46 fixed all six.
- The block-form guard also names TestBlockFormLoopsKeepTheirParens, which pins `if (COND) BLOCK` and its siblings in the same canon case.

The test is TestStatementModifierKeepsItsBody; TestStatementModifierForeachKeepsItsBody never existed. The old commands were bare backticks that verify does not run, and their perl fragments in parens would have run as shell.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] `$s += $_ foreach 1..3;` canons as `$s += $_ foreach 1 .. 3;`: the expression is the BODY, the list is the LIST, and the `;` survives (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] the same holds for the `for` spelling, a parenless call, an assignment, a postfix `++`, and a list operator, the five shapes measured wrong above (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] `print $_ foreach @a;` scores 0 Unknown and canons as `print($_) foreach @a;` (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] `if`, `unless`, `while` and `until` modifiers, which shared the defect, canon with the body first (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] REGRESSION GUARD: the BLOCK forms `for (LIST) { ... }`, `foreach my $x (LIST) { ... }` and `if (COND) { ... }` keep their shape (`go test ./internal/parse/ -run '^TestForLoopForms$' -count=1 -v | grep -q '^--- PASS: TestForLoopForms' && go test ./internal/parse/ -run '^TestBlockFormLoopsKeepTheirParens$' -count=1 -v | grep -q '^--- PASS: TestBlockFormLoopsKeepTheirParens'`)
- [ ] the perl.git `t/` ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] the whole tree is green (`go test ./... -count=1 -timeout=30m`)
