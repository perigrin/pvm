---
title: print of readline(*{...}) is typed a union including Void
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:17.585664576Z
updated: 2026-10-14T22:36:30.406732125Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T22:36:30.406732125Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:30.406732125Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-PRINT-READLINE-GLOB` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 1 moved line (B/Deparse.pm:946): `print readline(*{...})` -- legacy said `got List`, the lowered path says a union including Void; perl is silent. A readline result (or the deref of a glob block) must not carry Void into an argument.

#### r2_print_readline_glob

```perl
my $laststash = "main";
print readline(*{$laststash."::DATA"});
__DATA__
line
```

```perl
my $laststash = "main";
print readline(*{$laststash."::DATA"});
__DATA__
line
```
perl -w (rc=0): line
old: 2:7: error: call to print: argument 1 expects Str, got List [type-mismatch];
new: 2:7: error: call to print: argument 1 expects Str, got Undef|Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|NaN|Inf|Void [type-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestPrintOfReadlineIsNotVoid$' -count=1 -v | grep -q '^--- PASS: TestPrintOfReadlineIsNotVoid'`)
