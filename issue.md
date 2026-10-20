---
title: "Declarations the grammar cannot yet state: an optional invocant, a prototype with an invocant, a second package; and Net::Ping's OO interface"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T23:40:03.614353199Z
updated: 2026-10-20T23:40:03.614353199Z
---

Deferred from the IO and OO audit fixes (01a15ac0-6ca2, 2026-10-21), each documented in its .tpm comment; this is their destination.

- IO::Select::select: its invocant is conditionally absent (`shift if defined $_[0] && !ref($_[0])`); lib calls it both ways (IPC/Cmd.pm:1463 without a class). The one-invocant-slot declaration grammar cannot express it.
- IO::Handle/FileHandle `blocking`: its real prototype `($;$)` and an invocant colon cannot be stated together; left undeclared.
- File::Temp::Dir (dirname, newdir's return) and Tie::StdHash / Tie::ExtraHash: second packages inside File/Temp.pm and Tie/Hash.pm; a .tpm declares one package.
- Net::Ping's OO interface (new, ping, close, open, ack, nack, bind, ...): undeclared beyond pingecho/wakeonlan.

## Steps
- [ ] Decide (with perigrin) how a declaration states an optional invocant and a prototype alongside an invocant
- [ ] Decide whether a .tpm may declare a second package, or a second file is shipped for it
- [ ] Declare Net::Ping's OO interface, measured

## Acceptance criteria

### Positive Scenarios
- [ ] IO::Select::select is declared for both call shapes (`go test ./internal/infer/ -run '^TestIOSelectSelectBothShapes$' -count=1 -v | grep -q '^--- PASS: TestIOSelectSelectBothShapes'`)
- [ ] Tie::StdHash is declared (`go test ./internal/parse/ -run '^TestSecondPackageDeclared$' -count=1 -v | grep -q '^--- PASS: TestSecondPackageDeclared'`)

### Negative Scenarios
- [ ] A Net::Ping method perl does not define is refused (`go test ./internal/infer/ -run '^TestNetPingOOHeld$' -count=1 -v | grep -q '^--- PASS: TestNetPingOOHeld'`)
