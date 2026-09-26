---
title: "A bare closer is not a cascade: use/require and bare return each swallow the closing brace of their block"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T21:56:52.175899257Z
updated: 2026-09-26T23:40:27.271157152Z
sessions:
- start_sha: f7df2cbd6938f54be1c0e5835425fa26aaa770e9
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T23:40:27.271157152Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T23:40:27.271157152Z
---

## Measured finding

Bucketing every dirty perl.git `t/` file's FIRST Unknown span at `2257e08e` (620
files, 391 clean, 2822 nodes) by whether the span BEGINS WITH A CLOSER gives 67
files. The hypothesis on the table was that these are cascade — that a span
starting with `)` or `}` is downstream of an earlier refusal that swallowed the
opener, so the real cause lies elsewhere in the file.

**That hypothesis is false.** Walking BACKWARD from each bare closer to its
MATCHING opener (a depth-counting scan, not a nearest-open-brace scan, which is
the artifact that produced a wrong attribution on an earlier probe) puts the
opener a median of 4 lines away, and in 22 of 67 files on the SAME LINE. The
construct at the opener is almost always ordinary — `{`, `for (...) {`,
`sub {`. These are not cascades from far away. The refusal is inside the pair,
and the closer is the token the refusal reached.

Classifying instead by the construct the refusal actually stumbled over:

| files | nodes | cause |
|------:|------:|-------|
| 27 | 417 | trailing comma before `)` |
| 10 |  77 | `use`/`no`/`require` last in block, no `;` |
|  7 | 107 | other, `trailing_tokens`, statement ended `;` |
|  6 | 207 | other, `not_a_term`, statement ended `;` |
|  4 | 106 | other, `not_a_term`, statement ended `}` |
|  4 |  61 | bare `return` last in block, no `;` |
|  3 |  13 | other, `trailing_tokens`, statement ended `}` |
|  6 |  78 | six singletons |

The top cause is now fixed (see Related). The two named runners-up are not, and
each is a WRONG TREE and not merely a count — an Unknown count cannot see
either, which is why both were isolated through canon rather than through the
ratchet.

### B. `use`/`no`/`require` last in a block with no `;` — 10 files, 77 nodes

`parseUse` takes the module bareword, then reads an import list whenever the
next token is not a `Semicolon`. At the end of a block the next token is `}`,
so the import-list `parseExpr` consumes it:

```
{ require Errno }     canon: {require Errno }; }
{ use Errno }         canon: {use Errno }; }
{ no warnings }       canon: {no warnings }; }
{ require Errno; }    canon: {require Errno; }      correct, with the `;`
```

The closer is swallowed into the `use` node and canon then emits a spurious
`};`. `require "foo.pl"` and `require $x` are unaffected: neither reaches the
bareword-name branch. Affected: `comp/use.t`, `op/filetest.t`, `op/inccode.t`,
`op/mkdir.t`, `op/sprintf.t`, `op/sselect.t`, `op/stat.t`, `op/taint.t`,
`op/threads.t`, `re/regexp.t`. The idiom is `eval { require Foo }` as a
feature probe, which is why it clusters in the platform-dependent tests.

### C. bare `return` last in a block with no `;` — 4 files, 61 nodes

`return` with no argument hunts for an operand and consumes the `}`:

```
sub f { return }      canon: sub f {return ; }
sub f { return 1 }    canon: sub f {return 1; }    correct, argument present
sub f { last }        canon: sub f {last; }        correct
sub f { next }        canon: sub f {next; }        correct
sub f { redo }        canon: sub f {redo; }        correct
```

`last`, `next` and `redo` already stop at the closer; only `return` does not, so
the fix is to give `return` the same terminator set they have rather than to
invent one. Affected: `io/pipe.t`, `op/attrs.t`, `op/coreamp.t`, `uni/attrs.t`.
`op/attrs.t` and `uni/attrs.t` are the same file in two encodings and both open
with `sub A::MODIFY_SCALAR_ATTRIBUTES { return }`, so the construct is the
FIRST failure in each.

Both are confirmed against perl 5.42.0:

```
$ perl -e 'sub f { return } eval { require Errno }; print "ok\n"'
ok
```

## Acceptance criteria

- [ ] `{ require Errno }` scores 0 Unknown and canon's `{require Errno; }` — no swallowed `}`, no spurious `};` — and the same holds for the `use` and `no` spellings `go test ./internal/parse -run TestUseLastInBlockWithoutSemicolon -count=1`
- [ ] `sub f { return }` scores 0 Unknown and canon's `sub f {return; }` `go test ./internal/parse -run TestBareReturnLastInBlock -count=1`
- [ ] REGRESSION GUARD, passes today: `last`, `next` and `redo` bare at the end of a block already score 0 and must keep doing so, so a fix for `return` that widens the shared terminator set cannot silently change them `go test ./internal/parse -run TestLoopLabels -count=1`
- [ ] REGRESSION GUARD, passes today: `require "foo.pl"` and `require $x` take the literal/expression path and are unaffected by any change to the bareword-name branch `go test ./internal/parse -run 'TestUse' -count=1`
- [ ] the perl.git `t/` ratchet is regenerated in the same commit and every mover is verified individually `go test ./internal/parse -run TestPerlGitTRatchet -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
