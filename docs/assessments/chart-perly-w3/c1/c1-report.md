items-per-token: 4.302 4.242
items-dropped-per-token: 0.013
distinct-sets: 10500 11481
set-reuse: 0.501 0.502
cache-hit-rate: 64.8% 64.3%
verdict-agreement: 708/708
grouping-agreement: 695/695
resolutions-agreeing: 716/716
unaligned-sentences: 0
enumerations-cut: 0
unnarrowed-differences: 21
operator-share: 47.9%
leo-guard: pass
leo-walk-share: 1.8% to 11.0% over four non-quiet runs (10.3% in this run)
nullable-precedence-rules: 1

C1 report: precedence as a value on items (chart-recovery, plan §3 C1, ruling D7)
=================================================================================

Taken 2026-11-13 by agent:rce4e435-8 on origin/pu 7c12cd51 (P2's set-identity
golden merged), in a scratch worktree at
/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/chart-recovery/c1,
branch scratch/c1-precedence-spike, commit 9a6e84479 (amended from
3dfdca202 after gate pass 1 to drop its Implements: trailer, since no commit
implements C1). Nothing was pushed,
opened or merged. The worktree is removed once this report is filed, as C1's
last criterion requires. The commit stays reachable from the local branch
scratch/c1-precedence-spike in /home/user/pvm. That branch is never pushed.

In each pair above, the first figure is without values and the second is with
them. Every timing here was taken non-quietly: the 1-minute load average was
10 to 19 throughout, with other workers' runs on the machine.


What was built (scratch only)
-----------------------------

- recognize/prec.go holds the value algebra. A context is §8's exposed
  operator context:
  - the first terminal;
  - the junction tokens on the left spine (the first terminal of each second
    child down the left spine);
  - the rules on the right spine of the last child, and that child's junction
    tokens.
- Multiply runs at scan and at completion:
  - It kills the context where a rule on the right spine of the left part has
    a `shift` or `error` resolution against the next child's first terminal.
  - Completion kills it where the completing rule has a `reduce` or `error`
    resolution against a junction token on the left spine of its last child.
- Add is set union. Only the terminals and rules some resolution names are
  kept in a context, so the contexts stay finite.
- Values per dotted rule:
  - A kernel item carries a vector with one entry per dotted rule of its state.
    Each entry is an interned set of contexts.
  - A predicted item carries none.
  - Each transition maps source to target dotted rules through Tables.Items,
    dots moved over nullable symbols.
  - That map stayed tractable: it is built lazily per (state, symbol).
- An item whose every entry is empty is dropped.
- An item whose value grows is completed again: the builder's index loop
  became a queue. This happened 3,197 times over the referee and corpus runs.
- Sets are interned as (core, distances, values).
- Leo multiplies link by link along builder.leo's walk. The memo holds no
  value, and Chart.LeoLink is unchanged.
- Bison's resolutions are mapped onto table rules by rule text from
  `bison --report=solved` of the pinned perly.y. 1,303 of the 1,347 resolutions
  map. The other 44 all involve POSTJOIN, which the port excludes.
- The tree side evaluates the same algebra bottom-up on each derivation from
  the unchanged forest (forest.Build over the narrowed lattice,
  `Forest.Derivations` with a limit of 1000).
- Driver, run from internal/chart in the scratch tree (about 18 s for everything):

  ```
  S=/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/chart-recovery
  env -u T1_CORPUS -u TDT_FIXTURES C1_SPIKE=1 C1_BISON_OUTPUT=$S/bison/perly.output \
      C1_AGG=$S/c1-agg.txt go test . -run '^TestC1Spike$' -count=1 -v
  ```

  Optional control: `C1_MUTATE=none|flip|norequeue`.
- Inputs and where they live:
  - `$S/bison/perly.output`: bison 3.8.2, `bison -v --report=solved -o perly.c perly.y`,
    run on a copy of the pinned perly.y (/opt/corpora/perl5 at 94e50866) in `$S/bison/`.
    The same command sentences.sh uses for bison-conflicts.txt.
  - `$S/c1-agg.txt`: the aggregate's file list, written by the scratch
    ratchet's `C1_AGG_OUT` hook (cmd/ratchet/lex.go, right after
    `aggregateFiles`): `env -u T1_CORPUS -u TDT_FIXTURES C1_AGG_OUT=$S/c1-agg.txt go run ./internal/chart/cmd/ratchet`
    from the scratch tree's root.
  - The referee fixtures are read unchanged from internal/chart/testdata/precedence/.
