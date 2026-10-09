---
title: "T2 core is 51 of 56 files: ten Unknowns from two causes"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-c7d7-7d9c-9a85-60a59aa26bc0
blocks:
- 01a11eb2-2162-7f35-8d64-d28c72d24f5b
- 01a11eb2-21ac-7b07-b0f2-4e21a6b351ff
created: 2026-09-20T22:59:07.720149993Z
updated: 2026-10-09T03:44:03.65799767Z
sessions:
- start_sha: 8ec19ffcb8e44115e5b60423b138536be1add2ba
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T03:44:03.65799767Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T03:44:03.65799767Z
---

## Context

- paths: `internal/parse/t2_test.go` (`TestT2CoreParses`, its `shortfall` map)
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

Goal (perigrin, 2026-10-09): all 56 T2 files parse clean.

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

- [ ] Keep this survey true: each cause above has its own issue in this milestone (the BEGIN-time require; the file perl refuses), and the shortfall map's counts are regenerated as each lands

## Acceptance criteria

- [ ] each remaining cause has an issue (`sh -c 'git zhi issue list --milestone m1-parse-the-core --format json | grep -q "BEGIN block requires" && git zhi issue list --milestone m1-parse-the-core --format json | grep -q "perl refuses at end of file"'`)
- [ ] the T2 ratchet holds (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
