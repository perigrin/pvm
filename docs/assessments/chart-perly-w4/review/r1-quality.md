# crochet:review round 1: quality (prefix Q)

Participant: quality. Lens: is the code sound? (the paad:agentic-review taxonomy).

**How the lens ran.** `paad:agentic-review` v2.1.0 loaded, but its pre-flight stopped it. The session has prior history, the checkout is on `pu` (the integration branch), and the brief forbids writing under `paad/` in the repository. So I applied its six lenses inline, one reviewer, no subagents and no verifier pass:

1. logic and correctness;
2. error handling;
3. contract and integration;
4. concurrency and state;
5. security;
6. spec compliance, only as it bears on code soundness.

**Subject.** `git diff e3432c0f 8d8fff38`, read in a detached scratch worktree at 8d8fff38. `pu` HEAD has since moved to c383f861. The repository was not edited. Probes ran only in scratch worktrees, which are now removed or clean.

## Summary

**No blocking findings.**

- Four residual findings (Q1 to Q4) are listed below.
- I found nothing that meets a design §7 stop.
- The production paths I read are internally consistent:
  - the judged recognizer: `judge.go`, `recognize.go`, `leo*.go`, `expected.go`, `finished.go`, `slipper.go` and `narrow.go`;
  - Build's walk by context and refusal pairing: `forest.go` and `bycontext.go`;
  - the refusal reruns: `ChartRefusal`;
  - the `prectab` tables and `Feasible`;
  - the referee in `precedence.go`;
  - the ratchet stages in `speed.go`.
- The two risky areas the brief named came out clean:
  - **Caching keyed on a tables pointer.** Every per-tables cache goes through `grammar.Derived`. It is keyed per `*Tables` value, and a copy gets its own (`derivedFor`). `JudgesAsNew` compares `*Precedence` pointers that come from the same `Derived` slot, so the comparison is stable.
  - **Concurrency.** No `Derived` builder asks for its own type, so it cannot deadlock. The nesting is `derived` → `Parsed` and `derived` → `*grammarTables`. `narrowedSets` and `extend` publish copy-on-write under their locks. `Recognize` refreshes stale `ncls` and `cls` snapshots before validating edges.

## Test evidence

Every command ran with `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`, at 8d8fff38.

- `go test ./internal/chart/...`: every package `ok`, exit 0. The slowest were `chart` 77s, `cmd/ratchet` 105s, `forest` 291s, `prectab` 74s and `recognize` 278s.
- `go test -race ./internal/chart/...` (default timeout): 0 `DATA RACE` reports.
  - Four packages hit Go's default 10-minute timeout under race instrumentation, so this is not a data race but coverage stopped early:
    - `chart`;
    - `cmd/ratchet`;
    - `forest`;
    - `recognize`.
  - The tests still running when they hit the limit included `TestStartForms`, `TestPrecedenceCorpusMatchesReference` and `TestReferenceRecognizerPrecedence`.
  - `lex` failed `TestLexStreamAllocations`. That is an allocation bound under `-race`, and `lex` is unchanged by the delivery: the known failure.
  - `grammar`, `prectab` and `tdt/align` passed.
- Targeted race run, `go test -race -timeout 30m -run 'Concurren|Shared|ShareNothing|NotShared|Narrow|SetID|Derived|CopyTables|NewOverAny'` over `chart`, `recognize`, `forest` and `grammar`: all `ok`, 0 races. That set includes these tests:
  - `TestForestConcurrentBuildsOneRecognizer`;
  - `TestRecognizersExtendSharedTablesConcurrently`;
  - `TestPostdotIndexConcurrentRecognizers`;
  - `TestPrecedenceConcurrentCallsShareNothing`;
  - `TestDerivedConcurrentFirstCall`;
  - the `narrow` and `SetID` tests.
- A complete `-race` pass needs `-timeout` raised, which I did not run. Note this for whoever next runs a full `-race` pass; it is not a finding.
- `go vet ./internal/chart/...`: clean.
- `gofmt -l internal/chart`: clean.
- `staticcheck ./internal/chart/...`: two hits. One is pre-existing, `levels_test.go` S1007. The other is Q4.

## Findings

### Q1: RESIDUAL. `NewJudged` with an explicit precedence never checks that it was built over the same tables

- **Where:** `internal/chart/recognize/judge.go`, `NewJudged` and `newJudgement`.
- **What happens:** `newJudgement(tab, p)` indexes `tab`'s states, items and symbols against `p`'s tables. `Feasible(p)` and `PredictionContexts(p)` read `p.Tab`, not `tab`, and nothing compares the two.
- **Evidence:** a scratch test, since removed, called `NewJudged(toyTables("S ::= A","A ::= NUM"), pFromOpToy, Options{})`. It panicked with `runtime error: index out of range [3] with length 3` instead of returning an error. When the shapes happen to fit, the result would instead be a recognizer that judges silently with the wrong verdicts.
- **Grounds:**
  - Production `New` always derives `p` over `tab`, so this cannot happen on the shipped path, and `ParseWith` refuses explicit precedences over perly.y's tables.
  - Still, `NewJudged` is exported, and its contract says an error is "returned here". A mismatched `p` is a caller error that the constructor could report and does not.
- **Remedy:** at the top of `newJudgement`, return an error when `p.Tab != tab`. If copies must be accepted, compare the rule, symbol and item counts instead.

