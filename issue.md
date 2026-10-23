---
title: Classify dropped source parens by position, then write a Paren node's parens once
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5e5e-73cc-a210-0341ea517914
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
created: 2026-10-22T23:46:33.790255282Z
updated: 2026-10-23T01:09:02.566216851Z
---

## Context

- paths: the parser files a tree-defect fix touches (e.g. `internal/parse/term.go`), `internal/parse/parenpos_test.go` (new), `internal/parse/testdata/m3probes/dropped-paren-context.txt` (new), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/sites.txt`
- commands: `go test ./internal/parse/ -run '^(TestDroppedParenPositionsAreClassified|TestDroppedParenAfterBackslashIsContextChanging|TestDroppedParenFixturesCoverEveryContextPosition)$' -count=1 -v`
- entrypoints: `internal/parse/term.go` (`finishList`, ~723), `internal/parse/fidelity.go` (`Faithful`)

117 of the 169 M3-dir re-emission failures are `source has "(", canon has <something else>`: a source paren that canon does not write. Examples are `(shift)->SUPER::foo` (mro/basic_05_c3.t:32), `ok(($y eq '...'), ...)` (mro/c3_with_overload.t:40) and `if (($Config{...} !~ /.../) )` (lib/commonsense.t:12).

**perigrin's ruling (2026-10-22, revised after measurement).** Measured on perl 5.42.0 with B::Concise. Where parens change meaning, perl builds different ops: `my ($x) = @_` is aassign but `my $x = @_` is padsv_store, and `($x) x 3` is a list repeat (`lK/DOLIST`) but `$x x 3` is a string repeat. Before anything is forgiven, the 117 failures are classified by position. A failure in a context-changing position (before `=`, before `x`, after `my`/`our`/`local`/`state`, after `\`) is a tree defect, where psc read a list as a scalar. It is fixed, not forgiven. This issue does the classification and the fixes. The next issue then makes canon write a `Paren` node's parens once and makes `Faithful` forgive only extra nesting.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Concise` on `my ($x) = @_` vs `my $x = @_`, `($x) x 3` vs `$x x 3`, `($x) = f()`, `() = f()`, `local ($x) = 1`, and `\(@a)` vs `\@a`. Quote the output in the tests' comments.
- [ ] Classify every one of the 117 dropped-paren failures by position: context-changing (before `=`, before `x`, after my/our/local/state, after `\`) or not. Put the list in the commit message.
- [ ] Create `internal/parse/parenpos_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestDroppedParenPositionsAreClassified. For each context-changing failure, reduce it to a fixture asserting that psc's tree matches perl's ops (a list where perl builds a list). Run it and watch any tree defect fail.
- [ ] GREEN: fix each tree defect in the parser, one at a time, running the test after each. If there are none, the test still pins the measured positions as lists.
- [ ] Add every fixture, with perl's facts, as probe file `internal/parse/testdata/m3probes/dropped-paren-context.txt`. Run TestM3ProbesHeldToPerl.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Every context-changing dropped-paren position parses as perl's ops do (`go test ./internal/parse/ -run '^TestDroppedParenPositionsAreClassified$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenPositionsAreClassified'`)
- [ ] The context-changing fixtures are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/dropped-paren-context.txt`)

### Negative Scenarios

- [ ] A paren after `\` changes context, and the classification treats it so: `\(@a)` is `refgen` over `padav lPM`, while `\@a` is `srefgen`. A dropped paren there is a tree defect, not a forgivable one (`go test ./internal/parse/ -run '^TestDroppedParenAfterBackslashIsContextChanging$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenAfterBackslashIsContextChanging'`)
- [ ] The classification fixtures cannot pass empty: they hold at least one case from each context-changing position (before `=`, before `x`, after each of my/our/local/state, after `\`), and the test fails if any class has none (`go test ./internal/parse/ -run '^TestDroppedParenFixturesCoverEveryContextPosition$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenFixturesCoverEveryContextPosition'`)
