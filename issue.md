---
title: Thirteen T2 files still score WRONG, and it is not what the gate assumed
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-18T01:45:50.160422555Z
updated: 2026-09-18T01:46:10.75560066Z
---

Thirteen T2 files score WRONG through the subject contract, and the M1 gate cannot close until that is zero. The two easy causes are already fixed; what remains needs investigation rather than a known repair.

## Measured at 9669353d

    56 measured   33 exact (58.9%)   7 wider   13 WRONG   3 no-answer

    WRONG by marker:
      srefgen    7
      rv2hv      3
      anonhash   2
      match      1

Reproduce:

    PERL5_CORPUS=~/dev/perl5 PARSEORACLE_SHIM=/tmp/oracletree \
      go test ./internal/parseoracle -run TestGoParserWrongIsZero -count=1 -v -timeout 25m

## What this is NOT

Two hypotheses were checked and both are wrong. Recording them so the next
attempt does not start where this one did.

**Not class syntax.** The M1 gate's T2 note says class, field and method are
owned by no M1 issue, and three of the WRONG files are under `t/class/`. But
the syntax parses. Measured at 9669353d, every one of these is zero Unknown:

    class Point { field $x; }
    class Point;  field $x;
    class Point { field $x :param; }
    class Point { field $x = 0; }
    class Point { method zero { $x = 0 } }
    class Point { method set($v) { $x = $v } }
    class Point :isa(Base) { }
    field $x :reader;

The class files score WRONG on srefgen and anonhash, not on anything
class-shaped. The gate's note is stale on this point.

**Not the marker detection.** Every failing line decides correctly when
parsed on its own:

    t/class/field.t:329    ok(eq_hash({$obj->hash}, {}), '...')
                           -> decided=[anonhash anonhash]
    t/class/destruct.t:22  my $notifier = DestructionNotify->new( \$destroyed, 1 );
                           -> decided=[reference]

So the subject sees both anonymous hashes and the reference. Something about
how those sites are ATTRIBUTED in context is losing them.

## Where to look

`CompareFacts` matches perl's sites to the subject's by STATEMENT
(`compare_facts.go:210-232`): it finds the innermost statement covering
perl's reported line, then consumes one decided site per marker.

Two candidates, neither confirmed:

1. **Multi-line statements.** `Sites` reports one Line/EndLine pair per
   top-level statement (`sites.go:56`), taken from the statement's own span.
   A statement spanning lines 30-32 with perl reporting a site at line 31
   should match, but `innermost` may not treat the range the way this
   assumes.

2. **`method m { 1 }` splits into two statements.** Measured: it produces
   `[statement declaration] [statement term]` rather than one declaration
   with a block. If a method body becomes its own top-level statement, the
   sites inside it are attributed to the wrong statement. This is a real
   parser defect regardless of its effect here, and it is the same shape as
   `sub { 1 };` in issue 01a0afc1.

One file resists both explanations: `t/opbasic/concat.t:100` is
`*pi = \undef;`, and `\undef` decides reference correctly in isolation.

## Scope

Find out why the 13 score WRONG and fix what the finding says to fix. The fix
may be in the subject (as both previous ones were), in the parser, or in how
statements are reported to the harness. Do not tune the gate around them.

## Acceptance

- [ ] The cause of each of the 13 is named, in a comment or an issue (`go test ./internal/parseoracle -run TestGoParserWrongIsZero -count=1 -v -timeout 25m 2>&1 | grep -qE 'PASS|score WRONG'`)
- [ ] `method m { 1 }` is one statement holding a declaration and a block (`go test ./internal/parse/ -run '^TestMethodBodyIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestMethodBodyIsOneStatement'`)
- [ ] A statement spanning several lines attributes its sites correctly (`go test ./internal/parse/ -run '^TestSitesSpanMultilineStatements$' -count=1 -v | grep -q '^--- PASS: TestSitesSpanMultilineStatements'`)
- [ ] WRONG is zero across T2 (`PERL5_CORPUS=/home/perigrin/dev/perl5 PARSEORACLE_SHIM=/tmp/oracletree go test ./internal/parseoracle -run TestGoParserWrongIsZero -count=1 -v -timeout 25m | grep -q '^--- PASS: TestGoParserWrongIsZero'`)
- [ ] Exact does not fall below 58.9% (`PERL5_CORPUS=/home/perigrin/dev/perl5 PARSEORACLE_SHIM=/tmp/oracletree go test ./internal/parseoracle -run TestGoParserExactFloor -count=1 -v -timeout 25m | grep -q '^--- PASS: TestGoParserExactFloor'`)
- [ ] T1 does not move: 408 clean, 4,944 Unknown (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The last two are the negative scenarios: a change that fixes attribution by
making the subject quieter would move WRONG down and exact down with it, and
a change to statement boundaries could move the T1 numbers. Both must hold.

Blocks 01a0a70f (the M1 gate), whose last open criterion this is.

Spec: §7.1.2, §7.2. Plan: M1 metrics.
