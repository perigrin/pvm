---
title: "&M::f(...) and goto &M::f with a None signature are not exits"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-17T03:41:56.695763478Z
updated: 2026-10-17T03:41:56.695763478Z
---

internal/infer/flow.go:632, internal/lower/call.go:111-114: `&Carp::confess("x")` and bare `&confess("x")` lower to an Ampersand Call with no signatures and exits returns false for any Ampersand call; `goto &Carp::confess` is not an exit either. Perl 5.42 dies on all. & bypasses the prototype, not the sub, so a None return still makes it an exit; goto &sub never returns. Found by da9d's review gate (75, 70), 2026-10-17.
