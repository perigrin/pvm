---
title: "Discuss type libraries: declaration files for what a module defines"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-30T21:55:19.14961642Z
updated: 2026-09-30T21:55:19.14961642Z
---

perigrin (2026-09-30): "eventually we should discuss type libraries similar to TypeScript". Some modules define their interface in ways no static reading of their source can recover:
- XS subs in an installed perl with no .xs (prototypes, packages);
- imports built by a module's own `import`: Mojolicious::Lite's `get`/`plugin` (monkey_patch into the caller), Moose::Role's `requires` (Moose::Exporter), Filter::Simple's `FILTER`;
- keyword plugins: Future::AsyncAwait's `async sub` / `await` syntax.

These are the last refusals in PerlOnJava T1: mojolicious_mount_lifecycle_regression.t, mojolicious_skipped_closure_lifecycle_regression.t, moose_role_constant_method.t, source_filter_subroutine_lexical_storage.t, custom_warning_async_suspend_state.t.

A declaration file per module, like TypeScript's `.d.ts`, could state what a `use` defines (names, prototypes, packages, keywords) for the parser to read instead of the module's code. This issue is the discussion; the design is open.

## Acceptance criteria

- [ ] A design for module declaration files is written and agreed (`sh -c 'ls docs/plans/*type-librar* >/dev/null'`)
