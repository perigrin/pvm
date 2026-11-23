# crochet:assess round 8: codebase lens

**Subject.** `docs/plans/2026-11-23-chart-stateless-precedence.md`, blob **ed91e46239b29283ad5c82cd390411621b52a2cd** (`git hash-object`). The worktree HEAD is 146411199, the minute on top of 336067853. This is revision 8.

**Order of reading.** The subject's §3 and §4 first. Then `cmd/ratchet/speed.go` (`Protocol.pass`, `Ratio`, `Aggregate`, `measureStages`, `chartStages`) and `forest/resolve.go` (`libraryOperators`). Then `chart/precedence.go` (`judge`, `narrow`) and the `timeout-minutes` in `.github/workflows/chart-ratchet.yml`. Read-only.

## 1. C38

**Released.** It is released against the in-tree row of §3.4's table: "Every recognized file of the ratchet's in-tree corpora (T1, Guacamole parses and parsent, std5k, edit5k), perl-refused files included (new test in `cmd/ratchet`, using §4's new API)."

The row now names `edit5k` and `parsent` explicitly, and it drops the population, so it no longer needs labels or RD verdicts. `corpus()` (lex.go l.40-70) yields exactly that in-tree set without perl.git. Including perl-refused files that recognition accepts widens the refused-path coverage, which helps.

I raised C38 and am also releasing it, so I may be biased toward finding it addressed; the release rests on the quoted text.

## 2. Revision 8's new text against the code

### The per-pass preparation hook: accurate and workable

`Protocol.pass` (speed.go l.54-65) does this:
1. calls `p.GC()`;
2. starts the clock;
3. builds runners (`side()`, once, or per file when cold);
4. runs each file.

Nothing runs per pass before the clock today, so the decision's "Today `Protocol.pass` starts the clock before anything per pass runs" is right. A hook placed between the GC and `start := p.Now()` that copies one pass's forests matches "holding one pass's copies at a time".

The hook must apply only to the chart side. `Ratio` alternates `p.pass(p.Chart…)` and `p.pass(p.Ref…)`, and `Aggregate` also calls `pass` for RD.

### "One measurement is one ratchet run's printed value": accurate

`Ratio` (speed.go l.68-86) runs `measurements = 3` measurements. Each takes the median of `timedPairs = 9` ratios after `warmupPairs = 2`. It returns the median of the three. `measureStages` times every stage key through `p.Ratio` (speed.go l.255). So the printed value is "the protocol's own median of its three internal measurements", as stated.

### The CI timeout

`chart-ratchet.yml` sets `timeout-minutes: 20`. §4 requires the keys' cost to keep the run inside it, and to be reported. I did not measure the run's current duration. That stays the adding commit's check, as the decision says.

### The referee compares "Build's forest narrowed to the sentence's terminals": accurate

`judge` (precedence.go l.315-327) does three things in order:
1. `builtForest`;
2. `narrow(f, want, skip)`, which unlinks alternatives in place;
3. `f.Constrain(r.p)`.

The comparison has to give the reference and production two copies of the forest *after* `narrow`. Either use §4's exported copy or narrow two builds. `narrow` is unexported in package `chart`, which is where the referee tests live, so this works.

### "The forest after `libraryOperators`": accurate, and it corrects an earlier claim of mine

`libraryOperators` (resolve.go l.630-656) is not read-only. When `scope.LibraryOperators` marks a plugin terminal dead, it calls `r.prune(f)`. In rounds 3 and 6 I wrote that it "only reads the forest" and that "Build's output is exactly the constraint's input". That was wrong for any file whose scope names library operators. Revision 8 says "the forest after `libraryOperators`" for `ParseWith`, for the corpus checks, and for the precedence stage key. That is the correct boundary.

### C39: Low, editorial. Say which scope the corpus checks pass to `libraryOperators`

The constraint's input depends on `Scope.LibraryOperators`. A reference-against-production comparison is valid for any scope, because both sides receive the same forest. But only the pipeline's own scope reproduces the pipeline's input. In the ratchet that is the RD-derived `rd(path).scope` from `chartInputs` (verdicts.go). The `cmd/ratchet` tests should use it, and the issue should say so. Refinement can carry this; it is not a round finding.

## 3. What revision 8 might have broken

I found nothing.

- "Only `internal/chart` changes" now admits the Makefile and CI by ruling. That is consistent with the split of round 7.
- The verdict reading's definition agrees with `Ratio`.
- The referee's narrowed-forest boundary agrees with `judge`.
- The caller list and the new-API list are as in round 7, and still complete.

## 4. Held findings, and the fixed point

**I hold nothing.** From this lens:

- C1 to C38 are released, moot or carried.
- C33 is issue-level, for refinement.
- C39 is an editorial note for refinement.

I **confirm a fixed point** from the codebase lens.

## 5. What I did not examine

- The plans diff beyond §3 and §4.
- §5 and §6.
- The ratchet's current wall time against its 20-minute timeout.
- The runtime of the new `make test` checks.
- Whether CI's perl5 pin is 94e50866.
- The other assessors' round-7 reports.

## 6. Recommendation: **accept**
