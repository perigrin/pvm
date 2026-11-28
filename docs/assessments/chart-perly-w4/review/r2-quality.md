# crochet:review round 2: quality (prefix Q)

**Blocking held: 0. Promoted: none. New: none.**

The subject is unchanged: `git diff e3432c0f 8d8fff38`. I read all 13 residual rows: C1 and C2, Q1 to Q4, and P1 to P7.

## One correction to my round 1 report (Q4)

My round 1 report said the delivery removed the last caller of `constrainSource`. That is wrong, and P's report noted it correctly ("Noted, out of this delivery").

- **Command:** `git grep -n "constrainSource(" <rev> -- internal/chart`, run at both e3432c0f and 8d8fff38.
- **Result:** at each revision the only match is the definition itself (`precedence_test.go:84`, then `:63`). There was no caller at either end.

So Q4 is pre-existing dead code, not this delivery's debt. I withdraw it as a delivery finding. It belongs with P's "out of this delivery" list: `ownDir`, `itemText`, `labelsDir`, `lookup`, `store`.

## Rows I considered for promotion, and why none is promoted

My test for promotion: does accepting the delivery with this row left open produce a wrong forest, a wrong accept or refuse decision, or a check that passes falsely? None does.

- **Q1** (`NewJudged` does not check that the precedence's tables are its own) **and P4** (three checks of "are these perly.y's tables?", each with a different strictness):
  - Both are traps for API misuse. Only tests can reach them.
  - Production's `New` and `ParseWith` derive the precedence over the same tables. `ParseWith` panics on any recognizer whose judging does not match `New`'s.
  - One shared remedy closes both: one predicate in `prectab`, and an error from `newJudgement` when `p.Tab != tab`.
  - Kept RESIDUAL.
- **Q2** (the slipper's closure runs no reduce side):
  - Today's perly.y probe found 0 non-allow verdicts before `;`, `}` and EOF.
  - So no forest or verdict differs today. It is a latent hole that depends on a property of the grammar.
  - Kept RESIDUAL. The cheapest remedy is two more pins, for `}` and EOF, beside the `;` pin.
- **Q3** (refusal records are keyed by parent state, and their readers ignore context):
  - It is unconfirmed: 0 cross-state instances among 51 refused alternatives over T1.
  - The forest-equality differentials are green on waypoint2, the referee sentences and perl.git, where coverage reports 620 of 620 compared.
  - Kept RESIDUAL.
  - **P5** (the pairing step is written twice in `complete`) touches the same code. The Q3 fix and the P5 dedup should be done together, so that one copy of the pairing rule feeds both the adding of items and the recording of refusals.
- **P1, P2, P3, P6, P7:**
  - These are test scaffolding in production files, stale comments, and test-only exports.
  - P3's two false comments are the same kind as the four stale V1 comments already known. A reader could be misled, but no behaviour is.
  - All kept RESIDUAL.
- **C1** (the chain export due at close):
  - This is pending work that a live document says is pending. It belongs with the postmortem.
  - Not a code-soundness matter. Kept RESIDUAL.
- **C2** (an unrelated PR, #649, sits in the delivery range):
  - It concerns how the scope is stated, not a defect.
  - I did not review #649's Windows rate-limit change under this lens.
  - Kept RESIDUAL.

## New findings

None. The coverage and ponytail reports turned up nothing that changes my round 1 reading of the production code.

## Release

I hold no blocking finding, so there is nothing to release.
