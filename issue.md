---
title: Read typed variables with :export and :export_ok in declaration files
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
blocks:
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
created: 2026-10-16T20:02:15.387011621Z
updated: 2026-10-16T20:03:18.495603975Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one

## Context

- paths: `internal/parse/declarations.go`, `internal/parse/exports.go`, `internal/parse/typedour_test.go` (new), `internal/parse/testdata/declarations/` (test fixtures), `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -run '^TestTypedOur' -count=1 -v`
- entrypoints: `readDeclarationWith` (`declarations.go`), `readExportTags` and `moduleFacts.exports`/`defaults` (`exports.go`), the typed reader that already refuses typed Perl outside a `.tpm`

RFC 0001 "Core-module libraries", "Typed variables" (perigrin, 2026-10-16): a declaration states an exported variable with its type before it and the export as an attribute, `our Hash[Str] %Config :export;`. `:export` places the variable in `@EXPORT` (a bare `use` brings it in); `:export_ok` places it in `@EXPORT_OK` (it must be asked for), mirroring Exporter's lists. A typed `our` is typed Perl, legal only in a `.tpm`, as a typed signature is. Nothing reads this today. The parser records the variable's declared type in a table beside `signatures`, for infer to read later; this issue does not make infer consume it. It blocks the variable-exporting declarations (Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat, and Socket's `$CR $LF $CRLF`).

## Steps

- [ ] Write failing test `TestTypedOurExport`: a fixture `.tpm` with `our Hash[Str] %Config :export;`; `use M;` imports `%Config`, and the module's facts record its type `Hash[Str]`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestTypedOurExport$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestTypedOurExportOk`: `our Str $CRLF :export_ok;` is not imported by `use M;` and is by `use M qw($CRLF);`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestTypedOurOnlyInDeclarations`: the same statement in a module's own source is not read as typed Perl (it stays what perl reads it as), as `sub f (Ref $x) { }` does
- [ ] Implement; run to verify pass; commit
- [ ] Run the m1 gates named in the final-gate issue; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] `:export` puts a typed variable in @EXPORT with its type (`go test ./internal/parse/ -run '^TestTypedOurExport$' -count=1 -v | grep -q '^--- PASS: TestTypedOurExport'`)
- [ ] `:export_ok` puts a typed variable in @EXPORT_OK (`go test ./internal/parse/ -run '^TestTypedOurExportOk$' -count=1 -v | grep -q '^--- PASS: TestTypedOurExportOk'`)
- [ ] a typed `our` is typed Perl only in a .tpm (`go test ./internal/parse/ -run '^TestTypedOurOnlyInDeclarations$' -count=1 -v | grep -q '^--- PASS: TestTypedOurOnlyInDeclarations'`)
