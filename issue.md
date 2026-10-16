---
title: Read typed variables with :export and :export_ok in declaration files
state: in-progress
urgency: normal
milestone: core-module-types
blocks:
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
- 01a1464e-de34-7403-95d2-25f4855404b2
- 01a14657-682e-7432-8ea9-828f56cd2f69
- 01a14683-8375-7810-8a75-b772680603c6
- 01a1464e-dc07-754e-9ef4-637df15d51a9
created: 2026-10-16T20:02:15.387011621Z
updated: 2026-10-16T22:01:43.608215392Z
sessions:
- start_sha: 2fb0e62ffdc068b717c1b1a211c77e72d09ba51a
  end_sha: a9ce48f0047db49c938b19bdad67387cc27a0bd5
  commits: 2
  started_at: 2026-10-16T21:09:48.22713089Z
  ended_at: 2026-10-16T21:52:27.385557942Z
- start_sha: a9ce48f0047db49c938b19bdad67387cc27a0bd5
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T22:01:43.608215392Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T21:09:48.22713089Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T21:52:27.385557942Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-16T22:01:42.96008993Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T22:01:43.608215392Z
observed_paths:
- internal/infer/precision_test.go
- internal/lexer/proto.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/exports.go
- internal/parse/parse.go
- internal/parse/testdata/declarations/My/Exports.tpm
- internal/parse/testdata/declarations/My/Vars.tpm
- internal/parse/typedour_test.go
- internal/parseoracle/goparser_test.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/psc/lib_baseline_test.go
- internal/psc/testdata/lib_baseline.txt
- internal/psc/testdata/lib_true_positives.txt
---

## Prerequisites

- none. The baseline issue (01a1464e-d912) is not a prerequisite: no shipped `.tpm` uses the attributes until the conversion issue that follows this one.

## Context

- paths: `internal/parse/declarations.go`, `internal/parse/exports.go`, `internal/parse/typedour_test.go` (new), `internal/parse/exportattr_test.go` (new), `internal/parse/testdata/declarations/` (test fixtures), `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -run '^TestTypedOur|^TestSubExport' -count=1 -v`
- entrypoints: `readDeclarationWith` (`declarations.go`), `readExportTags` and `moduleFacts.exports`/`defaults` (`exports.go`), the typed reader that already refuses typed Perl outside a `.tpm`

