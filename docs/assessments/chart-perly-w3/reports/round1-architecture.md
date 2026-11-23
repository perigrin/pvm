# crochet:assess round 1: architecture and direction lens

**Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`. I checked the blob with
`git hash-object`: 52c82af16a738eca6e03cdb050e9c167903212c7. Worktree cr-close, at pu 8c3317079 plus that
file and `internal/chart/offline/precedence/`, both untracked.
**Lens:** whether the design fits the decisions now in force and the direction of the chart work, and
whether its measurement and exit plan are sound. I did not write the subject.
**Recommendation: modify.**

## Read

- The subject, all of it. `internal/chart/offline/precedence/count.sh` and `count.pl`.
- The waypoint 2 spec, v2 (`refine-w2/spec.md`): §2, §3, §8, §9.3–9.5, §10, §11, §12 and §13.
- `docs/plans/2026-10-29-chart-parser-design.md`, the first attempt: §3 and §9, and its section list.
- `internal/chart/README.md`: Status, the output contract, The rules, the key table, The speed protocol,
  Leo, The forest and The precedence constraint.
- `internal/chart/testdata/ratchet/baseline.txt`.
- The chart-perly-w2 milestone body, through `git zhi milestone show`: Design Rationale, Records and Acceptance
  Criteria. I also grepped chart-perly-w2-inputs for the 2026-11-23 ruling.
- Postmortems:
  - `chart-perly-w2.md`, both parts, with the What Will We Change sections in full;
  - the What Will We Change sections and verdicts of `chart-recovery.md`, `chart-waypoint.md` and `chart-perly-w1.md`.
- `docs/assessments/chart-waypoint.md`: its structure and round 1.
- The chart code:
  - `internal/chart/precedence.go`: `referee.judge` and `CountPrecedence`;
  - `internal/chart/forest/forest.go`: `parseStream`;
  - the headers and test list of `internal/chart/recognize/reference_test.go`, `agree_test.go` and `fuzz_test.go`.
- Chalk origin/HEAD 287e5da1, the first 150 lines of `lib/Chalk/Bootstrap/Semiring/Precedence.pm`.

## Not examined

- I did not run `count.sh`, so I did not reproduce the 670/667/10 and global/state-dependent counts.
  I read `count.pl`'s logic only. That is the codebase lens's job.
- I did not check the claims about the chart's states (51 of 772; state 439).
- I did not run the ratchet, any test or `make test`, as the brief said.
- I did not read the Afroozeh et al. papers or Marpa. My point about the 2016 paper's parameters at
  prediction (A3) is from memory of the paper. It is not checked.
- I did not read the full assessment minutes for v2 (`docs/assessments/chart-perly-w1.md` sits on the branch
  `docs/chart-parser-v2-design`, not on pu).
- I did not read the chart-recovery plan (D1–D12). I took D1 ("No limit, report only") from the chart-perly-w2
  Records of 2026-11-15 and the README.
- I did not find where "Option B" and "Option C" are defined, and I did not search the transcripts. The
  orchestrator's scratchpad was too large to grep in time.
- I did not search for the 2026-11-23 rulings outside the chain.

## What the design does well

- §6 puts stage keys for precedence, resolve and readings before any recognizer change. The chart-perly-w2
  postmortem asks for exactly this (part 2, What Will We Change: "add stage keys for precedence, resolve and
  readings before any speed work"). The chart-recovery postmortem shows why it works: "Stage keys first made
  every gain attributable".
- §5.1 says that a value per item changes set identity, and that the loss of reuse is "a measurement, not a
  guess". That is the honest framing of the first attempt's §3 rule 2 ("facts in Earley items would make
  otherwise-identical sets differ, which kills set interning").
- §4.3 keeps `forest.Constrain` until something shows it redundant, rather than removing it at once.
- §4.2 narrows Chalk's mechanism to perly.y's shape and gives the reason. Chalk carries one value per side for
  one `BinaryExpression` rule; perly.y names each operator's rule. Keeping the OR of the set, where Chalk keeps
  the more constraining value, is a sound fix: no derivation is lost.
- §7 excludes Option A with a ruling and a reason.

## Findings

### A1. §1, §4.1 and §4.2: the design reverses more accepted decisions than §1 lifts, and the ruling it rests on is not recorded where rulings live

§1 lifts one thing: "waypoint 2's stop condition 'a fix needing facts in Earley items' ... only for the
precedence value of §4.2". The accepted v2 spec decided more than that stop, and §4.1, which "carries no
value", reverses those decisions too:

- v2 §2: "`perly.y`'s conflicts are kept as ambiguity and resolved in the forest (§8), not inherited as an
  LALR table". §4.1 and §4.2 generate the recognizer's filters "from bison's resolutions", which is that LALR
  resolution, moved into recognition.
- v2 §8, under the heading "Precedence: in the forest, as bison resolves it": "precedence leaves recognition
  (perigrin, 2026-10-30): recognition uses `perly.y`'s rules as written, is Boolean, and is ambiguous on
  operator chains". It also says "Imitating that resolution inside an Earley recognizer would be faithful, not
  fitted." That faithful-versus-fitted test came from the chart-waypoint postmortem's first change: "Assess
  asks a component-fit question."
- v2 §3 rule 12: "The recognizer stays Boolean and its items stay (state, distance)." The first attempt's §3
  rule 2 says "Never inside the chart".
- v2 §10 lists "Facts in items" as a kill of up-front mode, the spike's question. It is not only a waypoint 2
  stop. The w2 milestone restates it as a stop (Design Rationale, "Stops (§10)").

The design names none of v2 §2, §8 or rule 12. The ruling it quotes ("Rather than half steps let's implement
the full Option B and Option C plan") is not in chart-perly-w2's Records, which end at the 2026-11-23 rulings on
the breaches, nor in chart-perly-w2-inputs. The document does not define Option B or Option C, so a reader
cannot tell whether the ruling covers §4.1, §4.3 or the removal of `Constrain`.

**Consequence if it ships:** refinement and review will judge the issues against v2. They will find §4.1
reversing §8 with no ruling behind it, or they will let it through and reverse an accepted ruling silently. The
waypoint 2 design chose the forest deliberately, for fidelity to bison's state-dependent resolutions (README,
"The parser states are bison's ... never the chart's split states"; the `indirob`, `twostates` and
`indirobsort` toys). The design replaces that with global tables, on the strength of a count no participant has
yet reproduced.

**Modify:**
- List every accepted clause the design reverses: v2 §2's sentence, §8's 2026-10-30 ruling, rule 12, and the
  §10 kill.
- Give each one the ruling text and where that ruling is recorded.
- Define Options A, B and C in the document.
- Record the 2026-11-23 rulings in a milestone's Records.

### A2. The header and §7: the governing decision is misidentified

The header says "Follows: `docs/plans/2026-10-29-chart-parser-design.md`". That is the first attempt, which v2
supersedes (v2's status line: "the successor ... not an amendment"). §7 cites "design §13", but the 2026-10-29
document has 12 sections and its future work is §11. §13 exists only in v2. v2 is not on pu: it is on the branch
`docs/chart-parser-v2-design`, as the w2 milestone body says. So the decision this design amends is neither
cited nor reachable from pu.

**Consequence:** A1's check, against the clauses actually in force, cannot be made from the document. A reader
following "Follows" reads superseded rules.

**Modify:** cite v2 by path and branch, or bring it onto pu first, and state the amends relation.

### A3. §2 and §4.2: the mechanism does not deliver the property §2 claims, so "expected to fall" is unfounded for the commonest chains

§2 justifies the move this way: published approaches "apply precedence before or during recognition, so a
forbidden grouping is never derived". The design's checks all run at completion, bottom-up. Prediction is
unchanged and context-free.

Take a left-associative chain such as `$a . $b . $c . $d`:
- `term` is still predicted at every operand's origin.
- Each sub-chain from origin i to j is a legal standalone term (`($b . $c) . $d`), and bison reduces it before
  `.` (left-associative). So neither the reduce-side filter nor the value kills it.
- Only the parent completion `$a . [$b . $c . $d]` is blocked, by the shift-side value.

Items from every origin to every later position survive, so the item count for these chains stays quadratic.
Only right-associative chains (`$a = $a = …`, the README's probe) and chains where a tighter operator follows
lose items, and they lose them through §4.1's lookahead filter. This is my reasoning. I have not measured it.
The 2016 data-dependent approach, as I recall it (not checked), passes the constraint down at prediction, which
this design does not do.

**Consequence:** `items-per-token` and `set-reuse` may move far less than §6 expects for `.`, `+`, `&&` and `||`
chains, which are common in real Perl. The exit could then fail for a reason the design did not anticipate.

**Modify:**
- State the invariant the design achieves. A candidate: "no item completes into a parent through an operand
  bison would not build there". Then say what is not cut: items from inner origins.
- Before refinement, add a probe beside the README's `$a = $a` counts: item counts for a left-associative chain
  of 50, 100 and 200 operators.

### A4. §6: three new ratchet keys with no ruling, and no lifecycle

The output contract says "no key is added, renamed or retired after it", and every exception names a ruling:
- `lex-overhead-aggregate`, perigrin 2026-11-07;
- the stage, allocation and GC keys, recovery plan D2 and D5;
- the ten w2 keys, KEYS, "Yes, in one issue", 2026-11-10;
- the retired std5k cold key, 2026-11-14.

The postmortem's request is not a ruling, and §6 cites none.

The design also does not say what `speed-stage-precedence-aggregate-longlived` times once `Constrain` is removed
in the waypoint's last issue (§4.3). There is then no precedence stage. The contract says "a number may not
become `na`", and retiring the key needs a ruling.

Until that last issue, the timed pipeline (`forest.ParseWith`, which calls `parseStream`, then `f.Constrain`)
pays for precedence three times: in recognition, in `Build`'s tables and in `Constrain`. Every intermediate
speed reading carries that double work.

**Consequence:** the key-set change is unauthorised under the contract, and the precedence stage key's end state
is undefined.

**Modify:** ask perigrin for the key addition explicitly, with all three keys in one issue as KEYS did, and
include the precedence key's retirement or redefinition when `Constrain` goes.

### A5. §2 and §6: the 2.5 exit is ungrounded, has no reading protocol, and quietly re-gates speed

**The figure.** In the baseline, the three stage keys already sum to 2.107 (0.6551 + 0.7903 + 0.6620). At 2.5,
0.39 is left for resolve, readings, any `Constrain` residue, and any rise in recognition from lost set reuse. The
design names that rise as a risk (§5.1), and its own stop allows recognize up to 1.0, which by itself would
exceed the exit. Resolve and readings are untimed today, so nobody knows whether 0.39 holds them. §7 also puts
speed work on them out of scope, so if they need more than 0.39 the exit cannot be reached by any work in scope.

**The precedent.** chart-recovery grounded its exit in a projection: §4's factors projected 1.98–2.29, against a
quiet before-reading on the same host, and it ended with a three-run verdict reading. The chart-perly-w2
postmortem asks for the same: "Spend the speed budget at refinement. Give each stage a share of the remaining
ratio", and "Check whether a successor can fit at all".

**The inference.** §2's residual of "2.7 is after the forest" comes from subtracting stage medians from a
whole-pipeline median. The README says the stage sum "is the sum of the three keys' medians, each against its own
recursive-descent passes, not one ratio". So the 2.7 is an inference, and the design's premise rests on it before
the keys that would test it exist. The 115 ms probe was on std5k, not the aggregate.

**The regate.** D1, extended by perigrin on 2026-11-15 ("dropping the speed gate was universal"), removed speed
ceilings. An exit at 2.5 re-introduces one. The design defers the figure to perigrin but does not say:
- that it re-gates speed;
- what a missed exit means (a stop, or a report);
- how it is read: one ratchet run, or checkpoint-3 style, two runs with load recorded.

The postmortem's puzzle "Which speed is the waypoint's" (std5k 7.72 against 8.39–8.56) shows that last question
is not academic.

**Consequence:** a number with no derivation becomes the waypoint's verdict, read by an unstated protocol.

**Modify:** take the three keys on pu first, as §6 already orders. Then project each stage's share, set the exit
from the projection, and name the reading protocol.

### A6. §6 stops: they compare timing keys strictly against a single-run baseline, and fire late

**The rules.** "Timing keys ... are never compared with the baseline, and no speed key has a limit (D1)" (The
rules). v2 §9.3: "timing keys are never compared strictly, since unchanged code varies from run to run".
chart-recovery's What Will We Change: "Compare only interleaved pairs from one session."

**Stop 4.** "does not fall below 4.79" compares against the baseline's single run, 4.7915. Checkpoint 3's two
readings were 5.18 and 5.11. Run-to-run spread of about 0.4 exceeds what this stop can detect, so a run at 4.78
with no real gain passes it.

**Stop 3.** "`speed-stage-recognize` rises above 1.0" is a speed limit with no stated ground. It stands in for
the first attempt's §9 signal, "Set reuse falling as steps add features", without reading `set-reuse` itself.

**Timing.** Both speed stops fire only "once both tables are in". That is late in exactly the way the w2
postmortem regretted: a per-stage criterion "would have stopped the waypoint at FOREST's gate rather than after
its merge".

**Consequence:** the stops detect noise or nothing, and they detect it after the work has merged.

**Modify:** give each issue a criterion on its own stage, as a before-and-after pair of interleaved runs in one
session (the chart-recovery practice), with a bound set at issue start ("Fill bounds at issue start"). Ground any
threshold that remains, and get perigrin's ruling for it.

### A7. §6 held keys: recognition-side keys change meaning, and the list omits gated keys

The README says `items-per-token`, `set-reuse`, `leo-disagree`, `terminals-per-token` and `lex-unmapped` "stay
recognition's" (`TestRecognitionKeysUnchangedByPipeline`), and that `items-per-token` and `set-reuse` are measured
"over files recognition accepts". Once recognition refuses precedence errors (the ten `%nonassoc` pairs,
`$a <=> $b <=> $c`), that population changes. Part of any fall in `items-per-token` is then a change of
definition. The README records the same trap once already: `speed-ratio-aggregate-cold` was renamed because it
"fell from 4.99 to 2.72 by redefinition".

§6's held list also omits gated keys: `collapse-over-limit`, `tables-drift`, `ledger-gaps` and `lex-unmapped`.
`precedence-covered`, which may not fall, appears only in §5.5. §6 also makes `oracle-exact-chart` a floor,
though the README marks it "reported". That is a new rule and should be called one.

**Consequence:** a gain could be claimed from a smaller population, and a gated key left out of the stops.

**Modify:** say how the recognition keys' population is held constant, or report the population beside them,
and list the held keys from the README's rules.

### A8. §5: a missing integration point, the reference recognizer

Rule 11 of v2 §3 requires an outside reference for algorithmic components, and the chart-waypoint postmortem's
third change kept one. That reference is `recognize/reference_test.go`: "A textbook Earley recognizer ... the
oracle compares acceptance and Expected under every Options combination". It is used by `TestReferenceRecognizer`,
`TestReferenceAgreesOnTruncatedFiles`, `TestWarmExpectedMatchesReference` and the fuzz oracle tests.

It has no precedence. After §4.1 and §4.2 the recognizer's acceptance and `Expected` diverge from it on perly.y.
The choice is then between two bad outcomes:
- the reference gains the same generated tables, and stops being independent of what it checks;
- the comparison is weakened.

`Expected` is also the interface v2 §13 names for interleaved lexing ("the lexer asks the parser what is
expected"). A precedence-aware `Expected` changes that contract, and possibly for the better. The design does
not decide either question.

**Consequence:** the suite goes red, or the recognizer loses its outside oracle. Both break a rule the earlier
postmortems put in on purpose.

**Modify:** add this as §5.6 and decide it at design time. One option: compare the reference with precedence off
against the recognizer with precedence off, keep that switch, and give the precedence filter its own outside
reference (bison).

### A9. §4.3 and §5.5: the referee cannot prove what §5.5 says it proves

`CountPrecedence`, which produces `precedence-disagree` and `precedence-covered`, judges each sentence through
`referee.judge` (`internal/chart/precedence.go`). That builds the forest with `builtForest`, which runs
recognition, then calls `f.Constrain(r.p)`.

While `Constrain` stays, `precedence-disagree 0` cannot detect an under-filtering table, because `Constrain`
removes what the tables let through. It detects only over-filtering. §4.3's "with `Constrain` reporting and not
removing" is a mode the key's definition does not have.

"Agree with it on every corpus" names no key and no test. The referee's fixtures are generated sentences, and the
corpora have no precedence referee beyond the oracle and verdict keys. When `Constrain` is removed, the referee's
own code path changes. So "a divergence is a defect in the tables, never in the referee" (§5.5) does not hold
across the removal: the referee's subject is different code.

**Consequence:** the waypoint's equivalence proof is partial until the last issue, and that issue changes the
prover.

**Modify:**
- Define the agreement measure: the tables' verdict against `Constrain`'s, per derivation, over the corpora and
  the referee's sentences. Make it a test, or a key with a ruling.
- State `judge`'s path after the removal.
- Keep `TestPrecedenceToyReferees` and `TestPrecedenceConstraintOrderInvariant`, which runs with Leo off and the
  cache off, against the new path.

### A10. §4 and §8: implementation is specified where the property should be

§4 names functions (`builder.complete`, `awaiting(src, lhs)`), a representation (`uint32`, a reserved bit) and a
table key ("per (state, LHS)"). That is issue-level detail.

The keying also sits oddly with §3's own finding: every pair but one is global, which suggests (rule, terminal)
alone would do. And the README records that keying by a chart-side context over-refused before ("An earlier form
keyed the state on the one symbol before the span, which over-refused").

What refinement needs instead is the property each component must hold. For example: Build's forest after the
tables equals `Constrain`'s output on every input; recognition's acceptance equals `forest.Parse`'s precedence
verdict. Criteria can be written against properties like those.

§5.3's choice between "a check at the next scan or a residue for the forest's tables" decides where correctness
lives, but it is missing from §8's open questions.

**Consequence:** issues inherit a mechanism with no invariant to test it against. The w2 postmortem's lesson
"Hedge by default ... a rule decides only where it follows a named perl routine" applies here too: the tables
should be stated as bison's resolution, keyed where bison keys it.

**Modify:** move the representation to the issues, state the invariants, and add §5.3 to §8.

### A11. §7: the direction the work points in is unstated

v2 §2 says what a verdict decides: a positive verdict makes the chart "a candidate engine for M2's incremental
work, and the interleaved spike (§13) is designed next". The design does not say which question this work
answers, or how its result bears on that verdict.

The first attempt's rule 2 kept facts out of items in order to protect set interning. v2 §13's convergence "on
state equality" depends on the same set identity. Values in sets make convergence and reuse harder, and §7 puts
M2 out of scope without recording that cost.

§7 also does not restate isolation: v2 §11 says only `internal/chart` changes.

**Consequence:** the waypoint could reach its speed exit and leave the next decision less reachable, with nothing
on record weighing the trade.

**Modify:** add one paragraph saying what the result decides, and the cost to set identity it accepts for M2 and
interleaving. Confirm `internal/chart` only.

## Summary

| ID | Section | Severity, my view |
|---|---|---|
| A1 | §1, §4.1, §4.2 | must fix: reverses v2 §2, §8 and rule 12 unlisted; ruling unrecorded; Options undefined |
| A2 | header, §7 | must fix: wrong governing decision; v2 not on pu |
| A3 | §2, §4.2 | must address: left-associative chains keep quadratic items; premise overstated (reasoned, unmeasured) |
| A4 | §6 | must fix: key addition without a ruling; precedence key's end state undefined |
| A5 | §2, §6 | must fix: 2.5 ungrounded; no reading protocol; re-gates speed against D1 unacknowledged |
| A6 | §6 | must fix: timing compared strictly to a single-run baseline; stops fire late |
| A7 | §6 | should fix: recognition keys' population shifts; held list incomplete |
| A8 | §5 | must fix: the reference recognizer and `Expected` contract are unaddressed |
| A9 | §4.3, §5.5 | must fix: the referee runs through `Constrain`, so it cannot see under-filtering; it changes at removal |
| A10 | §4, §8 | should fix: implementation in the design, invariants absent; §5.3 not an open question |
| A11 | §7 | should fix: what the result decides, and the M2 cost, unstated |

**Recommendation: modify.** The idea is plausible, and §6's keys-first ordering is right. But the document
reverses accepted decisions it does not name (A1, A2). Its proof of equivalence and its oracles do not survive
the change as described (A8, A9). Its exit and stops break the ratchet's own rules on timing and keys, and are
not grounded the way chart-recovery's were (A4, A5, A6).
