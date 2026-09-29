---
title: "`$ok = 1 when undef;` refuses: when is not a statement modifier"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:00:59.6107833Z
updated: 2026-09-29T14:01:00.717262011Z
sessions:
- start_sha: 7d2278da40ea70a59ac626769fb90220fd166c0e
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T14:01:00.717262011Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:01:00.717262011Z
---

`when` as a statement modifier refuses: `$ok = 1 when undef;` inside a given block. perly.y admits `expr KW_WHEN condition` under the switch feature, and the parser's modifier table has if, unless, while, until, for and foreach only -- though it reads the `when (...) {...}` statement form unconditionally.

Measured with perl 5.42:

    $ perl -e 'use feature "switch"; no warnings; my $ok;
          given (undef) { $ok = 1 when undef; } print "ok=$ok\n"'
    ok=1

perl.git t/op/switch.t:810 and after, 18 refusals.

## Acceptance criteria

- [ ] `EXPR when COND` is a conditional statement (`go test ./internal/parse/ -run '^TestWhenModifier$' -count=1 -v | grep -q '^--- PASS: TestWhenModifier'`)
