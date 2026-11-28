# chart-precedence-v2: rulings

These are perigrin's rulings on the V2 design, recorded on the day each was made. Only perigrin's words are rulings. Each entry gives:

- the question as it was put to perigrin;
- perigrin's answer, verbatim;
- what follows from the answer.

## 2026-11-25, while the design was being written

- **Direction.**
  - Asked how V1 lines up and what V2 is for, perigrin said: "so lets make v2 happen".
  - perigrin then chose approach 1 of three: "1".
  - Approach 1 is full reduce-side and shift-side judgement in the recognizer. The production precedence pass is retired. V1's `Constrain` and `ConstrainReference` are kept as test referees.
- **The §6 correctness fixes are out of V2.** The orchestrator proposed this as a working assumption, and perigrin did not object. It is recorded as an assumption, not a ruling.
- **Library-operator narrowing:** "Narrow before recognition (Recommended)".
  - `libraryOperators`' dead (token, terminal) pairs are applied to the lattice before recognition.
  - A scope refusal keeps its reason `scope`, but its byte may move.
  - Per-file verdicts must not change.
- **Precedence refusals:** "Keep 'precedence' (Recommended)". The recognizer records when a set died only because of a precedence refusal. It reports reason `precedence` at the byte where the set died, not at V1's byte 0.
- **Rule 12:** "Lift rule 12 for refusals". Recognition records, beside the sets, which parent each completion was allowed for, and Build reads that record. The rule is lifted for these refusal records only.
- **Guards:** "Regenerate everything".
  - There is no precedence off-switch.
  - Every guard runs with precedence on: the reference recognizer, the fuzzer, the set and forest digest goldens, and the brute-force outside reference. Each is updated to match.
  - `items-per-token` and `set-reuse` move by design.

## 2026-11-25, after assess round 1

The questions came from round-1 findings A1, A3, A9, A10, C1, P1, C9 and P6 (see the minute).

- **Superseding the governing design:** "Supersede, bounded (Recommended)".
  - V2 explicitly supersedes these parts of `docs/plans/2026-10-30-chart-parser-v2-design.md`, for precedence:
    - §8, "precedence leaves recognition (perigrin, 2026-10-30)";
    - rule 6's "after the chart";
    - rule 12's "items stay (state, distance)".
  - Items may carry the canonical spine top and nothing else.
  - Adding any other fact to an item is a stop. It replaces the §10 kill criterion "Facts in items".
- **What "a file's verdict" means:** "Accept/refuse only (Recommended)".
  - The verdict is accepted or refused.
  - A refused file's reason and byte may differ from V1's: they become positional.
  - `refused-where.txt` is regenerated.
  - This narrows the earlier "Precedence refusals" ruling: what must match V1 is accept or refuse, not the reason.
- **The exit:** "Relative only (Recommended)".
  - V2 must beat V1 past noise.
  - A correct V2 that is no faster is reported to perigrin as a stop.
- **Guards and keys:** "Hold golden, retire key (Recommended)".
  - `resolved-forest-digest.golden` stays byte-identical, as a V2 guard. This narrows "Regenerate everything".
  - `speed-stage-precedence-aggregate-longlived` is retired.

## 2026-11-25, after assess round 2

The question came from round-2 findings A20, C16 and P14 (see the minute).

- **The refusal reason:** "Diagnose by rerun (Recommended)".
  - When a file is refused, and only then, recognition is rerun without precedence, then rerun with narrowing. This reproduces V1's order exactly: grammar, then scope, then precedence.
  - `lex-unmapped` and resolve's `scope` are unchanged.
  - Reasons match V1's. Only a `precedence` refusal's byte moves, to where the chart died.
  - The cost falls on refused files only.
  - Production keeps a no-precedence mode, used only to classify refusals. This narrows "Regenerate everything"'s "no precedence off-switch". The mode classifies refusals and is not a way to parse.

## 2026-11-25, during assess round 3

- **Process:** "The goal is alignment and consensus". perigrin said this while round 3's reports were arriving, after the orchestrator suggested that a contradiction three participants had raised (C18, P17 and A27) might need perigrin's call.
  - A finding the subject can answer goes back to its raisers to settle.
  - Only views that do not reconcile go to perigrin.

