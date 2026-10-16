---
title: A .tpm accepts a bare :prototype attribute, which perl refuses
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T23:16:11.407308243Z
updated: 2026-10-16T23:16:11.407308243Z
---

In a .tpm, `sub f :prototype;` (bare, no argument) is accepted by the attribute allowlist (internal/parse/declarations.go:285), while perl 5.42 refuses it: "Invalid CODE attribute: prototype". It declares nothing harmful. Found by 8375's review gate (confidence 80), 2026-10-16. Fix: accept `:prototype` only with an argument; a negative test.
