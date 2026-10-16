---
title: Tests import the system perl into the user's real pvm registry
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T16:44:37.591147775Z
updated: 2026-10-16T16:44:37.591147775Z
---

## Context

- found by: fixing PR #486's CI (2026-10-16); related to 01a136e9 (tests fill the real cache)
- paths: `internal/...` tests that auto-import the system perl; `.github/workflows/ci.yml` (runs `go test ./...` without isolating HOME)

Running `go test ./internal/... ./cmd/...` with a fresh HOME leaves `~/.local/share/pvm/registry.json` holding `{"version":"5.38.2","install_path":"/usr/bin","source":"system"}`: some test auto-imports the system perl into the REAL user registry. On developer machines that mutates the user's registry; on CI it made Linux pass PR #486's e2e tests by luck (ubuntu-24.04's /usr/bin/perl is 5.38.2, which test/e2e/.perl-version names). Find the test (internal/pm's Issue201 alone did not), isolate HOME/XDG for every test that can touch the registry, and add a guard.

## Steps

- [ ] Find the writer; isolate HOME/XDG in it and its siblings; a test that `go test ./internal/...` leaves the real registry untouched (run under a temp HOME and diff); `make test`; commit

## Acceptance criteria

- [ ] the test suite never writes the user's registry (`go test ./internal/testing/ -run '^TestSuiteLeavesUserRegistryAlone$' -count=1 -v | grep -q '^--- PASS: TestSuiteLeavesUserRegistryAlone'`)
