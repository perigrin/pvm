---
title: Socket's getaddrinfo declares a Str element perl never returns
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T21:06:05.361634112Z
updated: 2026-10-21T21:06:05.361634112Z
---

crochet:review round 3 V16 (stale isdual comments); destination given by 01a16016-481d.

Socket.tpm and IO/Socket.tpm declare getaddrinfo's list `List[Str|HashRef|DualVar]`, keeping Str because the return-type oracle's observer once had no isdual check (01a15aaa). It has one now. Measured 2026-10-21 on fix/481d: with Str removed from Socket.tpm's getaddrinfo, the return-type oracle still passes every getaddrinfo case (TestConstantExportersReturnTypes), and with DualVar also removed it fails on all four with "returned DualVar". So no getaddrinfo element perl returns is a plain Str: the declaration is wider than perl. getnameinfo's host and service are Str and keep it.

## Steps
- [ ] Narrow getaddrinfo's element type to HashRef|DualVar in Socket.tpm and IO/Socket.tpm
- [ ] Run the r2probes/r3probes Socket probes against pu and round 2 before and after

## Acceptance criteria

### Positive Scenarios
- [ ] Socket's return types hold (`go test ./internal/infer/ -run '^TestConstantExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersReturnTypes'`)

### Negative Scenarios
- [ ] getaddrinfo declares no Str element (`! grep -h 'sub getaddrinfo' internal/parse/declarations/Socket.tpm internal/parse/declarations/IO/Socket.tpm | grep -q 'List\[Str|'`)
