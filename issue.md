---
title: "A sub declared in a package is known in every package: the sub table ignores the current package"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:35:13.046766756Z
updated: 2026-10-11T02:48:24.314868738Z
---

A sub declared inside a package is recorded under its bare name, so it is "known" in every package. `package Pack; sub method {...} package main; method Pack ("a")` then declines the indirect reading, because `method` looks like a known sub in main, where perl finds none and makes the indirect call:

    $ perl -MO=Deparse -e 'package Pack; sub method { "m" } package main; my $x = method Pack ("a");'
    my $x = 'Pack'->method('a');

declareSub keys the sub table with subKey(name), which strips `main::` but adds no package for an unqualified name. The parser does not track the current package, so a lookup cannot ask which package a call is in.

No perl.git t/ file is known to depend on this yet; op/method.t declares `sub Pack::method` qualified.

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `package Pack; sub method { "m" } package main; my $x = method Pack ("a");` parses as the indirect call to `method`, matching perl's `'Pack'->method('a')`; declared in main, `method Pack ("a")` is `method(Pack("a"))`, matching perl's `main::method(Pack('a'))`. Fixed by 0b13f919, which tracks the current package and keys an unqualified sub by it.

The example above does not tell a scoped sub table from a package-blind one: `package Pack;` makes Pack a known package, and perl reads `method Pack (...)` as indirect for that alone. It still passes with package tracking removed. The cases that do tell them apart declare the sub in another package, Foo: from main, perl reads `method Pack ("a")`, `method $obj` and `f Bar` as indirect calls. A package-blind table made each a direct call to Foo's sub at Unknown=0. At pu the parser refuses them (Unknown=1) rather than reading the indirect call, so no wrong tree is built, but perl's reading is not reached either.

TestSubKnownOnlyInItsPackage was named here but did not exist. It was written in test/pgt-reconcile; it asserts the example, its control, and that none of the three Foo cases calls Foo's sub directly. It fails on all three when 0b13f919's package tracking is removed. TestRefScalarSlotIsPackageScoped, from 0b13f919, asserts the same scoping through `\$` prototypes.

## Acceptance criteria

- [ ] A sub declared in a package is known only there (`go test ./internal/parse/ -run '^TestSubKnownOnlyInItsPackage$' -count=1 -v | grep -q '^--- PASS: TestSubKnownOnlyInItsPackage' && go test ./internal/parse/ -run '^TestRefScalarSlotIsPackageScoped$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotIsPackageScoped'`)
