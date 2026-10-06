---
title: Read IO::File's computed export list, so T1's indirect_object_constructor.t returns to 0
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T11:43:33.863867274Z
updated: 2026-10-06T11:43:33.863867274Z
---

## Context

- paths: `internal/parse/exports.go` (export-list reading), `internal/parse` T1 ratchet baseline
- commands: `go test ./internal/parse/ -run '^TestParseRatchet$' -count=1`; perl 5.42 `/home/perigrin/.local/bin/perl`

The package-scoping fix for 01a10dc1-3530-762f-9a03-177445629c6b (0b13f919) raised T1's `indirect_object_constructor.t` from 0 to 1 Unknown, accepted because the old 0 was a WRONG tree: another package's `new ($$;$)` prototype leaked into main and parsed `new Win32::Process::Info->Subprocesses(1)` as `new(Win32::Process::Info->Subprocesses(1))`; perl reads `'Win32::Process::Info'->new->Subprocesses(1)`. Now `new` is undeclared in main, but the file's `use IO::File` has an export list pvm cannot read (computed), so `new` might be imported and the statement is refused rather than guessed.

## Steps

- [ ] Measure what IO::File (and its parents IO::Handle, IO::Seekable, Exporter) actually export, and how the list is computed
- [ ] Read that form of computed export list, with a failing test; the T1 file returns to 0 with perl's tree (verify with Deparse); lower the baseline; `make test`; commit

## Acceptance criteria

- [ ] indirect_object_constructor.t parses with no Unknown and perl's tree (`go test ./internal/parse/ -run '^TestIOFileExportListIsRead$' -count=1 -v | grep -q '^--- PASS: TestIOFileExportListIsRead'`)
