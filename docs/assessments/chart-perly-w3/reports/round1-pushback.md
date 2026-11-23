# Assess round 1: pushback lens

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` in
`/home/user/pvm/.claude/worktrees/cr-close`, blob
52c82af16a738eca6e03cdb050e9c167903212c7 (checked with `git hash-object`).
Method: paad:pushback v2.0.2, run non-interactively. The findings are listed
here and not walked through one at a time, because this round airs views and
does not decide them. Nothing in the worktree was edited. `make test` was not
run. To check §3, I ran `count.sh` against `/opt/corpora/perl5`, which is at
94e50866 with a clean tree. The script works in its own temporary directory.

## Recommendation: MODIFY

The direction is sound, and the evidence it gives is accurate where I
checked it (see "Verified" below). As written, though, the decision does not
yet say what must become true. Three things block acceptance:

- The value semantics in §4.2 are under-defined in two places, P1 and P2.
  The predecessor plan had already worked out both.
- The exit and stops in §6 can't be derived or measured as written (P5, P6,
  P7). §8.3 and §8.4 are preconditions of the exit, so they can't stay open
  questions.
- The decision amends a binding design rule without naming or rewriting it
  (P8).

## Reality check (source control)

The last 50 commits on this worktree hold nothing that contradicts the
spec's code claims. `builder.complete`, `builder.awaiting`, `builder.leo`,
`forest.Build`, `forest.Constrain` and `forest.NewPrecedence` all exist as
the spec describes them. The baseline keys in §2 match
`internal/chart/testdata/ratchet/baseline.txt` exactly.

One process point: the evidence for §3, `internal/chart/offline/precedence/`,
is untracked. It shows as `??` in `git status`, and it has no expected-output
file (see P12).

## Scope shape

The spec covers one feature, so it is cohesive. It is large, though:
recognizer, set interning, Leo, `forest.Build`, a table generator, three new
timing keys, a report-only mode for `Constrain`, and a removal gate. There is
a meaningful split:

1. **§6's stage keys and the derivation of an exit.** These are worth having
   on their own: the postmortem says "add stage keys for precedence, resolve
   and readings before any speed work". They also decide whether 2.5 can be
   reached at all (P5).
2. **The reduce side (§4.1).** It adds no value to items, so it needs no
   amendment of rule 12. It can ship and be measured alone.
3. **The shift side (§4.2), the forest tables (§4.3) and the removal of
   `Constrain`.** These need the rule-12 amendment (P8).

## Findings, by severity

### P1. §4.2: one value per item cannot express a constraint per rule (high)

**If it ships:** the implementer has to choose between two models. One
keeps a value per item, which drops or admits derivations wrongly wherever
an LR(0) state holds several dotted rules. The other keeps a value per
dotted rule, which is a different design with a different cost to set
identity.

**Because:** an item is (state, distance), and a state holds many dotted
rules. 51 of 772 chart states complete more than one rule. I verified that
count from `grammar/testdata/tables/states.txt`; §3 states it itself. §4.2
defines the value as "the set of top-operator precedence levels of the
item's derivations". In a state that completes both a unit rule and an
operator rule, that set mixes the two. "The parent's slot keeps the
members its table admits" also does not say which waiting dotted rule's
table applies, since one awaited item's state can wait on the same symbol
in several rules.

The predecessor plan worked through exactly this and decided the other way:
`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, on branch
`docs/chart-recovery-plan`, §C, C1, "What rides on an item". It says "A
single value per item could not kill one rule inside a state. So … the value
is per dotted rule." The decision reverses that without an argument.

**What would change my mind:** an argument that, in perly.y's tables, no
state where a shift-side check fires holds two kernel dotted rules waiting
on the completed LHS at different operand positions. That would need a
count like §3's.

### P2. §4.2: a value that widens after an item completes is never propagated (high)

**If it ships:** a derivation the table admits can be lost. Its parent was
already completed with the narrower value, so the parent never sees the
wider set. The result is a false refusal, which shows up as
`rd-accepted-chart-refused` above 0, or a stop.

