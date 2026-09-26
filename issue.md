---
title: no feature signatures turns signatures ON, so a prototype is read as a signature
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T11:17:53.425614128Z
updated: 2026-09-26T11:20:37.427904675Z
sessions:
- start_sha: f178e76df654244563b62e3b3f68f49cdfef252b
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T11:20:37.427904675Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T11:20:37.427904675Z
---

`no feature "signatures"` turns signatures ON. `noteSignatures` has no path to
false, so a `($)` prototype after that pragma is read as a signature and the
sub's `return` body goes Unknown. One corpus case, `adjacency-07_subroutines`,
whose `refuses:` cites a CLOSED issue.

## Measured, and the bisect is the finding

`adjacency-07_subroutines.md`'s case has ONE Unknown. Bisected by progressive
prefix, it appears at line 14, `sub g ($) { return "g[" . $_[0] . "]" }` -- but
that line ALONE parses clean. The refusal needs three things at once, and
dropping any one of them clears it:

    body                              alone   with `use v5.36; no feature "signatures";`
    sub g ($) { return "a" . $_[0] }    0       1
    sub g ($) { return $_[0] }          0       1
    sub g ($) { return "a" . "b" }      0       1
    sub g ($) { return "g[" }           0       1
    sub g ($) { "g[" . $_[0] . "]" }    0       0     <- no `return`
    sub g    { return "g[" . ... }      0       0     <- no prototype
    sub g () { return "g[" . ... }      0       0     <- EMPTY prototype
    sub g ($$) { return "g[" . ... }    0       1

So: a NON-EMPTY prototype, plus `return` in the body, plus the pragma pair.
An empty `()` is clean because there is nothing in it to misread as a
parameter list.

## Perl's reading, for comparison

Perl parses it the same with or without the pragma -- the pragma changes
nothing about how the sub reads, because a prototype is a prototype:

    perl -MO=Deparse -e 'use v5.36; no feature "signatures"; sub g ($) { return "a" . $_[0] } print g 1;'
      use feature ':5.36';
      no feature 'signatures';
      sub g ($) { return 'a' . $_[0]; }
      print g(1);

    perl -MO=Deparse -e 'sub g ($) { return "a" . $_[0] } print g 1;'
      sub g ($) { return 'a' . $_[0]; }
      print g(1);

Identical. So we diverge from perl on valid input.

## Root cause

`internal/lexer/ident.go:199`'s `noteSignatures` returns immediately unless
`l.pendingPragma == 1`:

    if l.pendingPragma != 1 {
        return
    }

`pendingPragma` is 1 for `use` and 2 for `no` (`ident.go:158-161`), so the
`no` spelling is silently ignored. And every assignment in the function is
`l.signatures = true` -- five of them, at lines 220, 226, 232, 250 and 257.
There is no `= false` anywhere in the lexer.

`proto.go:32` is the consumer and it is RIGHT about perl's rule, citing
toke.c:5923: the feature decides, not the contents, because `($a, $b)` is a
valid prototype spelling too. It just never sees the feature go off.

The fix has a sibling two cases above it. `utf8` handles both spellings:

    case "utf8":
        switch l.pendingPragma {
        case 1:
            l.utf8Pragma = true
        case 2:
            l.utf8Pragma = false
        }

`signatures` needs the same shape.

## The stale citation, which is the second half

The case carries `refuses: 01a0c432-fbd5`. That issue is state DONE,
milestone m3-conformance-corpus, titled "Tier 07 call-form slice: the
smallest useful corpus gate" -- a corpus-construction issue that never owned
a parser gap. Repoint it here or drop both lines when the fix lands.

This is the SECOND corpus case found citing a closed m3 issue; the first was
`undef`, now `01a0dd43`. Worth checking whether the others do too:
`adjacency-11_oo.md` cites `01a0c35f-a487`, also DONE, also m3.

## Scope

1. Give `noteSignatures` a false path for `pendingPragma == 2`, matching
   `utf8`'s shape. Measure whether the `use feature 'signatures'` arm and the
   version-bundle arms all need it, or only the quoted-feature arm -- perl
   permits `no feature ':5.36'` and it is not obvious we should.
2. Decide whether this should be lexically scoped. `noteSignatures`'s own
   docstring says the tracking is deliberately file-level "like utf8Pragma
   and for the same reason", and the corpus case re-enables with
   `use feature "signatures"` on the next line, so file-level plus a false
   path is enough for it. Record the decision either way.
3. Repoint or remove the case's `refuses:`/`refusal:` lines.

## Context

- paths: `internal/lexer/ident.go` (`noteSignatures` at 199, the pragma
  switch at 156), `internal/lexer/proto.go:32` (`scanPrototype`)
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: `noteSignatures`'s `pendingPragma != 1` guard

## Acceptance criteria

- [ ] `no feature "signatures"` turns the feature off, so a prototype stays a prototype (`go test ./internal/lexer/ -run '^TestNoFeatureSignaturesTurnsItOff$' -count=1 -v | grep -q '^--- PASS: TestNoFeatureSignaturesTurnsItOff'`)
- [ ] The adjacency-07 case parses (`go test ./internal/conformance/ -run 'TestMdtestTopicsMakeTheSameClaims/adjacency-07' -count=1 | grep -q '^ok'`)
- [ ] The case no longer cites a closed issue (`grep -q '01a0c432' conformance/mdtest/adjacency-07_subroutines.md; test $? -ne 0`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
