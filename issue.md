---
title: "Lower calls: five sites to one Call, and the filehandle slot from the Handle flag"
state: pending
urgency: normal
milestone: m2-lower-the-tree
created: 2026-09-20T23:07:00.666688859Z
updated: 2026-09-20T23:07:00.666688859Z
---

## Context

- paths: `internal/lower/call.go` (new), `internal/lower/call_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Calls"; §4.8.1, §4.8.3

This slice retires a measured consumer defect and collapses five duplicated
consumer sites into one.

**The five sites.** §4.14.2 records that five places in `internal/infer` do the
same `GetBuiltin` + arity + argument-type check and differ only in where the
name hides: `infer.go:693` (`function_call_expression`,
`ambiguous_function_call_expression`), `:881` (`func1op_call_expression`),
`:1009` (`func0op_call_expression`), `:4004` and `:4048` (`function` or
`bareword` child). All five lower to one `Call{Name, Args, Resolved}`. That
duplication is the proof the erasure is VARIATION and not structure (§4.14.5).

**`Call.Handle` is no longer fragile.** §4.14.2 marks it "LOWERING by child
shape, fragile; a bit at `call.go:167` is the right fix". **That bit SHIPPED**:
`parse.Node.Handle` exists (`parse.go:299`) and is set at `call.go:143`.
Verified at HEAD: `print $fh "x"` gives `call "print"` over `term "$fh" handle`
and `term "\"x\""`. So read the flag; do not infer the slot from child count.

The defect this removes is measured and documented in the consumer itself:
`infer.go:1155-1165` records that counting the handle as argument 1 made every
typed-handle `print` a false `Str` mismatch, and it works around that by
skipping `indirect_object`. With `Handle` on the lowered node, `Call.Handle` is
a separate field from `Call.Args` and the workaround has no successor.

`print {$fh} "x"` is STILL `Unknown` at HEAD — `call.go:177` has the branch but
the lexer is not marking that `{` as `OpensBlock`. Emit a `Hole`; do not guess.

Remaining rows:

- `Call.Parenthesized` LOWERING — `f(1,2)` gives a `call "f"` whose span
  includes the `(`; `map $_+1, @a` does not. §4.8.1: this affects the paren
  cliff and prototypes.
- `Call.Ampersand` / `.BareAmpersand` LOWERING — `&foo;` gives `term "&foo"`;
  `&foo();` gives `[index "(" [term "&foo"]]`. Bare iff there is no enclosing
  `index "("`. §4.8.3: `&foo` passes the caller's `@_`, which is why the bare
  form is STRUCTURE and not a spelling.
- `Call.Fn` (call through a reference) LOWERING for two shapes, verified at
  HEAD: `$code->(1)` gives `[binary "->" [term "$code"] [term "1" paren]]` and
  `&{$code}(1)` gives `[index "(" [index "{" [term "&"] [term "$code"]] ...]`.
  `&$code(1);` is still `Unknown` — emit a `Hole`.
- `MethodCall` LOWERING, three shapes verified at HEAD: `$obj->method(1)` gives
  `[binary "->" [term "$obj"] [call "method" [term "1"]]]`; `Foo->new` gives
  `[binary "->" [call "Foo"] [call "new"]]`; `$obj->$name(1)` gives
  `[binary "->" [term "$obj"] [index "(" [term "$name"] [term "1"]]]`.
  `HasParens` comes from the `call` span. `Indirect`/`Ambiguous` stay PARSER —
  `new Foo(1)` is still `Unknown` at HEAD. Whether those two fields exist at
  all was settled by the vocabulary issue per §4.14.6; honour that decision.

`Call{Resolved}` is SHIPPED and is §4.14.1 rule 3's honest-uncertainty channel:
`Resolved:false` means "a call to something I have not seen", and PSC widens to
`Any` at exactly those points rather than guessing. The four non-call overloads
of that flag are handled by the literals/barewords issue.

## Steps

- [ ] Write `TestLowerCallHandleFromFlag`: `print $fh "x"`, `print STDERR "x"`
      and `say $fh "x"` lower to a `Call` with `Handle` set and `Args` holding
      only the real arguments; `print $fh, "x"` lowers with no `Handle` and two
      args; `print {$fh} "x"` lowers to a `Hole`.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerCallFlags`: `f(1,2)` sets `Parenthesized`, `map $_+1, @a`
      does not; `&foo;` sets both `Ampersand` and `BareAmpersand`; `&foo();`
      sets `Ampersand` only. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerCallThroughReference`: `$code->(1)` and `&{$code}(1)`
      lower to `Call` with `Fn` set and `Name` empty; `&$code(1)` lowers to a
      `Hole`. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerMethodCall`: the three shapes above lower to `MethodCall`
      with the right `Invocant`, `Name`/`Dynamic` and `HasParens`; `new Foo(1)`
      lowers to a `Hole`. Run to verify failure, implement, run to green,
      commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The filehandle slot comes from the Handle flag and is not an argument (`go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v | grep -q "^--- PASS: TestLowerCallHandleFromFlag"`)
- [ ] Parenthesized, Ampersand and BareAmpersand are derived correctly (`go test ./internal/lower/ -run "^TestLowerCallFlags$" -count=1 -v | grep -q "^--- PASS: TestLowerCallFlags"`)
- [ ] Calls through a code reference lower with Fn set (`go test ./internal/lower/ -run "^TestLowerCallThroughReference$" -count=1 -v | grep -q "^--- PASS: TestLowerCallThroughReference"`)
- [ ] All three method-call shapes lower to one MethodCall kind (`go test ./internal/lower/ -run "^TestLowerMethodCall$" -count=1 -v | grep -q "^--- PASS: TestLowerMethodCall"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
