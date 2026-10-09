---
title: Record what T2's ten remaining Unknowns are, beside the shortfall map
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
- 01a0c13f-97f5-7f98-b32d-07245ec6ddfe
- 01a0c13f-aaf8-7c2f-a037-754feea1cf77
blocks:
- 01a0c10b-7148-77e5-8923-16fbce93c64b
created: 2026-09-20T23:56:17.751068121Z
updated: 2026-10-09T03:27:24.566924324Z
sessions:
- start_sha: 287e85b834f36ff072bd9398794ceec01d1eec61
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T03:27:24.566924324Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T03:27:24.566924324Z
---

## Context

- paths: `internal/parse/t2_test.go` (the shortfall map is the inventory)
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

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

This issue's job is now to put that attribution beside the counts, so the next reader does not re-measure: the shortfall map's comment names each remaining node's construct and file:line.

## Steps

- [ ] Record the per-node attribution above in `t2_test.go`'s shortfall comment; change no parse behaviour; `make test`; commit

## Acceptance criteria

- [ ] the shortfall map names the BEGIN-time require as its cause (`grep -q "require './test.pl'" internal/parse/t2_test.go`)
- [ ] the shortfall map names the file perl refuses (`grep -q 'final_line_num.t ends' internal/parse/t2_test.go`)
- [ ] the T2 ratchet still holds -- no parse behaviour changed (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
