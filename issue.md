---
title: "A file a BEGIN block requires declares its subs: BEGIN { require './test.pl' } makes is, like, plan, skip and refcount_is known"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c10b-7148-77e5-8923-16fbce93c64b
blocks:
- 01a0c13f-e0e2-7b4f-a9d2-06f7edad1e7f
created: 2026-10-09T03:25:52.865747111Z
updated: 2026-10-09T03:26:29.862291719Z
---

## Context

- paths: `internal/parse/module.go` (`ParseFile`, `ParseFileFrom`, how `use` reads a module), `internal/parse/use.go`, `internal/parse/t2_test.go` (parses with `parse.Parse(src)`, no path)
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

Nine of T2's ten remaining Unknowns (01a0c10b): perl knows test.pl's subs at compile time because `require './test.pl'` runs inside BEGIN, so `plan tests => 4`, `like $@, qr//`, `is eval {...}, undef`, `skip "...", 1` and `refcount_is $obj, 1` parse as list-operator calls -- `is` and `like` with test.pl's prototype `($$@)`. pvm reads modules a `use` loads, but not a file a BEGIN-time `require` loads. Measure perl first: a BEGIN `require` of a relative path resolves against @INC and the current directory (perl's tests run from t/); `require` at run time (outside BEGIN) declares nothing at parse time. T2's test parses source with no path, so it must parse each file with its path and t/ as the directory a relative require resolves from, as perl's harness runs them.

## Steps

- [ ] Measure on 5.42: BEGIN-time vs run-time `require` of a file declaring a prototyped sub; how a relative path resolves
- [ ] Write failing test `TestBeginRequireDeclaresSubs`: a BEGIN-time `require` of a file makes its subs (and prototypes) known to the rest of the parse; a run-time `require` does not (negative)
- [ ] Implement; parse T2's files with their path and t/ as the require directory; regenerate the shortfall counts; ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] a BEGIN-time require declares its subs (`go test ./internal/parse/ -run '^TestBeginRequireDeclaresSubs$' -count=1 -v | grep -q '^--- PASS: TestBeginRequireDeclaresSubs'`)
- [ ] the T2 ratchet holds with those nine nodes gone (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