**Because:** `builder.add` (recognize.go:376) deduplicates with `isNew` and
never revisits an item. §4.2 says alternatives "combine by bitwise OR, so
items stay merged and no derivation is lost". That holds only if an item
whose value grows is completed again. The recovery plan states this
requirement outright: "An item whose value grows must complete again …
The spike re-queues such an item until values stop changing." The decision
leaves it out, and it does not mention the cost of reaching a fixed point.

### P3. §4.2: "operator rule" and "unit rule" overlap, and one reading gives the wrong grouping (high)

**If it ships:** under one reading, a forbidden grouping survives
recognition.

**Because:** §4.2 says "a completion of an operator rule sets the value to
that rule's level; unit rules (`term: termbinop`) pass it up unchanged".
perly.y has unit rules that carry `%prec`, and so have a level:

- `termbinop: termrelop %prec PREC_LOW` (perly.go:440)
- `termbinop: termeqop %prec PREC_LOW` (perly.go:442)
- `termrelop: relopchain %prec PREC_LOW` (perly.go:464)
- `listexpr: term %prec PREC_LOW` (perly.go:360)
- `expr: listexpr %prec PREC_LOW` (perly.go:354)

Each of these is both a unit rule and an operator rule.

Take `$a << $b < $c`. Bison resolves the SHIFTOP rule against CHRELOP as
reduce: CHRELOP has 26 reduce pairs in my bison run. The shift-side table
therefore removes CHRELOP's level from the right operand of
`term SHIFTOP term`. The operand `$b < $c` reaches that slot through
`termbinop: termrelop %prec PREC_LOW`:

- If that rule sets PREC_LOW, the value is PREC_LOW and the check passes,
  so `$a << ($b < $c)` is derived.
- If it passes the value up, the value is CHRELOP and the grouping is
  blocked, as bison blocks it.

Nothing in the spec chooses between the two. The §3 count does not settle
it either. It classifies (rule, token) pairs. It does not check that
"token t's level" equals "the level of the operand's top rule" for every
pair, and the two can differ wherever a rule's level comes from `%prec`.

### P4. §4.2, §5.4, §5.5: no error path when the tables cannot be generated or cannot be checked (high)

**If it ships:**

- A resolution the generator cannot map is silently dropped or
  mis-mapped.
- 43 resolutions have table entries that nothing checks.
- §5.5's "a divergence is a defect in the tables" assumes a check exists
  that, for those 43, does not.

**Because:**

- Bison has 373 rules and the chart has 350 (§5.4; `rules.txt` has 350
  lines). The resolutions name rules the chart left out. One reduce
  resolution names `POSTJOIN`, which `grammar/perly.go:796` omits on
  purpose ("POSTJOIN is left out"). That is from my bison run: 1 reduce
  pair on POSTJOIN.
