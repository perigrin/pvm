---
title: "User library .pmt files: decide their search path (RFC open question 6) and have psc report their errors"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T15:04:15.752368845Z
updated: 2026-10-16T18:48:33.695990031Z
---

## Context

- paths: `internal/parse/declarations.go` (embedded `declarations` FS), `internal/parse/module.go` (resolver), `internal/psc/check_command.go` (`checkFile`), RFC 0001 open question 6
- commands: `make test`

From 01a0fd9b-5698-70cf-b6e4-8899a83b0fe0, whose gate item (1) could not be built: library .tpm files come only from the embedded `declarations` FS, so no user-written .tpm -- and so no broken one -- can reach `psc check`, which parses with internal/parser (tree-sitter) and never calls internal/parse. RFC 0001 open question 6 (where user .tpm files are found) is undecided; 5698 put it out of scope.

## Steps

- [ ] perigrin decides open question 6: the search path for user .tpm files (beside the module, a project dir, PERL5LIB-style, ...)
- [ ] Implement the search path with a failing test
- [ ] Have `psc check` parse with internal/parse as well and print `parse.DeclarationErrors(root)`; a test that `psc check` on a file importing a module whose user .tpm is in error prints the error; `make test`; commit

## Acceptance criteria

- [ ] psc check reports a user library .tpm's declaration errors (`go test ./internal/psc/ -run '^TestCheckReportsLibraryTpmErrors$' -count=1 -v | grep -q '^--- PASS: TestCheckReportsLibraryTpmErrors'`)