**Note on the round-2 entry.** Two clauses in the round-2 entry quote the option perigrin chose; perigrin did not say them. Round 3 measured both:

- **"Only a `precedence` refusal's byte moves."** Finding A26 shows this is false for a `scope` refusal from narrowing. Measured: `foo plus 2 plus 3;` V1 byte 4, rerun byte 11. The earlier ruling "Narrow before recognition" already says that byte may move, and the subject follows that ruling.
- **"Used only to classify refusals."** The referees also use the mode, in tests. The subject now says so, and the raisers judge whether that fits.

## 2026-11-25, during assess round 6

- **Acceptance:** "I think if 7 of 8 participants agree and there are no blockers we could move forward yeah".
  - perigrin said this while round 6's last report, from the new fresh reader assess-fresh6, was still out.
  - At that point:
    - seven participants recommended accept: assess-codebase, assess-architecture, assess-pushback, assess-fresh2, assess-fresh3, assess-fresh4 and assess-fresh5;
    - no blocking finding was held.
  - Revision 7 is accepted.
  - assess-fresh6's report is recorded in the minute when it arrives. If it raises a blocking finding, the finding goes to perigrin and is not answered silently.

## 2026-11-25, after assess round 6

- **The late findings:** "Revise to rev 8, run round 7 (Recommended)". assess-fresh6's report arrived after the acceptance ruling. It recommended modify and raised three blocking findings: F6-C1, F6-C2 and F6-C3. F6-C3 is a defect for perly.y itself.
  - Revision 8 answers all three.
  - Round 7 runs within the bound of eight, to confirm consensus.
  - The acceptance of revision 7 is held until round 7 is done.

## 2026-11-25, close of assessment

There was no new ruling. The hold set by "Revise to rev 8, run round 7" ends because the assessment reached a fixed point at round 8, on revision 9:

- all ten participants recommended accept;
- no blocking finding was held or raised.

## 2026-11-26, chain-review bound for chart-perly-w4

- **Closing chain-review:** "Apply fixes, accept (Recommended)".
  - Chain-review on milestone `chart-perly-w4` reached its bound of eight rounds without a fixed point. Five of nine participants recommended accept.
  - Three blocking items remained, each with a one-line fix its raisers had proposed:
    - r7-fresh7 FR7-1: item 8 of the docs issue would have written a false README sentence.
    - FR6-6: d166's cache test fixture could not fail.
    - r6-fresh6 FR6-3: the corpus differentials in d218 and 3282 ran on un-narrowed input.
  - The three fixes were applied exactly as proposed. Chain-review passed by this ruling.
  - The reports are in `docs/assessments/chart-perly-w4/chain-review/`.

## 2026-11-26, execute on chart-perly-w4

- **The git-zhi version:** "Upgrade to v0.7.2 (Recommended)". The container had git-zhi 0.5.2, below crochet's minimum of 0.7.2. 0.5.2 ignores `ZHI_ACTOR` on writes, so every execute worker failed the ownership check. v0.7.2 was installed with git-zhi's own `install.sh` into `~/.local/bin`, ahead of 0.5.2 on PATH.

## 2026-11-26, execute: follow-up issue `01a215b8-9e3a`

- **The follow-up issue's two defects:** "Fix name, drop EOF (Recommended)". The orchestrator wrote two defects into this issue, and the worker stopped on both:
  - Criterion 3 named a test that does not exist, `TestFeasibilityNullableRulesAllow`. The real test is `TestFeasibilityAllNullableRulesAllow`.
  - A step asked for a toy whose non-allow verdict is decided only at EOF. No such toy can exist: EOF is a pseudo-terminal that no rule names, so the generator always writes verdict 0 there. The worker measured this over perly.y and every toy.

  The ruling:
  - criterion 3's test name is corrected;
  - the EOF step is removed from the issue.

## 2026-11-26, execute: issue `01a21447-d10d`

