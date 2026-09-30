---
title: Faithful counts a dropped trailing comma as a misread
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.861463591Z
updated: 2026-09-30T19:54:22.526268991Z
---

Faithful failed every file with a trailing comma, even when canon's tree was right. Canon dropped the separator, as perl does (`f(1, 2,)` deparses as `f(1, 2)`), and Faithful forgave only an added paren after a call name and a `;` before a closing `}`.

perigrin decided (2026-09-30) that canon keeps trailing commas where the source wrote them, rather than Faithful forgiving the difference. The parser records the separators it drops on the element before them (Node.TrailingComma), and canon writes them back. When an element is followed by several separators, all of them are kept in source order and canon writes no separator of its own after it, so `(1, => 2)` stays `(1 , => 2)`.

A second limit showed up once PerlOnJava unit/print_indirect_method_postfix.t parsed: Faithful forgave a call's added parens only at the top level, not inside another call's forgiven parens (`print(create Foo(sub {...}))`). The same forgiveness now applies one level down.

## Acceptance criteria

- [ ] Canon keeps trailing commas and stays faithful (`go test ./internal/parse/ -run '^TestCanonKeepsTrailingComma$' -count=1 -v | grep -q '^--- PASS: TestCanonKeepsTrailingComma'`)
- [ ] A nested call's added parens are forgiven (`go test ./internal/parse/ -run '^TestFaithfulNestedCallParens$' -count=1 -v | grep -q '^--- PASS: TestFaithfulNestedCallParens'`)
