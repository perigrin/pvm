---
title: "Statements inside a deref's braces refuse: `${no strict; \\$_}` is a block"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T11:14:23.1199479Z
updated: 2026-09-29T11:14:23.1199479Z
---

The braces after a deref sigil hold a BLOCK in perly.y (`indirob: BAREWORD | scalar | block | PRIVATEREF`), and the parser reads them as one expression, so statements inside refuse: `${no strict; \$_}`, `${; do { j: \(my $foo = "foo") } }`.

Measured with perl 5.42:

    $ perl -MO=Deparse -e '$_ = "r"; my $a = ${no strict; \$_};
          my $b = ${; do { j: \(my $foo = "foo") } }; my @c = @{; [1,2] };'
    my $a = ${\$_;};
    my $b = ${do {

perl.git t/op/lex.t:99 and t/op/goto.t's `${; do { LABEL: ... } }` tests (lines 674-742).

## Acceptance criteria

- [ ] Statements in a deref's braces parse as a block, round-trip, and canon to a fixpoint; `${$x}` stays one expression (`go test ./internal/parse/ -run '^TestDerefHoldsStatements$' -count=1 -v | grep -q '^--- PASS: TestDerefHoldsStatements'`)
- [ ] REGRESSION GUARD: every refusal site keeps a code (`go test ./internal/parse/ -run '^TestEveryUnknownSiteHasACode$' -count=1 -v | grep -q '^--- PASS: TestEveryUnknownSiteHasACode'`)
