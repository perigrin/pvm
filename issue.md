---
title: Faithful counts a dropped trailing comma as a misread
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.861463591Z
updated: 2026-09-29T22:20:52.861463591Z
---

Faithful fails every file with a trailing comma, even though canon's tree is right. Canon drops the separator, as perl does (`f(1, 2,)` deparses as `f(1, 2)`). Faithful forgives only an added paren after a call name and a `;` before a closing `}`:

    f(1, 2,);            faithful=false  token 5: source has ",", canon has ")"
    my @a = (1, 2,);     faithful=false
    sub f {} f 1, 2,;    faithful=false
    print 1, if $x;      faithful=false

So t1fidelity counts these files as misread. PerlOnJava unit/prototype_trailing_comma_infix.t went 0 -> 1 when its lines began to parse: before, they were Unknown and compared verbatim.

A third forgiven difference would be: a source `,` that canon omits, where the next source token is canon's next token and cannot begin an element (a closer, `;`, `,`, `=>`, a modifier, or an operator). That condition keeps it from hiding two elements merged into one. This changes what the fidelity harness measures, so it needs perigrin's decision before code.

## Acceptance criteria

- [ ] `f(1, 2,);` is faithful (a test asserting parse.Faithful on the trailing-comma forms, named in the fix)
