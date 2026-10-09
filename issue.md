---
title: A bareword class name before -> parses and canons as a sub call
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-01T05:24:19.533594975Z
updated: 2026-10-09T09:18:14.061836201Z
sessions:
- start_sha: 4f8842ca84c6c91bf81d9192c8c4eac456a3f55b
  end_sha: 13c3009925f756eda5650f5e3f134fc0bbf9ef5a
  commits: 3
  started_at: 2026-10-09T08:08:45.597426889Z
  ended_at: 2026-10-09T09:18:14.061836201Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T08:08:45.597426889Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T09:18:14.061836201Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/barewordinvocant_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/emptybraces_test.go
- internal/parse/lastindexblock_test.go
- internal/parse/parse.go
- internal/parse/postfixhashslice_test.go
- internal/parse/sites.go
- internal/parse/smartmatchaggregate_test.go
- internal/parse/stmt_test.go
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
- internal/parse/wrongtree_test.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
---

A bareword class name before `->` is parsed as a Call, and canon renders it as a call: `Future::AsyncAwait->import` canons as `Future::AsyncAwait() -> import` (PerlOnJava unit/custom_warning_async_suspend_state.t:9). perl 5.42.0 reads it as a string invocant:

    $ perl -MO=Deparse -e 'Foo::Bar->new; Foo->new(1);'
    'Foo::Bar'->new;
    'Foo'->new(1);

The tree says "call the sub Foo::Bar and invoke new on its result", which is a different program, unless a sub by that name is declared. Fidelity doesn't catch it. 95 of perl.git t/'s canon files hold the `NAME() -> ` pattern. Found while landing declared keywords; not fixed there.

## Acceptance criteria

- [ ] A bareword class invocant is not a call (`sh -c 'go test ./internal/parse/ -run "^TestBarewordInvocant" -count=1 -v | grep -q "^--- PASS"'`)
