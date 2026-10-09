---
title: "A C-style for head canonicalises as garbage: the clauses have no slots"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:20:34.136271682Z
updated: 2026-10-09T23:54:16.85003825Z
sessions:
- start_sha: d965c9877e9c5dfe0d0e3361dfc095beab6acbb1
  end_sha: 419a99b64470f2c2a547bcd922fd613573eb6dbf
  commits: 1
  started_at: 2026-10-09T23:03:08.940686655Z
  ended_at: 2026-10-09T23:54:16.85003825Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T23:03:08.940686655Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T23:54:16.85003825Z
observed_paths:
- internal/parse/canon.go
- internal/parse/commafat_test.go
- internal/parse/control.go
- internal/parse/cstylefor_test.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/testdata/fuzz/FuzzParser/8cbe7cdacf7dc963
- internal/parse/testdata/fuzz/FuzzParser/a162c0187a1bd0a1
- internal/parse/testdata/t1fidelity.ratchet
---

A C-style `for` head canonicalises as garbage at Unknown=0, because the tree keeps its clauses as plain children with no record of which slot each fills:

    for ($i = 0; $i < 3; $i++) { 1 }        canon: for ($i = 0) $i < 3 $i++ {1;}
    for (my $i = 0; $i < 3; $i++) { 1 }     canon: for (my $i = 0) ($i < 3) $i++ {1;}

parseForHead knows the form (its cStyle result) and drops the empty slots of `for (;;)`, so neither the separators nor an empty init, condition or step can be written back. The Loop needs to carry the three slots -- an empty one included -- for canon to emit `for (INIT; COND; STEP)`.

Measured with perl 5.42, `perl -MO=Deparse -e 'for ($i = 0; $i < 3; $i++) { 1 }'` prints `for ($i = 0; $i < 3; ++$i) {`.

## Acceptance criteria

- [ ] A C-style for head's canon is `for (INIT; COND; STEP)`, empty slots included, and is a fixpoint (`go test ./internal/parse/ -run '^TestCStyleForCanon$' -count=1 -v | grep -q '^--- PASS: TestCStyleForCanon'`)
