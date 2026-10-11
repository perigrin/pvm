---
title: The sub table is keyed by spelling, so ok, main::ok and ::ok are three subs
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T20:13:05.035618513Z
updated: 2026-10-11T15:40:03.560711743Z
sessions:
- start_sha: 5ded07671843d54f646205a7c0ccd2efa7f728e1
  end_sha: b845d9b4e8940e6f86d5c0d3f74acc5510290083
  commits: 1
  started_at: 2026-09-29T02:07:45.457821878Z
  ended_at: 2026-09-29T02:16:05.204210182Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T02:07:45.457821878Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T02:16:05.204210182Z
observed_paths:
- internal/parse/decl.go
- internal/parse/shape.go
- internal/parse/subspelling_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
---

The in-file sub table is keyed by the LITERAL SPELLING of a name, so a
parenless call resolves only when the call site spells the callee exactly as
the declaration did. `ok`, `main::ok` and `::ok` are one sub in perl and three
distinct keys here.

Found while fixing the leading `::` (01a0de97-3c5d). That issue's lexer fix is
in and is not the cause: the token stream is correct now, and the remaining
refusal is at the table lookup.

## Measured 2026-09-26, with the leading-`::` lexer fix applied

Each program is the whole file. `Parse`, Unknown count and refusal code:

    sub ok {1} ok 1;                 U=0              canon "sub ok {1;} ok(1);"
    sub main::ok {1} main::ok 1;     U=0              canon "sub main::ok {1;} main::ok(1);"
    sub Foo::bar {1} Foo::bar 1;     U=0              canon "sub Foo::bar {1;} Foo::bar(1);"

    sub ok {1} main::ok 1;           U=1 trailing_tokens
    sub ok {1} ::ok 1;               U=1 trailing_tokens
    sub main::ok {1} ok 1;           U=1 trailing_tokens
    sub main::ok {1} ::ok 1;         U=1 trailing_tokens

The three that pass are the three where the two spellings are byte-identical.
Every cross-spelling pair refuses, in BOTH directions, and the declaration
being the qualified one does not help.

Valid perl 5.42.0 in every case:

    /home/perigrin/.local/bin/perl -e 'sub main::ok {print "A\n"} ::ok(1);'   -> A
    /home/perigrin/.local/bin/perl -MO=Deparse -e 'sub main::ok {1} ::ok(1);'
        sub ok { 1; }
        ok 1;

perl's own deparse collapses `main::ok` to `ok`, which is the canonical form
the table should key on.

## Where it is

`internal/parse/decl.go:322` `declareSub` stores `p.imports[name]` under the
name exactly as `declaredSub` read it out of the source. The call site looks up
the same map with whatever it was spelled. Nothing normalises either side.

    func (p *parser) declareSub(n *Node) {
        name, proto := declaredSub(n)
        ...
        p.imports[name] = Import{Name: name, ...}
    }

The normalisation perl applies: strip a leading `::`, then strip a leading
`main::`. Both spellings then land on the bare name, and a name qualified into
any OTHER package keeps its qualification.

## Reach

Not measured against perl.git t/ as a file count, because it needs the pair to
co-occur in one file. The construct is live in that corpus: `t/comp/colon.t`
and `t/opbasic/magic_phase.t` both declare their own `sub ok` and call it
without parens throughout, and 57 of the 620 files use a leading `::` call.
Scoping the count is part of the work.

### Criteria named (2026-10-11)

Measured at `971b9031`: the three cross-spelling pairs parse at Unknown=0 and
canon as resolved calls, and `sub Foo::bar {1} bar 1;` still refuses, as perl
5.42.0 does. The fix's test, TestSubNameSpellings, checks only Unknown counts
and uses `1, 2` rather than the issue's spellings. TestSubNameSpellingsResolve
is new in internal/parse/subspelling_test.go and asserts the tree shape and
canon of the issue's six spellings and the refusal of the seventh; it fails
with `subKey` returning the name unchanged, and fails on the negative when
`subKey` strips every package.

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Every criterion below ran the whole parse
package, the whole conformance package or `go test ./...`, which only restates
"make test passed" -- and make test gates every commit. Each claim now names
the test that holds it. Claims are reworded only to keep parentheses out of
the claim text, since git-zhi runs the first paren span on the line.

The two regression guards keep their claims and name gate tests instead:
TestCorpusRatchet for the corpus ratchet, and TestParseRatchet and
TestPerlGitTRatchet -- the T1 and perl.git t/ parse ratchets -- for the suite.

## Acceptance criteria

- [ ] `sub ok {1} ::ok 1;` parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubNameSpellingsResolve$' -count=1 -v | grep -q '^--- PASS: TestSubNameSpellingsResolve'`)
- [ ] `sub main::ok {1} ok 1;` parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubNameSpellingsResolve$' -count=1 -v | grep -q '^--- PASS: TestSubNameSpellingsResolve'`)
- [ ] `sub ok {1} main::ok 1;` parses with zero Unknown (`go test ./internal/parse/ -run '^TestSubNameSpellingsResolve$' -count=1 -v | grep -q '^--- PASS: TestSubNameSpellingsResolve'`)
- [ ] A sub qualified into another package does NOT collapse: `sub Foo::bar {1} bar 1;` still refuses (`go test ./internal/parse/ -run '^TestSubNameSpellingsResolve$' -count=1 -v | grep -q '^--- PASS: TestSubNameSpellingsResolve'`)
- [ ] REGRESSION GUARD: the three byte-identical pairs still parse clean, `sub ok {1} ok 1;` among them (`go test ./internal/parse/ -run '^TestSubNameSpellingsResolve$' -count=1 -v | grep -q '^--- PASS: TestSubNameSpellingsResolve'`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds, named as the T1 and perl.git t/ parse ratchets (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet' && go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
