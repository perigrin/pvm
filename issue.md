---
title: Remove conformance test code the adjacency refactor left dead, and its false refusal claims
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T23:49:20.285028918Z
updated: 2026-10-24T00:09:08.649774522Z
---

Found in the crochet:review of m3-conformance-corpus (round 1, 2026-10-23), by the ponytail lens and the fix-pass worker. Refactor e8ee9001 left this residue, and the milestone's diff did not create it.

## Context

- paths: `internal/conformance/*_test.go`, `conformance/04_operators/README.md`

1. **68 unused symbols** (staticcheck U1000) remain in internal/conformance. They are the `.t`-era per-tier adjacency and spelling machinery, for example tier07_callforms's `adjacencyExempt`, tier09's `tierRegexDependsOn`, tier01's `regenerateNumbering`, and most regexes and form helpers in tiers 03, 05, 06 and 08 to 14. Their comments describe checks that no longer run. The full list is in the round-1 fix report (scratchpad ccx/rev1, 774a3860).
2. **False claims about refusals.** The corpus is 270 clean with 0 refusing, so these are false:
   - tier04_test.go:906, "The fifth, `00_adjacency.t`, refuses"
   - tier11_test.go:468, "EXACTLY ONE REFUSING FILE"
   - tier12_test.go:209
   - 04_operators/README.md:285, "The fifth refusal"

   Renaming the file they mention would keep the false claims alive. They need rewriting, or deleting along with their dead code.

3. **Stale names in live comments**, found in review round 2: tier04_test.go:716, tier11_test.go:468 and tier12_test.go:209 still name `00_adjacency.t`. tier01_test.go:199 and :222 do as well, inside or next to the dead regenerateNumbering.

4. **More findings from review round 2:**
   - `regenerateNumbering` and `reNumbered` (tier01_test.go:196) are dead code, and they name `00_adjacency.t`. Delete them.
   - `adjacencyCaseOps` (adjacency_test.go) is a 6-line wrapper over fileOps and countOp with one caller. Inline it.
   - The READMEs still name about 125 `NN_name.t` files from before the topic refactor. One example is 01_literals/README.md:32-34 (`07_negative.t`). Rename each to the topic it now lives in.

## Acceptance criteria

### Positive Scenarios

- `staticcheck -checks U1000 -tests ./internal/conformance/ 2>&1 | grep -c 'is unused' | grep -qx 0` (no unused symbols remain)
- `! grep -rnE 'ONE REFUSING FILE|The fifth refusal|00_adjacency\.t., refuses' internal/conformance conformance/04_operators/README.md` (the false refusal claims are gone)
- `! grep -rn '00_adjacency\.t' internal/conformance` (no live code or comment names the deleted file)
- `! grep -rnE '\b[0-9][0-9]_[a-z_]+\.t\b' conformance/*/README.md` (no README names a .t file)
