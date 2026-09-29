---
title: A trailing comma is a separator with no element, and the closer it swallowed made a wrong tree
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T22:01:56.90867962Z
updated: 2026-09-29T02:48:03.110341418Z
---

## Measured finding

A trailing comma before a closer was read as an infix operator missing its right
operand. The operand hunt reached the closer and `parseTerm` consumed it as
`not_a_term`.

Bucketing every dirty perl.git `t/` file's FIRST Unknown span at `2257e08e` by
the construct the refusal stumbled over makes this **the single largest cause in
the 620-file corpus: 27 files, 417 nodes**, against 10 files for the runner-up.
`class/construct.t`'s note in `t2_test.go` has named it since the day it was
first seen — "Untracked: nothing in the chain owns trailing commas in argument
lists" — and this is that owner.

The Unknown count was the cheap half of the damage. The expensive half is that
the swallowed closer was GONE from the construct that owned it, which then read
the next token as its own — a WRONG TREE that no Unknown count can see:

```
f($a,)                      canon: f($a , ))              spurious `)`
f($a, $b,)                  canon: f($a , $b , ))
is(index($a, "bar",  ), 5)  canon: is(index($a , "bar" , ) , 5))
foreach my $x ($a,) { 1 }   canon: foreach my $x ($a , ){1})
$h{a,}                      canon: $h{a() , }}
f(a => 1,)                  canon: f(a => 1 , ))
```

`foreach my $x ($a,) { 1 }` is the worst: the loop BODY became a hash subscript
on the comma expression. Verified against perl 5.42.0 — the trailing comma
yields no element in every position:

```
$ perl -MO=Deparse -e 'sub f{} f($a,); f($a,$b,); foreach my $x (1,) { print 1 }'
f $a;
f $a, $b;
foreach my $x (1) { print 1; }

$ perl -MO=Deparse -e 'my %h; my $x = $h{a,};'
my $x = $h{'a'};
```

### The fix

One guard at the top of `parseInfix` in `internal/parse/expr.go` — NOT in
`parseCallArgs`, because the call argument list is only one of three callers
that reach it (`parseSubscript` and the loop list are the others), and a guard
per caller is a guard forgotten at one of them:

```go
if text == "," || text == "=>" {
    if next, ok := p.peekAfter(tok); ok && next.Kind == lexer.CloseBracket {
        p.advanceTo(tok)
        return left
    }
}
```

`lexer.CloseBracket` is one kind for `)`, `]` and `}`, which is exactly the three
closers the measurement found. The comma is CONSUMED before returning so the
closer is what the caller peeks at next; returning without consuming would spin,
because the comma still binds above the caller's `minBP`. `f(,)` — a comma with
no element in front of it — never reaches the loop at all, so that refusal is
untouched and stays correct.

### Measured effect (this change alone, isolated from concurrent work)

| ratchet | before | after | movers |
|---|---|---|---|
| `t1.ratchet` | 587 clean (59.5%), 1773 nodes | **679 clean (68.9%), 1456 nodes** | 138 better, 1 worse |
| `perlgitt.ratchet` | 391 clean (63.1%), 2822 nodes | **406 clean (65.5%), 2211 nodes** | 61 better, 3 worse |
| `t1canon.ratchet` | 758 clean (76.9%), 228 mismatches | **767 clean (77.8%), 219 mismatches** | 9 better, 0 worse |

All nine `t1canon` movers are 1 → 0 with no risers — direct evidence the
wrong-tree corruption is gone. `t1fidelity.ratchet` does not move.

Two further gates move and are HAND-MAINTAINED pins rather than regenerated
files, so they need source edits in the same commit:

- `TestT2CoreParses`, `t2_test.go`: `comp/parser_run.t` now parses cleanly and
  must be REMOVED from the shortfall map; `class/field.t` goes 2 → 1 and its
  entry must be updated. T2 core reaches 29 of 56 clean (51.8%).
- `TestClassCorpusRatchet`, `use_test.go:395`: `const want = 14` becomes `13`.

`perlgitt` moves 61 files better and 3 worse. **Each riser is explained and none
is a regression** — all three are the documented "A COUNT IS NOT A SIZE" effect,
the parser reaching bytes it previously never attempted:

- `mro/package_aliases.t` 24 → 37 and `mro/package_aliases_utf8.t` 23 → 36. The
  `for(...)` list at the top of each ends `},\n)`. With the `)` swallowed the
  loop BODY was a hash subscript and its contents never parsed as statements;
  now the body is reached and its five `fresh_perl_is q~...~` calls refuse
  individually. Proved independent of this change: `for ({name=>'a'}) {...}`
  with NO trailing comma scores the same 2 Unknowns, and the bare
  `fresh_perl_is q~x~, 'y', {}, 'z';` scores 2 on its own. After the fix the
  trailing-comma spelling parses IDENTICALLY to the no-comma spelling.
