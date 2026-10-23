---
title: Report a ->{...} hash read's rv2hv exactly when perl keeps it out of multideref
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
- 01a1664d-6fdb-7a09-8211-882e814db87c
blocks:
- 01a16602-61a2-78de-a798-da4cef3208c5
created: 2026-10-09T11:07:58.504119457Z
updated: 2026-10-23T01:56:31.027315515Z
---

## Context

- paths: `internal/parse/sites.go` (`hedgeDerefElement`, ~183-205, and the call sites at ~262, ~346, ~377), `internal/parse/arrowhashread_test.go`, `internal/parse/hashhedge_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `go test ./internal/parse/ -run '^TestArrowHashReadFollowsMultideref$' -count=1 -v`; `go test ./internal/parseoracle/ -run '^TestRv2hvWiderFellFromM3Start$' -count=1 -v`
- entrypoints: `internal/parse/sites.go` (`hedgeDerefElement`)

This is the real fix that the m1 hedge for arrow hash reads defers to (filed 2026-10-09). Whether perl compiles a `->{...}` hash read as a visible `rv2hv` or folds it into `multideref` depends on the chain's op_lvalue context. push, keys, foreach, assignment and sub args keep rv2hv; a plain read folds into multideref. The chain's shape and keys matter too: a computed key like `->{shift()}`, a read inside a block dereference `@{$r->{k}}`, and `$#{ $x->{k} }`. internal/parse commits to one answer. m1 hedges the rv2hv marker for the `->{...}` reads it cannot predict, and the oracle scores them `wider`. This issue models perl's rule and reports rv2hv exactly. Seen in japh/abigail.t, op/gmagic.t, op/ref.t and porting/libperl.t. The site ratchet (TestSiteRatchetCorpus, sites.txt) records the rv2hv wider count. The starting value is `rv2hv_wider_after_magic` in m3start.txt, recorded under its own key by the op/magic.t issue once that file is answered. This issue therefore follows that one. This issue lowers the count and records the new value as `rv2hv_wider_after_arrow` in m3start.txt.

**How much this can lower (measured 2026-10-23, perl 5.44.0, scratch copy at 054e75e6).** The oracle ran over the 67 corpus files containing `->{`, twice: as is, and with `hedgeDerefElement` disabled. As is, it found 157 rv2hv wider findings (per line, before the per-site counting change). Disabled, it found 154 wider and 3 WRONG. So the arrow hedges account for 3 line-level findings, and every one of them is a site where perl DOES emit rv2hv. Most rv2hv wider comes from the named-hash `Index` hedges, which the next issue owns. The strict decrease below is reachable, but only by committing those few sites correctly. Dropping the hedge without deciding them turns wider into WRONG, which fails the build.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Concise` on `->{...}` chains in each context: plain read, push, keys, foreach, assignment (both sides), sub argument, computed key, inside `@{...}`, and inside `$#{...}`. Quote the results in the test's comment.
- [ ] RED: write TestArrowHashReadFollowsMultideref in arrowhashread_test.go, one case per measured context. Run it and watch it fail.
- [ ] GREEN: decide rv2hv from op_lvalue context in `hedgeDerefElement`, hedging only the shapes the measurement cannot settle statically. Update or retire TestArrowHashReadHedgesRv2hv to match, and keep TestArrowHashReadStaysDecided green. Run the tests and watch them pass.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Record the new count as `rv2hv_wider_after_arrow` in m3start.txt. The commit message gives it and names each file that moved.
- [ ] RED/GREEN TestRv2hvWiderFellFromM3Start in internal/parseoracle: it reads sites.txt's `total all rv2hv wider` row and fails unless that count is below m3start.txt's `rv2hv_wider_after_magic`. It also fails if either value is missing.
- [ ] Add every measured context, with perl's facts, as probe file `internal/parse/testdata/m3probes/arrow-hash-read.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (e.g. `local $r->{k}`, `delete $r->{a}{b}`, `exists`, `\$r->{k}`, lvalue sub args, slices off an arrow), and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.
- [ ] Update docs/specs/perl-parser/00-findings.md §0.11 ("A `->{...}` read whose multideref folding is not in the text hedges its rv2hv"): replace "Sites now hedges rv2hv... The exact rule is issue 01a12059-30a8" with the decided rule

