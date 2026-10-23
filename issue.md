---
title: Decide a named hash's rv2hv from its declaration
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a1664d-7bc8-7498-af02-7eaf187b43ed
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.530218409Z
updated: 2026-10-23T01:11:07.794088139Z
---

## Context

- paths: `internal/parse/sites.go` (the `Index` case, ~236+, and the stale "state 1" comment at ~669-673), `internal/parse/hashscope.go` (the resolver, read), `internal/parse/hashhedge_test.go` (and its stale comment at ~35-37), `internal/parse/testdata/m3probes/named-hash-decl.txt` (new), `internal/parse/sites_decided_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/baseline.txt`
- commands: `go test ./internal/parse/ -run '^(TestNamedHashRv2hvFollowsDeclaration|TestEveryRv2hvHedgeNamesItsReason)$' -count=1 -v`
- entrypoints: `internal/parse/sites.go` (`Index` case, `hedge`)

Named hashes and hash slices still hedge rv2hv. This issue decides rv2hv from the declaration in effect, using the hash-scope resolver from the previous issue, and hedges only when no declaration is visible (a package variable from elsewhere, a string eval, a glob assignment). perl's rule, measured on 5.42.0 and re-measured here at the 5.44.0 pin, is that a `my %h` use is padhv with 0 rv2hv. For `our %h; keys %h` the count is 2, counting the declaration's own `rv2hv vK/OURINTR` and the `keys`; for `local %h; keys %h` it is also 2, counting `rv2hv vKM/LVINTRO`. Done issue 01a0bae1's "our 1; local 2" counted the declaring statement inconsistently. `delete local` follows the hash's scope (`our %h` 1, `my %h` 0), and slice counts follow the scopes of the hashes sliced. **perigrin's ruling (2026-10-23): goal (7)'s wording on `state` is overridden by perl.** With the feature on, `state %h; keys %h` is padhv with 0 rv2hv. Under `no feature "state"`, `state %h` is not a declaration (perl reads `state % h`), and the later `%h` is a package hash with 1 rv2hv. `use v5.10` and later bundles turn the feature on. A test names the reason for every hedge that remains. The rv2hv wider count must fall below the value the arrow-read issue recorded (`rv2hv_wider_after_arrow` in m3start.txt).

## Steps

- [ ] Re-measure the 01a0bae1 table with `perl -MO=Concise` on the pinned perl 5.44.0 (my/our/local/delete local, state with the feature on, off, and via `use v5.10` and `use v5.36`, keys/values/each/exists/delete, and slices across scopes), and quote it in the test's comment.
- [ ] RED: write TestNamedHashRv2hvFollowsDeclaration in hashhedge_test.go, one case per row. Run it and watch it fail.
- [ ] RED: write TestNamedHashStateFeatureOffIsPackageHash (`no feature "state"; state %h; keys %h` gives 1 rv2hv) and TestNamedHashStateViaFeatureBundle (`use v5.10` and `use v5.36` each make `state %h` padhv, 0 rv2hv). Run them and watch them fail.
- [ ] GREEN: decide rv2hv in the `Index` case from the resolver's answer. Keep TestDecidableHashStillCommits green, and update TestUndecidableHashHedges to cover only undeclared hashes. Run the tests and watch them pass.
- [ ] RED/GREEN: write TestEveryRv2hvHedgeNamesItsReason. Every remaining rv2hv hedge in sites.go carries a reason string from a closed set, and a fixture for each reason produces it. Implement the reason field.
- [ ] Correct the stale "state 1" comments at sites.go:669-673 and hashhedge_test.go:35-37 ("behaves like a package hash here (1 rv2hv)"). They hold only with the feature off.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/share/pvm/versions/5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] RED/GREEN TestRv2hvWiderFellBelowArrowFix in internal/parseoracle: sites.txt's `all rv2hv wider` must be below m3start.txt's `rv2hv_wider_after_arrow`.
- [ ] Add every table row, with perl's facts, as probe file `internal/parse/testdata/m3probes/named-hash-decl.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (e.g. `our %h` shadowed by `my %h` in an inner block, `local %h` under a sub, `delete local @h{...}`, a slice mixing `my` and `our` hashes), and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A named hash's rv2hv count follows its declaration as perl measures it (`go test ./internal/parse/ -run '^TestNamedHashRv2hvFollowsDeclaration$' -count=1 -v | grep -q '^--- PASS: TestNamedHashRv2hvFollowsDeclaration'`)
- [ ] Every remaining rv2hv hedge names its reason (`go test ./internal/parse/ -run '^TestEveryRv2hvHedgeNamesItsReason$' -count=1 -v | grep -q '^--- PASS: TestEveryRv2hvHedgeNamesItsReason'`)
- [ ] A `state %h` with the feature off is a package hash, with 1 rv2hv (`go test ./internal/parse/ -run '^TestNamedHashStateFeatureOffIsPackageHash$' -count=1 -v | grep -q '^--- PASS: TestNamedHashStateFeatureOffIsPackageHash'`)
- [ ] `use v5.10` and `use v5.36` turn on state, making `state %h` padhv (`go test ./internal/parse/ -run '^TestNamedHashStateViaFeatureBundle$' -count=1 -v | grep -q '^--- PASS: TestNamedHashStateViaFeatureBundle'`)
- [ ] The committed rv2hv wider count is below the count recorded after the arrow-read fix (`go test ./internal/parseoracle/ -run '^TestRv2hvWiderFellBelowArrowFix$' -count=1 -v | grep -q '^--- PASS: TestRv2hvWiderFellBelowArrowFix'`)
- [ ] The stale claim that `state %h` behaves like a package hash is gone from both files (`sh -c 'test -f internal/parse/sites.go && test -f internal/parse/hashhedge_test.go && ! grep -q "like a package hash here" internal/parse/sites.go && ! grep -q "is the trap: a LEXICAL declaration" internal/parse/hashhedge_test.go'`)
- [ ] The table is committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/named-hash-decl.txt`)

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
