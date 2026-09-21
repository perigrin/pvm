---
title: The lint cannot see inside a sub, so five tier-07 ops and methstart are unclaimable
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:00.587037252Z
updated: 2026-09-21T18:43:00.587037252Z
---

`opsOf` in `internal/conformance/lint.go:197` runs

    exec.Command("perl", "-MO=Concise,-exec", path)

with no sub named, so B::Concise prints the MAIN program's optree only. A
subroutine body is a separate CV and needs `-MO=Concise,-exec,<subname>`.

Two tiers hit this independently while writing their corpus files at
0ce515cb:

- **Tier 07** dropped from seven claimed ops to two. `argcheck`,
  `argdefelem`, `argelem`, `leavesub` and `return` are all really emitted
  by `conformance/07_subroutines/*.t` and all live in a CV. Its adjacency
  file is the sharpest case: source carrying a signature, a parameter
  default, a `for` loop, a `next`, three `return`s and an ampersand call
  compiles a main optree containing exactly `anoncode` and `entersub`.
- **Tier 11** dropped `methstart`, which opens every `method` body under
  the `class` feature. No file can be written that puts it where the lint
  looks.

Both READMEs record the real op streams in prose so the measurements are
not lost, and both say why the op is unclaimed.

The fix is to teach `opsOf` to enumerate the file's subs and dump each.
It is deliberately NOT done here: it changes the instrument for all
fourteen tiers at once, and several tiers' declared sets were measured
against the current behaviour. It wants measuring before being assumed an
improvement -- specifically, whether widening the visible op set causes
any tier to start emitting an op claimed by a LATER tier, which is the
direction the lint exists to catch.
