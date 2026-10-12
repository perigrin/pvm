---
title: "Lower calls: five sites to one Call, and the filehandle slot from the Handle flag"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c111-8c3e-7451-97d9-10307441ba2d
blocks:
- 01a12a85-32cf-7d76-96f7-b814935374e1
created: 2026-09-20T23:07:00.666688859Z
updated: 2026-10-12T19:03:34.878590183Z
sessions:
- start_sha: 93874fbfe9d9e6c38b8b145da0b41e96452b674a
  end_sha: ""
  commits: 0
  started_at: 2026-10-12T19:03:34.878590183Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T19:03:34.878590183Z
---

## Context

- paths: `internal/lower/call.go` (new), `internal/lower/call_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Calls"; §4.8.1, §4.8.3

This slice retires a measured consumer defect and collapses five duplicated
consumer sites into one. Shapes below were measured at 24bfac4a; re-measure
at execution time. Line numbers in `internal/infer` drift.

**Arguments arrive as one comma chain.** `f(1,2)` gives
`[call "f" [binary "," [term "1"] [term "2"]]]`, not two argument children,
and `print $fh, "x"` gives `[call "print" resolved [binary "," ...]]`. The
operators slice (01a0c111-8c3e) lowers a `,`/`=>` chain to one flat `List`;
`Call.Args` is that list's elements. That is why this issue depends on the
operators issue.

**The five sites.** §4.14.2 records that five places in `internal/infer` do the
same builtin lookup + arity + argument-type check and differ only in where
the name hides. At 24bfac4a the lookups are the `parse.CoreBuiltin` calls at
`infer.go:701` (`function_call_expression`,
`ambiguous_function_call_expression`), `:890` (`func1op_call_expression`),
`:1011` (`func0op_call_expression`), `:4198` and `:4242` (`function` or
`bareword` child). All five lower to one `Call{Name, Args, Resolved}`. That
duplication is the proof the erasure is VARIATION and not structure (§4.14.5).

**`Call.Handle` is no longer fragile.** §4.14.2 marks it "LOWERING by child
shape, fragile; a bit at `call.go:167` is the right fix". **That bit
shipped**: `parse.Node.Handle` (`parse.go:359`). `print $fh "x"` gives
`call "print"` over `term "$fh" handle` and `term "\"x\""`. So read the
flag; do not infer the slot from child count.

The defect this removes is measured and documented in the consumer itself:
the `indirect_object` comment at `infer.go:1144` records that counting the handle as argument 1 made every
typed-handle `print` a false `Str` mismatch, and it works around that by
skipping `indirect_object`. With `Handle` on the lowered node, `Call.Handle` is
a separate field from `Call.Args` and the workaround has no successor.

**`print {$fh} "x"` now parses**: it gives
`[call "print" resolved [block handle [statement [term "$fh"]]] [term "\"x\""]]`.
The handle is a block with `Handle` set. Lower `Call.Handle` to the block's
single expression when it has exactly one, and to the lowered block
otherwise.

Remaining rows:

- `Call.Parenthesized` LOWERING — `f(1,2)` gives a `call "f"` whose span
  includes the `(`; `map $_+1, @a` does not. §4.8.1: this affects the paren
  cliff and prototypes.
- `Call.Ampersand` / `.BareAmpersand` LOWERING — `&foo;` gives `term "&foo"`;
  `&foo();` gives `[index "(" [term "&foo"]]`. Bare iff there is no enclosing
  `index "("`. §4.8.3: `&foo` passes the caller's `@_`, which is why the bare
  form is STRUCTURE and not a spelling.
- `Call.Fn` (call through a reference) LOWERING for three shapes:
  `$code->(1)` gives `[binary "->" [term "$code"] [term "1" paren]]`,
  `&{$code}(1)` gives `[index "(" [index "{" [term "&"] [term "$code"]] [term "1"]]`,
  and **`&$code(1)` now parses** as `[index "(" [term "&$code"] [term "1"]]`.
- `MethodCall` LOWERING, four shapes: `$obj->method(1)` gives
  `[binary "->" [term "$obj"] [call "method" resolved [term "1"]]]`;
  `Foo->new` gives `[binary "->" [term "Foo"] [term "new"]]` (both sides are
  now `term`, not `call`); `$obj->$name(1)` gives
  `[binary "->" [term "$obj"] [index "(" [term "$name"] [term "1"]]]`; and
  the indirect form. `HasParens` comes from the `call` span. **`Indirect`
  now has a source**: `parse.Node.Indirect` (`parse.go:373`). With
  `package Foo; sub new {1} package main;` above it, `new Foo(1)` gives
  `[call "new" resolved indirect [term "Foo"] [term "1"]]`: invocant first,
  then the arguments. Lower it to a `MethodCall` with `Indirect` set, if the
  vocabulary issue kept that field. An undeclared `new Foo(1)` is still one
  Unknown and lowers to a `Hole` like any other.

`Call{Resolved}` is SHIPPED and is §4.14.1 rule 3's honest-uncertainty channel:
`Resolved:false` means "a call to something I have not seen", and PSC widens to
`Any` at exactly those points rather than guessing. The four non-call overloads
of that flag are handled by the literals/barewords issue.

**Hole tests use input perl rejects** (confirmed with `perl -c` on 5.42).
The earlier fragments `print {$fh} "x"` and `&$code(1)` parse now, and
`new Foo(1)` parses whenever `Foo` is declared, so none of them can test a
`Hole`. Use `foo(;`, which gives `[call "foo" [unknown refusal=not_a_term]]`,
and `Foo->new(;`, which gives the `->` over `call "new"` with an Unknown
argument.

## Steps

- [ ] Write `TestLowerCallHandleFromFlag`: `print $fh "x"`, `print STDERR "x"`
      and `say $fh "x"` lower to a `Call` with `Handle` set and `Args` holding
      only the real arguments; `print $fh, "x"` lowers with no `Handle` and two
      args; `print {$fh} "x"` lowers with `Handle` holding `$fh` and one arg.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerCallFlags`: `f(1,2)` sets `Parenthesized`, `map $_+1, @a`
      does not; `&foo;` sets both `Ampersand` and `BareAmpersand`; `&foo();`
      sets `Ampersand` only. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerCallThroughReference`: `$code->(1)`, `&{$code}(1)` and
      `&$code(1)` lower to `Call` with `Fn` set and `Name` empty. Run to
      verify failure, implement, run to green, commit.
- [ ] Write `TestLowerMethodCall`: the four shapes above lower to
      `MethodCall` with the right `Invocant`, `Name`/`Dynamic`, `HasParens`
      and `Indirect`. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerCallOverBrokenInput`: `foo(;` lowers to a `Call` named
      `foo` whose one argument is a `Hole`; `Foo->new(;` lowers to a
      `MethodCall` on `Foo` whose one argument is a `Hole`. No panic. Run to
      verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The filehandle slot comes from the Handle flag and is not an argument (`go test ./internal/lower/ -run '^TestLowerCallHandleFromFlag$' -count=1 -v | grep -q '^--- PASS: TestLowerCallHandleFromFlag'`)
- [ ] Parenthesized, Ampersand and BareAmpersand are derived correctly (`go test ./internal/lower/ -run '^TestLowerCallFlags$' -count=1 -v | grep -q '^--- PASS: TestLowerCallFlags'`)
- [ ] Calls through a code reference lower with Fn set (`go test ./internal/lower/ -run '^TestLowerCallThroughReference$' -count=1 -v | grep -q '^--- PASS: TestLowerCallThroughReference'`)
- [ ] All four method-call shapes lower to one MethodCall kind, the indirect one with Indirect set (`go test ./internal/lower/ -run '^TestLowerMethodCall$' -count=1 -v | grep -q '^--- PASS: TestLowerMethodCall'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A call or method call with an argument perl rejects lowers with a Hole argument and its name intact, and nothing panics (`go test ./internal/lower/ -run '^TestLowerCallOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerCallOverBrokenInput'`)