- With values off, the spike leaves the recognizer as on pu:
  TestSetDigestGolden passes at 9a6e84479 (rerun after the amend, unregenerated).


Nullable check (done first)
---------------------------

`nullable-precedence-rules` is 1. One precedence-bearing rule derives the
empty string: `optlistexpr: empty %prec PREC_LOW` (perly.y:1787, bison's rule
336). Bison reports no resolution on that rule, so the ε-DFA skipping past it
loses no multiply that bison would have made.

The issue asked for this check first, so a nonzero count was a stop to
report. I continued on my own reasoning (no conflict names rule 336), without
a ruling. perigrin should confirm that waiver. Three other flagged rules have a nullable
left-hand side but a non-nullable right-hand side, so the ε-DFA never skips
them.


Agreement with the bison referee
--------------------------------

- **Verdicts (narrowed input):** 708 of 708 agree.
- **Grouping:** 695 of 695 accepted sentences agree. Each has exactly one
  surviving tree inside the sentence's span, and that tree's rule-and-span
  pairs are bison's.
- **Resolutions:** all 716 covered resolutions agree.
- No sentence was unaligned once the leading GRAMPROG was dropped, and no
  enumeration was cut.
- **Cross-check:** the recognizer with values accepted a sentence exactly
  where some tree survived the tree-side evaluation. There were 0
  disagreements.
- **Un-narrowed input:** 21 sentences differ, and every difference is in
  grouping:
  - Verdicts are identical. Each differing sentence keeps 2 or 3 trees
    because a word has another reading: `foo` in 19, `print` in 1, `use` in 1.
  - Under the plan this is scope's to collapse later, not §10's stop.

**Controls.** These show that the agreement figures can fail:

- **Every resolution removed:** verdicts 695/708, grouping 0/695,
  resolutions 0/716.
- **shift and reduce swapped:** verdicts 159/708, grouping 0/695,
  resolutions 12/716.
- **Widened items not completed again:** the referee figures stay at 708, 695
  and 716, but the Leo guard fails on one sentence. So on this corpus the Leo
  guard catches a missing re-completion. The cross-check and the agreement
  counts do not.


Leo
---

The guard passes on both the referee sentences (narrowed and un-narrowed) and
all 1,391 aggregate files. Leo on and Leo off agree on acceptance and on the
value of every item present in both runs.

`leo-walk-share` is not stable on this machine. My run read 10.3%. The gate's
three runs read 1.8%, 11.0% and 2.4%, at load 11 to 16. So the observed
range is 1.8% to 11.0%. A quiet median of three was not possible.


Measured on the ratchet's aggregate
-----------------------------------

The aggregate is 1,391 files: the superset population less std5k.pl,
edit5k.pl and T1/code_too_large.t. It was dumped from the scratch ratchet's
own `Aggregate` call.

How each figure was taken:

- **items-per-token, set-reuse and items-dropped-per-token:** a fresh
  recognizer per file, over accepted files, as the ratchet takes its keys.
- **distinct-sets:** sets interned by one long-lived recognizer over all files.
- **cache-hit-rate:** the long-lived recognizer's cached nodes as a share of
  its nodes with a set.

What the figures show:

- **Items per token** fall slightly, from 4.302 to 4.242. A zero value drops
  0.013 items per token.
- **Distinct sets** rise 9.3%, from 10,500 to 11,481.
- **Set-reuse** is unchanged: 0.501, then 0.502.
- **The cache hit rate** falls by half a point, from 64.8% to 64.3%.
- **The value interners** end with 4,048 contexts, 5,429 context sets and
  4,953 vectors.
- **Recognition time** across the aggregate is longer with values. The ratio
  is not stable: my run read 2.2× (0.708 s without, 1.581 s with), and the
  gate's three runs read 2.4×, 2.9× and 1.7×. So the observed range is 1.7×
  to 2.9×, all non-quiet at load 10 to 19. The memos are unoptimised maps.
  Time is not on plan §3's which-way list. It is reported for context, not
  as an argument.