- **d10d's invalid criterion:** "Drop 2nd alternative (Recommended)". The orchestrator wrote this defect into the issue at refinement, and chain-review did not catch it.
  - The negative criterion's pattern was `compared ([1-9][0-9]*) of \1( |$)|compared ([1-9][0-9]*) of \1 files`.
  - In the second alternative, `\1` refers to a group in the first. GNU grep 3.11 rejects the whole pattern with "Invalid back reference", so the criterion can never exit 0.
  - The worker stopped, without editing the criterion. The test logs `compared 1881 of 1881 files`.
  - The other 25 `grep -E` patterns in the chain were validated mechanically, and all are valid.

  The ruling: drop the second alternative and keep `compared ([1-9][0-9]*) of \1( |$)`. That still matches the logged line, and still fails when fewer files are compared than loaded.

## 2026-11-26, execute: the repository's git hooks

The d166 gate found an intermediate commit that was not gofmt-clean. Its cause: none of the repository's pre-commit hooks were installed in this container, because `pre-commit` was not on PATH and `.git/hooks/pre-commit` was absent.

- **Install the hooks:** "Yes". The orchestrator installed:
  - `pre-commit`;
  - the hook tools, at the last releases that build on Go 1.24: goimports, gocritic, staticcheck and govulncheck.

  It then installed the pre-commit stage. With the stage installed, `pre-commit run --all-files` on pu fails only on a pre-existing trailing-whitespace line in a testdata file, which a normal commit does not touch.
- **The commit-msg stage:** "Add chart, then install (Recommended)". `commitlint.config.js`'s type list had no `chart`, although 20 of pu's last 40 commits used it. `chart` is added by a `ci:` PR, and the commit-msg hook is installed after that PR merges.
- **The pre-push stage:** "Leave pre-push off (Recommended)". With `vuln.go.dev` allowed, govulncheck reports 42 reachable vulnerabilities on pu, in the Go 1.24.4 standard library and golang.org/x/net v0.33.0. Those predate this milestone. A toolchain and dependency upgrade is separate work, outside V2.
- **Persistence:** "Add to setup script (Recommended)". The install commands go into the environment's setup script.

## 2026-11-26, execute: the d199 gate's out-of-scope findings

The gate on `01a21447-d199`, the reduce side, reopened it for two test gaps. It also raised three findings that change issues not yet started.

- **O1, d218's count of feasible toys:** "Amend d218 now (Recommended)". On pu, `prectab.Feasible` refuses 5 of the 13 toys: `empty`, `emptyred`, `ladder`, `nonassoc` and `prefix`. d218 said 11 were feasible, and its `TestInfeasibleToysJudgedByConstrain` criterion did not name `ladder`, `nonassoc` or `prefix`. d218 was pending. Its count is corrected to 8, and the three toys are added as named subtests.
- **O2, the slipper's zero-width `;`:** "New issue blocking d218 (Recommended)". Nothing judges the slipper's closure, and no open issue carried it. Issue `01a21945-d1e6` adds counters showing that no reduce before, and no shift of, the inserted `;` is forbidden on perly.y. A non-zero count is a stop. The issue blocks d218.
- **O4, Leo's judged walk:** "Measure before switch (Recommended)". The judged recognizer's Leo step walks the whole memo chain, and both cost bounds measure only the default recognizer. Issue `01a21945-d35d` runs both bounds, at unchanged constants, on the judged recognizer, and checks that Leo's walk is linear. A red bound is a stop. The issue blocks 3262.

## 2026-11-26, execute: the judged Leo walk (issue `01a21945-d35d`)

The gate on `01a21945-d35d` raised a design §7 STOP. Its linearity test could not fail on what it claims, because its toys take Leo only once per input.

On the shape the issue exists for, the unchanged judged code walks n(n−1)/2 memo links. That shape is a right-recursive nonterminal whose FOLLOW holds the token that continues it: `S ::= L T ; L ::= a L | a ; T ::= a T | b`, on aⁿ b. The walk is 4851, 19701 and 79401 links at n = 100, 200 and 400, which is quadratic where the default recognizer is linear. A search of the corpus sample and 26 synthetic Perl families found no perly.y input with this shape.

