---
title: "A comma with no element after it refuses unless a closer follows: f(1, , 2) and skip ..., if $m"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T02:42:29.097984945Z
updated: 2026-09-29T03:03:53.930152719Z
sessions:
- start_sha: ae2fb4786754099203add719feedfd96fedd7208
  end_sha: 5018701de8e42bd5acfd54be73995bc173201a25
  commits: 1
  started_at: 2026-09-29T02:47:25.258627448Z
  ended_at: 2026-09-29T03:03:53.930152719Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T02:47:25.258627448Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T03:03:53.930152719Z
observed_paths:
- internal/parse/emptycomma_test.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

A comma with no element after it is dropped by perl wherever it stands, not
only before a closer. The guard 652f4792 added to parseInfix (issue
01a0dfbd-3fac) covers a comma before `)`, `]` or `}`; these still refuse:

    sub f {} f(1, , 2);                  perl: f 1, 2;
    sub skip {} skip "x", 2, if $m;      perl: skip 'x', 2 if $m;
    close $fh, or die;                   perl: die unless close $fh;
    open my $fh, , "<", $f;              perl: open my $fh, '<', $f;

(each is `perl -MO=Deparse` on 5.42.0). So a comma before another comma, before
a `;`, and before a statement modifier or a low-precedence word operator is a
separator with nothing to separate.

Reach: the first Unknown in 11 dirty perl.git t/ files -- `skip ..., if
is_miniperl` in op/aassign.t, op/each.t, op/for.t, op/state.t and op/sub_lval.t,
`like($@, , qr/.../)` in re/reg_nc_tie.t, `close $manifest_fh, or die` in
porting/podcheck.t, and more.

A declaration followed by a comma -- `my $aa, $bb, $cc;` in op/or.t and
opbasic/cmp.t -- is a separate path (parseVarDeclNoSemi) and is included here:
01a0eab6 left the comma out of the declaration-as-left-operand rule, and perl
reads `(my $aa), $bb, $cc`.

## Acceptance criteria

- [ ] A comma before another comma, a terminator, a modifier or a low-precedence word is dropped, and each case canons to perl's reading (`go test ./internal/parse/ -run '^TestEmptyCommaSlot$' -count=1 -v | grep -q '^--- PASS: TestEmptyCommaSlot'`)
- [ ] A declaration with no initialiser is the left operand of a following comma (`go test ./internal/parse/ -run '^TestEmptyCommaSlot$' -count=1 -v | grep -q '^--- PASS: TestEmptyCommaSlot'`)
- [ ] NEGATIVE: a comma with nothing before it still refuses (`go test ./internal/parse/ -run '^TestTrailingCommaBeforeCloserIsNotATerm$' -count=1 -v | grep -q '^--- PASS: TestTrailingCommaBeforeCloserIsNotATerm'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
