---
title: "Attributes before a signature split the sub: `sub f :prototype($) ($a) {...}`"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:36:04.080118176Z
updated: 2026-09-29T08:49:15.048981007Z
sessions:
- start_sha: 09b4134381cb566dc9d94e715c6e742875f8d50e
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T08:49:15.048981007Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T08:49:15.048981007Z
---

Under the signatures feature, a sub's attributes come BEFORE its signature; perl rejects them after ("Subroutine attributes must come before the signature"). parseSubDecl reads the parens first and the attributes second -- the prototype order -- so `sub t105 :prototype($) ($a) { ... }` stops after the attribute list and the signature and body are left over. At Unknown=0 the tree was wrong; since 01a0ec21-837a-73e2-9f6d-dba44e242bd3 it refuses.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "signatures"; sub t105 :prototype($) ($a) { $a || "z" }'
    sub t105 : prototype($) ($a) {

perl.git t/op/signatures.t:1174, 1186.

## Acceptance criteria

- [ ] Attributes before a signature parse as one sub declaration (`go test ./internal/parse/ -run '^TestAttributesBeforeSignature$' -count=1 -v | grep -q '^--- PASS: TestAttributesBeforeSignature'`)
