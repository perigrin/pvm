---
title: "`*|` (the glob of the punctuation variable $|) parses as bitwise or"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:15.169407428Z
updated: 2026-10-14T23:01:31.608371719Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: 3d6ed48fd3ca446333a6c4b012f22e9ca9518666
  commits: 1
  started_at: 2026-10-14T22:36:30.22735741Z
  ended_at: 2026-10-14T23:01:31.608371719Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:30.22735741Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T23:01:31.608371719Z
observed_paths:
- internal/lexer/brackets.go
- internal/lexer/expect.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/parse/globpunct_test.go
- internal/parse/parse.go
- internal/parse/term.go
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-GLOB-PUNCT` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 1 added line (English.pm:169). `*OUTPUT_AUTOFLUSH = *|;` -- perl's Deparse keeps `*|`, a glob; the lowered path reads `|` as bitwise-or (`right operand of |`). A parser defect in internal/parse: `*` followed by a punctuation variable name is a glob, as `$|` is a scalar. The m1 gates must hold.

#### r2_glob_pipe

```perl
*OUTPUT_AUTOFLUSH = *|;
*OFS = *,;
```

```perl
*OUTPUT_AUTOFLUSH = *|;
$OUTPUT_AUTOFLUSH = 1;
```
perl -w (rc=0): 
Deparse: (*OUTPUT_AUTOFLUSH = *|); ($OUTPUT_AUTOFLUSH = 1); 
old: 
new:

```perl
*OUTPUT_AUTOFLUSH = *|;
*OFS = *,;
```
perl -w (rc=0): Name "main::OFS" used only once: possible typo at r2_glob_pipe.pl line 2. Name "main::OUTPUT_AUTOFLUSH" used only once: possible typo at r2_glob_pipe.pl line 1.
Deparse: (*OUTPUT_AUTOFLUSH = *|); (*OFS = *,); 
old: 
new: 1:23: warning: right operand of "|": expected Int, got Int|Num|Inf [coercion-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/parse/ -run '^TestGlobOfPunctuationVariable$' -count=1 -v | grep -q '^--- PASS: TestGlobOfPunctuationVariable'`)
