---
title: After require M with no import, M::f(...) carries no signatures
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-17T03:41:56.166246042Z
updated: 2026-10-17T03:41:56.166246042Z
---

internal/parse/use.go:174-177,309-318: `require M` never reaches resolveImports, so moduleSubs stays empty and a parenthesised `Carp::confess("x")` after `require Carp` carries no signatures; perl 5.42 dies at runtime on it (the parenless form is a compile error), psc reports a spurious coercion-mismatch. moduleSubs is the compile-time shape table, so record signatures for the parenthesised call without making the parenless shape known (a separate table or a flag on the Import). Found by da9d's review gate (85), 2026-10-17.
