---
title: "`@119797` lexes as `@1` and a number: a digit name takes one digit"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:58:30.305728928Z
updated: 2026-09-29T14:58:31.831352717Z
sessions:
- start_sha: 180c36303e06048aeb3fb31e9a8d771c7dc70293
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T14:58:31.831352717Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:58:31.831352717Z
---

A variable named by digits takes only the first digit: `@119797` lexes as `@1` followed by the number 19797, and `$99` as `$9` and 9. scan_ident's parse_ident runs with STOP_AT_FIRST_NON_DIGIT, so the name is the whole run of digits.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my @x = @119797; my $y = $119797[0]; my %h = %12;'
    my(@x) = @119797;
    my $y = $119797[0];
    my(%h) = %12;

perl.git t/op/sub_lval.t:1035 and after; t/io/open.t:376 `open $99, "foo"`.

## Acceptance criteria

- [ ] A digit name is the whole run of digits (`go test ./internal/lexer/ -run '^TestDigitVariableName$' -count=1 -v | grep -q '^--- PASS: TestDigitVariableName'`)
