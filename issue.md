---
title: package Foo v1.2.3 { } is refused, though perl accepts a v-string version
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T11:48:36.042127887Z
updated: 2026-10-16T11:48:36.042127887Z
---

## Context

- found by: the PR #486 review's S10 fix (2026-10-16)

`package Foo v1.2.3 { }` -- a package block whose version is a v-string -- is refused by internal/parse as `trailing_tokens`, so it lowers to a Hole. perl 5.42 accepts it and `Foo->VERSION` is `v1.2.3`. S10 keeps a numeric version in `Decl.Version`; a v-string version should land there too.

## Steps

- [ ] Failing test first; parse a v-string package version as perl does; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] a v-string package version parses and lowers to Decl.Version (`go test ./internal/lower/ -run '^TestPackageVStringVersion$' -count=1 -v | grep -q '^--- PASS: TestPackageVStringVersion'`)
