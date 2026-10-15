---
title: The parser ignores use lib, so imports from a use lib directory go unresolved
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-15T06:03:18.615387465Z
updated: 2026-10-15T06:03:18.615387465Z
---

## Context

- paths: `internal/parse/use.go` (`quietPragmas` lists "lib"), `internal/parse/` loader (`Session.ParseFileFrom`), `internal/perl/inc.go` (static @INC)
- commands: `go test ./internal/parse/`

Found 2026-10-14 by cutover 1 (01a0c115-a44b): `psc check` now resolves imports through a statically built @INC (perigrin ruled perl is never executed for it), but internal/parse does not honour `use lib` -- `use.go` only lists "lib" as a quiet pragma. perl applies `use lib 'dir'` at compile time, prepending to @INC before later `use` statements in the same file, so `use lib 't/lib'; use My::Helper;` resolves My::Helper from t/lib. Without it, a parenless call to a sub imported from a `use lib` directory is refused, and its statement becomes a Hole.

## Steps

- [ ] Failing test first: a module found only through a preceding `use lib` resolves its imports; `use lib` with a constant string or list (relative to the file's directory, as perl resolves relative to cwd -- measure and decide); `no lib` removes; never execute anything; `make test`; commit

## Acceptance criteria

- [ ] a preceding `use lib` adds its directory to the loader's search path (`go test ./internal/parse/ -run '^TestUseLibExtendsSearchPath$' -count=1 -v | grep -q '^--- PASS: TestUseLibExtendsSearchPath'`)
