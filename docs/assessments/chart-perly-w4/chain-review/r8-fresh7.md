# chain-review chart-perly-w4, round 8: review-fresh7 (resumed)

## 1. Recommendation

**Accept.**

- I hold no blocking finding and promote nothing.
- My one new finding, FR8-1, is residual.
- Both blocking items the round-7 repair answered have landed, and the repair broke nothing I could find.

**Checks I made before reading `repair-r7.md`.**

I re-read the live chain and diffed it against my round-7 snapshot (`fr7z/issues.json`, `fr7z/ms.json`).

- **What changed.** The chain still has 19 issues, and no id, edge (`blocks`), label or title changed. Only three issue bodies changed: d166, d218 and 3262.
- **Milestone body.** The only addition is a "Chain-review round 7 repair" entry in Records. The residual mapping is untouched.
- **New criteria are red at e3432c0f.** I ran both, with `PERL5_CORPUS=/nonexistent`, and each exits 1:
  - d166's `TestNarrowedSetIdsStableAcrossFiles`;
  - d218's five-test "verdict only" criterion.
- **3262's joint criterion stays green after d218.** Its command is unchanged. `TestWaypoint2RefusalsMatchReference` keeps its name, and d218 now makes it pass on the judged forest under the verdict-only rule.
- **d218's verdict-only rule has rows to count.** The waypoint2 table does contain the precedence rows d218 relies on for its N > 0 count: `$a <=> $b <=> $c;`, `$a == $b <=> $c;`, `1 .. 2 .. 3;` and `$a <=> $b + 1 <=> $c;` (`waypoint2_refusals_test.go:20–37`).
- **d218 agrees with d247.** It cites the verdict-only rule that d247 states for the referee. The two do not conflict.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR7-1 (bb9b item 8) | **held, residual** | See "FR7-1 and FR7-2" below. |
| FR7-2 (3282's fixture against d0ba's) | **held, residual** | See "FR7-1 and FR7-2" below. |
| FR7-3 (no Leo test under 3282's pairing rule or shift side) | held, residual | Unchanged. 3282's seed-44 differential (15,000 Leo chains) still covers it incidentally. |
| FR7-4 (`slipper.go`'s `slip()` adds items outside `builder.scan`, unnamed) | held, residual | d1c2's and 3282's paths are unchanged. |
| FR7-5 (d166 and d2a2 both own narrowing the stage harness; `TestRecognizeSetAddedAfterNew` is unlisted) | held, residual | d166 was edited, but not on this point. Its new stable-id property makes the scoping of `TestRecognizeSetAddedAfterNew`'s "id beyond the table is an error" clause more pressing, not less. |
| FR7-6 (no answer-key check at d218 or d247) | held, residual, narrowed | d218 now runs `make chart-precedence-reference` where perl.git is present. That checks the differential, not the 27 keys. d247 is unchanged. |
| FR7-7 (`drive.sh` does not unset `T1_CORPUS` and `TDT_FIXTURES`) | held, residual | Unchanged. `T1_CORPUS=/opt/corpora/PerlOnJava` is still exported in this sandbox. |
| FR7-8 (test locations versus the `forest` test helpers) | held, residual | Unchanged. |
| FR7-9 (d199's size) | held, residual | Unchanged. I do not promote it. See section 3. |
| FR7-10 (where the Constrain entry-counting hook lives, and who added it) | held, residual | Unchanged. |
| FR7-11 (`-race` needs cgo) | held, residual | Unchanged. d166's new criterion does not use `-race`. |

### FR7-1 and FR7-2: are they defects in what the chain says?

As the raiser, I judge that both are defects in the chain's text, not in the builder's latitude. Neither is blocking.

**FR7-1 is a defect in bb9b's text.**

- Item 8 does not leave the question open. It gives its own answer: the limit "now applies (Constrain only, since production's Build walks by context)".
- That answer is false for the forest V2 must produce.
  - d0ba's `TestForestEqualityUnionShortcut` and d218's real differential require V2's forest to equal V1's node for node.
  - So a node reached under two live contexts keeps the union of their alternatives in V2 as well.
  - That union is exactly the limit's stated consequence: it "can only add a derivation to what a later reader enumerates".
- Walking by context removes only the alternatives of a context whose chain dies (C6). V1 also lacks those, because they are unreachable.
- **Why it is not blocking.** The consequence is a false sentence in a live document. No code goes wrong. bb9b already says "a sentence this issue cannot make true … is reported, not papered over", and crochet:review reads the live documents against the diff. Proceeding does not make anything wrong that a later gate cannot see.
- **Fix:** drop the parenthetical. Say that the union is a property of the shared-node forest, which holds for production too, and that walking by context keeps out only a dead context's alternatives.

**FR7-2 is a defect in 3282's text: an under-specified cross-issue reference.**

- An input can satisfy both issues at once: a separate site where a context dies, beside d0ba's two-live-context node.
- So the texts do not contradict each other. But the obvious reading, "add a dead third context at the same node", turns d0ba's closed criterion red, because its log line `contexts: 2, distinct non-empty alternative sets: 2` is exact.
- That would amend a started issue, which §7 makes a stop.
- **Why it is not blocking.** A builder has a clear correct option, and the cost of misreading is a reported stop, not a wrong result.
- **Fix:** let 3282 use its own constructed input for the dead-context test, or a second site that leaves d0ba's node unchanged.

## 3. Promotions

**None.** I read the round-7 reports in full and the held-row tables of r1 to r6.

- **r6-fresh6 FR6-5 (`resetKeys` in `main_test.go`).** I do not promote it. d2a2's close runs `make test`, which turns `TestResetKeySet` red, and the fix is the same key removal, not tuning. Adding `main_test.go` to d2a2's paths is the residual fix.
- **The "tables declare no levels, so the failure is silent" family (r6-fresh2/3/4 FR6-1 and FR6-2, r6-fresh6 FR6-7, r6-pushback N15).** I do not promote it. On perly.y the failure stays loud: a failed derivation refuses every input, and `ParseWith`'s guard rejects any substitute precedence. The silent branch is reached only over tables that are not perly.y's, and no production path builds those.
- **My own FR7-9.** I do not promote it. d199 is large, but its 15 criteria are concrete and each is red before the work. Running out of session leaves a partly done issue with red criteria, which is visible, not wrong. At the bound, a split costs a repair round that the gain does not justify.
- **All other residual rows.** Unchanged in substance, and none makes proceeding wrong.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR8-1 | residual | 01a21447-d218 | The repointed differentials' candidate is "`recognize.New` over perly.y's tables … then Build". The text does not say whether that is a narrowed lattice, as `parseStream` uses, or the test's existing un-narrowed lattice with `LibraryOperators` applied afterwards. "Where a test already applied `LibraryOperators` … it still does" can be read as applying to the candidate side. | On an un-narrowed lattice the judged reduce side keeps the mixed-set completions V1 drops: the 117 tokens of design §4.1, `listexpr ::= term` before the PLUGIN readings. If a kept library operator's live reading is the one that forbids, pruning afterwards leaves an alternative that V1's `Constrain` would drop, which is a false "forest differs" stop. On the inputs these tests read today, the corpus, perl.git and waypoint2 under `Scope{}`, no library operator is kept (C12). So it cannot fire there, which is why this is residual. | d218's repointed paragraph; `feasibility.md` lines 120–133; design §4.1 ("A recognizer given an un-narrowed lattice … keeps completions V1 drops"). | One clause in d218: the candidate recognizes the lattice narrowed under the file's scope, as `parseStream` does, and `LibraryOperators` applies to the reference side only. |

## 5. What I did not re-examine

- I did not re-run the overlay experiments from round 7. No changed text touches them: the referee leaf rule, the `recognizeSource` sweep, the measured bytes and the ratchet timing.
- I did not run the old criteria again, beyond the three affected ones (d166's and d218's new criteria, and d247's existing one).
- I did not read r1 to r5 in full. I read their held-residual tables and the round-7 reports.
- I did not try to construct the dead-context fixture that FR7-2 concerns.
- I did not check d218's six named tests against each test's present use of `LibraryOperators`, beyond `TestWaypoint2RefusalsMatchReference`.
