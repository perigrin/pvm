---
title: "Declarations the grammar cannot yet state: an optional invocant, a prototype with an invocant, a second package; and Net::Ping's OO interface"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T23:40:03.614353199Z
updated: 2026-10-21T02:39:29.375874769Z
---

Deferred from the IO and OO audit fixes (01a15ac0-6ca2, 2026-10-21), each documented in its .tpm comment; this is their destination.

- IO::Select::select: its invocant is conditionally absent (`shift if defined $_[0] && !ref($_[0])`); lib calls it both ways (IPC/Cmd.pm:1463 without a class). The one-invocant-slot declaration grammar cannot express it.
- IO::Handle/FileHandle `blocking`: its real prototype `($;$)` and an invocant colon cannot be stated together; left undeclared.
- File::Temp::Dir (dirname, newdir's return) and Tie::StdHash / Tie::ExtraHash: second packages inside File/Temp.pm and Tie/Hash.pm; a .tpm declares one package.
- Net::Ping's OO interface (new, ping, close, open, ack, nack, bind, ...): undeclared beyond pingecho/wakeonlan.
- Symbol::qualify's Ref candidate return (01a15ac0-6ccc, 2026-10-21): measured, `ref qualify({})` is HASH and `ref qualify([])` is ARRAY, not GlobRef as declared -- the reference comes back exactly as given, Ref, not coerced. Left as GlobRef rather than corrected to Ref: IO/Handle.pm's own `tell qualify($_[0], caller) if ref($_[0])` only proves $_[0] is some reference, never which kind, so a Ref return there would make tell's own Glob-only parameter see a Ref psc cannot narrow further, though $_[0] is in fact always a GlobRef at that call site -- a new diagnostic perl -w does not confirm. Needs a ruling on how psc treats an argument whose union type only partly fits a parameter; IO/Handle.pm:516 is the example.

- Getopt::Long (01a15ac0-6bfd, perigrin 2026-10-21: leave as-is, follow up here): `use Getopt::Long qw(GetOptionsFromArray)` installs $REQUIRE_ORDER, $PERMUTE and $RETURN_IN_ORDER too, which perl does not; its import cuts the list at a literal ":config", which a declared import cannot state. Fix: a construct for "everything from this literal on is configuration".
- POSIX (01a15ac0-6bfd): `use POSIX qw(math_h)` (bareword, no colon) installs the math_h tag in perl, whose import rewrites it to ":math_h"; psc installs nothing.
- parameterExempt granularity (01a15ac0-6cfa): an entry exempts every argument of the sub, not the one argument its reason names (the stat_cando entry for $effective hid a first-argument misfit).

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
