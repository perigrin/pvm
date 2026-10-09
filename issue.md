---
title: T2 holds all 56 files clean, the file perl refuses refused alike
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
- 01a11eb2-2162-7f35-8d64-d28c72d24f5b
- 01a11eb2-21ac-7b07-b0f2-4e21a6b351ff
blocks:
- 01a11eb2-21e9-75a7-ae19-ce6f6a358be2
created: 2026-09-20T23:56:24.16253937Z
updated: 2026-10-09T03:26:32.497353689Z
---

## Context

- paths: `internal/parse/t2_test.go`
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

Ruling (perigrin, 2026-10-09): all 56 T2 files parse clean -- no file is excluded as another milestone's. This issue's earlier aim, a scoped denominator excluding what M1 does not own, is withdrawn by that ruling. A file perl itself refuses counts as clean only when internal/parse refuses it the same way.

## Re-measured, 2026-10-09 at 287e85b8

T2 core is **51 of 56 files clean (91.1%)** -- not the 30.4% this milestone's surveys were filed against. Ten Unknown nodes remain, in five files, from two causes:

    class/destruct.t:51,54          refcount_is $obj, 1, ...;
    class/field.t:333               ... or skip "XS::APItest not available", 1;
    class/gh23511.t:21              like $@, qr/.../, "...";
    comp/filter_exception.t:8       plan tests => 4;
    comp/filter_exception.t:28,30   like $@, qr/.../;
    comp/filter_exception.t:29 (x2) is eval { require "t/Foo.pm" }, undef;
    comp/final_line_num.t:13        +

- **Nine nodes: a file a BEGIN block requires declares subs.** All four files open `BEGIN { require './test.pl'; ... }`, so perl knows `plan`, `is ($$@)`, `like ($$@)`, `refcount_is` and `skip` (test.pl:55, 395, 510, 536, 570) when it parses their parenless calls; pvm does not read a file a BEGIN-time `require` loads.
- **One node: a file perl refuses.** comp/final_line_num.t ends `print 1+` at end of file on purpose: perl reports "syntax error at ... line 13, at EOF", which its own BEGIN `$SIG{__DIE__}` handler checks. Under perigrin's ruling (2026-10-09) such a file is clean only when internal/parse refuses it the same way.

The four comp files c7d7 was filed to measure -- comp/parser.t (65), comp/hints.t (36), comp/colon.t (25), comp/redef.t (21) -- are clean: all 147 nodes are gone.

## Steps

- [ ] After the two cause issues land: write `TestT2HoldsAll56`, asserting the shortfall map is empty, all 56 files are clean, and comp/final_line_num.t is clean because internal/parse refuses it at end of file as perl does; remove the map's entries; `make test`; commit

## Acceptance criteria

- [ ] T2 holds all 56 files (`go test ./internal/parse/ -run '^TestT2HoldsAll56$' -count=1 -v | grep -q '^--- PASS: TestT2HoldsAll56'`)
- [ ] the T2 ratchet holds with an empty shortfall (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '56 of 56 files clean'`)
