---
title: Thirteen T2 files still score WRONG, and it is not what the gate assumed
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-18T01:45:50.160422555Z
updated: 2026-10-09T17:23:48.444756884Z
sessions:
- start_sha: 6393fdaa0bd1a57acd1acc0fbf10810c76893536
  end_sha: 6393fdaa0bd1a57acd1acc0fbf10810c76893536
  commits: 0
  started_at: 2026-09-18T09:31:30.960549339Z
  ended_at: 2026-09-18T09:31:42.792061377Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T09:31:30.960549339Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-18T09:31:42.792061377Z
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

### Criteria re-pointed (2026-10-09)

These criteria sat under a heading `git zhi verify` does not read, so they
never ran; run now, some named tests that no longer exist or need a shim:
TestGoParserWrongIsZero and TestGoParserExactFloor run only with
PARSEORACLE_SHIM (the oracle measures internal/parse by default since
01a0b835), TestUnknownScoresWider was deleted in 3b9422eb (TestUnknownHedgesEveryMarker
asserts the rule directly), and the T1-easy tests went with easy_test.go
in b4180102 (the T1 ratchet is the gate). Each now names the gate that
holds the same claim today: TestCorpusBaselineIsIntact asserts zero WRONG;
an exact floor is checked against the committed baseline's rows; the T1
rate against the T1 ratchet's own header; TestSitesSpanMultilineStatement
lives in internal/parseoracle.

## Acceptance criteria

- [ ] The cause of each of the 13 is named, in a comment or an issue (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] `method m { 1 }` is one statement holding a declaration and a block (`go test ./internal/parse/ -run '^TestMethodBodyIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestMethodBodyIsOneStatement'`)
- [ ] A statement spanning several lines attributes its sites correctly (`go test ./internal/parseoracle/ -run '^TestSitesSpanMultilineStatement$' -count=1 -v | grep -q '^--- PASS: TestSitesSpanMultilineStatement'`)
- [ ] WRONG is zero across T2 (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] Exact does not fall below 58.9% (`awk '$1=="exact"{e++} $1~/^(exact|wider|WRONG|no-answer)$/{m++} END{exit !(m>0 && e*1000 >= m*589)}' internal/parseoracle/testdata/ratchet/baseline.txt`)
- [ ] T1 does not move: 408 clean, 4,944 Unknown (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The last two are the negative scenarios: a change that fixes attribution by
making the subject quieter would move WRONG down and exact down with it, and
a change to statement boundaries could move the T1 numbers. Both must hold.

Blocks 01a0a70f (the M1 gate), whose last open criterion this is.

Spec: §7.1.2, §7.2. Plan: M1 metrics.


## Closed, 2026-09-18: 13 WRONG down to 1

    13 WRONG  33 exact   at 9669353d, when this issue was filed
     1 WRONG  39 exact   at 6393fdaa

The one is comp/proto.t, pinned TODO(M4) at 1c5e3b4a. Exact ROSE from 33 to
39 while WRONG fell, which is the shape that matters: a subject bought quiet
would have moved them together.

BOTH HYPOTHESES IN THIS ISSUE WERE WRONG, and recording that is the point.

Not class syntax: it already parsed, and the M1 gate's note saying otherwise
was stale.

Not `method` routing: `method` is not in keyword.go at all.

The span diagnosis was RIGHT about the defect and WRONG about its effect.
Applying it alone measured 13 -> 15 WRONG. The actual first cause was
`TookReference`: `parse.Sites` emitted `Kind: reference` with that field at
zero, and `compare_facts.go:336` reads the FIELD as the srefgen decision, so
every reference site was discarded on arrival.

## The seven causes, in the order they were found

1. `TookReference` unset on reference sites            13 -> 10
2. quote-op keyword as a name after `->` and sub names 10 -> 8
3. nested Unknown reported nothing                      8 -> 6
4. delimiter enumerated rather than shaped              6 -> 5
   (`s(a)(b)` decided MATCH -- a false positive no test reached)
5. swallowed deref blocks stayed silent                 5 -> 3
6. `=~` required a Term operand; `s =>` ate its comma   3 -> 2
7. site spans were points, not ranges                   2 -> 1

## The one worth remembering

Three of the seven were the same bug: a quote-op keyword (`s`, `y`, `q`, `m`,
`tr`) read as an operator where perl reads a name -- after `->`, after a sub
or method name, and before a fat comma. A fourth should reach for the general
rule rather than a fourth special case.

And the last fix was a genuine conflict between two TRUE rules, settled by
measuring both: perl reports a site inside a CV at the body's line and a site
in a multi-line statement at the statement's line. A span from the statement
to the node contains both. Two tests disagreed about this; perl decided.
