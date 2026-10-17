---
title: A module loaded only transitively gives its qualified calls no signatures
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-17T03:41:56.347791752Z
updated: 2026-10-17T03:41:56.347791752Z
---

internal/parse/use.go:309-318, module.go:338: each module is read by its own parser and only resolver facts are shared, so `use Tr;` where Tr does `use Carp ()` leaves the main parser's moduleSubs without Carp: `Carp::confess "x"` (perl 5.42 accepts and dies) gets no signatures and a diagnostic. Fix: add the qualified subs of every module in res.loaded to moduleSubs. Found by da9d's review gate (85), 2026-10-17.
