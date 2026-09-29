---
title: "A comma before a fat comma refuses: `tie $@, => \"main\", 1`"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T13:11:42.079218768Z
updated: 2026-09-29T13:11:43.168905374Z
sessions:
- start_sha: a5b8c9b1d7843b16fbf0cabe247695a68e21f8a6
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T13:11:43.168905374Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T13:11:43.168905374Z
---

A comma directly followed by a fat comma refuses: `tie $@, => 'main', 1;`, `(1, => 2)`. perl drops a separator with no element after it, and a fat comma is a separator; the parser's empty-slot rules -- parseInfix's elementlessAfterComma and parseParenList's empty slot -- took only a plain `,` there.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'tie $@, => "main", 1; my @a = (1, => 2);'
    tie $@, 'main', 1;
    my(@a) = (1, 2);
    $ perl -MO=Deparse -e 'sub foo {7} my @x = (foo, => 1);'
    my(@x) = (foo(), 1);

The fat comma quotes nothing there: the word before it is behind the dropped comma.

perl.git t/op/tie_fetch_count.t:331-335; t/re/reg_mesg.t's `'m/\xABC/' => "",\n => '...'` lists.

## Acceptance criteria

- [ ] A comma before a fat comma is an empty slot, and the fat comma quotes nothing (`go test ./internal/parse/ -run '^TestCommaBeforeFatComma$' -count=1 -v | grep -q '^--- PASS: TestCommaBeforeFatComma'`)