- The referee covers 1,304 of 1,347 resolutions (README, "The perly.y
  referee"). The other 43 are on `uncovered-reviewed.txt`.
- The spec does not say what the generator does with a resolution it
  cannot map:
  - Does the build fail?
  - Does the resolution go on a reviewed list?
  - Is the table regenerated byte-identically, with a key like
    `tables-drift`?
  - Is there a key for "table entries no sentence exercises"?

**Modify:** state what must become true. The generator refuses any
resolution it cannot map unless that resolution is on a reviewed list. The
tables are committed and drift-checked. No resolution is uncovered and
unlisted.

### P5. §6 and §8.4: the exit of 2.5 has no derivation and may be out of reach by the spec's own scoping (high)

**If it ships:** the waypoint can't finish, or it finishes after the exit
is lowered partway through, which is "tuned toward" in all but name.

**Because:** baseline aggregate 4.7915 = lex 0.655 + recognize 0.790 +
forest 0.662 + about 2.68 after the forest (§2's arithmetic, which I
checked). To reach 2.5, 2.29 has to come out:

- Lex is untouched by this work.
- §7 puts resolve and readings out of scope "beyond what the smaller
  forest gives them".
- Even if recognize and forest each halved, the post-forest share would
  have to fall from 2.68 to at most about 1.1, a cut of about 59%. Most of
  that has to come from removing `Constrain`, which only happens in the
  last issue (§4.3).

Nobody knows whether `Constrain` makes up 59% of the 2.68, because no key
times it. §6 itself orders those keys taken first. The only figure is the
postmortem probe's 115 ms on std5k, which is one file and not the
aggregate.

**Modify:** derive the exit from the stage keys once they are taken. For
example: precedence stage at or below some value, and the ratio at or
below the baseline minus the precedence stage plus a stated allowance.
Alternatively, make taking the keys and setting the exit a precondition
issue. §8.4 is a precondition, not an open question.

### P6. §4.3 against §6, stop 4: the stop fires on the behaviour the design predicts, and has no protocol for measuring it (medium-high)

**If it ships:** stop 4 is either a false alarm or never checked.

**Because:** §4.3 keeps `Constrain` running "until the ratchet shows the
tables alone agree", and removes it "in the waypoint's last issue". Until
then, `Constrain`'s cost stays, and §4.3 adds table work to `forest.Build`
on top. So "once both tables are in", the ratio is not expected to fall
below 4.79. Stop 4 checks exactly that moment.

It also gives no measurement protocol. Repeated quiet readings in the
postmortem varied: 5.18 against 5.11, and 8.56 against 8.39. The v2 design
uses the median of three readings. Stop 3's "set-reuse falls far enough
that…" is a causal claim. The key `speed-stage-recognize` above 1.0 can be
measured; the "because" cannot.

**Modify:**

- Place stop 4 after `Constrain` is removed, or measure it with
  `Constrain` off.
- Name the protocol, for example the median of three quiet readings.
- Write stop 3 as the threshold on the key alone.

### P7. §4.3 and §8.3: the gate for removing `Constrain` cannot be measured as written (medium-high)

**If it ships:** the last issue has no runnable criterion, or it removes
`Constrain` on evidence that covers 1,296 generated sentences and no real
corpus.

**Because:**

- §4.3 asks for the tables to agree with `Constrain` "on every corpus
  (`precedence-disagree` 0 with `Constrain` reporting and not removing)".
  `precedence-disagree` is defined over the kept generated sentences in
  `testdata/precedence/`, not over the corpora (README key table, line
  536).
- `Constrain` has no mode that reports without removing. It returns a
  bool and unlinks alternatives (README, "Refuse"; precedence.go:322).
- So both the key and the mode are unstated work.
- §8.3, whether a residue is left, decides whether `Constrain` can be
  removed at all. That makes it a precondition of P5's exit.
- "The waypoint's last issue" doesn't say which waypoint. w2 has closed,
  and this spec names no milestone.

**Modify:**

- Add a corpus key, for example "alternatives `Constrain` unlinks after
  the tables", which must be 0 before removal.
- Name the report-only mode as a deliverable.
- State what happens if a residue is found: is that a stop, or does
  `Constrain` stay?

### P8. §1 and §5.1: a binding design rule is lifted by quoting a ruling, not amended, and its evidence is not cited (medium-high)

**If it ships:** the merged code contradicts the live README and the
binding design, and no superseding record says so. Later review rounds can't
check what was ruled.

**Because:**

- `internal/chart/README.md` says the v2 design (2026-10-30) and the
  recovery plan bind. The spec's "Follows" line cites the v1 design
  (2026-10-29).
- The v2 design's §3 ("never inside Earley items") and its §10 kill "Facts
  in items" are still in force. The README states as live fact: "No fact
  is added to an Earley item (spec §3 rule 12, §10)".
- The recovery plan (§C, C1) names this exact move: "A successor design
  that amends the spec (rule 12, and §10's 'facts in items') … The spike's
  numbers are its evidence." C1 was to measure what §5.1 now defers:
  `set-reuse`, distinct sets, cache hits, and the operator share of the
  forest. The decision does not cite C1 or its result.
- §1 rests the lifting on "the full Option B and Option C plan". "Option
  B" and "Option C" appear in no committed document except this one, so
  nobody can check the claim "and only for the precedence value of §4.2".

**Modify:**

- Cite the binding documents correctly.
- State the amended text of rule 12 and of §10's kill.
- Define Options B and C, or quote the record that does.
- Cite C1's figures, or say C1 was not run.
- Under doc-first, the README invariant changes in the same pull request.

**What would change my mind:** a C1 report exists and these definitions
are in the milestone Records.

### P9. §5.2 and §8.2: Leo is a precondition, and the spec treats it as a shift-side issue only (medium)

**If it ships:** with Leo on, the reduce-side filter is skipped for the
intermediate links of a chain. Leo on and Leo off can then disagree on the
items they derive, and `leo-disagree` compares acceptance only, so it
would not notice.

**Because:**

- `builder.complete` calls `leo()` before `awaiting`, and a memo adds the
  top of the chain in one step.
- The existing lookahead case avoids this through a stated monotonicity:
  "FOLLOW only shrinks up a chain" (leo.go, comment on `leo`). The spec
  claims no matching property for precedence blocks.
- The recovery plan already settled the shift side: a memo stores no
  value, multiply along the walk, and a guard compares values with Leo on
  and off.
- The spec leaves the choice open in §8.2 and holds only `leo-disagree`.

**Modify:** pick the mechanism. State what makes the reduce filter at the
top of a chain sound, or complete link by link. Add a guard comparing
values with Leo on and off.

### P10. §4.2 against §8.1: 33 bits in a `uint32` (medium; the value contradicts itself)

**If it ships:** the "no operator" bit collides with a level, or the type
changes partway through implementation.

**Because:** `grammar/perly.go:744-806` declares 32 levels, and the
`PLUGIN_*` levels are already among them: `PLUGIN_LOW_OP`,
`PLUGIN_REL_OP` and `PLUGIN_HIGH_OP` have levels of their own, and the
other plugin tokens share a level. §4.2 adds a reserved bit for "no
operator" on top of those 32, which needs 33 bits. §8.1 can be answered
from the repository, and the answer is that the plugin levels are counted
and the type is too small.

**Fix:** use `uint64`, or a bit only for the levels some reduce or error
pair actually names.

### P11. §5.3: the ambiguous-lexeme residue is left undecided between two homes (medium-low)

**If it ships:** "a check at the next scan or a residue for the forest's
tables" are two designs. Only the second interacts with P7's removal gate,
and if neither is built, a derivation through a blocked terminal survives
into the forest.

**Fix:** choose one, and state which check proves the derivation dies.

### P12. §3: the cited evidence is not committed (low)

**If it ships:** `count.sh` and `count.pl` are untracked and have no
committed expected output, so §3's figures cannot be reproduced from `pu`.

I reproduced all of §3 with the script: 1347 pairs, 57 states, 0
inconsistent, shift 669 global and 1 state-dependent, reduce 667 global,
error 10 global, and the state-dependent pair rule 370 PERLY_BRACKET_OPEN
at state 153. The script needs a perl5 checkout at the pin and bison 3.8.2.

**Fix:** commit both files with an expected-output file in the same pull
request, as `offline/referee` does.

## Material only an implementing issue needs

- The details of the semiring in §4.2 ("bitwise OR", "multiplies") and the
  name of `builder.complete` and `awaiting(src, lhs)` in §4.1 are
  implementation detail.
- The decision should instead state the properties that must hold:
  - every derivation bison allows is kept;
  - every grouping bison forbids is absent from the chart and the forest;
  - Leo on and off agree on values;
  - every resolution is either in a table or on a reviewed list.
- Each of §5's points 1 to 5 is either a property of that kind or a
  measurement, and should be written as one.

## Verified and discarded (8 candidates)

- §2: the baseline values and the arithmetic (stage sum 2.107; remainder
  about 2.68). Correct.
- §3: the bison count, all figures. Reproduced.
- §3: 51 of 772 states complete more than one rule; state 439 reduces 271
  and 275. Correct.
- §4.1: `builder.complete` has the completing state, its LHS and the
  lookahead bitset. Correct (recognize.go:419).
- §5.1: `set-reuse` is 0.704. Correct.
- §4.2: Chalk commit 287e5da1 exists in `/home/user/perigrin/chalk`. I did
  not check the semantics of Chalk's `add`.
- "Held" omits `superset-recognized`. Discarded: its population is files
  recursive descent accepts, so `rd-accepted-chart-refused` 0 covers it.
- The waypoint-2 completion date is given as 2026-11-23, and the postmortem
  was generated on 2026-11-20. Discarded: the milestone may have closed
  later.
