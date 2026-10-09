---
title: "A trailing infix operator perl refuses at end of file is refused, not an Unknown: comp/final_line_num.t"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c10b-7148-77e5-8923-16fbce93c64b
created: 2026-10-09T03:25:52.940372472Z
updated: 2026-10-09T03:26:28.261559631Z
---

## Context

- paths: `internal/parse/expr.go` (infix parsing), `internal/parse/refusal.go` (refusal codes), `internal/parse/t2_test.go`
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

The tenth T2 Unknown (01a0c10b): comp/final_line_num.t ends `BEGIN { $last_line_num = __LINE__; } print 1+` at end of file. perl refuses it -- "syntax error at ... line 13, at EOF" -- and the file's BEGIN `$SIG{__DIE__}` handler checks that line number. internal/parse leaves a lone `+` Unknown instead of refusing. Under perigrin's ruling (2026-10-09) the file is clean only when internal/parse refuses it the same way: a refusal (with its code, per RFC 0001 "Refusing what perl refuses") at end of file on line 13, where perl's is.

## Steps

- [ ] Measure perl on a trailing infix operator at EOF and before `;`, `}` and `)`
- [ ] Write failing test `TestTrailingInfixAtEOFRefused`: `print 1+` at end of file is refused at its line, as perl refuses it; `print 1+2` is not (negative); `TestRefusalCodeInventory` stays in step
- [ ] Teach T2's test that a file perl refuses is clean when internal/parse refuses it at the same line; `make test`; commit

## Acceptance criteria

- [ ] a trailing infix operator at end of file is refused as perl refuses it (`go test ./internal/parse/ -run '^TestTrailingInfixAtEOFRefused$' -count=1 -v | grep -q '^--- PASS: TestTrailingInfixAtEOFRefused'`)
- [ ] the refusal inventory holds (`go test ./internal/parse/ -run '^TestRefusalCodeInventory$' -count=1 -v | grep -q '^--- PASS: TestRefusalCodeInventory'`)
