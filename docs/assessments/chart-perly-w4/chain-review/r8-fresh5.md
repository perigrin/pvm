# chart-perly-w4 chain-review, round 8 (the bound): review-fresh5 (resumed)

I read the live chain first and diffed it against my round-7 dump (`fr7x/` against `fr8x/`). I ran my checks, then read `repair-r7.md`.

**What the repair changed.** It edited d166, d218 and 3262, and appended one Records entry. Nothing else moved:
- no dependency edge, title or label changed;
- the residual mapping is unchanged;
- the new criteria (`TestNarrowedSetIdsStableAcrossFiles` and d218's joint reference criterion) name tests that do not exist yet, so they are red today.

## 1. Recommendation

**Modify.** My own blocking finding, FR5-3, is settled and released. I promote one residual row to blocking: r7-fresh7 FR7-1, the README claim in bb9b's item 8. Its fix is one sentence, but it is a defect in what the chain asserts. Since this is the bound, it goes to perigrin as a held objection.

## 2. Status of my findings

I am biased toward finding my own bar met, so I scored the repair against what I wrote, not against what landed near it.

### FR5-3: released

**What I predicted the fix would also have to change:**
- name each existing reference test;
- give a rule for inputs the judged chart refuses;
- deal with `TestWaypoint2RefusalsMatchReference`'s scope rows and precedence rows;
- keep 3262's joint criterion green at verify;
- say which instruments stay on the no-precedence mode;
- add a criterion.

**What I checked:**
- **Named tests.** d218 names six repointed tests: three in `forest/precedence_reference_test.go`, `TestPrecedenceCorpusMatchesReference`, `TestPrecedencePerlGitMatchesReference` and `TestWaypoint2RefusalsMatchReference`.
- **Refused inputs.** It gives them d247's verdict-only rule. The reference must refuse too, the input is counted in a logged `<test> verdict only: N` line, and the test never errors "recognition leaves no forest".
- **Precedence rows.** The criterion requires N > 0 for waypoint2. It can be met: `waypoint2Refusals` holds 10 rows that perl refuses for precedence (`$a <=> $b <=> $c;`, `1 .. 2 .. 3;`, `print $fh [];`, the `isa` chains and others).
- **Scope rows.** `waypoint2ScopeRefusals` are resolve-scope rows (`field`, `method`, `ADJUST`, `:lvalue` signatures), not library operators. The test runs them under `Scope{}`, which `LibraryOperators` already passes today. So narrowing does not kill them, and they stay full comparisons.
- **The no-precedence instruments.** d218 names them: the harness tests, `precedence_statefree_test.go` and `precedence_pinned_test.go`. These are the tests that test `Constrain` or the comparison itself.
- **3262's joint criterion.** Its command is unchanged and only checks PASS. Its prose now says the test runs on the no-precedence mode at 3262 and on the judged forest from d218. It stays green at verify.
- **The perl.git half.** It sits behind its build tag, outside the joint criterion. The milestone-end `make -s chart-precedence-reference` criteria in d089 and d10d run it wherever perl.git is present.

My bar was to name the tests, give a rule for refused inputs, and add a criterion. That bar is met.

### FR5-2: held, residual

3262's rerun prose is unchanged. d166 still states the no-rerun shortcut for an unmapped token in its own text, and 3262 can be built to keep it.

### FR7-1 (`ParseWith` guard when the derivation failed): held, residual

The repair did not address it. The guard can be built correctly.

### FR7-2 (the "13 toys, less `empty` and `emptyred`" count): held, residual

The repair did not address it. The general exception still covers any further toy that `Feasible` refuses.

### FR5-1, FR6-1, FR5-4, FR5-5

These were released in earlier rounds. Re-reading them against round 8's text, I found nothing that undoes their fixes: the `ParseWith` guard paragraph, the three-way list and the cost bounds are byte-identical to round 7.

## 3. Promotions

I read every residual row in r1 to r7, including the two rows the orchestrator flagged.

### r7-fresh7 FR7-1 (bb9b, README item 8): promoted to blocking

**What item 8 says.** bb9b tells the docs issue to rewrite the "Known limits" paragraph to say that "an alternative kept in one context is kept in all" now applies to `Constrain` only, "since production's Build walks by context".

**Why that is false for V2.** The README paragraph (`internal/chart/README.md:2716-2720`) says the shortcut "can only add a derivation to what a later reader enumerates". V2 requires its production forest to equal V1's, node for node and alternative for alternative:
- design §2's goal;
- d0ba's `TestForestEqualityUnionShortcut`, on a node reached under two live contexts with two different non-empty sets;
- d218's "every production differential is green for real".

So V2's forest carries exactly the derivations the shortcut adds.

**What walking by context actually does.** It keeps a *dead* context's alternatives out (C6, 3282). V1 drops those too, so V2 does not lift the shortcut.

**Why it blocks.** bb9b instructs the README to state something two other issues' criteria contradict. That is a false claim in the subject, and a contradiction between its parts, so it is blocking however small the fix.

**Fix.** Reword item 8 to say three things:
- the union is a property of the shared-node forest and holds for V2's production forest as for V1's;
- walking by context keeps a dead context's alternatives out, which V1 also does;
- `TestPrecedenceShortcutStated`'s phrase stays.

### r7-fresh7 FR7-2: not promoted

I would stand aside on it alone. 3282's dead-context test needs a constructed input that d0ba's two-live-context fixture lacks. The repair can be "a second input in 3282's own test, beside the fixture", which amends nothing in d0ba. The phrase "on the union-shortcut fixture" makes that awkward but does not forbid it.

### All other rows: not promoted

I promote none of the other rows. These are each true and residual on their own terms:
- AL5 to AL11, AL15, AL16, AL18;
- PB8, PB12 to PB15, PB18 to PB20, N2 to N4, N8 to N11, N13, N15 to N17;
- FR3-2, r7-fresh2 FR6-1 and FR7-1;
- r7-fresh4 FR7-1;
- r7-fresh6 FR6-3 to FR6-9, FR7-1 and FR7-2;
- r7-fresh7 FR7-3 to FR7-11.

Two of them are close calls:
- **r7-fresh7 FR7-9 (d199's size)** is the closest. d199 is cohesive around one constructor pair, so I would stand aside on it.
- **r7-fresh6 FR6-6** is now answered by d166's stable-id test.

## 4. New findings

None.

## 5. Not re-examined

- I ran no `go test` criteria.
- I did not re-trace the issues the repair left byte-identical. Those are every issue except d166, d218 and 3262.
- I did not verify `TestNarrowedSetIdsStableAcrossFiles`'s fixtures beyond reading the criterion.
