---
title: "Redesign the export syntax: a declaration derives or declares the module's import itself"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T17:04:27.601832636Z
updated: 2026-10-21T21:06:56.579569197Z
---

perigrin, 2026-10-21: "I think maybe we need to re-think this export syntax and possibly just derive/create the import function itself" -- then: back out :always now, redesign later.

Motivating cases the :export / :export_ok marks plus a declared import's signature cannot state:
- Moose: strips -traits/-metaclass/-meta_name from anywhere, installs the remaining names Exporter-style, and installs meta on every import renamed by -meta_name (measured perl 5.42.0, Moose 2.4000). Until this lands Moose over-installs on a named list: `use Moose qw(has); blessed();` gives a false call_arity that pu and cf08c0ac do not.
- Getopt::Long: truncates its list at a literal ":config"; `qw(GetOptionsFromArray)` also installs GetOptions (deferred, 01a15baf).
- POSIX: rewrites a bareword tag (`math_h`) to ":math_h" (01a15baf). Every `*_h` tag is affected, not only math_h: measured 2026-10-21, `use POSIX qw(fcntl_h); F_GETFD(1,2)` dies "Too many arguments for main::F_GETFD" under perl 5.42, and psc check (692a560c and fix/481d) reports nothing.
- IO::Socket::IP: its import forwards a named list to Socket's Exporter, so `use IO::Socket::IP qw(inet_aton)` installs inet_aton alone (measured 2026-10-21); IP.tpm's `sub import (Str $class: List @args) Void;` installs all 170 Socket names for any list. r3probes contract/iplist3-5.pl draw a false arity error at 692a560c that pu and 2db182e6 do not (round-3 review V10).
- Named parameters (`:$name`): a .tpm parses them (686b146c), but nothing past the parser's ordering check reads `Param.Named` -- arity, prototype derivation, `candidatesDifferInArity` and `alteringKey` treat one as positional, so a .tpm declaring one gets false call_arity errors (round-3 review V11/L2). Either the redesign reads them, or the parser refuses them until it does.
- Mojolicious::Lite, Test::More, Locale::Maketext::Simple: any list installs every default.

## Steps
- [ ] Brainstorm and amend RFC 0001: a .tpm derives or declares the import itself (what it installs, under which names, from which use-list shapes), replacing or subsuming :export/:export_ok and the declared-import signature keys
- [ ] Convert the 55 shipped declarations; held_to_perl's install check still holds every one
- [ ] Moose's named list installs its names and meta; Getopt::Long, POSIX cases fixed

## Acceptance criteria

### Positive Scenarios
- [ ] Moose's named list installs those names and meta (`go test ./internal/parse/ -run '^TestMooseNamedListInstallsNamesAndMeta$' -count=1 -v | grep -q '^--- PASS: TestMooseNamedListInstallsNamesAndMeta'`)
- [ ] Moose's -meta_name renames meta (`go test ./internal/parse/ -run '^TestMooseMetaNameRenamesMeta$' -count=1 -v | grep -q '^--- PASS: TestMooseMetaNameRenamesMeta'`)
- [ ] use Moose qw(has) then blessed() gives no false call_arity (`go test ./internal/infer/ -run '^TestMooseNamedListNoFalseArity$' -count=1 -v | grep -q '^--- PASS: TestMooseNamedListNoFalseArity'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
