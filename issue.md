---
title: "Module exports: a bare use imports @EXPORT_OK too, :DEFAULT is read as a name, @EXPORT without our is missed"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:44:30.046420853Z
updated: 2026-09-29T17:44:31.900569489Z
sessions:
- start_sha: 4aa05bb40fab4acab33c9adf4eaeb6384ccfa24c
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T17:44:31.900569489Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T17:44:31.900569489Z
---

Three gaps in reading a module's exports:

1. A bare `use Mod;` imported @EXPORT_OK as well as @EXPORT: the two were recorded as one list.
2. `:DEFAULT` in an import list was taken as a name, so `use File::Spec::Functions qw(:DEFAULT splitdir ...)` imported `:DEFAULT` and not @EXPORT.
3. An export list assigned without `our` -- `@EXPORT = qw(...)`, as Devel::Peek and lib/Benchmark.pm write it -- was not read at all.

Measured with perl 5.42, against a module with `@EXPORT = qw(alpha); @EXPORT_OK = qw(beta);`:

    $ perl -I. -e 'use My::Mod; print alpha(), " ", defined(&beta) ? "beta imported" : "beta not imported", "\n";
          package Q; use My::Mod qw(:DEFAULT beta); print alpha(), beta(), "\n";'
    a beta not imported
    ab

perl.git t/benchmark/gh7094-speed-up-keys-on-empty-hash.t (`use Benchmark; timethese ...`), t/op/chdir.t:28.

## Acceptance criteria

- [ ] A bare use imports @EXPORT, :DEFAULT names it, and a bare assignment is read (`go test ./internal/parse/ -run '^TestExporterDefaults$' -count=1 -v | grep -q '^--- PASS: TestExporterDefaults'`)
