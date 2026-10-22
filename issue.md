---
title: Decide a named hash's rv2hv from its declaration
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a12059-30a8-722e-bbc3-52066b8610cb
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.530218409Z
updated: 2026-10-22T23:58:50.106997762Z
---

## Context

- paths: `internal/parse/sites.go` (the `Index` case, ~236+), `internal/parse/hashhedge_test.go`, `internal/parse/sites_decided_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/baseline.txt`
- commands: `go test ./internal/parse/ -run '^(TestNamedHashRv2hvFollowsDeclaration|TestEveryRv2hvHedgeNamesItsReason)$' -count=1 -v`
- entrypoints: `internal/parse/sites.go` (`Index` case, `hedge`)

Named hashes and hash slices still hedge rv2hv. Done issue 01a0bae1 measured perl's rule against the hash's declaration: `my %h` keys give 0 rv2hv (padhv); `our` gives 1; `local` 2; `delete local` follows the hash's scope (`our %h` 1, `my %h` 0); and `state` is padhv, 0 rv2hv, when the feature is on (re-measured 2026-10-22 -- 01a0bae1's "state 1" was measured without `use feature "state"`, where `state %h` is not a declaration: perl reads `state % h` as a modulus and the later `%h` is a package hash); and a slice counts follow the scopes of the hashes sliced. This issue decides rv2hv from the declaration when the declaration is in the file. It hedges only when the declaration is not visible (a package variable from elsewhere, a string eval, a glob assignment). A test names the reason for every hedge that remains, so a hedge cannot survive without one. It edits sites.go after the arrow-read issue and re-baselines sites.txt below 389 rv2hv wider, from the lower count that issue recorded.

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

### Negative Scenarios

- [ ] `state %h` is lexical and must not commit rv2hv. Re-measured on perl 5.42.0, `use feature "state"; state %h; my @k = keys %h` compiles to `padhv[%h:2,4] lRM` + `keys`, with 0 rv2hv, at file scope and inside a sub. That contradicts the 01a0bae1 table's "state 1"; the test holds perl's measurement (`go test ./internal/parse/ -run '^TestNamedHashStateIsPadhv$' -count=1 -v | grep -q '^--- PASS: TestNamedHashStateIsPadhv'`)
- [ ] `delete local` follows the hash's scope, not a fixed count. perl compiles `our %h; delete local $h{a}` to `gv[*h]` + `rv2hv sKR` + `delete vK/LVINTRO`, but `my %h; delete local $h{a}` to `padhv sR` with 0 rv2hv (`go test ./internal/parse/ -run '^TestNamedHashDeleteLocalFollowsScope$' -count=1 -v | grep -q '^--- PASS: TestNamedHashDeleteLocalFollowsScope'`)
- [ ] A declared package hash in a nested subscript chain folds. perl compiles `our %h; my $v = $h{a}{b}` to `multideref($h{"a"}{"b"})` with 0 rv2hv, so `our` alone does not commit an rv2hv (`go test ./internal/parse/ -run '^TestNamedHashOurChainFoldsToMultideref$' -count=1 -v | grep -q '^--- PASS: TestNamedHashOurChainFoldsToMultideref'`)
- [ ] Shadowing resolves to the innermost declaration. In `our %h; { my %h; my @k = keys %h }`, perl compiles the inner `keys %h` to `padhv`, and a decision keyed on the first declaration seen commits WRONG (`go test ./internal/parse/ -run '^TestNamedHashInnerMyShadowsOur$' -count=1 -v | grep -q '^--- PASS: TestNamedHashInnerMyShadowsOur'`)
- [ ] A use before the declaration is the package hash. perl compiles `my @k = keys %h; my %h; @k = keys %h` to `gv[*h]` + `rv2hv` for the first `keys` and `padhv` for the second, and a later `my` never decides an earlier use (`go test ./internal/parse/ -run '^TestNamedHashUseBeforeMyIsPackageHash$' -count=1 -v | grep -q '^--- PASS: TestNamedHashUseBeforeMyIsPackageHash'`)
- [ ] A `my %h` in a block that has closed does not decide a later `%h` outside it (`go test ./internal/parse/ -run '^TestNamedHashClosedBlockMyDoesNotLeak$' -count=1 -v | grep -q '^--- PASS: TestNamedHashClosedBlockMyDoesNotLeak'`)
- [ ] A slice mixing a `my` hash and an `our` hash in one statement counts each by its own scope (`go test ./internal/parse/ -run '^TestNamedHashMixedScopeSliceCountsEach$' -count=1 -v | grep -q '^--- PASS: TestNamedHashMixedScopeSliceCountsEach'`)
- [ ] The declaration table in the test runs perl rather than copying the 01a0bae1 numbers, so a stale row fails (`go test ./internal/parse/ -run '^TestNamedHashRv2hvTableMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestNamedHashRv2hvTableMatchesPerl'`)
- [ ] A hedge with no reason, or with a reason outside the closed set, fails; and a reason in the set that no fixture produces fails as dead (`go test ./internal/parse/ -run '^TestRv2hvHedgeReasonsAreClosedAndLive$' -count=1 -v | grep -q '^--- PASS: TestRv2hvHedgeReasonsAreClosedAndLive'`)
