---
title: "`$ok = 1 when undef;` refuses: when is not a statement modifier"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:00:59.6107833Z
updated: 2026-09-29T14:12:58.809278772Z
sessions:
- start_sha: 7d2278da40ea70a59ac626769fb90220fd166c0e
  end_sha: 580b9d981afc942c4fa785b856dca63174928fdb
  commits: 1
  started_at: 2026-09-29T14:01:00.717262011Z
  ended_at: 2026-09-29T14:12:58.809278772Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:01:00.717262011Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T14:12:58.809278772Z
observed_paths:
- internal/parse/control.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/whenmod_test.go
---

`when` as a statement modifier refuses: `$ok = 1 when undef;` inside a given block. perly.y admits `expr KW_WHEN condition` under the switch feature, and the parser's modifier table has if, unless, while, until, for and foreach only -- though it reads the `when (...) {...}` statement form unconditionally.

Measured with perl 5.42:

    $ perl -e 'use feature "switch"; no warnings; my $ok;
          given (undef) { $ok = 1 when undef; } print "ok=$ok\n"'
    ok=1

perl.git t/op/switch.t:810 and after, 18 refusals.

## Acceptance criteria

- [ ] `EXPR when COND` is a conditional statement (`go test ./internal/parse/ -run '^TestWhenModifier$' -count=1 -v | grep -q '^--- PASS: TestWhenModifier'`)
