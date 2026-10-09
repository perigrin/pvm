---
title: The lexer splits -e into a minus and a word, and the whole file-test family with it
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ce57-db92-78e5-bbe4-7c28e74db6f3
created: 2026-09-23T17:51:06.293983513Z
updated: 2026-10-09T17:13:24.366373733Z
sessions:
- start_sha: 6ae2242dc5addd89075b7c29d0a0777e69d757ef
  end_sha: 3ef0ae208b19053e94b91fbac2d6ddf19b4f4af4
  commits: 4
  started_at: 2026-09-26T09:53:36.362796482Z
  ended_at: 2026-09-26T11:06:23.503008191Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T09:53:36.362796482Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T11:06:23.503008191Z
observed_paths:
- conformance/mdtest/adjacency-06_control.md
- conformance/mdtest/argument-extent.md
- conformance/mdtest/unary.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/lexer/testdata/lex.t.tokens
- internal/parse/call.go
- internal/parse/decl.go
- internal/parse/expr.go
- internal/parse/fidelity.go
- internal/parse/infile_test.go
- internal/parse/keyword.go
- internal/parse/parse.go
- internal/parse/shape.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parseoracle/parenless_test.go
---

`-e $f` is a file-test operator and our lexer splits it into a minus and
a word. Found by the conformance corpus while writing tier 04's unary
slice (issue 01a0cc04-341c).

## Measured

    my $f = "/etc/hostname"; print "[", (-e $f), "]\n";

      Operator(-) Word(e) Variable($f)

Perl lexes `-e` as ONE file-test operator; we produce two tokens.

## Why it matters beyond this one operator

perlop's level 10 holds the whole family -- `-e -d -f -s -z -r -w -x -M
-A -C` and more -- so this is one rule covering a dozen spellings.

The minus is the ambiguity. A single `-` has at least three readings
decided by what follows it:

    -e $f        file test
    -$x          arithmetic negation
    -bareword    the STRING "-bareword", which tier 01's
                 07_negative.t already measures as two tokens

A lexer that leaves a bare `-` in front of a word has not decided which
of those it meant, and the decision is exactly what the level exists for.

## Where it is recorded

`conformance/04_operators/31_file_test.t` carries the refusal with a
token fact asserting what a correct lexer produces: `no operator whose
text is "-"`, because the whole `-e` should be the operator's name.

## Criteria added after the fact (2026-10-09)

This issue was closed (merge 3ef0ae20) with no criteria section, so
`git zhi verify` checked nothing for it. Each criterion names the test the
fix landed. The conformance `.t` file named above was replaced by the
mdtest corpus, which `TestCorpusRatchet` holds.

## Acceptance criteria

- [x] a file test is one operator, its minus part of the name (`go test ./internal/lexer/ -run '^TestFileTestIsOneOperator$' -count=1 -v | grep -q '^--- PASS: TestFileTestIsOneOperator'`)
- [x] the file-test family parses as unary operators (`go test ./internal/parse/ -run '^TestFileTestOperators$' -count=1 -v | grep -q '^--- PASS: TestFileTestOperators'`)
