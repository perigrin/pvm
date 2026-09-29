---
title: "A sub declared in a package is known in every package: the sub table ignores the current package"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:35:13.046766756Z
updated: 2026-09-29T16:35:13.046766756Z
---

A sub declared inside a package is recorded under its bare name, so it is "known" in every package. `package Pack; sub method {...} package main; method Pack ("a")` then declines the indirect reading, because `method` looks like a known sub in main, where perl finds none and makes the indirect call:

    $ perl -MO=Deparse -e 'package Pack; sub method { "m" } package main; my $x = method Pack ("a");'
    my $x = 'Pack'->method('a');

declareSub keys the sub table with subKey(name), which strips `main::` but adds no package for an unqualified name. The parser does not track the current package, so a lookup cannot ask which package a call is in.

No perl.git t/ file is known to depend on this yet; op/method.t declares `sub Pack::method` qualified.

## Acceptance criteria

- [ ] A sub declared in a package is known only there (`go test ./internal/parse/ -run '^TestSubKnownOnlyInItsPackage$' -count=1 -v | grep -q '^--- PASS: TestSubKnownOnlyInItsPackage'`)
