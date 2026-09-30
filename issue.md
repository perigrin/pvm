---
title: "An exported sub built out of sight refuses: assume it has no prototype"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-30T00:34:50.695819919Z
updated: 2026-09-30T00:34:50.695819919Z
---

A name a module exports but never declares is imported with its prototype unknown, so every parenless call to it refuses. perl.git t/op/chdir.t:188, `$ENV{$key} = catdir $Cwd, 'op';`: File::Spec::Functions lists catdir in @EXPORT and builds it with `*{$meth} = ...` in a loop over @EXPORT. perl 5.42 reports `prototype "main::catdir"` as undef, so the call is `catdir($Cwd, 'op')`.

perigrin decided (2026-09-30) that a sub a module exports but builds out of sight has no prototype. The exception is a module that loads C (XSLoader::load, a DynaLoader bootstrap) and whose .xs was not read. Its sub may be XS with a prototype, as List::Util's `first` is `&@`, so it stays unknown. The default applies only to names the module exports, because an import list also carries non-sub arguments (`use feature 'defer'`). Without that restriction op/defer.t went 0 -> 40 when `defer` became a sub.

## Acceptance criteria

- [ ] An exported-but-undeclared sub of a non-XS module has no prototype (`go test ./internal/parse/ -run '^TestUnseenExportHasNoPrototype$' -count=1 -v | grep -q '^--- PASS: TestUnseenExportHasNoPrototype'`)
