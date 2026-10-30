# Milestone chart-waypoint Postmortem

Stopped by the §9 speed kill on 2026-10-30, one day into a 15-day estimate
(buffer burned 6.0 of 12.5). 14 of 25 issues done. Final measurement after
the token-mapping memo: std5k.pl 2.83–3.85× recursive descent, corpus
aggregate 3.92–5.00× (four runs, median of 9 each), against a 3× line.
perigrin ruled the spike re-attempted without Guacamole; the generator,
recognizer, Leo and ratchet carry over.

## What Worked Well

- **The kill fired on time and on evidence.** The speed check (2f46) ran as
  soon as the recognizer accepted std5k.pl, before the forest, error recovery
  and the oracle adapter were built: 11 of 25 issues were unfinished, 9 of
  them never started. Repeated runs made the verdict robust to load: the
  lowest-load run gave 3.915.
- **Per-issue gates caught real defects that every criterion missed.**
  2eec: a branch set's id −1 keyed the transition cache, so a warm cache
  flipped acceptance; `Expected` was emptied by lookahead. 2f18: Leo wrongly
  rejected 16 of 758 random grammars; the shipped chain walk was cubic.
  306a: the unreproduced count of 2 depended on a map mixing two rules
  (3–4 of 4 when one rule is applied). Reviewers' adversarial inputs caught
  each; no issue criterion did.
- **The milestone review produced an outside oracle.** The quality lens wrote
  a textbook Earley recognizer and found 0 disagreements over 770 corpus
  files, 3,000 mutated files and 11 adversarial grammars, with a sabotage
  control that went red 45 times. The carried-over core is verified
  independently, not only against itself.
- **Stop rules held.** Every kill or stop condition that fired (2e94's count,
  the parsent `< $fh >`, speed) stopped work and went to perigrin rather than
  being worked around. Leo's untestable criterion was redefined by the
  orchestrator and recorded for perigrin to overrule, not escalated.

## What Didn't Work

- **The grammar was translated faithfully, not fitted to its components.**
  Guacamole is scanless; pvm's lexer is not. The token cut (2e3a) compiled
  each G0 lexeme to an anchored regexp — literals included — and chart/lex
  ran every predicate of a token's kind on every token: 110 ms of 150 on
  std5k.pl, against the production lexer's 9 ms. Most of the milestone's
  amendments trace to the same mismatch: `*` as a bare glob sigil, `x::foo`
  as one word, the operator ladder spread across LR(0) states, `< $fh >`
  admitted, `StatementSeq` defeating Leo, `Program` not deriving the empty
  program. Nothing in assess or chain-review asked whether the grammar fit
  the lexer.
- **Recognizer hot spots were left for the speed check to find.** `fmt.Sprint`
  class keys took 74% of recognition (std5k 34.9 → 18.3 ms warm when keyed
  as bytes). No issue carried a benchmark, so the kill measured two
  implementation choices together with the algorithm.
- **2e94 took five definition rounds.** The issue pinned "seven conflicts"
  from Standard Perl; tagging gave 108, three orchestrator rules gave 20 with
  `*` lost, and only perigrin's "observe the lexer" settled it. The
  orchestrator amended the definition twice, the second time with three rules
  that each needed an exception, before escalating.
- **Refinement defects reached execute.** A criterion glob matched
  parseoracle's baseline twice (2da1 N3, 2f46); 2e6b N2 checked an
  uncommitted copy; 2ec0 required `lex-unmapped 0` against open lexer
  issues; 2e94's Leo eligibility ("a state whose single item…") made
  2f18's own test unpassable.
- **Branch rules were unstated until violated.** #536 and #544 merged `pu`
  into their branches before perigrin ruled "Branches should be FF on pu".

## What Puzzles Us

- The mapping memo more than halved std5k.pl but cut the aggregate only by a third.
  The corpus files are small and the memo was per call; the aggregate after
  the fix was never profiled.
- Worker 3's 20,000-grammar probe could not reproduce Leo's F1; the
  reviewer's 758 grammars found it 16 times. The generators differed in
  whether the start symbol may appear on a right-hand side.
- Recursive descent's own std5k medians ranged 18.6–30.8 ms on one machine,
  a spread larger than some of the effects being measured.

## Where A Human Had To Act

- **Collaboration (the protocol reserves these):** the speed-kill response,
  the parsent kill, the redesign, FF-on-pu, ponytail in the briefs, the 2ec0
  fourth review pass, commitlint style.
- **Friction:** the 2e94 rounds (four of five could have been one question
  with the lexer-observation option on it); the disk filling (19 GB Go build
  cache, stalled worker 4); "seven" written into an accepted spec as a fact
  rather than a hypothesis; git-zhi has no terminal state for a killed
  milestone, so closing out needs cancellations and an open milestone.

## What Will We Change

1. **Assess asks a component-fit question.** For any grammar or format
   adopted from elsewhere: what is the interface between it and our
   components, and is the translation fitted to that interface or faithful to
   the source? The re-attempt states its token interface (lexer kind and
   text to terminal id, one lookup) before any grammar work.
2. **Benchmarks and an overhead gate from the first performance-bearing
   issue.** The re-attempt carries `lex-overhead ≤ 1.3×` as a gate and
   benchmarks for recognition and `Expected` from its first recognizer
   issue, not a single speed checkpoint.
3. **Algorithmic components get an outside reference in tests from the
   start.** Keep the review's reference Earley recognizer and use it in the
   fuzz target.
4. **A definition amended twice goes to perigrin.** A third orchestrator
   amendment of the same definition is a stop, not a fix.
5. **Hypothesised counts are criteria against a reviewed list, never a pinned
   number in the spec.**
6. **Refinement lints criterion globs for anchoring** (`internal/chart/*…`)
   and rejects criteria that can't pass while another milestone's issue is open.
7. **For crochet:** the worker briefs now carry FF-on-`pu` and ponytail, and
   the gate brief carries FF-on-`pu`; the missing killed-milestone state goes to git-zhi.

---
Generated by crochet:postmortem on 2026-10-30
Data sources: issue list, milestone telemetry, ## Records (which
holds the source of every figure here), the review gate's three lenses, 2f46's measurements. Not used: sanbao (available, not run), docs health.
