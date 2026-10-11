---
title: "A sub attribute ends the sub early: sub :lvalue leaves its body trailing"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:41:11.595550955Z
updated: 2026-10-11T15:40:03.344194746Z
sessions:
- start_sha: 2257e08ee261e0a0303f9201d0210d81f034b326
  end_sha: 2257e08ee261e0a0303f9201d0210d81f034b326
  commits: 0
  started_at: 2026-09-26T21:16:21.035437975Z
  ended_at: 2026-09-26T21:16:29.932580807Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T21:16:21.035437975Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T21:16:29.932580807Z
---

A sub with an attribute -- `sub :lvalue { 1 }`, `sub :prototype($) { 1 }` --
refuses, anonymous and named alike. The same sub without the attribute parses
clean, so the defect is the attribute list between the name and the body.

## Measured 2026-09-26 at `2dc301fd`

Each program is the whole file:

    my $s = sub :lvalue { 1 };           Unknown=1  trailing_tokens ":lvalue { 1 }"
    my $s = sub :prototype($) { 1 };     Unknown=1  trailing_tokens ":prototype($) { 1 };"
    sub f :lvalue { 1 }                  Unknown=1  trailing_tokens ":lvalue { 1 }"

    my $s = sub { 1 };                   Unknown=0  <- no attribute is fine
    sub x () { 8 }                       Unknown=0  <- a prototype in parens is fine

The span shows the parser finishing the sub at `sub` and leaving the attribute
and the whole body as trailing tokens, which is why one attribute costs the rest
of the statement.

Valid perl 5.42.0:

    /home/perigrin/.local/bin/perl -e 'my $s = sub :lvalue { 1 }; print $s->(),"\n"'     -> 1
    /home/perigrin/.local/bin/perl -e 'sub f :lvalue { 1 } print f(),"\n"'               -> 1
    /home/perigrin/.local/bin/perl -e 'my $s = sub :prototype($) { 1 }; print $s->(1)'   -> 1

## Reach

Appears in 26 of perl.git t/'s 620 files, 25 of them dirty. `t/op/sub_lval.t`
and `t/op/lex.t` are the dense cases; `:lvalue` dominates, with `:prototype`
and `:method` in the tail.

## Not to be confused with the cascade

`not_a_term` is 1,584 nodes but only 41 are the first Unknown in their file; the
other 1,543 are resync fallout. This construct is one of the 41-node cause
surface. See the `not_a_term` section of
`docs/plans/2026-09-26-perl-git-t-parses.md`.

### Criteria named (2026-10-11)

Measured at `971b9031`: every spelling below parses at Unknown=0 and canons with
the attribute inside the sub. TestSubAttribute, added by the fix 684229e0 in
internal/parse/subattr_test.go, already asserts the canon of each claimed
spelling, so no test was written.

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Every criterion below ran the whole parse
package, the whole conformance package or `go test ./...`, which only restates
"make test passed" -- and make test gates every commit. Each claim now names
the test that holds it. Claims are reworded only to keep parentheses out of
the claim text, since git-zhi runs the first paren span on the line.

The two regression guards keep their claims and name gate tests instead:
TestCorpusRatchet for the corpus ratchet, and TestParseRatchet and
TestPerlGitTRatchet -- the T1 and perl.git t/ parse ratchets -- for the suite.

## Acceptance criteria

- [ ] `my $s = sub :lvalue { 1 };` parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubAttribute$' -count=1 -v | grep -q '^--- PASS: TestSubAttribute'`)
- [ ] `sub f :lvalue { 1 }` parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubAttribute$' -count=1 -v | grep -q '^--- PASS: TestSubAttribute'`)
- [ ] An attribute taking an argument, `:prototype` with a `$` argument on an anonymous sub, parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubAttribute$' -count=1 -v | grep -q '^--- PASS: TestSubAttribute'`)
- [ ] `sub x` with an empty prototype, and `sub { 1 }`, still parse, no regression (`go test ./internal/parse/ -run '^TestSubAttribute$' -count=1 -v | grep -q '^--- PASS: TestSubAttribute'`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds, named as the T1 and perl.git t/ parse ratchets (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet' && go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
