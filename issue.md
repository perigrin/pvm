---
title: An opener unclosed at end of file parses with no Unknown, though perl refuses it
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T06:03:13.464345631Z
updated: 2026-10-09T07:49:22.76891418Z
sessions:
- start_sha: 0cbc6dfb48de6e5fb1f7641ecefb081702ab84f9
  end_sha: 0f8697ba4cff25a3304a02f6478bb57511bfe8ef
  commits: 4
  started_at: 2026-10-09T06:33:25.015715771Z
  ended_at: 2026-10-09T07:49:22.76891418Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T06:33:25.015715771Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T07:49:22.76891418Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/formateofcanon_test.go
- internal/parse/parse.go
- internal/parse/protoslot_test.go
- internal/parse/refscalar.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/sites.go
- internal/parse/term.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/unaryplus_test.go
- internal/parse/unclosed_test.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); `make fuzz-parser` (FuzzParser, with the canon-fixpoint invariant from 01a0de03); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; crashers from the finding run are in the session scratchpad `fuzz-crashers/` (not committed)
- NEVER use pkill or killall

Found by FuzzParser's canon-fixpoint invariant (01a0de03, 2026-10-09): for input that parses with no Unknown, canon(canon(x)) must equal canon(x). `make fuzz-parser` fails on this within about a second until it is fixed, which hides the fuzzer's deeper invariants.

`0(` parses to an Index with one child and canonicalises `0(;`, then `0(;);`; `for(0A0` canonicalises `for (0) A0() `, which re-reads as two statements. perl refuses both, and also refuses `{`, `sub f {` and `my $x = (1` -- which parse clean here (they happen to be fixpoints). An opener unclosed at end of file must be refused (or Unknown) as perl refuses it, with the refusal at the line perl reports; measure perl on each.

## Steps

- [ ] Measure perl; write failing test `TestUnclosedOpenerAtEOFRefused` (commit the minimised crasher as a regression seed under testdata/fuzz if it is the clearest form); fix; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] an unclosed opener at end of file is refused (`go test ./internal/parse/ -run '^TestUnclosedOpenerAtEOFRefused$' -count=1 -v | grep -q '^--- PASS: TestUnclosedOpenerAtEOFRefused'`)