- **The walk:** "Fix it in d35d (Recommended)". The started issue's scope widens from tests only to a production fix that makes `leoCrosses` linear. The linearity test moves onto the quadratic toy, and it must be red before the fix and green after.
- **A bound on the walk:** "Add it with the fix (Recommended)". Neither existing cost bound can see the walk. d35d adds `TestLeoLinksPerTokenBoundedJudged`, a corpus bound on Leo links per token with a new constant set from the measured value after the fix.

## 2026-11-27, execute: the d1c2 gate's findings (issue `01a21447-d1c2`)

The gate on the item-context issue passed. It raised two findings for perigrin.

- **F1, §4.3 at predictor granularity:** "Delete it, keep as oracle (Recommended)".
  - `PredictionContexts` gives each predictor one context. At that granularity §4.3's collapse is identical to §4.2's for every grammar, because a canonical top forbids shifts only in its own predictor's region.
  - So the `forbidsAt` clause in production is dead code.
  - A finer per-state collapse is not built. The worker and the gate judged it unsound under 3282's pairing, or else needing a §7 item fact.
  - Issue `01a21a56-1251` removes the clause, keeps `forbidsAt` as the test's oracle, and has the README state the granularity.
- **F2, the context through the slipper and `Expected`:** "Add to 3282 (Recommended)".
  - The context flows through both, but no test observes it there.
  - 3282 had not started. It was amended to add `TestShiftRefusalThroughSlipper`, on a toy, since perly.y forbids no shift of the slipper's `;`, and `TestShiftRefusalThroughExpected`.
  - The two stale comments the gate found (F3) were also folded into 3282.

## 2026-11-27, execute: the switch-over's Step 1 probe (issue `01a21463-3262`)

The 3262 worker applied the switch-over alone, with no test edits: `recognize.New` judges. It then ran `make test`. Two red tests were not covered by the issue's "Tests this change moves" list, so the worker stopped before editing anything. No forest or acceptance test went red other than those the list names, and every cost bound held.

- **`TestRecognizersShareGrammarTables` times out:** "Share one copy (Recommended)".
  - The test's "alone" baseline builds a fresh tables copy for each of its 794 sample files.
  - Once `New` judges, each copy derives perly.y's precedence again. That runs past the 10-minute test timeout. Production derives once per tables value.
  - The per-file baseline recognizers now share one `ownTables` copy, so the test still checks sharing against a separate tables value.
  - 3262 is amended to list the test as rewritten.
- **`TestReadingsCollapse` goes red:** "Pin it, amend 3262 (Recommended)".
  - `builtReadings` builds the forest before precedence with `recognize.New`. That forest is V1's built forest.
  - It is pinned to `NewNoPrecedence`, following the list's "reproduce V1" rule.
  - 3262 is amended to list the test as pinned.

## 2026-11-27, execute: the speed exit at risk (issues `01a21447-d218` and `01a21463-3262`)

The d218 gate took rough timings, neither interleaved nor on a quiet machine. With Constrain gone from production, `speed-ratio-aggregate-longlived` gave after-medians of 3.35–3.46, against V1 before-minimums of 2.84–3.02.

The cost is `forest.Build` of the judged chart, not Constrain. `speed-stage-forest` went from about 0.65 to about 1.45 at the switch-over (3262).

The exit in design §7 needs the after-median below the before-minimum, so it looks likely to be unmet. The remaining issues are d247, d278, d2a2 and bb9b.

- **Order of the verdict reading:** "Take a quiet reading now (Recommended)". Once d218 and d1eb merge, the formal interleaved B A B A B A reading runs before the remaining issues. If the exit is unmet, execution stops and the `forest.Build` cost is reported, without spending effort on d247, d278 or d2a2 first.

## 2026-11-27, execute: the early reading misses the speed exit

The early reading followed the ruling "Take a quiet reading now". It is kept in `docs/assessments/chart-perly-w4/early-reading/`.

| | |
|---|---|
| before-minimum of `speed-ratio-aggregate-longlived` | 3.2143 |
| after-median | 3.3991 |

All 27 answer keys held, and every differential is green.

