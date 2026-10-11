---
title: A caret control variable lexes as two tokens, so 179 files carry a wrong tree
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T18:35:44.781455309Z
updated: 2026-10-11T15:17:19.934232076Z
sessions:
- start_sha: 20c20c3c9f6b7772b719bd7affc1d50f419e1e75
  end_sha: b2084901525559c23fe0eef21f146e164b5c4cc9
  commits: 2
  started_at: 2026-09-26T18:36:10.630614992Z
  ended_at: 2026-09-26T19:40:26.614724955Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T18:36:10.630614992Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T19:40:26.614724955Z
observed_paths:
- conformance/mdtest/loops.md
- conformance/mdtest/names.md
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/caret_test.go
- internal/lexer/scan.go
- internal/lexer/testdata/lex.t.tokens
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/parse.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

`$^O` lexes as `Variable($^)` followed by a separate Word, so every caret
control variable is two tokens where perl has one. 179 of perl.git t/'s 620
files use one, and in most spellings the node count is ZERO -- the tree is
wrong and nothing reports it.

Found by tracing `op/local.t`'s first failure after the require and label fixes
took the goal to 318 of 620.

## Measured at `20c20c3c`

    source                              Unknown  last var token  canon
    my $x = $^O;                           0     Variable($^)    $^;O()
    my $r = $^O eq 'x';                    0     Variable($^)    $^;O() eq 'x'
    $s = 0 if ($^O eq 'VMS');              1     Variable($^)    ($^ O eq 'VMS')
    $s = 0 if ($^O);                       1     Variable($^)    ($^ O)
    $s = 0 if $^O eq 'VMS';                0     Variable($^)    $^ O() eq 'VMS'
    $s = 0 if ($o eq 'VMS');               0     Variable($o)    fine, the control

The last var token is `$^` in EVERY caret case, including the four that score
Unknown=0. Perl reads it as one variable:

    perl -MO=Deparse -e 'my $x = $^O;'   ->  my $x = $^O;

So `my $x = $^O;` produces a confident wrong tree: `$^` then a call to `O()`.
That is the fifth zero-Unknown-and-wrong finding today, and the first where the
count is zero for the MAJORITY of spellings -- only the parenthesised forms
refuse, which is why it survived the goal's bucketing passes.

## Reach

    179 of 620 files contain a caret variable (grep -lE '\$\^[A-Z]')

`$^O` dominates -- it is how perl's suite tests for VMS, cygwin, MSWin32 and
dec_osf. The three span buckets `O eq 'VMS')`, `O eq 'dec_osf')` and
`O eq 'MSWin32')` together were 49 nodes with 25 of them a file's FIRST
Unknown, which is how this was found.

I first probed `$^O` in isolation, saw Unknown=0 on four of six spellings, and
concluded it was NOT a cause. That was wrong, and the reason is the thing worth
recording: an Unknown count cannot see a mis-lex that still yields a parseable
tree. The token stream had `Variable($^)` the whole time.

## What already works, so the fix is narrow

`${^TAINT}` is already correct -- `bracedNameFollowsAt` has an explicit caret
branch (`internal/lexer/scan.go`, the `l.src[i] == '^'` case) added because
`${^TEST}` splitting was caught by `TestLexDotTGoldenStream`. So the BRACED
spelling was fixed and the BARE one was not.

That is the same shape as the trivia set diverging in four copies
(`01a0dd25`-era work): one spelling of a rule was repaired and its sibling was
left, because nothing tied them together.

## Scope

1. `scanVarName` should take `^` plus one following identifier character as the
   whole name when the sigil is `$`. Measure the boundary first: `$^` alone is a
   real punctuation variable (the format top-of-page name), so a bare `$^` with
   no letter after it must stay one token.
2. Measure which letters are legal. perlvar lists `$^A` through `$^X` plus
   `$^_`, and `${^WORD}` for the long names -- check whether a lowercase letter
   after `$^` is a caret variable or `$^` followed by a word.
3. A corpus case. `$^O` is tier 02's territory (a punctuation variable) and the
   case should assert the TOKEN as well as the output, since the output alone
   passes today for `my $x = $^O;`.
4. Re-measure perl.git t/ after: the 49 nodes in the three `$^O` buckets are the
   visible part, but 179 files carry the construct and most are scoring clean on
   a wrong tree.

## Context

- paths: `internal/lexer/scan.go` (`scanVarName`, `bracedNameFollowsAt`'s caret
  branch at the `l.src[i] == '^'` case), `internal/lexer/scan.go`'s
  `scanVariable`
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`
- related: `01a0deeb` (no ratchet gates the t/ goal, so this cannot regress
  visibly), `01a0de11` (lintOps sees op names, not flags -- the same
  instrument-blindness in a different layer)

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] `$^O` lexes as ONE Variable token (`go test ./internal/lexer/ -run '^TestCaretControlVariable$' -count=1 -v | grep -q '^--- PASS: TestCaretControlVariable'`)
- [ ] A bare `$^` with no letter after it is still one token (`go test ./internal/lexer/ -run '^TestCaretControlVariable$' -count=1 | grep -q '^ok'`)
- [ ] `${^TAINT}` still lexes as it does today (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