## Acceptance criteria

### Positive Scenarios

- [ ] An arrow hash read reports rv2hv exactly when perl keeps it (`go test ./internal/parse/ -run '^TestArrowHashReadFollowsMultideref$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadFollowsMultideref'`)
- [ ] The committed rv2hv wider count is below the count recorded once op/magic.t was answered (`go test ./internal/parseoracle/ -run '^TestRv2hvWiderFellFromM3Start$' -count=1 -v | grep -q '^--- PASS: TestRv2hvWiderFellFromM3Start'`)
- [ ] The measured contexts are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/arrow-hash-read.txt`)
- [ ] The arrow-read files named in this issue lose rv2hv wider sites. Each of japh/abigail.t, op/gmagic.t and op/ref.t that had an rv2hv wider row when op/magic.t was answered has fewer now, and at least one had a row (`sh -c 'f=internal/parseoracle/testdata/ratchet; c=$(git log -1 --format=%H -G "^rv2hv_wider_after_magic" -- $f/m3start.txt); [ -n "$c" ] || exit 1; n=0; for p in japh/abigail.t op/gmagic.t op/ref.t; do a=$(git show $c:$f/sites.txt | awk -v p=$p "\$1==\"site\" && \$2==p && \$3==\"rv2hv\" && \$4==\"wider\" {print \$5}"); [ -n "$a" ] || continue; n=$((n+1)); b=$(awk -v p=$p "\$1==\"site\" && \$2==p && \$3==\"rv2hv\" && \$4==\"wider\" {print \$5}" $f/sites.txt); [ "${b:-0}" -lt "$a" ] || exit 1; done; [ $n -gt 0 ]'`)

### Negative Scenarios

- [ ] A hash element passed as a sub argument does not keep rv2hv. perl 5.42 compiles `foo($r->{a})` to `multideref($r->{"a"})` with 0 rv2hv; only a whole hash deref argument keeps it, and `foo(%{$r->{a}})` shows 2 rv2hv and no multideref. Committing rv2hv for the element read is WRONG (`go test ./internal/parse/ -run '^TestArrowHashReadSubArgElementFoldsToMultideref$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadSubArgElementFoldsToMultideref'`)
- [ ] The outer deref's context decides, not the chain alone. perl compiles `my %x = %{$r->{a}}` to `multideref` plus 1 rv2hv, `keys %{$r->{a}}` to 2 rv2hv, and `push @{$r->{a}}, 1` to 1 rv2hv, and the decision gives each its own count (`go test ./internal/parse/ -run '^TestArrowHashReadCountFollowsOuterDerefContext$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadCountFollowsOuterDerefContext'`)
- [ ] A non-constant key keeps rv2hv. perl compiles `$r->{"a$x"}` and `$r->{shift()}` each to 1 `rv2hv` and no multideref, so a rule that folds every `->{...}` read into multideref commits WRONG here (`go test ./internal/parse/ -run '^TestArrowHashReadComputedKeyKeepsRv2hv$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadComputedKeyKeepsRv2hv'`)
- [ ] Lvalue-looking element contexts still fold. perl compiles `local $r->{a} = 1`, `\$r->{a}`, `for ($r->{a}) {}`, `exists $r->{a}{b}` and `delete $r->{a}{b}` each to one `multideref` with 0 rv2hv; none of these commits rv2hv (`go test ./internal/parse/ -run '^TestArrowHashReadLvalueElementContextsFold$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadLvalueElementContextsFold'`)
- [ ] `$#{$r->{a}}` keeps 1 rv2hv, as perl measures, and it is not folded with the plain-read rule (`go test ./internal/parse/ -run '^TestArrowHashReadArrayLastIndexKeepsRv2hv$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadArrayLastIndexKeepsRv2hv'`)
- [ ] A hedge removed from `hedgeDerefElement` is named in a test case with the perl measurement that justified removing it. A hedge that remains names the shape it covers, so no hedge disappears or survives without a stated reason (`go test ./internal/parse/ -run '^TestArrowHashReadHedgesNameTheirShape$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadHedgesNameTheirShape'`)
- [ ] Regression guard: previously decided arrow reads stay decided and agree with perl after the change (`go test ./internal/parse/ -run '^TestArrowHashReadStaysDecided$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadStaysDecided'`)