The cost is `forest.Build` of the judged chart:
- the forest stage went from 0.66 to 1.47;
- forest allocations per token went from 0.063 to 0.113.

The likeliest cause is 3282's walk by context and its `RuleRefused` filtering. That is inferred from the timings, not profiled. This is design §7's stop, "the checks pass but the exit is unmet".

- **How to proceed:** "Profile, then new issue (Recommended)".
  - Profile `forest.Build` on the judged chart, and report the cause before changing anything.
  - A fix becomes a new issue blocking d2d1, and it must hold every forest byte-identical.
  - The remaining chain stays paused until perigrin rules on the fix.
- **The raw runs:** "Commit as a docs PR (Recommended)". They are kept beside a short README holding the medians, the minimum and the exit line.

## 2026-11-27, execute: the profile of `forest.Build`

The profile found the cause of the missed speed exit.

**The walk.** 3282 added the walk by context, `(*builder).byContext`. On judged charts it runs in almost every file: 1597 of 2211 files, holding 99.4% of the tokens. It doubles `forest.Build`, from 463 to 1030 ms per aggregate pass, and its `seen` hash map alone is 27% of Build.

**Direction A.** The walk without the hash map was prototyped. Its forests are byte-identical over all 2211 files, and it uses only the context and the refusal records already kept. It cuts Build to about 800 ms. On a rough, unmeasured estimate, that puts the headline ratio near 3.03 against the before-minimum of 3.21.

- **The fix's scope:** "A, then re-read (Recommended)". A single issue, `01a21db4-4792`, labelled `model:sonnet` and blocking d2d1, carries direction A:
  - the walk without a per-node hash map;
  - the `keep` closure created once per walk;
  - a pre-sized `leoAt`.

  Forests must stay byte-identical. A quiet, interleaved reading follows. If the exit is still unmet, it is reported before direction B is taken up.

## 2026-11-28, after close: the four open questions

These are perigrin's answers to the four questions left open during chart-perly-w4.

- **O1, the judged Leo walk under union lookaheads:** "File a follow-up (Recommended)". The crossing-verdict cache is keyed per lookahead, so a never-repeating run of unions keeps the walk quadratic. The corpus has none.
  - Filed as `01a21f98-a8bf`, in `chart-perly-w4-followups`: fold each link's refusing terminals into the memo when it is built.
- **The `internal/cli/docs` live-GitHub tests:** "File an issue to stub (Recommended)". They timed out in CI on #667 and #673.
  - Filed as `01a21f99-e2ce`, in its own milestone `cli-docs-offline-tests`: serve the calls from a local test server, and keep one opt-in live test.
- **The `Co-Authored-By` trailer against "no model identifiers in commits or PRs":** "Drop the trailer".
  - From now on, commits carry no `Co-Authored-By` line; the `Claude-Session` trailer alone attributes the work.
  - Commits already merged are not rewritten.
- **Golden-alone commits against CLAUDE.md's "`make test` passes before every commit":** "Test + golden together (Recommended)".
  - A test change and the golden it regenerates are committed together, and the commit message names the golden.
  - Refinement no longer writes golden-alone criteria.
  - d278's two red commits stay in history.

## 2026-11-28, after close: what a difference from RD or V1 means

These rulings came from discussing the prune-at-recognition prototype, whose pre-resolve forest differs from V1's on six inputs. perigrin's words:

- "The forest differing isn't a problem if the CST is the same"
- "Even the CST differing isn't bad as long as our CST is more accurate to the actual source"

This is recorded in `internal/chart/README.md`'s precedence section, under "What a difference means":
- The arbiter is perl's own parse.
- Differences from RD or V1 are investigated, and accepted when they agree with perl at least as well, where the oracle compares.
- A claim of greater accuracy where the oracle does not compare adds an oracle case.
- A parse perl would not make is a defect.
- A decision that makes forest equality its stop, as V2's §7 does, still binds within its milestone.

On the string-token classes, perigrin ruled: "Honestly we should do B because it's more accurate regardless of the RD CST". This is filed as `01a2200c-112e` in `chart-string-token-class`.
