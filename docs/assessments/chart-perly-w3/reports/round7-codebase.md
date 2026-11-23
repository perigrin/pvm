# crochet:assess round 7: codebase lens

**Subject.** `docs/plans/2026-11-23-chart-stateless-precedence.md`, blob **ce616721c9ed717dc119cf4c5f3329daf802f4cd** (`git hash-object`). The worktree HEAD is 1ff6e3f6a, the minute on top of 9afff6c63. This is revision 7.

**Order of reading.** The subject first. Then, in `cmd/ratchet`: `lex.go` (`corpus`), `population.go`, `verdicts.go` (`labelVerdicts`, `realInputs`, `realDirs`) and `main.go`. Then `testdata/labels/`, and every caller of `.Constrain(`. The plans diff last. Read-only.

## 1. Disposition of C34 to C37

I raised these findings and I am also releasing them, so I may be biased toward finding the bar met. Each release quotes the text it rests on.

| Id | Disposition |
|---|---|
| C34 | **Released**, against §3.4's split table and its notes: "The perl.git target is not run by `make test`. CI's corpus-backed Ratchets job runs it with `PERL5_CORPUS` set, and fails unless it passes", and "A developer without perl.git runs `make test`." |
| C35 | **Released**, against §3.4: "Equal kept sets give equal readings." The readings count is no longer compared at the boundary, so no switch inside `ParseWith` is needed. |
| C36 | **Released**, against §3.4: "asserts agreement on every sentence, and that at least one is refused. It does not assert the 185 and 989 of today's grammar". Runtime is handled by "If it proves slow, the bulk moves behind the build tag". |
| C37 | **Released**, against §2: "Building the tables costs about 90 ms, once per process." |
| C33 | Unchanged: issue-level, carried to refinement. |

## 2. The split: can the in-tree part of the population be built without perl.git?

**Yes.**

- `corpus()` (lex.go l.40-70) reads only in-tree files:
  - Guacamole `parses/*.pl` and `parsent/*.pl`;
  - `std5k.pl`;
  - `edit5k.pl`;
  - T1, from `T1_CORPUS` or the in-tree copy.
- perl.git is added separately by `perlGit(perl5Root(), …)` (main.go l.68). That is the only part that needs the pinned checkout.
- `supersetPopulation` (population.go l.43) needs, per file:
  - a perl label, from the committed `testdata/labels/in-tree.txt` and `t1.txt`;
  - an RD verdict;
  - the lexer-list exclusion, also committed (`lexerListed`).
- `realDirs` gives T1 files `t1Dir()` and other in-tree files their own directory (verdicts.go l.227-236). So RD needs no perl5 tree for them.
- `labelVerdicts` also reads the committed `perlgit.txt`. That is harmless without the corpus.
- The population is deterministic, as §3.4 says.

### C38: Low, editorial. The table's in-tree row omits two sets

The row reads "T1, Guacamole, std5k". `corpus()` also yields:

- `internal/psc/testdata/latency/edit5k.pl`, labelled `accept` in `in-tree.txt` l.908;
- Guacamole's `parsent/` files, 124 of them labelled.

If the implementer builds the in-tree part from the row rather than from `corpus()`, it silently drops those files from the differential. Fix: write "the superset population over `corpus()`'s in-tree files (T1, Guacamole parses and parsent, std5k, edit5k)". Refinement can make this edit; it does not need another round.

The in-tree part needs `cmd/ratchet`'s unexported `corpus`, `realInputs` and `supersetPopulation`. So the `make test` part lives in `cmd/ratchet` as an ordinary test, and the perl.git part sits beside it behind the build tag. `go test ./...` already covers `cmd/ratchet`.

## 3. The caller list for the renamed method

**Complete and correct.** Every non-offline caller of `.Constrain(`:

- `forest/forest.go:138`, `ParseWith`;
- `chart/precedence.go:327`, the referee's `judge`;
- `chart/precedence_test.go:95` and `:242`;
- `forest/precedence_test.go:88`, the toys' `constrained`.

That is exactly "ParseWith, the referee, and the forest and chart precedence tests". The README's code quote, `f.Constrain(forest.NewPrecedence(…))` (README l.2445), keeps its meaning under the rename. Its prose about edge contexts and bison's states describes the reference, not production, and §3's "live documents change first" covers it.

## 4. The new-API list in §4

**Complete for what the stage keys need.** All four of these are unexported today:

- **Copying a forest.** `Forest` has no copy method. Its exported fields allow a hand copy, but the unexported `steps` field and the `Readings` value make an exported method the right call.
- **`libraryOperators`.**
- **`resolve`.** It takes `(tab, f, s *lex.Stream, src, scope)`, and `cmd/ratchet` holds all of those.
- **The readings count.** `countReadings` needs `buildFactsGrammar(tab).fullstmt`, which is unexported. The exported wrapper must supply it.

§3.4's differential also uses the exported `libraryOperators`: it compares "after `libraryOperators`", and a file that operator check refuses never reaches the constraint. That is consistent with §4's commit coming first.

## 5. Fixed point

**Nothing remains held.** From this lens:

- C1 to C37 are released, moot or carried.
- C33 is issue-level, for refinement.
- C38 is an editorial note for refinement.

I raised nothing new that needs another assessment round.

## 6. What I did not examine

- The plans diff beyond §§2–4.
- §5 and §6 of the subject.
- The runtime of the in-tree differential and the generators in `make test`.
- Whether CI's `perl542.pin` names 94e50866.
- The other assessors' round-6 reports.

## 7. Recommendation: **accept**

C38 is an editorial fix that refinement can carry. C33 goes to the issues. From the codebase lens the assessment has reached its fixed point.
