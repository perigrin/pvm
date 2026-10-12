---
title: An anonymous sub with a signature opens a node-identity anchor named after its parameter
state: done
urgency: normal
milestone: m2-lower-the-tree
created: 2026-10-12T17:13:20.907907353Z
updated: 2026-10-12T17:58:35.44984818Z
sessions:
- start_sha: 34b03d318ec7428fe87316e7752b9ef1f1f6164e
  end_sha: 74ef5d96a45bfa0f13ccbde7572cce6b943e1329
  commits: 2
  started_at: 2026-10-12T17:14:12.368426573Z
  ended_at: 2026-10-12T17:58:35.44984818Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T17:14:12.368426573Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T17:58:35.44984818Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/expr.go
- internal/lower/id.go
- internal/lower/id_test.go
- internal/lower/node.go
---

## Context

- paths: `internal/lower/id.go` (`anchorKey`), `internal/lower/id_test.go`
- commands: `go test ./internal/lower/ -run '^TestAnonSubWithSignatureOpensNoAnchor$' -count=1 -v`

Found 2026-10-11 by the declarations slice (01a0c112-8a65). Node identity (01a0c10e-c81e) opens an anchor at a NAMED `sub`/`method` -- a Declaration whose first child is a Term with text. An anonymous sub with a one-parameter signature, `sub ($x) { ... }`, parses as `[declaration "sub" [term "$x" paren] [block ...]]` (§4.14.2, re-measured at 971b9031), so its parameter is taken for its name and it opens an anchor "sub $x". Anonymous subs are meant to open no anchor. With this bug, two anonymous subs
with the same one-parameter signature get the anchors "sub $x#0" and
"sub $x#1", so inserting one above the other renumbers it -- the instability
anchors exist to prevent.

## Steps

- [ ] Failing test: in `use v5.36; my $f = sub ($x) { 1 };` every node's anchor is "" (anonymous subs open none), while `sub f ($x) { 1 }` still anchors "sub f"; fix anchorKey to require that the first child is not a parenthesised signature; `make test`; commit

## Acceptance criteria

- [ ] an anonymous sub with a signature opens no anchor, and a named sub with one still does (`go test ./internal/lower/ -run '^TestAnonSubWithSignatureOpensNoAnchor$' -count=1 -v | grep -q '^--- PASS: TestAnonSubWithSignatureOpensNoAnchor'`)
