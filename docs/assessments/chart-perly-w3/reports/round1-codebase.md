# crochet:assess round 1: codebase lens

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`, blob 52c82af16a738eca6e03cdb050e9c167903212c7 (checked with `git hash-object`). Worktree `cr-close` at pu 8c3317079. Lens: does the design fit the code as it stands, and will its mechanism work on the chart's split LR(0) states and lattice input?

Method: I read recognize.go, sets.go, cache.go, leo.go, leo_memo.go, forest.go (build, leoFacts, split, alt), forest/precedence.go (NewPrecedence header), the README sections on the recognizer, Leo and the precedence constraint, and the ratchet baseline. I reproduced §3 with count.sh. I ran bison 3.8.2 on the pinned perly.y in the scratchpad and wrote three Perl scripts over its report (`shapes.pl`, `xr.pl`, `pred.pl`, `la.pl`) and one over the chart's committed tables (`mix.pl`). I also ran three scratch Go tests in package `recognize` through `go test -overlay`, which wrote nothing into the worktree: terminal sets per token, Leo memo shapes, and FOLLOW contents. Every script is in `.../scratchpad/assess-w3/codebase/`.

## 1. Factual claims in §2 and §3

| Claim | Result | Evidence |
|---|---|---|
| The seven §2 keys and values | Correct | `testdata/ratchet/baseline.txt` |
| "sum about 2.1, remaining 2.7" | Arithmetic correct (0.6551+0.7903+0.6620 = 2.107; 4.79 − 2.11 = 2.68). The stage keys are separate long-lived measurements, so the remainder is only an estimate of the post-forest cost (README "The speed protocol"). | README l.636-650 |
| Postmortem probe, 115 ms against 25 ms | Correct | `docs/postmortems/chart-perly-w2.md` l.229-230 |
| README: `$a = $a = …` is quadratic "with Leo on or off" | **The quote drops a qualifier.** The README says "at lookahead 0", and that is the one mode in which §4.1's lookahead filter does nothing (see C3). | README l.1941-1944 |
| §3 table (670/669/1, 667/667/0, 10/10/0, 57 states, inconsistent 0, the pair 370 before `[` reducing in state 153) | Reproduced exactly | `count.sh /opt/corpora/perl5` output |
| 772 chart states, 51 completing more than one rule, state 439 completing rules 271 and 275 (`term ARROW CAST_STAR`) | Correct | `grammar/testdata/tables/states.txt`, `rules.txt` |
| bison has 373 rules, the chart 350 | Correct (bison 0..372) | bison report; `rules.txt` |
| perly.y declares 32 levels | Correct (32 `%left/%right/%nonassoc` lines). The PLUGIN_* tokens are already inside them: 3 have a line of their own, the rest share a line. | perly.y l.150-182 |
| Chalk's `add` keeps one value and prefers the more constraining | Correct at 287e5da1 | `Semiring/Precedence.pm` `method add` l.574-620 |
| "precedence-disagree 0 and precedence-covered 1304 are the equivalence proof" | Overstated (C10) | README l.473-476, l.2510-2516 |

## 2. Findings

### C1: Blocking. §4.2: setting the value to "an operator rule's level" refuses `++ ++$x`, which bison accepts

bison resolves (rule 256 `PREINC term`, POSTINC/POSTDEC/POSTJOIN) and (rule 257, the same three tokens) as **error**, because `%nonassoc PREINC PREDEC POSTINC POSTDEC POSTJOIN` puts them on one level. Under §4.2, the completion of the inner `++$x` sets the value to that level, and the "error" pairs remove that level from rule 256's last operand. The outer `++` therefore gets an empty value, and `++ ++$x` is refused. bison accepts it: toy `testdata/precedence/toys/prefix.txt` records `accept … PREINC PREINC NUM`, and the README says "Prefix after prefix is no conflict and is kept (`- - 1`, `++ ++ $x`)". The perly referee would not catch this: `testdata/precedence/sentences.txt` has no sentence with two prefix increments.

Root cause: the shift side concerns only rules whose operator follows a left operand. Running `shapes.pl` over every one of the 677 reduce and error resolutions shows the shift item is always `X: Z . t …` with the dot at 1. `xr.pl` lists the 33 X rules: binary operators, `expr ANDOP/OROP`, the comparison chains, the postfix `term POSTINC/POSTDEC/POSTJOIN`, and the ternary. Each X rule's level equals its token's level.

Fix for the decision text: the value is set only by left-operand rules. Prefix rules (`PREINC term`, `PERLY_MINUS term %prec UMINUS`, `UNIOP term`, …) set "no operator". Chalk makes the same separation (UnaryExpression at level −1 at 287e5da1). If this ships as written, a correct input is refused and no referee notices.

### C2: Partial. §4.1 does not say which terminals the filter tests, and the literal reading does not fire on §2's own example

The lexer gives the common operators two terminals each. My scratch test, `TestScratchAssessSets`:

- `=` is {ASSIGNOP, PERLY_EQUAL_SIGN}
- `*` is {MULOP, PERLY_STAR}
- `+` is {ADDOP, PERLY_PLUS}, and `-` is {ADDOP, PERLY_MINUS}
- `++` is {POSTINC, PREINC}

Over std5k.pl plus perl5 `lib/*.pm` (37 files, 46,404 tokens), 2,595 of 7,455 operator tokens have more than one terminal.

The second terminal is never in a bison resolution pair. So a table holding §3's "shift" pairs, applied as written ("skips an LHS when every terminal the lookahead admits is in that set"), never fires on `$a = $a = …` or `$a + $b * $c`.

It does fire if either of these holds:
- (a) the test runs over `la ∩ FOLLOW(lhs)`. FOLLOW(termbinop) and FOLLOW(term) hold MULOP, ASSIGNOP and ADDOP, but none of PERLY_EQUAL_SIGN, PERLY_STAR, PERLY_MINUS, PERLY_PLUS or PREINC (`TestScratchAssessFollow`).
- (b) the table is the complement of bison's reduce lookaheads for the rule, which is what "never reduces it" literally says. `la.pl` shows bison never reduces rules 215, 219 or 221 on any of those second terminals.

Under (b) the filter is an LALR(1) lookahead filter, which is more than precedence, and §3 does not count it.

The decision should name one of the two, and the generator's test should include `$a = $b = $c` with the real terminal sets. With the definition fixed, §5.3's residue shrinks to the word operators that share a set with ATTRLIST or BAREWORD (about 20 tokens in the sample).

### C3: Missing. A lookahead filter breaks the Options contract, and interleaved mode loses it

`recognize.Options` documents that "Neither changes acceptance; the tests and the fuzz target hold them to that". `TestLookaheadKeepsAcceptance` and the fuzz `oracleDisagreement` enforce this: the latter compares every combination in `optionCombos` against `refRun`, a plain Earley reference.

§4.1 places the reduce side in `builder.complete`'s lookahead path, which `NoLookahead` switches off (`follows = nil`, recognize.go l.276-280). So, as written:
- acceptance differs between lookahead on and off;
- the reference recognizer no longer agrees with either;
- at lookahead 0, the mode interleaved lexing will use (README l.1931-1934), half of precedence vanishes.

The decision must say three things:
- whether the reduce-side table applies regardless of `NoLookahead` (a separate check from the FOLLOW prune);
- what the fuzz oracle becomes (a reference with precedence, or Constrain on the reference forest);
- whether `Expected` (expected.go, built "as recognition without lookahead builds it") should reflect precedence.

If this ships as written, the fuzz target and `TestLookaheadKeepsAcceptance` fail, or someone weakens them.

### C4: Partial, and it reshapes §5.1. The value need not live in items at all

The value is read in exactly one place: when a completed child at node `pos` completes into its parent at `pos`. Everything else either sets it or never reads it:
- A child's value is fixed by completions ending at `pos`, all of them in that node's closure. An operator rule sets its level, a unit rule takes its child's value (same end node), and a scan sets "no operator".
- A parent item's own value is never read.
- No later set reads any value from an earlier set. The scan advances by state, and `memoAt` reads states.

So the value can be a per-builder annotation, discarded once `relative()` is taken. Then:
- set identity, `intern`, the transition cache key and its dependencies stay as they are, because the result is still determined by (source set, token, la, origin sets);
- §5.1's loss of `set-reuse` does not happen;
- the "facts in Earley items" stop condition lifted by the §1 ruling may not need lifting.

If the value is stored instead, values on in-progress items (for example the left-operand state after `$a + $b` against after `$a * $b`) split sets for no reason.

Either way there is a real requirement the decision implies but does not state. `builder.complete` visits each item once (`for i := 0; i < len(b.items); i++`), and `isNew` drops a duplicate (state, origin). An OR-merged value that grows after its item was processed must be re-propagated. Otherwise §4.2's "cannot lose a derivation" is false, and the failure shows as false refusals. This is issue-level, but the guarantee sits here.

### C5: Partial. §5.2 aims at the wrong Leo interaction

Operator chains never form Leo memos. At the node after an operator, the predicted state's items also await `term`, so `memoAt` finds n ≥ 2. `TestScratchAssessLeo` over the same 37 files:
- memos with more than one link: only `else` (14) and `optlistexpr` (239);
- single-link memos: termbinop 1,652, myattrterm 471, listop 808, termeqop 187, termrelop 97, termunop 110, among others.

The single-link termbinop, termeqop, termrelop and termunop memos are the unit links (`term ::= termbinop` from a predicted item) that must pass the value up. `builder.leo` returns only `top`, and `complete` adds it and `continue`s before the `awaiting` loop. So the value must travel through the Leo path, or the shift side loses it on most of its traffic.

Answer to open question 2: link-by-link completion of mixed-level chains is moot. A memo must carry the value (a pass-through for unit links; for `optlistexpr`/`else` chains the value is constant "no operator").

### C6: Partial. §4.3 is right that the forest must re-apply the tables, for a broader reason and with a gap

`forest.build` makes a fact for every final item in each set (`for _, r := range tab.States[it.State].Reduces`), including items whose completion the recognizer skipped. `factNode` and the work loop then split every rule with the same (sym, from).

So the reduce side leaks too. For example, at `$x = $y + $z + $w`, both `termbinop` facts over `$x..$z` exist, and the assignment one is packed under the outer `+` even though its completion was blocked.

The gap: "applies the same two tables when it packs a derivation" cannot be done per parent, because a child node (sym, from, to) is shared by parents with different admitted sets. The decision should either:
- adopt Constrain's documented shortcut ("an alternative kept in one parser state is kept in all", README l.2527-2531), or
- key forest nodes by admitted set.

Cheaper: drop a blocked final item rather than only skipping its completion. Each of the 35 chart states that complete an operator rule holds exactly that one reduce and no non-final kernel item (`mix.pl`), so a dropped item loses nothing else and produces no fact.

### C7: Ready, with a simplification. The `indirob` pair and how the tables are keyed

The chart's split states already separate the indirob case:
- state 373 holds {`subscripted ::= scalar [ …`, `term ::= scalar`, `indirob ::= scalar`} and matches bison state 193, where the conflict is;
- state 232 holds {`indirob ::= scalar`} and matches bison state 153, where the reduction is allowed.

So a (state, LHS) key expresses the pair without the parent-item key. The parent-item key also works: `listop ::= LSTOP . indirob listexpr` is not a Leo item, so that completion goes through `awaiting`.

But §5.4 says the generator reuses `NewPrecedence`, whose states are bison's LR(0) automaton and "never the chart's split states" (precedence.go doc comment). The decision should say whether the reduce table is keyed:
- by rule, which `resolve(r, t)` gives directly and works for 669 of 670 pairs; or
- by chart state, which needs a chart-to-bison state map that does not exist today.

### C8: Partial. §4.2's value width and identity element

32 levels plus a reserved "no operator" bit is 33 bits, which does not fit a `uint32`. Only the levels of left-operand rules can be on top (about 20 distinct levels, `xr.pl`), so it fits once C1 is applied.

"The empty set's identity, a reserved bit" contradicts itself. The semiring's zero has to be ∅, which is what "an empty result is zero" says, so "none" must be a real member.

Answer to open question 1: the PLUGIN_* tokens are already among the 32 levels.

### C9: Partial. §4.2: "unit rule" needs a definition that includes `%prec` units

These must pass the value up, not set PREC_LOW:
- `termbinop: termrelop %prec PREC_LOW`
- `termbinop: termeqop %prec PREC_LOW`
- `termrelop: relopchain %prec PREC_LOW`
- `listexpr: term %prec PREC_LOW`
- `expr: listexpr %prec PREC_LOW`

Without that, `$a + $b < $c` and `$a and $b and $c` group wrongly. bison does use these rules' own precedence, but only in reduce-side pairs: for example rule 236 against CHRELOP in bison state 68, which the (state, LHS) filter handles. Folding this into C1's definition settles it.

### C10: Partial. §5.5 and §4.3: there is no equivalence proof, and the residue has no key

`precedence-covered` 1,304 is the number of the 1,347 resolutions that the 1,296 generated sentences cover. Forty-three resolutions are uncovered (reviewed), and C1's case is absent.

§4.3's removal condition ("the tables alone agree with it on every corpus") has no key. `precedence-disagree` reads only the generated fixture.

Needed:
- a corpus-wide key counting what `Constrain` still removes after the tables (open question 3's residue);
- running the 13 toy referees against the generated tables, not only against `Constrain`. That would have caught C1.

### C11: Ready, as support. Checking the top operator alone is sound for perly.y, but only alongside the reduce side

This supports §4.2's design and should be written down, because it ties the two sides together.

**Every shift-side parent is R.** `pred.pl` looked at each pair of (R, the bison state before R's last operand). In 43 of 46, R's item is the only kernel item that predicts the operand. In the other 3 (`KW_DO block`, `UNIOP block`, `KW_MY REFGEN myterm attrlist`), the alternative cannot begin a left-operand rule. So in perly.y the parent of a shift-side derivation is always R.

**Deeper spine nodes are covered by the reduce side, not by the shift side.** Along the left spine of R's last operand, each inner left-operand node completes before the next operator token. The reduce-side check therefore forces each inner level to bind at least as tight as the one above it. That makes checking only the top node enough.

Consequence: any reduce-side residue (C2's undecided terminals, or C3's lookahead-off mode) also opens a shift-side hole.

## 3. Classification summary

| Area | Classification |
|---|---|
| §2 and §3 facts | Ready (one quote missing "at lookahead 0") |
| §4.1 reduce-side filter | Partial (C2, C3, C7) |
| §4.2 shift-side value | Blocking (C1), plus Partial (C4, C8, C9) |
| §4.3 forest | Partial (C6, C10) |
| §5.1 interning and cache | Partial: avoidable by design (C4) |
| §5.2 Leo | Partial: aimed at the wrong interaction (C5) |
| §5.3 ambiguous lexemes | Partial: mostly resolved once C2 defines the filter |
| §5.4 rule mapping | Partial (C7) |
| §5.5 referee | Partial (C10) |

## 4. Recommendation: **modify**

The mechanism fits this recognizer, and my measurements back the design's core bet. Every reduce and error resolution has the left-operand shape, the parent is always R, and the states that complete an operator rule are clean.

Blocking before acceptance:
- C1: define which rules set the value.

Should be decided in the decision text:
- C2: which terminals the filter tests.
- C3: the relation to `NoLookahead` and the fuzz oracle.
- C4: whether the value is builder-local, and the re-propagation requirement.
- C6: how the forest handles shared nodes.
- C10: a residue key and running the toys against the tables.

Issue-level, not for the decision:
- C5: carrying the value through the Leo path.
- C7: how the table is keyed.
- C8: the integer width.
- C9: the list of unit rules.

## 5. What I did not examine

- Performance: I made no timing runs.
- The slipper's closure and value interaction (slipper.go).
- Branch sets and interleaved lattices with more than one out-edge.
- How `Expected` and the refusal byte should change.
- `cache.go` dependency recording under a builder-local value, beyond the argument in C4.
- `resolve` and `readings`.
- `Constrain`'s internals beyond its README section.
- PerlOnJava and perl.git `t/` corpora: the Leo and lexer samples are std5k.pl plus perl5 `lib/*.pm` only.

I did not run `make test`, and I did not modify the worktree.
