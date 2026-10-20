---
title: Narrow undef after ref, defined and blessed guards, so typed Scalar::Util and List::Util returns add no false positive
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T01:19:39.205726373Z
updated: 2026-10-20T02:37:16.577092645Z
sessions:
- start_sha: e80f4f920667615dc28874844b338224612fffe3
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T01:19:55.217000079Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T01:19:55.217000079Z
---

perigrin ruled on 2026-10-17 (issue dba9's gate): fix narrowing first. Declaring Scalar::Util, List::Util and Cwd's returns as `...|Undef` (which perl 5.42 measures) makes psc check report 76 new lib sites. dba9's reviewer read 26 of them: about 11 are real (undef reaches eq/=~ unguarded), about 13 are false positives because the code already rules undef out, 2 are unknowable. Goal (5) allows no added false positive, so dba9 cannot merge until psc sees through those guards.

## Context
- Branch from exec/dba9 (the .tpm files that expose the sites).
- The reviewer's classification, with file:line and reason for each site, is in dba9's issue (01a1464e-dba9) gate record; the sites are under perl 5.42's privlib.
- The guard shapes that must narrow, from that sample:
  - `ref($x)` truthy, including `return ... unless ref $x;`, `if (!ref($x)) {...} elsif ...`, and `ref($x) && reftype($x) eq 'CODE'` (short-circuit): after it, `$x` is a Ref, so a multi like reftype/refaddr/blessed selects its Ref candidate (Str/Int), not Undef.
  - `if (!$t) {...} elsif ($t eq 'HASH')`: the falsy branch consumes Undef.
  - A value assigned from reftype($item) after `$item` was proven a ref (threads/shared.pm:100-177).
- Perl decides each declaration: reftype, refaddr and blessed return a defined value for a reference argument and undef otherwise, so declare them as multis selected by argument type (`(Ref $r) Str` beside `(Any $x) Undef`, measured), rather than one `Str|Undef`. `List::Util::sum` with one or more arguments returns a number and with none returns undef: declare by arity.
- Selection by argument type and arity already exists (types.Select, da0a, 5700).

## Steps
- [ ] For each reviewer-classified site, a test fixture reproducing its guard shape; the GUARDED ones must give no diagnostic, the REAL ones must still be reported.
- [ ] Declare reftype, refaddr, blessed and sum (and any other dba9 sub with the same shape) as multis by argument type or arity, each measured with perl 5.42.
- [ ] Narrow a variable after `ref`, `defined` and `blessed` guards, short-circuit `&&`, `unless ... return`, and an if/elsif chain whose earlier branch consumes the falsy value.
- [ ] Re-run psc check over perl 5.42's lib on the dba9 tree: no line is added that the reviewer classified GUARDED; every remaining added line is in lib_true_positives.txt with perl -w output from the real lib site's path, not a minimal repro of the construct.
- [ ] Precision stays >= 65/76 with WRONG 0.

### Added 2026-10-20: Symbol (from db4c)

db4c's worker parked Symbol.tpm (scratchpad/db4c-symbol-deferred/) because measuring `Symbol::qualify` accurately (Str for a name, the argument unchanged for a ref: Str|GlobRef on perl 5.42) adds `IO/Handle.pm:515`, `tell qualify($_[0], caller) if ref($_[0])`, as a new psc diagnostic ("argument 1 expects Glob"). perl -w is silent there: `$_[0]` is the invocant, proven a ref by the `if ref($_[0])` guard. Same root cause as this issue: narrow `$_[0]` after the `ref` guard, and declare qualify / qualify_to_ref as multis by argument type (a Ref argument returns that Ref; a Str returns Str), measured. Add IO/Handle.pm:515 to TestNarrowGuardedLibSites; db4c ships Symbol.tpm after this lands.

### Added 2026-10-20: Encode::encode (from dbd3)

dbd3 declared `Encode::encode` as plain `Str` (commented, pointing here), though perl 5.42 measures `Str|Undef`, because the accurate type adds one guarded site, Pod/Text.pm's `encode($encoding, $text, $check)` (both arguments defined in practice). After the narrowing lands, restore `encode` to its measured `Str|Undef` (or a multi by argument type, if undef is returned only for an undef input) and confirm the Pod/Text.pm site stays silent. The declared type must match perl before this issue closes.

## Acceptance criteria

### Positive Scenarios
- [ ] Guarded sites give no diagnostic `go test ./internal/infer/ -run '^TestNarrowGuardedLibSites$' -count=1 -v | grep -q '^--- PASS: TestNarrowGuardedLibSites'`
- [ ] Real sites are still reported `go test ./internal/infer/ -run '^TestNarrowKeepsRealLibSites$' -count=1 -v | grep -q '^--- PASS: TestNarrowKeepsRealLibSites'`
- [ ] reftype of a proven reference is a Str `go test ./internal/infer/ -run '^TestReftypeOfRefIsStr$' -count=1 -v | grep -q '^--- PASS: TestReftypeOfRefIsStr'`
- [ ] sum with arguments is a number `go test ./internal/infer/ -run '^TestSumWithArgsIsNum$' -count=1 -v | grep -q '^--- PASS: TestSumWithArgsIsNum'`

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`
- [ ] Precision holds with WRONG 0 `go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`
