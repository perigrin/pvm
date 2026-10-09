---
title: A format at end of file with no trailing newline grows a space every canon
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T06:03:13.557942667Z
updated: 2026-10-09T06:03:13.557942667Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); `make fuzz-parser` (FuzzParser, with the canon-fixpoint invariant from 01a0de03); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; crashers from the finding run are in the session scratchpad `fuzz-crashers/` (not committed)
- NEVER use pkill or killall

Found by FuzzParser's canon-fixpoint invariant (01a0de03, 2026-10-09): for input that parses with no Unknown, canon(canon(x)) must equal canon(x). `make fuzz-parser` fails on this within about a second until it is fixed, which hides the fuzzer's deeper invariants.

`format=\n0\n.` with no newline after the `.`: canon writes `' '` after the format and the FormatBody token absorbs it into `. `, so each emission grows by one space and canon never settles. A canon or lexer defect; measure what perl accepts as a format terminator at end of file.

## Steps

- [ ] Measure perl; write failing test `TestFormatAtEOFCanonSettles` (commit the minimised crasher as a regression seed under testdata/fuzz if it is the clearest form); fix; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a format at end of file is a canon fixpoint (`go test ./internal/parse/ -run '^TestFormatAtEOFCanonSettles$' -count=1 -v | grep -q '^--- PASS: TestFormatAtEOFCanonSettles'`)
