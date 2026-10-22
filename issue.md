---
title: Decide a named hash's rv2hv from its declaration
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-22T23:46:34.530218409Z
updated: 2026-10-22T23:46:34.530218409Z
---

## Context

- paths: `internal/parse/sites.go` (the `Index` case, ~236+), `internal/parse/hashhedge_test.go`, `internal/parse/sites_decided_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/baseline.txt`
- commands: `go test ./internal/parse/ -run '^(TestNamedHashRv2hvFollowsDeclaration|TestEveryRv2hvHedgeNamesItsReason)$' -count=1 -v`
- entrypoints: `internal/parse/sites.go` (`Index` case, `hedge`)

Named hashes and hash slices still hedge rv2hv. Done issue 01a0bae1 measured perl's rule against the hash's declaration: `my %h` keys give 0 rv2hv (padhv); `our` gives 1; `state` 1; `local` 2; `delete local` 1; and a slice counts follow the scopes of the hashes sliced. This issue decides rv2hv from the declaration when the declaration is in the file. It hedges only when the declaration is not visible (a package variable from elsewhere, a string eval, a glob assignment). A test names the reason for every hedge that remains, so a hedge cannot survive without one. It edits sites.go after the arrow-read issue and re-baselines sites.txt below 389 rv2hv wider, from the lower count that issue recorded.

## Steps

- [ ] Re-measure the 01a0bae1 table with `perl -MO=Concise` on perl 5.42 (my/our/state/local/delete local, keys/values/each/exists/delete, and slices across scopes), and quote it in the test's comment.
- [ ] RED: write TestNamedHashRv2hvFollowsDeclaration in hashhedge_test.go, one case per row. Run it and watch it fail.
- [ ] GREEN: decide rv2hv in the `Index` case from the resolved declaration. Keep TestDecidableHashStillCommits green, and update TestUndecidableHashHedges to cover only undeclared hashes. Run the tests and watch them pass.
- [ ] RED/GREEN: write TestEveryRv2hvHedgeNamesItsReason. Every remaining rv2hv hedge in sites.go carries a reason string from a closed set, and a fixture for each reason produces it. Implement the reason field.
- [ ] Re-baseline sites.txt (and baseline.txt where whole files move) with `-parseoracle.update` in the same commit, giving the new rv2hv wider count and naming each moved file. WRONG must stay 0.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the Concise table and writes NEW adversarial probes outside the corpus (e.g. `our %h` shadowed by `my %h` in an inner block, `local %h` under a sub, `delete local @h{...}`, a slice mixing `my` and `our` hashes). Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A named hash's rv2hv count follows its declaration as perl measures it (`go test ./internal/parse/ -run '^TestNamedHashRv2hvFollowsDeclaration$' -count=1 -v | grep -q '^--- PASS: TestNamedHashRv2hvFollowsDeclaration'`)
- [ ] Every remaining rv2hv hedge names its reason (`go test ./internal/parse/ -run '^TestEveryRv2hvHedgeNamesItsReason$' -count=1 -v | grep -q '^--- PASS: TestEveryRv2hvHedgeNamesItsReason'`)
- [ ] The site ratchet records the further-lowered rv2hv wider count (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestSiteRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -q '^--- PASS: TestSiteRatchetCorpus'`)
