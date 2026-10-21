---
title: Back out the :always mark until the export syntax is redesigned
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-21T17:04:58.342574399Z
updated: 2026-10-21T17:28:15.907300256Z
sessions:
- start_sha: 377d3518b97962c50d9b4ddc091ba26bb5cc037c
  end_sha: 692a560c81fe680ed09244599f0e1bd216c5a075
  commits: 3
  started_at: 2026-10-21T17:05:27.183922622Z
  ended_at: 2026-10-21T17:28:15.907300256Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T17:05:27.183922622Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T17:28:15.907300256Z
observed_paths:
- .github/workflows/ci.yml
- docs/rfcs/0001-type-libraries.md
- internal/conformance/perl.go
- internal/data/help/types.md
- internal/infer/declared_test.go
- internal/infer/parametertypes_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/oracle/Cwd.pl
- internal/parse/declarations.go
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/Getopt/Long.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/Moose.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/exportattr_test.go
- internal/parse/exports.go
- internal/parse/heldtoperl_test.go
- internal/parse/use.go
- internal/perltest/constantdifferences.go
- internal/perltest/fixture.go
- internal/perltest/missingsubs.go
- internal/perltest/perltest.go
- internal/perltest/staleexempt.go
- internal/perltest/staleexempt_test.go
---

perigrin, 2026-10-21: back out the `:always` mark (added by 01a15efc for Moose's meta) now; the export syntax is to be redesigned (01a15f6b). Moose returns to its plain declared import, `sub import (Str $class: List @args) Void;` as 01a15ac0-6bfd shipped it, accepting the known over-install on a named list (`use Moose qw(has); blessed();` false call_arity) until 01a15f6b lands. The reader for perl 5.44 named parameters and the RFC's two rulings on it stay.

## Steps
- [ ] Remove `:always` from the declaration language (declarations.go readExportMarks/declarationSubAttrs/alwaysMark/alwaysAttr, exports.go moduleFacts.always, use.go's always-install loop) and from the RFC
- [ ] Restore Moose.tpm's import and meta's mark as 01a15ac0-6bfd shipped them
- [ ] Keep `Param.Named` and the named-parameter reading, and resolveNamedImport only if a shipped .tpm still uses it; otherwise remove that branch too and say so
- [ ] psc check over perl 5.42's lib adds no unlisted line

## Acceptance criteria

### Positive Scenarios
- [ ] No declaration uses :always (`! grep -rq ':always' internal/parse/declarations`)
- [ ] The parser no longer reads :always (`go test ./internal/parse/ -run '^TestAlwaysMarkRefused$' -count=1 -v | grep -q '^--- PASS: TestAlwaysMarkRefused'`)
- [ ] A .tpm still reads named parameters (`go test ./internal/parse/ -run '^TestDeclarationReadsNamedParameters$' -count=1 -v | grep -q '^--- PASS: TestDeclarationReadsNamedParameters'`)
- [ ] -traits alone installs every default (`go test ./internal/parse/ -run '^TestMooseTraitsInstallsSugar$' -count=1 -v | grep -q '^--- PASS: TestMooseTraitsInstallsSugar'`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)

### Negative Scenarios
- [ ] The RFC no longer describes :always (`! grep -q ':always' docs/rfcs/0001-type-libraries.md`)
