# crochet:assess round 2: codebase lens

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` at c270bbfa3, blob **68371487a68e3f8c1661a80ea830356228a3a163** (`git hash-object`). Read-only. I read the subject and ran my checks first, then read the diff (`7b264fa9c..c270bbfa3`) last. I also read the round-1 minute and the C1 spike report.

Bias note: I hold these findings and I am also the one who releases them. Where I release, I name the exact text I release against, so a reader can check whether my bar was really met.

## 1. Round-1 findings

| Id | Disposition | Against / why |
|---|---|---|
| C1 | **Carried; I agree.** It is about which rules set the value. Revision 2 has no value (§2: "items stay (state, distance) and carry nothing new"). | §6 carries it. |
| C2 | **Released.** | §4.5: "refused only when every terminal it admits, among those that can follow the rule, refuses it". That is round 1's fix (a), `la ∩ FOLLOW(lhs)`. FOLLOW(termbinop) excludes the second terminal of `=`, `*`, `+`, `-` and `++` (`TestScratchAssessFollow`), so the filter fires on `$a = $a = …`. Caveat: §4's opening says "every terminal that can follow it *there*", and "there" could be read as a per-state LALR set. The issue should bind it to `grammarTables.follow`, the set the recognizer already uses. |
| C3 | **Released.** | §4.3: the filter does not run with `NoLookahead`. Acceptance with lookahead off is unchanged. The Options promise is rewritten. Checked against the code: `Expected` reruns `complete(nil)` (expected.go l.39-46), so a filter that lives inside `la != nil` leaves `Expected` alone. `TestLookaheadKeepsAcceptance` runs on `toyStmt`, which has no precedence table. The agree-test's corpus sample (agree_test.go l.63-67) is all accepted files, which the filter cannot refuse if §4.1 holds. The interleaved-mode loss is stated and is out of scope. |
| C4 | **Carried; I agree.** It concerns where the value lives and its re-propagation. | §6 |
| C5 | **Carried; I agree.** The reduce-side Leo concern is now §4.5's "A Leo memo never completes a link the filter refuses". That is achievable: the filter runs on the child's LHS in `complete` before `b.leo` is called (recognize.go l.425-431). The only multi-link memos I measured are `optlistexpr`/`else`, whose rules carry no pair. What is left (passing a value through unit-link memos) is value-only. | §4.5, §6 |
| C6 | **Released.** | §4: "The forest build refuses the same completions, so it never packs a derivation through one." On the reduce side, refusal depends only on the fact's rule and the token at the fact's end (`to`). So it is a property of the fact, not the parent, and the shared-node problem I raised cannot arise. `forest.build` must filter facts (a fact is made for every final item, recognized or skipped) and the `leoFacts` links; §4's sentence covers both. Optional, for an issue: dropping a refused single-rule final item in the recognizer makes the forest filter unnecessary for the 35 operator states. |
| C7 | **Held.** | Revision 2 deleted the indirob mechanism (round 1's "keyed by the parent item…" line). §3 and §6 still name 370 before `[` as the exception, and §6's stop says "beyond `indirob`", which implies it is applied with state. But §4 states only a rule-global table. Applied globally, (370, `[`) refuses `T1/array.t`'s `@$array[1..3]`. That is exactly the over-refusal the C1 spike recorded, and it would break §4.1 Exactness. No referee sentence would catch it: `testdata/precedence/uncovered-reviewed.txt` lists `370 PERLY_BRACKET_OPEN` as uncovered. The decision needs one sentence choosing between two options. (a) Leave the pair out and let the constraint handle it. That is the simplest choice and exact by construction. (b) Key it by chart state: state 373 holds the conflict and matches bison 193, state 232 matches 153; I verified both in `states.txt`. |
| C8 | **Carried; I agree.** | §6 |
| C9 | **Carried; I agree.** | §6 |
| C10 | **Held, narrowed.** | The removal gate is gone (§7: the constraint is unchanged), so that part is moot. What remains: §4.1 claims verdict and resolved forest are unchanged "on every corpus", but no test or key measures that. The gated verdict keys count disagreements with RD and perl. A file perl and RD both refuse could flip chart verdict without moving any of them. The resolved digest covers only the sample. A cheap, exact check exists because of §4.3: with lookahead off the filter does not run. So per file, `forest.Parse`'s verdict and resolved-forest digest with lookahead on must equal those with it off, over every corpus. Name that test, or an equivalent, as the evidence for §4.1. |
| C11 | **Released (moot).** | The constraint stays the authority (§2, §7). A reduce-side residue now only means less speed, never a correctness gap. |

## 2. Are §4's five properties true, testable and achievable?

### 4.1 Exactness

**True in part.** The subset claim holds for the 669 global pairs. For each such pair, count.pl shows no bison state reduces the rule on that token without a shift resolution. So no context of the constraint keeps what the filter refuses. This assumes `NewPrecedence`'s automaton agrees with bison's, which the README says it does on every toy. For indirob it is false unless C7 is fixed.

**Not true as written for "the forest digests (`forest/testdata`) are unchanged".** See C12.

### 4.2 Independence of the reference

**False as written.** See C13.

### 4.3 Lookahead

**True, testable and achievable.** C3 is released against it.

### 4.4 Every resolution accounted for

**Achievable.** `NewPrecedence` builds its states over the chart's own rules plus the refused ones. So rule numbering is already the chart's, and round 1's "373 against 350" mapping worry dissolves. The generator still has to do two things itself. It must enumerate conflicts over those states, because `NewPrecedence` exposes `resolve(r, t)` and the states but no enumeration. And it must decide, for each pair, whether it holds in every state, as count.pl does.

For the "cannot map" case, the generator has to read bison's own list (`testdata/precedence/bison-conflicts.txt`). Its own automaton never sees POSTJOIN (the chart's symbols lack it; 44 resolutions name it in `resolutions.txt`). So it cannot fail on something it does not see. Issue-level.

### 4.5 Leo and ambiguous tokens

**True and achievable.** See C2 and C5 above.

## 3. New findings

### C12: Blocking. §4.1 and §6: "the forest digests are unchanged" contradicts §4

There are two goldens in `forest/testdata`, and they cover different forests:

- `forest-digest.golden` (`TestForestDigestGolden`) hashes `built()`, which is `forest.Build` straight after recognition, *before* the constraint (resolve_test.go l.304-314, forestdigest_test.go l.77-79).
- `resolved-forest-digest.golden` hashes `forest.Parse`'s forest after the constraint and resolve.

§4 makes `Build` refuse completions. Take `$x = $y + $z;`, which the sample certainly contains in some form: the pre-constraint forest no longer holds the `(x=y)+z` alternative. So `forest-digest.golden` must move. §6's first stop ("a forest digest changes") then fires on the first correct build.

Fix:
- Hold `resolved-forest-digest.golden` unchanged.
- Expect `forest-digest.golden` to move, and regenerate it under a check that every alternative it loses is one the constraint unlinks. The lookahead-off build of C10 gives that comparison for free.

Consequence if it ships as written: the exit is unreachable, or the stop gets "interpreted" away.

### C13: Blocking. §4.2: the reference contract is false

The text: "the recognizer accepts exactly what the reference accepts and the precedence constraint leaves a derivation for".

The reduce side removes only one grouping of a `%nonassoc` or prefix/postfix conflict. Two counterexamples from the committed fixture, both "refuse" in bison:

- `s0568 qw(a) .. qw(a) .. qw(a) ;`: the filter refuses the left grouping ((DOTDOT rule, DOTDOT) is an error pair). The right grouping survives recognition, and the constraint refuses it.
- `s0395 ++ qw(a) ++ ;`: (rule 256, POSTINC) is an error pair, so `(++a)++` is filtered. `++(a++)` survives recognition.

So the recognizer accepts more than "reference ∧ constraint".

The true contract has three parts:
- the recognizer accepts a subset of what the reference accepts;
- it accepts at least everything that "reference ∧ constraint" accepts;
- with lookahead off it equals the reference.

The oracle as written needs a forest from the reference, which `refRun` does not build. `oracleDisagreement` currently requires equality with `refRun` under every Options combination (fuzz_test.go l.317 and after). Say which oracle replaces it.

Consequence if it ships as written: the oracle fails on the first such input, or it is weakened during execution without a decision.

### C14: Question for perigrin. §6: the exit's speed clause is already met

The exit requires median `speed-ratio-aggregate-longlived` ≤ 5. The committed baseline reads 4.7915, and the postmortem reports 4.95 under VERDICT. The clause can be met with no recognizer change; only the stop ("not below the before-reading's minimum") tests the decision. This was perigrin's ruling, so I raise it rather than hold it. State the stop as the real test, or set the exit relative to the before-reading.

### C15: Low. §3's citation of C1 overstates "operator groupings"

The C1 report says 47.9% of packed alternatives "come from a rule that is an operator rule or carries %prec". That includes unit `%prec` rules such as `term: scalar %prec PERLY_PAREN_OPEN`, which are not groupings. The share of groupings is lower than cited. Reword, or cite it as the report states it.

### C16: Low, issue-level. §4.1: refusal reason and byte may move from `precedence` to `grammar`

Wherever every chart derivation dies on the reduce side, recognition now refuses (reason `grammar`, with a byte) where `forest.Parse` used to return `Refusal{Reason: "precedence"}` at Byte 0. The verdict is unchanged, which is all §4.1 claims. I found no such case among the obvious ones: `$a <=> $b <=> $c` keeps its right grouping, so `TestParseAppliesPrecedence` still passes by my reasoning. The issue should still say whether a moved reason counts as a change.

## 4. What I did not re-examine

- Timing and the new stage keys.
- The slipper beyond its lookahead bits. A `;` or `}` is in no resolution pair, so it can never make a refusal.
- Branch sets and lattices with forks.
- The other assessors' findings (P*, A*).
- The v2 design and the chart-recovery plan cited in §1 and §2.

I did not build the filter. C12, C13 and C16 rest on reading the code and the committed fixtures, not on running a modified build.

## 5. Recommendation: **modify**

Two blocking findings, each fixable in a sentence:
- C12: hold the resolved digest, and let the pre-constraint digest move under a check.
- C13: restate the reference contract as the three-part bracket above.

Held:
- C7: decide how indirob is handled; excluding it is simplest.
- C10: name the lookahead-on/off equivalence test as the evidence for §4.1.

Released: C2, C3, C6, C11. Carried with agreement: C1, C4, C5, C8, C9.

Raised for perigrin: C14. Low: C15, C16.

The narrowing to the reduce side itself fits the code well. It needs no item change and no shared-node problem in the forest, and lookahead off gives an exact oracle for free.
