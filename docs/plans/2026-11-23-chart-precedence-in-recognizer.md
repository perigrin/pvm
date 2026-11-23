# Chart parser: precedence in the recognizer

Status: proposed, for crochet:assess. Date: 2026-11-23.
Follows: `docs/plans/2026-10-29-chart-parser-design.md` and waypoint 2
(`chart-perly-w2`, completed 2026-11-23, postmortem
`docs/postmortems/chart-perly-w2.md`).

## 1. Rulings this design rests on

- perigrin, 2026-11-23: "Rather than half steps let's implement the full
  Option B and Option C plan". This lifts waypoint 2's stop condition "a fix
  needing facts in Earley items" for this work, and only for the precedence
  value of §4.2.
- perigrin, 2026-11-23, on how to run it: "Full pipeline": assess to a fixed
  point, refine into a milestone, chain-review, execute, review, postmortem.
- perigrin, 2026-11-23: Chalk's speed numbers are not a comparison for this
  work ("it's implemented in pure Perl and its grammar isn't as
  comprehensive as ours"). Chalk is a reference for the mechanism only.

## 2. Problem

Waypoint 2 met its correctness exit and missed on speed. From the committed
ratchet baseline on `pu` (`internal/chart/testdata/ratchet/baseline.txt`),
chart time over recursive descent's, long-lived:

| Key | Value |
|---|---|
| `speed-ratio-aggregate-longlived` | 4.7915 |
| `speed-ratio-std5k-longlived` | 7.2212 |
| `speed-stage-lex-aggregate-longlived` | 0.6551 |
| `speed-stage-recognize-aggregate-longlived` | 0.7903 |
| `speed-stage-forest-aggregate-longlived` | 0.6620 |
| `readings-collapsed` | 835 |
| `alloc-per-token-chart` / `-rd` | 9.327 / 2.077 |

Each timed stage alone is under recursive descent's whole parse; their sum is
about 2.1. The remaining 2.7 is after the forest: the precedence constraint,
resolve and the readings count, which no key times separately. The
postmortem's probe put about 115 ms of `std5k` in the precedence constraint
against about 25 ms for lex and recognize.

The cause is where precedence is applied. The chart recognizes perly.y with
its precedence stripped, so every operator chain is ambiguous: the README
records `$a = $a = …` at quadratic item counts with Leo on or off, "because
precedence is not applied yet". The forest packs every grouping, and
`forest.Constrain` then removes all but bison's. The published approaches
(Afroozeh et al. 2013, 2016; Marpa's prioritized rules; Chalk's precedence
semiring) apply precedence before or during recognition, so a forbidden
grouping is never derived.

## 3. The count

`internal/chart/offline/precedence/count.sh` runs bison 3.8.2 on the pinned
perly.y (94e50866) with default reductions off and classifies each of bison's
precedence resolutions, a (rule, token) pair, by what applying it needs.

| bison's resolution | Blocks | Pairs | Global | State-dependent |
|---|---|---|---|---|
| shift | completing the rule before the token: the reduce side | 670 | 669 | 1 |
| reduce | the rule's last operand having the token's operator on top: the shift side | 667 | 667 | 0 |
| error (`%nonassoc`) | both | 10 | 10 | 0 |

A pair is global when applying it in every state blocks nothing bison allows
in another state. No pair is inconsistent across the 57 states that hold a
resolution. The one state-dependent pair is rule 370, `indirob: scalar`,
before `[`: bison resolves it in the state after `print` and the same
reduction is legal after `@` (`@$array[1..3]`, T1's `array.t`).

On the chart's own tables (`grammar/testdata/tables/states.txt`), 51 of 772
states complete more than one rule, and one completes two rules with one LHS
(`term ARROW CAST_STAR`, state 439, no precedence). So a completion step
knows the rule it completes.

## 4. Design

### 4.1 The reduce side: a lookahead filter at completion (no value)

perly.y names each binary operator in its own alternative
(`termbinop: term ADDOP term`), so when such a rule completes, the rule is
the operator. `builder.complete` already has the completing state, its LHS
and the lookahead bitset. The filter is one table, generated from the
grammar alone: per (state, LHS), the terminals before which bison never
reduces it. Completion skips an LHS when every terminal the lookahead admits
is in that set.

The `indirob` pair is keyed by the parent item, which completion has in hand
(`awaiting(src, lhs)` enumerates them): the reduction is blocked into
`LSTOP indirob …` and allowed into `@ indirob`. No item carries anything new.

### 4.2 The shift side: a top-operator set carried by items

The shift side needs, when a parent rule completes, the operator on top of
its last operand. perly.y loses it at the unit rule `term: termbinop`: by the
time `term` completes into the parent, the item says "a termbinop" and not
which. So the operand's top operator travels as a value.

- **The value** is the set of top-operator precedence levels of the item's
  derivations: perly.y declares 32 levels, so one `uint32`. A derivation
  with no operator on top contributes the empty set's identity, a reserved
  bit for "no operator".
- **Semiring operations.** Alternatives of one item combine by bitwise OR,
  so items stay merged and no derivation is lost. A completion multiplies:
  the parent's slot keeps the members its table admits; an empty result is
  zero and the item is not added.
- **Where the value is set.** A completion of an operator rule sets the
  value to that rule's level. Unit rules (`term: termbinop`) pass it up
  unchanged. Any other completion, and a scan, sets "no operator".
- **The table** is per (parent rule, operand position): the levels a top
  operator may have there. It is generated from bison's resolutions, not
  written by hand: "reduce" and "error" pairs remove levels from the last
  operand's slot.

This is Chalk's mechanism (`Chalk::Bootstrap::Semiring::Precedence`, read at
287e5da1) narrowed: Chalk's grammar has one `BinaryExpression` rule for every
operator, so it needs a value on both sides; perly.y needs it only on the
shift side. Chalk's `add` keeps a single value per item, preferring the more
constraining; this design keeps the set, so it cannot lose a derivation a
later check would have allowed.

### 4.3 What the forest does

The forest is built from the chart's items by span. If a parent item exists
through one admitted operand, a forbidden operand of the same span and LHS
would be packed beside it. So `forest.Build` applies the same two tables
when it packs a derivation. `forest.Constrain` stays as the referee's check
until the ratchet shows the tables alone agree with it on every corpus
(`precedence-disagree` 0 with `Constrain` reporting and not removing); it is
removed in the waypoint's last issue.

## 5. Integration points the assessment must check

1. **Set identity and the transition cache.** Sets are interned by their
   relative items, and `set-reuse` is 0.704. A value per item changes set
   identity: two sets with equal items and different values are different.
   How much reuse survives is a measurement, not a guess.
2. **Leo.** A Leo memo completes a right recursion in one step and skips the
   intermediate completions where the shift-side check runs. Either Leo
   memos carry the value, or a chain whose levels differ is completed link
   by link, as the lookahead case already is.
3. **Ambiguous lexemes.** A token's `TerminalSet` can admit several
   terminals. The reduce-side filter skips a completion only when every
   admitted terminal blocks it; when they disagree it completes, and the
   derivation through the blocked terminal must still die. That is either a
   check at the next scan or a residue for the forest's tables.
4. **Rule mapping.** bison's rule numbers (373) are not the chart's (350).
   `forest.NewPrecedence` already reads levels, `%prec` and the refused rules
   from the grammar and builds bison's states from the tables. The generator
   reuses it, so the tables have one source.
5. **The precedence referee.** `precedence-disagree` 0 and
   `precedence-covered` 1304 are the equivalence proof. A divergence is a
   defect in the tables, never in the referee.

## 6. Measurement and exit

Existing keys carry the result; the precedence stages get their own keys
first, as the postmortem asks ("add stage keys for precedence, resolve and
readings before any speed work"):

- New timing keys: `speed-stage-precedence-aggregate-longlived`,
  `speed-stage-resolve-aggregate-longlived`,
  `speed-stage-readings-aggregate-longlived`, taken on `pu` before any
  recognizer change, so the gain is measured against them.
- Held at their baseline: `oracle-wrong` 0, `rd-accepted-chart-refused` 0,
  `chart-accepted-both-refused` 0, `precedence-disagree` 0,
  `precedence-uncovered-unlisted` 0, `leo-disagree` 0,
  `oracle-exact-chart` not below 8524.
- Expected to fall, and reported: `items-per-token`, `readings-collapsed`,
  `alloc-per-token-chart`, every speed ratio.

Proposed exit, for perigrin to set: `speed-ratio-aggregate-longlived` at or
below 2.5, with the correctness keys above unchanged.

Proposed stops, reported and never tuned toward:

- any correctness key above rises;
- a precedence resolution the two tables cannot express without the parser
  state, beyond the `indirob` pair;
- `set-reuse` falls far enough that `speed-stage-recognize` rises above 1.0;
- `speed-ratio-aggregate-longlived` does not fall below 4.79 once both
  tables are in.

## 7. Out of scope

- Grammar rewriting (Option A): it grows the grammar, which is Earley's cost
  (perigrin, 2026-11-23).
- Speed work on resolve or the readings count, beyond what the smaller
  forest gives them.
- Interleaved lexing, incremental reparse and M2 (design §13).

## 8. Open questions

1. Is a `uint32` enough once `PLUGIN_*` levels are counted, or are they
   separate levels?
2. Does the Leo interaction (§5.2) need a value in the memo, or is
   link-by-link completion of mixed-level chains cheap enough?
3. Does the forest-side table (§4.3) make `forest.Constrain` redundant on
   every corpus, or is a residue left, and what is in it?
4. What exit ratio does perigrin want? 2.5 is a proposal.