**Operator share.** In the unchanged forest, 47.9% of the packed alternatives
come from a rule that is an operator rule or carries %prec. Packed
alternatives here are those at nodes with two or more alternatives: 7,589 of
15,829. Over every alternative the share is 45.5% (144,753 of 318,165).


One corpus over-refusal: a finding, not a fixed defect
------------------------------------------------------

With values, the recognizer refuses one aggregate file that it accepts
without them: `T1/array.t`. `perl -c` reports `syntax OK`.

- **Where it dies:** line 52, `my @slice = @$array[1..3];`, at the `[`.
- **The resolution that kills it:** rule 370 against PERLY_BRACKET_OPEN,
  resolved as shift. Rule 370 is `indirob: scalar %prec PREC_LOW`. The
  resolution forbids reducing `$array` to indirob with `[` next.
- **Why bison does not refuse the line:** bison's conflict exists only in
  states where a shift on `[` is also viable, such as after LSTOP
  (`print $fh [`). After a cast `@` only indirob is expected, so bison reduces
  with no conflict.
- **Why the spike does:** its value is keyed on (rule, token) alone, so it
  applies the resolution in every state.
- **Bison's states:** in perly.output, rule 370 `indirob: scalar .` sits in
  state 153 alone (default reduce, no conflict) and in state 193 beside
  `subscripted: scalar . PERLY_BRACKET_OPEN …`. Only in state 193 is there a
  conflict on `[`.
- **What this means for §8 itself, not only the successor design:**
  - The tree-side evaluator (treeValue, the §8-as-written path) uses the same
    Mul and Complete as the recognizer.
  - It kills array.t's only reading too, so `@$array[1..3]` would be refused
    as a file under §8 as written.
  - So §8's value ("the context at both edges") is not bison-exact on its
    own. The prohibition also needs to know whether a shift was viable there:
    which LR state the reduction happened in.
  - This bears on both outcomes. Flagged to perigrin as evidence on §10's stop
    "precedence not expressible", in the form written: with (rule, token)
    contexts, the edge value is not enough.
- **Possible fixes, not built:**
  - **Tree side (§8 as written):** the remedy is cheaper there. Forest
    evaluation can ask the chart, once per reduced node, whether the set at
    its right edge holds an item awaiting the token with the reduced node's
    origin (bison's state 193 against 153). That is a read of a finished
    chart, outside any item.
  - **Recognizer side (successor design):** the same test must run inside
    multiply, which sees only values. It needs either the set's contents at
    the reduction, or the shift's viability carried as one more fact in the
    context. The second grows the value. That is exactly what the spike
    exists to measure, and it is not measured.
- **Why the referee missed it:** no referee sentence reaches this state pair,
  so 716 of 716 does not exercise it.
- **What it is:** on the plan's terms, a spike defect at the corpus level. It
  is reported, not fixed here. It is not a disagreement on narrowed referee
  input, so it is not the cross-check's §10 signal.


Which way the numbers argue (plan §3: no threshold)
---------------------------------------------------

The numbers barely move:

- sets +9.3%;
- cache hits −0.5 points;
- set-reuse flat;
- items per token slightly down.

Meanwhile about half (47.9%) of the forest's packed alternatives are operator
or %prec alternatives. By the plan's reading, those argue for a successor
design that allows one finite precedence value on items.

Against that, or on both sides:

- **The state-insensitive value over-refuses** the `@$array[...]` form. This
  holds on both sides, since the tree evaluator over-refuses it too. The
  remedy is cheaper on the forest side, where the finished chart can be
  read. On the recognizer side the value must grow to carry the state.
- **Time:** recognition with values took 1.7× to 2.9× as long, and the Leo
  walk took 1.8% to 11.0%. Both are non-quiet and outside plan §3's
  which-way list.

Whether that refinement stays finite and cheap was not measured. perigrin
rules between §8 as written and a successor design (D7).


What was not done or not measured
---------------------------------

- No quiet timing. leo-walk-share and the time ratio are given as observed
  ranges over four non-quiet runs (mine and the gate's three).
- The over-refusal above is not fixed.
- The context is approximate where a rule's last right-hand side symbols are
  nullable and empty: the reduce check reads the last nonempty child.
  Agreement does not show this mattering.
- PLUGIN_* and POSTJOIN resolutions are not exercised. No rendering reaches
  them.
