---
title: "`undef..2` refuses: a range operator after a word is read as its argument"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:40:09.962534186Z
updated: 2026-09-29T07:40:18.24426033Z
sessions:
- start_sha: d6662dfe252501db6de1354bab1263b50feaa372
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T07:40:18.24426033Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T07:40:18.24426033Z
---

endsArgumentList names the infix operators that cannot start a term (`.`, `==`, `->` and the rest), but not the range operators. So a named unary or sub name before `..` or `...` took the operator as the start of its argument: `(undef..2)` canonicalised to `undef(..);2)` and refused twice.

Measured with perl 5.42:

    $ perl -MO=Deparse,-p -e 'sub g{1} sub f { my @x = (undef..2);
          my @y = (shift .. 2); my @z = (lc .. 2); my @w = (g ... 2) }'
    (my(@x) = ((undef) .. 2));
    (my(@y) = ((shift()) .. 2));
    (my(@z) = (lc($_) .. 2));
    (my(@w) = (g() ... 2));

perl.git t/op/range.t:137 has `is(join(":",undef..2), '0:1:2');`, and the file refuses 15 times.

## Acceptance criteria

- [ ] A range operator after a word ends its argument list (`go test ./internal/parse/ -run '^TestRangeEndsTheArgumentList$' -count=1 -v | grep -q '^--- PASS: TestRangeEndsTheArgumentList'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
