---
title: "`substr $x, 0, 1, = \"a\"` refuses: a comma before `=` looks for an element"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T19:51:01.564381681Z
updated: 2026-09-29T19:51:01.564381681Z
---

`substr $x, 0, 1, = "\x{100}"` refuses. The trailing-comma rule drops a comma only when the next token is a closer, a terminator, a modifier or a low-precedence word operator. For `=` it goes looking for an element that does not exist. `=` cannot start an element, and because it binds tighter than the comma, the list ends at the comma and the list operator becomes the assignment's left side.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'substr $x, 0, 1, = "\x{100}";'
    substr($x, 0, 1) = "\x{100}";

perl.git t/op/utf8cache.t:70.

## Acceptance criteria

- [ ] A trailing comma before an assignment operator ends the list (`go test ./internal/parse/ -run '^TestTrailingCommaBeforeAssignment$' -count=1 -v | grep -q '^--- PASS: TestTrailingCommaBeforeAssignment'`)