- `scalar_coderef_call_context.t` (T1) 2 → 4. Same shape: the
  `for my $case (...,)` list at lines 7-11 swallowed its `)`, so the loop body
  never parsed. Proved reach-only at the minimal case, where the fix REDUCES the
  count: `for my $c ([1],) { local $_ = $c->[0]; is scalar $f->($_), $c->[1], "x"; }`
  scores 2 before and 1 after — the same 1 as the no-comma spelling
  `for my $c ([1]) {...}` scores both before and after, and the same 1 that
  `is scalar $f->($_), $case->[1], "x";` scores alone with no comma anywhere.
  That residual 1 is a separate pre-existing defect in the parenless
  `is scalar $f->(...)` form.
- `porting/bench.t` 16 → 61. The `for (...)` list at line 384 ends `],` then
  comment lines then `)`. At base ONE 291-byte Unknown swallowed everything
  from line 548; now lines 384-548 are reached and refuse statement by
  statement. First-Unknown offset is 595 before AND after, so nothing regressed
  ahead of the newly-reached region.

The refused-BYTE total rises 167527 → 187519 for the same reason: bytes that
were never attempted were never counted as refused. Nodes fall by 611 and clean
files rise by 15, over strictly MORE source read.

### Why this is filed rather than committed

Another agent was working concurrently in the same physical worktree
(`.claude/worktrees/pu` — defect 01a0df51; `git worktree list` names the path but
does not reveal the sharing). During this session they modified `term.go`,
`keyword.go`, `call.go` and all four ratchets, reverted this change's `expr.go`
edit out of the shared tree once, and finished with their own commit STAGED
including a `perlgitt.ratchet` regenerated for their change. Committing would
have captured their staged `term.go` and `ampcall_test.go` under this message and
published a baseline nobody could attribute — the exact failure the ratchet
header comment forbids. The measurement was therefore taken in an isolated
`git archive` copy of `2257e08e`, and the patch is attached below for whoever
holds the tree next.

The patch is `internal/parse/expr.go` only. It does not touch `keyword.go`,
`call.go` or `term.go`, so it composes with the concurrent work — verified: the
test below passes against the shared tree carrying their change.

<details><summary>internal/parse/expr.go patch</summary>

Apply the guard shown above immediately after the `infix[text]` lookup and its
`op.BP <= minBP` early return, before the comparison-class check. Full diff and
the test file are in this session's scratchpad as `trailingcomma.patch` and
`trailingcomma_test.go`.

</details>

## Closing note, 2026-09-29

Landed in 652f4792 and never closed, which the chain review of this milestone
found. Closed with --force because the criteria below cannot be checked as
written: every command is in bare backticks, which git-zhi does not extract,
and the ratchet criteria pin the counts AT 652f4792 (406 clean, 2211 total),
which later fixes have moved on. The behavioural criteria were re-run on
ae2fb478 and pass: TestTrailingCommaBeforeCloserIsNotATerm (with its
only_a_comma and foreach_list cases), TestQuoteOp*, and
TestParenthesisedWordInfixParses. The follow-on for a comma before anything
else that cannot start an element is 01a0eb0a-ce6a-7065-826f-1ac6d2adf2e9.

## Acceptance criteria

- [ ] every trailing-comma spelling scores 0 Unknown and canon's the source without a spurious closer: call args, a named list operator, the loop list, a hash subscript, and the fat comma `go test ./internal/parse -run TestTrailingCommaBeforeCloserIsNotATerm -count=1`
- [ ] `foreach my $x ($a,) { 1 }` canon's with `{1;}` as the loop BODY and not as a subscript, which is the wrong-tree half an Unknown count cannot see `go test ./internal/parse -run TestTrailingCommaBeforeCloserIsNotATerm/foreach_list -count=1`
- [ ] the perl.git `t/` ratchet reads 406 clean and 2211 total, regenerated in the same commit, with each of the three risers explained in the commit message `go test ./internal/parse -run TestPerlGitTRatchet -count=1`
- [ ] the T1 ratchet reads 679 clean and 1456 total, with `scalar_coderef_call_context.t` 2 to 4 explained as reach `go test ./internal/parse -run TestParseRatchet -count=1`
- [ ] the canon ratchet reads 767 clean and 219 total, all nine movers 1 to 0 `go test ./internal/parse -run TestCanonRatchet -count=1`
- [ ] the two HAND-MAINTAINED pins are edited in the same commit: `comp/parser_run.t` removed from `t2_test.go`'s shortfall map, `class/field.t` set to 1 there, and `use_test.go`'s `const want = 14` set to 13 `go test ./internal/parse -run 'TestT2CoreParses|TestClassCorpusRatchet' -count=1`
- [ ] REGRESSION GUARD, passes today: `f(,)` — a comma with no element before it — still refuses, so the guard cannot turn a real syntax error into a silent empty list `go test ./internal/parse -run TestTrailingCommaBeforeCloserIsNotATerm/only_a_comma -count=1`
- [ ] REGRESSION GUARD, passes today: a regex quantifier `{2,}` is one Quote token and is untouched by a guard keyed on the comma, so `$f =~ s/(\\-){2,}/-+/g` keeps scoring 0 `go test ./internal/parse -run TestQuoteOp -count=1`
- [ ] REGRESSION GUARD, passes today: `and`/`or`/`xor` inside parens still bind above the comma, which is the other branch of the same loop `go test ./internal/parse -run TestParenthesisedWordInfixParses -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
