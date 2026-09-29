---
title: "`$ $name` is not a deref: a spaced second sigil is not seen"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:43:12.746102284Z
updated: 2026-09-29T12:43:14.236259989Z
sessions:
- start_sha: afacf7d206a75c7deb219d34136f542ea5929db1
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T12:43:14.236259989Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T12:43:14.236259989Z
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
