---
title: "`print $cfh $$code` refuses: a deref sigil does not count as starting a term"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T09:55:24.546565595Z
updated: 2026-09-29T10:10:41.663215031Z
sessions:
- start_sha: c1896bb6b4a8198ecab4c5a98e4aea747a6ceee6
  end_sha: 5f106d9d86951f7f7804461574c45246da5500d3
  commits: 1
  started_at: 2026-09-29T09:55:26.875064358Z
  ended_at: 2026-09-29T10:10:41.663215031Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T09:55:26.875064358Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T10:10:41.663215031Z
observed_paths:
- internal/parse/call.go
- internal/parse/handlederef_test.go
- internal/parse/testdata/perlgitt.ratchet
---

`print $cfh $$code;` refuses. toke.c takes the scalar after print as a handle when a `$` or `@` follows it (yyl_dollar's `memCHRs("$@\"'`q", *s)`) and the lexer agrees, but the parser's startsTerm, asked by parseFilehandleSlot whether what follows begins a term, did not count the DerefSigil token that `$$code` and `@$lines` begin with.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'print $cfh $$code; print $fh @$lines; print $x, $$y;'
    print $cfh $$code;
    print $fh @$lines;
    print $x, $$y;

perl.git t/porting/cpphdrcheck.t:276 and 340.

## Acceptance criteria

- [ ] A dereference after `print $fh` makes $fh the handle (`go test ./internal/parse/ -run '^TestHandleBeforeDereference$' -count=1 -v | grep -q '^--- PASS: TestHandleBeforeDereference'`)
