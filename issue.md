---
title: use Foo 1.2 qw(a) splits its import list into a separate statement
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-12T21:48:12.192258849Z
updated: 2026-10-12T21:48:12.192258849Z
---

## Context

- paths: `internal/parse/use.go` (use/no with a module version)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-12 by the m2 statements slice (01a12a85-2fa3). `use Foo 1.2 qw(a);` -- a module name, a version, then an import list -- parses with `qw(a)` split off into a separate statement. perl 5.42 reads it as one `use`: `use List::Util 1.2 qw(max); print max(1,3)` prints 3, and Deparse keeps `use List::Util 1.2 ('max')`. With the list split away, the import is lost: the lowered UseStmt has no Args, and `psc analyze` (01a12a85-34bf) would miss the imported name.

## Steps

- [ ] Measure `use Foo VERSION LIST`, `use Foo VERSION`, `no Foo VERSION LIST` with Deparse; failing test first; one statement, the list its import list; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a use with a module version keeps its import list in the same statement (`go test ./internal/parse/ -run '^TestUseVersionKeepsItsImportList$' -count=1 -v | grep -q '^--- PASS: TestUseVersionKeepsItsImportList'`)