RFC 0001 "Core-module libraries", "Every exported name" and "Typed variables" (perigrin, 2026-10-16). Every exported name in a `.tpm`, sub or variable, carries an attribute: `:export` means a bare `use` brings it in; `:export_ok` means it must be asked for; a tag is an attribute argument, `:export_ok(flock)` or `:export(tag, ...)`, and `use M ':flock'` brings in every name carrying that tag. Examples: `sub croak :export (List[Str|Boolean|DualVar|Ref] @message) None;`, `sub LOCK_EX :export_ok(flock) () Int;`, `our Hash[Str|Undef] %Config :export;`. This issue builds the ONE reader for those attributes, on subs and on typed variables alike, beside the existing export-list reader; the conversion issue that follows (01a14683-8375-7810-8a75-b772680603c6, Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm) converts the shipped files and then removes list reading from `.tpm`s. The conversion is a separate issue rather than merged into this one: this one is the reader and its tests over fixtures, that one touches every shipped declaration and the T1/T2 ratchets, and a reviewer can judge each alone. A typed `our` is typed Perl, legal only in a `.tpm`, as a typed signature is: perl refuses `our Hash[Str] %Config :export;` outside one at compile time (measured 2026-10-16 on 5.42.0: `No such class Hash ... syntax error`), so the parser refuses it there too. (A typed signature in a module's own source differs: perl reads `sub f (Ref $x) { }` as a prototype, so the parser does too.) The parser records the variable's declared type in a table beside `signatures`, for infer to read later; this issue does not make infer consume it. It blocks the variable-exporting declarations (Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat, Socket's `$CR $LF $CRLF`, POSIX's `%SIGRT`).

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

**Ruling 3 (perigrin, 2026-10-16):** `%Config`'s type is `Hash[Str|Undef]` (172 values are undef on 5.42).

## Steps

- [ ] Write failing test `TestSubExportAttribute`: a fixture `.tpm` with `sub f :export () Int;` and `sub g :export_ok () Int;` and no export list; `use M;` imports `f` and not `g`, `use M qw(g);` imports `g`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestSubExportAttribute$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestSubExportTag`: `sub LOCK_EX :export_ok(flock) () Int;` is imported by `use M ':flock'` and `use M qw(LOCK_EX)`, not by a bare `use M`; a name with `:export(a, b)` is in both tags
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestTypedOurExport`: a fixture `.tpm` with `our Hash[Str|Undef] %Config :export;`; `use M;` imports `%Config`, and the module's facts record its type `Hash[Str|Undef]`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestTypedOurExport$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestTypedOurExportOk`: `our Str $CRLF :export_ok;` is not imported by `use M;` and is by `use M qw($CRLF);`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestTypedOurOnlyInDeclarations`: the same statement in a module's own source is refused as a parse error, as perl refuses it at compile time (measured above)
- [ ] Implement; run to verify pass; commit
- [ ] Run the m1 gates named in the final-gate issue, with the commands it gives them; a ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin. `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-16, on a9ce48f0 (fresh reviewer subagent; lenses applied inline). VERDICT: REOPEN.

1. (F1, in scope, confidence 80) A list-form `our` in a .tpm silently drops its type and its export: `our Str ($ta, $tb) :export;` and `our ($pa, $pb) :export;` record no type, export nothing and raise no error (`declarations.go:247-249` looks only for a Term child; `decl.go:773` reaches declareVariable only for one variable). This breaks the rule at declarations.go:69-71 that an unread statement is an error, not a gap. Fix: when an `our` in a .tpm carries a type or an :export/:export_ok mark and declares a list, append the error "declare one variable per statement"; add a negative test. (perl itself dies "No such class Str" on `our Str ($a,$b);`.)
2. (F3, folded in because it is the same reader, confidence 65) A typed `our` in error still exports its name (`our Nope $x :export;`, `our Str %h :export;`), contradicting the comment at typedour_test.go:59-60 "a declaration in error records nothing". Fix: readExportMarks skips a name whose declaration is in error; extend the test to check exports.

## Acceptance Criteria

### Positive Scenarios
- [ ] `:export` and `:export_ok` on a sub in a `.tpm` decide what a bare `use` and `use M qw(name)` import (`go test ./internal/parse/ -run '^TestSubExportAttribute$' -count=1 -v | grep -q '^--- PASS: TestSubExportAttribute'`)
- [ ] an attribute argument is a tag: `:export_ok(flock)` is imported by `use M ':flock'` (`go test ./internal/parse/ -run '^TestSubExportTag$' -count=1 -v | grep -q '^--- PASS: TestSubExportTag'`)
- [ ] a bare `use` imports a typed `:export` variable, with its type (`go test ./internal/parse/ -run '^TestTypedOurExport$' -count=1 -v | grep -q '^--- PASS: TestTypedOurExport'`)
- [ ] a typed `:export_ok` variable is imported only when asked for (`go test ./internal/parse/ -run '^TestTypedOurExportOk$' -count=1 -v | grep -q '^--- PASS: TestTypedOurExportOk'`)
- [ ] a typed `our` outside a .tpm is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestTypedOurOnlyInDeclarations$' -count=1 -v | grep -q '^--- PASS: TestTypedOurOnlyInDeclarations'`)

### Negative Scenarios
- [ ] a typed `our` with no export attribute records its type but is imported by neither a bare `use M;` nor `use M qw(...)` naming it (`go test ./internal/parse/ -run '^TestTypedOurWithoutExportNotImported$' -count=1 -v | grep -q '^--- PASS: TestTypedOurWithoutExportNotImported'`)
- [ ] `use M ();` imports no `:export` variable (`go test ./internal/parse/ -run '^TestTypedOurEmptyImportListImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestTypedOurEmptyImportListImportsNothing'`)
- [ ] `use M ':nosuch'`, a tag no declared name carries, imports nothing from the declaration (`go test ./internal/parse/ -run '^TestSubExportUnknownTagImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestSubExportUnknownTagImportsNothing'`)
- [ ] a sub with no export attribute in a `.tpm` is imported by neither a bare `use` nor `use M qw(name)` (`go test ./internal/parse/ -run '^TestSubWithoutExportNotImported$' -count=1 -v | grep -q '^--- PASS: TestSubWithoutExportNotImported'`)
- [ ] an unknown attribute on a typed `our` in a `.tpm` (e.g. `:exported`) is an error, not ignored (`go test ./internal/parse/ -run '^TestTypedOurUnknownAttributeRejected$' -count=1 -v | grep -q '^--- PASS: TestTypedOurUnknownAttributeRejected'`)
- [ ] an unknown type name on a typed `our` in a `.tpm` is an error, as it is in a signature (`go test ./internal/parse/ -run '^TestTypedOurUnknownTypeRejected$' -count=1 -v | grep -q '^--- PASS: TestTypedOurUnknownTypeRejected'`)
- [ ] a hash or array `our` with a bare element type, such as `our Str %h :export;`, is refused, as a slurpy with a bare element type is (`go test ./internal/parse/ -run '^TestTypedOurAggregateWithBareTypeRejected$' -count=1 -v | grep -q '^--- PASS: TestTypedOurAggregateWithBareTypeRejected'`)
