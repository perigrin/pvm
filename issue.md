---
title: "`$ $name` is not a deref: a spaced second sigil is not seen"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:43:12.746102284Z
updated: 2026-09-29T12:55:21.068289909Z
sessions:
- start_sha: afacf7d206a75c7deb219d34136f542ea5929db1
  end_sha: db9808e688f56c2993d8a528c48cd171180e3067
  commits: 1
  started_at: 2026-09-29T12:43:14.236259989Z
  ended_at: 2026-09-29T12:55:21.068289909Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T12:43:14.236259989Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T12:55:21.068289909Z
observed_paths:
- internal/lexer/scan.go
- internal/lexer/spacedderef_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A deref sigil separated by spaces from the sigil it applies to -- `$ $name1`, `@ $r` -- does not lex as a deref: startsDerefExpression skips spaces before a brace but checks the byte right after the sigil for a second `$`.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $x = $ $name1; my @y = @ $r; my $z = $$ ;'
    my $x = $$name1;
    my(@y) = @$r;
    my $z = $$;

perl.git t/op/postfixderef.t:224, `is ($ $name1, undef, ...)`.

## Acceptance criteria

- [ ] A spaced second sigil makes a deref, and `$$ ;` stays the pid (`go test ./internal/lexer/ -run '^TestSpacedDeref$' -count=1 -v | grep -q '^--- PASS: TestSpacedDeref'`)
