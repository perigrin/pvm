---
title: "Filehandle builtins reject a lexical filehandle: expects Glob, got Scalar"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T05:30:01.903192468Z
updated: 2026-10-14T05:30:01.903192468Z
---

## Context

- paths: `internal/types/CORE.pmt` (filehandle slots typed Glob: close, seek, read, stat, ...), `internal/infer/lowered_calls.go`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-14 by cutover 1 (01a0c115-a44b): `call to seek: argument 1 expects Glob, got Scalar` fires on lexical filehandles on BOTH paths (142 such lines added and 395 removed in the cutover diff only because parse coverage moved; e.g. Archive/Tar.pm:1449, App/Prove/State.pm:531). A lexical filehandle is a Scalar holding a GlobRef (or an IO object): `open(my $fh, "<", $f); seek($fh,0,0); read($fh, my $b, 3); close($fh)` runs clean under perl 5.42 -w. perigrin ruled 2026-10-14 to file it and fix after the cutover (01a0c115-df67).

## Steps

- [ ] Failing test first; a filehandle slot accepts Glob, GlobRef and a Scalar (which may hold either), reporting only a type that cannot be a handle (e.g. HashRef, Int literal); `make test`; commit

## Acceptance criteria

- [ ] a lexical filehandle satisfies a filehandle slot (`go test ./internal/infer/ -run '^TestLexicalFilehandleSatisfiesHandleSlot$' -count=1 -v | grep -q '^--- PASS: TestLexicalFilehandleSatisfiesHandleSlot'`)
