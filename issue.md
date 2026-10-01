---
title: "Declaration files for subs a module builds out of sight: Moose, Mojolicious::Lite"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-01T04:42:54.861459514Z
updated: 2026-10-01T05:20:48.486320452Z
sessions:
- start_sha: 555b4ff88e27fe5ed3f19052f9d50091492432ac
  end_sha: a70a294e499e5fca37b1ed8035f11e3db6698468
  commits: 2
  started_at: 2026-10-01T04:42:55.491133661Z
  ended_at: 2026-10-01T05:20:48.486320452Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-01T04:42:55.491133661Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-01T05:20:48.486320452Z
observed_paths:
- internal/parse/declarations.go
- internal/parse/declarations/Mojolicious/Lite.pmd
- internal/parse/declarations/Moose.pmd
- internal/parse/declarations/Moose/Role.pmd
- internal/parse/declarations_test.go
- internal/parse/module.go
- internal/parse/perlgitt_test.go
- internal/parse/ratchet_test.go
- internal/parse/session_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
---

First slice of type libraries (01a0f450-9ded-7c8a-8c8b-c04734752b2d): declaration files for subs. Moose and Moose::Role export through Moose::Exporter's setup_import_methods, and Mojolicious::Lite through a monkey_patch loop in its own import, so reading their source can't recover what they define. A declaration file states it instead: `our @EXPORT` and `sub NAME (PROTO);`, read by the module reader the resolver already has, in place of the module's source, whether or not the module is installed. Files are embedded under internal/parse/declarations/, `Moose::Role` at `Moose/Role.pmd`.

The Mojolicious files load the module with `BEGIN { eval { require M; M->import; 1 } or plan skip_all ... }`. An import on a literal class name in BEGIN is now applied as `use M LIST`; a computed invocant still leaves the sub table incomplete. Measured on 5.42.0: after `BEGIN { require Test::More; Test::More->import() }`, `foo $f;` deparses as `$f->foo`.

PerlOnJava unit/moose_role_constant_method.t:21, unit/mojolicious_skipped_closure_lifecycle_regression.t:22, unit/mojolicious_mount_lifecycle_regression.t:54.

## Acceptance criteria

- [ ] A declaration stands in for a module's source (`go test ./internal/parse/ -run '^TestDeclarationStandsInForSource$' -count=1 -v | grep -q '^--- PASS: TestDeclarationStandsInForSource'`)
- [ ] A literal BEGIN import is applied as use (`go test ./internal/parse/ -run '^TestBeginImportIsUse$' -count=1 -v | grep -q '^--- PASS: TestBeginImportIsUse'`)
