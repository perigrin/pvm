---
title: Type CORE.pmt's I/O, file and directory builtins
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
- 01a10177-4e7b-7d2c-8d2a-2033647ffc4a
blocks:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-03T11:12:38.999799272Z
updated: 2026-10-05T14:20:10.821666785Z
---

## Prerequisites

- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print "@{$Kinds{q(I/O)}} @{$Kinds{Binary}} @{$Kinds{File}}"'`

Third batch, by perlfunc's categories "Input and output functions", "Functions for fixed-length data or records" and "Functions for filehandles, files, or directories" (Pod::Functions kinds `I/O`, `Binary`, `File`), intersected with CORE.pmt's prototyped lines and without names the string, numeric, regex, array, list and hash batches hold. A name perlfunc lists under several categories belongs to the first of the three remaining batches (I/O and files; processes, time and the rest; system and flow) that holds one of its categories. Measured on 5.42 that is 47 lines: `vec`; `read`, `syscall`, `sysread`, `sysseek`, `syswrite`; `chdir`, `chmod`, `chown`, `chroot`, `fcntl`, `glob`, `ioctl`, `link`, `lstat`, `mkdir`, `open`, `opendir`, `readlink`, `rename`, `rmdir`, `stat`, `symlink`, `sysopen`, `umask`, `unlink`, `utime`; `die`; `binmode`, `close`, `closedir`, `eof`, `fileno`, `flock`, `getc`, `readdir`, `readline`, `rewinddir`, `seek`, `seekdir`, `tell`, `telldir`, `truncate`, `warn`, `write`; `dbmclose`, `dbmopen` (`die` and `warn` were typed with the measured rows). Most take `*`, `$` or `;$`; a `*` slot is `Glob *` (RFC "A typed signature and a prototype say the same thing"), which is a bareword or scalar, not the aliased glob slot open question 11 leaves unspelled. Every type is measured on perl 5.42 and recorded as a comment above its line; element types the paper has not defined are written as plain `List`. Hold the batch's names in the test, computed once with the command above, never against a hand-kept table. Re-marks no RFC section.

## Steps

- [ ] Measure the batch's names with Pod::Functions and perl; write failing test `TestCoreIOFileBuiltinsTyped` in `internal/parse/coretypes_test.go`: every name in the batch has a typed CORE.pmt line
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreIOFileBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines, category by category, with their measurement comments; run the harness after each category; `make test`; commit per category

### Negative Scenarios

- a `*` slot takes a bareword handle or any scalar, measured (`binmode STDOUT`, `binmode $fh`), so its typed parameter derives `*` and never `\*` or `$` (`go test ./internal/parse/ -run '^TestCoreStarSlotDerivesStar$' -count=1 -v | grep -q '^--- PASS: TestCoreStarSlotDerivesStar'`)
- an all-optional line in the batch (`getc`, `eof`: prototype `;*`) is typed with an optional parameter, not a required one, so its derived prototype keeps the `;` (`go test ./internal/parse/ -run '^TestCoreOptionalHandleKeepsSemicolon$' -count=1 -v | grep -q '^--- PASS: TestCoreOptionalHandleKeepsSemicolon'`)

## Acceptance criteria

- [ ] CORE.pmt types every I/O, fixed-length data, file and directory builtin (`go test ./internal/parse/ -run '^TestCoreIOFileBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreIOFileBuiltinsTyped'`)
- [ ] every CORE.pmt prototype, derived from its types where typed, is still perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] a star slot derives a star (`go test ./internal/parse/ -run '^TestCoreStarSlotDerivesStar$' -count=1 -v | grep -q '^--- PASS: TestCoreStarSlotDerivesStar'`)
- [ ] an optional handle keeps its semicolon (`go test ./internal/parse/ -run '^TestCoreOptionalHandleKeepsSemicolon$' -count=1 -v | grep -q '^--- PASS: TestCoreOptionalHandleKeepsSemicolon'`)
