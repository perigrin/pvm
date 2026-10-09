---
title: Canon writes unary + as the word pos, a different program
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T06:03:13.36178496Z
updated: 2026-10-09T06:03:13.36178496Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); `make fuzz-parser` (FuzzParser, with the canon-fixpoint invariant from 01a0de03); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; crashers from the finding run are in the session scratchpad `fuzz-crashers/` (not committed)
- NEVER use pkill or killall

Found by FuzzParser's canon-fixpoint invariant (01a0de03, 2026-10-09): for input that parses with no Unknown, canon(canon(x)) must equal canon(x). `make fuzz-parser` fails on this within about a second until it is fixed, which hides the fuzzer's deeper invariants.

`prefixName` (`internal/parse/term.go`, ~line 557) names unary `+` "pos"; `unaryToken` (`internal/parse/canon.go`) has no inverse, so `+$x` canonicalises to `pos $x` -- perl's `pos` builtin, a different program (a fidelity defect as well as a fixpoint one). The fix is likely one `case "pos": return "+"`, but check every name `prefixName` produces has its inverse, and whether the canon ratchet's counts move.

## Steps

- [ ] Measure perl; write failing test `TestCanonUnaryPlusIsPlus` (commit the minimised crasher as a regression seed under testdata/fuzz if it is the clearest form); fix; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] canon writes unary + as + (`go test ./internal/parse/ -run '^TestCanonUnaryPlusIsPlus$' -count=1 -v | grep -q '^--- PASS: TestCanonUnaryPlusIsPlus'`)
