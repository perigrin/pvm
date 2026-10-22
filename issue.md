---
title: Report a ->{...} hash read's rv2hv exactly when perl keeps it out of multideref
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
blocks:
- 01a16602-61a2-78de-a798-da4cef3208c5
created: 2026-10-09T11:07:58.504119457Z
updated: 2026-10-22T23:47:10.262748111Z
---

## Context

- paths: `internal/parse/sites.go` (`hedgeDerefElement`, ~183-205, and the call sites at ~262, ~346, ~377), `internal/parse/arrowhashread_test.go`, `internal/parse/hashhedge_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `go test ./internal/parse/ -run '^TestArrowHashReadFollowsMultideref$' -count=1 -v`; the oracle sweep, `PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestSiteRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parse/sites.go` (`hedgeDerefElement`)

This is the real fix that the m1 hedge for arrow hash reads defers to (filed 2026-10-09). Whether perl compiles a `->{...}` hash read as a visible `rv2hv` or folds it into `multideref` depends on the chain's op_lvalue context. push, keys, foreach, assignment and sub args keep rv2hv; a plain read folds into multideref. The chain's shape and keys matter too: a computed key like `->{shift()}`, a read inside a block dereference `@{$r->{k}}`, and `$#{ $x->{k} }`. internal/parse commits to one answer. m1 hedges the rv2hv marker for the `->{...}` reads it cannot predict, and the oracle scores them `wider`. This issue models perl's rule and reports rv2hv exactly. Seen in japh/abigail.t, op/gmagic.t, op/ref.t and porting/libperl.t. The site ratchet (TestSiteRatchetCorpus, sites.txt) records rv2hv at 540 sites with 389 wider; this issue lowers the wider count and records the new number.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on `->{...}` chains in each context: plain read, push, keys, foreach, assignment (both sides), sub argument, computed key, inside `@{...}`, and inside `$#{...}`. Quote the results in the test's comment.
- [ ] RED: write TestArrowHashReadFollowsMultideref in arrowhashread_test.go, one case per measured context. Run it and watch it fail.
- [ ] GREEN: decide rv2hv from op_lvalue context in `hedgeDerefElement`, hedging only the shapes the measurement cannot settle statically. Update or retire TestArrowHashReadHedgesRv2hv to match, and keep TestArrowHashReadStaysDecided green. Run the tests and watch them pass.
- [ ] Re-baseline sites.txt (and baseline.txt where whole files move) with `-parseoracle.update` in the same commit. The commit message gives the new rv2hv wider count and names each file that moved. WRONG must stay 0.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the Concise measurements and writes NEW adversarial probes outside the corpus (e.g. `local $r->{k}`, `delete $r->{a}{b}`, `exists`, `\$r->{k}`, lvalue sub args, slices off an arrow). Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] An arrow hash read reports rv2hv exactly when perl keeps it (`go test ./internal/parse/ -run '^TestArrowHashReadFollowsMultideref$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadFollowsMultideref'`)
- [ ] The site ratchet records the lower rv2hv wider count, below 389 (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestSiteRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -q '^--- PASS: TestSiteRatchetCorpus'`)
