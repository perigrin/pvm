# chain-review round 7: refinement repair on chart-perly-w4

## Scope

This round answers the two blocking findings that remained open.

1. **r6/r7-fresh2 FR4-1, together with r6/r7-fresh5 FR5-3.** d218 repointed production-versus-reference differentials without naming them.
2. **r6-fresh6 FR6-6, promoted by r7-pushback.** Narrowed set ids in the transition cache.

No residual finding was answered.

- All edits went through the `git zhi` CLI, on git-zhi 0.5.2.
- The repository, git config and branches were not touched.
- No issue, edge, title or label changed. The milestone still has 19 issues.

## 1. d218 names what it repoints (r6/r7-fresh2 FR4-1, r6/r7-fresh5 FR5-3)

d218's "Tests this change moves" section now lists every test by name.

### Repointed

Design §5 says "'Production' now means the recognizer's forest", so these tests stop running on the no-precedence constructor:

- `TestPrecedenceToyForestsMatchReference`, `TestPrecedenceOperatorSentencesMatchReference` and `TestPrecedenceChainSentencesMatchReference`, in `forest/precedence_reference_test.go`;
- `TestPrecedenceCorpusMatchesReference`, in `cmd/ratchet/precedence_reference_test.go`;
- `TestPrecedencePerlGitMatchesReference`, in `cmd/ratchet/precedence_perlgit_test.go`, behind the `precedencereference` tag;
- `TestWaypoint2RefusalsMatchReference`.

3262 pinned all of these to the no-precedence constructor.

- **The candidate** is the judged recognizer's forest, then Build, with no `Constrain`.
  - Over perly.y's tables it comes from `recognize.New`.
  - Over a toy it comes from the explicit form, using the toy's own precedence.
- **The reference** is `ConstrainReference` over the no-precedence forest. Where a test already applied `LibraryOperators` under each file's scope, it still does.
- **The infeasible toys** stay `Constrain` referees on the no-precedence mode.

### Inputs the judged chart refuses

These are the precedence rows, and the scope rows once the lattice is narrowed. Each test applies d247's verdict-only rule to them:

- The reference must refuse the input too, either because `ConstrainReference` leaves the root no derivation or because the no-precedence chart or scope refuses it.
- The input is counted in a logged `<test name> verdict only: N` line.
- The test never errors "recognition leaves no forest".

Where both sides accept, the kept alternatives must still be equal.

### Staying on the no-precedence mode

These are instruments, not differentials, so they do not move:

- `TestPrecedenceDiffDetectsDifference`, `TestComparePrecedenceUsesScope` and `TestPrecedenceHarnessesDoNotAlias`;
- `precedence_statefree_test.go`;
- `precedence_pinned_test.go`, including the known-limit pins.

### 3262 stays green at milestone verify

`TestWaypoint2RefusalsMatchReference` keeps its name and passes. 3262's joint criterion runs it, and that criterion asserts only that it passes, so it stays green after d218.

- **3262's pinned list** now says that the Constrain-removal issue repoints these differentials, and that 3262 pins them only until then.
- **3262's joint criterion description** no longer claims the waypoint2 check runs on the no-precedence mode for good. Its command is unchanged.

### New in d218

**Step:** repoint the named tests, applying the verdict-only rule.

**Step:** where perl.git is present, run `make chart-precedence-reference`.

**Criterion:** run the five tests that need no perl.git, with `T1_CORPUS` set. It requires that:

- each test passes, with no skip;
- each test logs `<name> verdict only: N`;
- `TestWaypoint2RefusalsMatchReference` logs N > 0. It has precedence rows, among them `$a <=> $b <=> $c;`, `1 .. 2 .. 3;` and `$a == $b <=> $c;`.

## 2. Narrowed set ids keep their contents (r6-fresh6 FR6-6)

d166 now states:

- A narrowed terminal-set id keeps the same contents for the whole life of a recognizer.
- An id is never reassigned to a different set. So the transition cache and the intern table, which key on set ids, never serve one file's transition to another file whose narrowed set differs.
- Through one long-lived recognizer, each file gets the forest a fresh recognizer would give it.

**Step:** a TDD step for this property.

**Criterion:** `TestNarrowedSetIdsStableAcrossFiles`, with subtests `a-then-b` and `b-then-a`. It takes two files that narrow to different terminal sets: one with a library operator in scope, one with it out of scope. Each file is parsed through one long-lived recognizer, in both orders. Each must produce exactly the forest and refusal that a fresh recognizer gives.

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21447-d218 | Named the repointed differentials and the instruments that stay, stated the verdict-only rule, and noted that 3262's criterion stays green. Reworded the repoint step, added the perl.git step, and added the five-test criterion. | r6/r7-fresh2 FR4-1, r6/r7-fresh5 FR5-3 |
| 01a21463-3262 | The pinned list now gives the repointed tests' later owner. The joint criterion's description no longer claims "no-precedence mode" for good; its command is unchanged. | same |
| 01a21447-d166 | Added the set-id lifetime statement, a TDD step, and the `TestNarrowedSetIdsStableAcrossFiles` criterion. | r6-fresh6 FR6-6 |
| Milestone Records | Added the "Chain-review round 7 repair" entry. No mapping row changed. | both |

Every other line of each edited body is unchanged. All 3 stored bodies and the milestone body were re-read and match their source files. Labels and the resolution command are unchanged.

## Verification

- **New criteria are red today.** All 3 new or changed criterion commands were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and all 3 exit 1. 3262's command did not change, but it was re-run anyway.
- **The five named tests exist and pass today.** With `T1_CORPUS` set, each of them prints `--- PASS`. None logs `verdict only`, which is why d218's criterion is red.
- **Lint passes.** `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.