### Q2: RESIDUAL. The slipper's closure runs no reduce side, and only `;` is pinned

- **Where:**
  - `internal/chart/recognize/slipper.go`, `semiAt`: `v := &builder{...}` never sets `v.la`, so `complete` runs with `j == nil`;
  - `recognize.go:319`, whose comment says "the reduce side is judged under every option";
  - `prectab/semicolon_test.go`, `TestSlipperSemicolonReduceNeverForbidden`.
- **What happens:** completions inside the zero-width `;` closure are never judged. That closure is completed before the lookahead of the `}` token's set, or before end of input. The pin test's doc says such a rule "would be judged by ReducesFree(r, ;, pi)", but nothing judges it. The pin holds only because no verdict could forbid it.
- **What is pinned:** the pin covers `;` alone. The closure's lookahead also includes `}` (the `}` token's terminal set) and EOF (`eofBits`), and neither is pinned.
- **Evidence:** a scratch test, since removed, counted the rules whose `ReduceAt[r*NT+t]` is not 0 in perly.y's tables. The result is 0 for `PERLY_SEMICOLON`, 0 for `PERLY_BRACE_CLOSE`, and 0 for EOF (id 237, NT 238). So nothing differs today.
- **Relation to known residuals:** the milestone's d199 record O2 is about the `;` shift. `semiShift` now judges that shift. This finding is the closure's reduce side, which no record names.
- **Grounds:** this is a latent hole that rests on a property of perly.y. If a grammar change gives any rule a non-allow verdict before `}` or EOF, the slipper would accept or keep what the main path refuses, and no check would notice.
- **Remedy:** either judge the closure, by setting `v.la` and `v.laKey` in `semiAt` when `r.judge != nil`, or add pins for `}` and EOF beside the `;` pin. In the second case, also add a `shortcut:` comment in `semiAt` and correct the comment at `recognize.go:319`.

### Q3: RESIDUAL. Refusal records are keyed by parent state only, and two readers ignore context

- **Where:** `recognize/judge.go`, `(*builder).settled`, and `forest/forest.go`, `(*builder).alt`.
- **What happens:**
  - `settled` drops a parent record when another item from the same origin `completes` the same LHS into that parent state. It never checks that the item pairs with that parent under §4.3, that is, that `predicts(parent, parentCtx) == item ctx`, the condition `complete` itself applies.
  - `alt` checks `Refused` against `b.item(...)`, which is the first placed item at `k` holding the core item. It never checks other states at `k` that hold the same item.
- **When it could matter:**
  - A dropped record can let `alt` keep a pairing the recognizer refused. `byContext` re-checks only `RuleRefused`, so it does not catch parent-level refusals from the parent-dependent (`dep`) path.
  - The first-item check can drop an alternative that another state allowed.
- **Evidence:**
  - A scratch build, since removed, counted refused alternatives in `alt`, over all 986 T1 files through `forest.Parse`. It found 51 refused alternatives, and 0 where another placed state at `k` holding the core item was not refused.
  - The forest-equality differentials are green on waypoint2, on the referee sentences, and in the `exit: met` verdict.
  - I could not build an input that triggers either path, so this is unconfirmed.
- **Grounds:** the property rests on the corpus and is not stated as a Feasible counter or an invariant in the code.
- **Remedy:**
  - In `settled`, require the allowing item to pair with `x.parent`, by the same `predicts` test `complete` uses. Or add the child's context to the record.
  - In `alt`, refuse only when every placed state at `k` holding the core item is refused.
  - Alternatively, write the invariant down with a test that pins it.

### Q4: RESIDUAL. Dead test helper `constrainSource`

- **Where:** `internal/chart/precedence_test.go:63`, `constrainSource`.
- **Evidence:** `staticcheck ./internal/chart/...` reports `func constrainSource is unused (U1000)`. The function exists at e3432c0f with callers. The delivery removed its last caller. It differs from the known stale ABOUTME in the same file.
- **Remedy:** delete it.

## Lenses with nothing to report

- **Security:** there is no trust boundary in the delta. The input is Perl source that is lexed and recognized in memory. Malformed lattices return errors from `Recognize`, which checks edge bounds, direction and set ids. The one panic I found is the API misuse in Q1.
- **Error handling:**
  - `ChartRefusal` propagates rerun errors as refusals.
  - A failed derivation in `NewJudged` is returned from every `Recognize` and every `builtStream`, so the recognizer never runs unjudged.
  - `grammar.Derived` caches a failed derivation for the life of the tables. That is deterministic, and the doc states it.
- **Tests that cannot fail:** I sampled these:
  - `forest_equality_test.go`, which requires every row compared and some refused;
  - `TestLeoDisagreeZeroUnderPrecedence`, which fails when no `.t` files exist;
  - `refusal_test.go`'s rerun counter;
  - the fuzz target's judged seeds (`precDisagreement`);
  - `semicolon_test.go`, which pins its populations.

  Each has a liveness or denominator guard. I found no new vacuous test beyond those the milestone records already list, such as d35d's F1.
- **Shared global counters** (`reruns`, `constrainEntries`, `forestGrammarBuilds`): these are atomic, and no test in these packages calls `t.Parallel`, so their before-and-after deltas are reliable.
