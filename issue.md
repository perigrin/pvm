---
title: A bareword before => is a Call node; Sites may report a call site perl does not have
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T17:44:56.635363313Z
updated: 2026-10-09T17:44:56.635363313Z
---

## Context

- paths: `internal/parse/term.go:621,757` (where `Fat` is set), `internal/parse/sites.go` (call sites; never reads `Fat`), `internal/parse/canon.go:175`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl -MO=Concise`

Seen 2026-10-09 while writing TestWrongTreesAreGone: a bareword before `=>` -- `{a=>1}`, `(a => 1)` -- is a `Call` node (`call a`) with `Fat` set; perl reads it as the constant string 'a' (no entersub). canon honours `Fat`; `Sites` never reads it, so it may report a call site perl does not have. The corpus oracle would not show this as WRONG: scoring ignores surplus decided sites. Unmeasured -- measure first: does Sites report a call site for `my %h = (a => 1);`, and does a sub `a` declared before change perl's reading (01a11ca3 / RFC 0001's `:bareword` on `=>` says no)?

## Steps

- [ ] Measure perl and Sites; failing test first; a fat-comma bareword reports no call site (and, if the tree should not call it a Call at all, decide that with the `:bareword` rule); `make test`; commit

## Acceptance criteria

- [ ] a bareword before `=>` reports no call site (`go test ./internal/parse/ -run '^TestFatCommaBarewordIsNotACallSite$' -count=1 -v | grep -q '^--- PASS: TestFatCommaBarewordIsNotACallSite'`)
