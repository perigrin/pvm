---
title: Declared imports read perl 5.44 named parameters as keys; Moose's named list installs those names and meta
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-21T15:02:46.98508217Z
updated: 2026-10-21T15:14:43.776040816Z
sessions:
- start_sha: 9711a47ad2aaa4e8463f48604119e9b4ed2bf964
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T15:14:43.776040816Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T15:14:43.776040816Z
---

crochet:review round 2 (quality S15, found again by 01a15ce6-1959): `use Moose qw(has)` installs all 13 Moose subs, giving a false call_arity on `blessed()` that pu and cf08c0ac do not have.

Measured (perl 5.42.0, Moose 2.4000): Moose::Exporter's import (Moose/Exporter.pm:449-458, _strip_option) removes exactly three named options, each with its value, from anywhere in the list -- -traits, -metaclass, -meta_name (which renames the installed meta). Everything left goes to Sub::Exporter, which reads a leading "-" as a group (`use Moose -foo => 1, qw(has)` dies: group "foo" is not exported). Installs: `use Moose` and `-traits => []` all 13 + meta; `-meta_name => 'm'` all 13 with meta as m; `qw(has)` has + meta; `-traits => [], qw(has)` has + meta; `()` nothing.

Rulings: perigrin, 2026-10-21 -- the signature names the keys, written with perl 5.44's named-parameter syntax (`:$name = default`, perlsub "named parameters"); and a declared import's named parameter `:$traits` matches the use-list key `-traits` or `traits` (perl 5.44 itself matches only the exact name; this is a .tpm rule, recorded in RFC 0001 beside the import-signature ruling).

## Steps
- [ ] Read perl 5.44 named parameters (`:$name`, optional with `=`, after positionals, before a slurpy) in a .tpm signature
- [ ] A declared import's named parameters are consumed from anywhere in the use-list as `-name => value` or `name => value`; the remaining names install the Exporter way (empty: the defaults), and a named key that alters installs (meta_name) does so
- [ ] Declare Moose's import: `sub import (Str $class: ArrayRef :$traits = [], Str :$metaclass = undef, Str :$meta_name = 'meta', List @names) Void;`
- [ ] Record both rulings in RFC 0001
- [ ] psc check over perl 5.42's lib adds no unlisted line

## Acceptance criteria

### Positive Scenarios
- [ ] A named list installs those names and meta (`go test ./internal/parse/ -run '^TestMooseNamedListInstallsNamesAndMeta$' -count=1 -v | grep -q '^--- PASS: TestMooseNamedListInstallsNamesAndMeta'`)
- [ ] -traits alone installs every default (`go test ./internal/parse/ -run '^TestMooseTraitsInstallsSugar$' -count=1 -v | grep -q '^--- PASS: TestMooseTraitsInstallsSugar'`)
- [ ] -meta_name renames meta (`go test ./internal/parse/ -run '^TestMooseMetaNameRenamesMeta$' -count=1 -v | grep -q '^--- PASS: TestMooseMetaNameRenamesMeta'`)
- [ ] A .tpm reads named parameters (`go test ./internal/parse/ -run '^TestDeclarationReadsNamedParameters$' -count=1 -v | grep -q '^--- PASS: TestDeclarationReadsNamedParameters'`)
- [ ] The RFC records the dash rule (`grep -q 'named parameter' docs/rfcs/0001-type-libraries.md`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)

### Negative Scenarios
- [ ] use Moose qw(has) then blessed() gives no false call_arity (`go test ./internal/infer/ -run '^TestMooseNamedListNoFalseArity$' -count=1 -v | grep -q '^--- PASS: TestMooseNamedListNoFalseArity'`)
