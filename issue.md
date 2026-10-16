---
title: "Ponytail cuts in infer and psc from the PR #486 review: one acceptance rule for the second-chance checks"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T01:16:38.13979476Z
updated: 2026-10-16T18:48:35.925290487Z
---

## Context

- found by: the agentic review of PR #486, ponytail lens (half B); report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`

perigrin ruled 2026-10-16: the ponytail simplifications are filed as refactor issues, not applied on PR #486. Apply them after #486 merges, each in its own commit with tests unchanged in behaviour. Items a criterion-named test depends on stay. The specialist's findings:

#### Ponytail review, half B (PR #486 at 8a0c85c6)

Each collapse marked "verified" was applied to a scratch copy
(scratchpad/ptB, a tar of the tree without .git) and run with
`go test -count=1 ./internal/infer/ ./internal/psc/` and
`go test ./test/e2e/ -run 'PSC|TypeInference'`. All of them passed, and the
precision report still showed WRONG 0. The diffs are in
scratchpad/ponytail-B-checks.diff and scratchpad/ponytail-B-symbols-guards.diff.

#### 1. The second-chance satisfaction checks collapse into one rule (verified)

The rule: a slot accepts a value if any view perl takes of it silently
satisfies the slot. Those views are the value as it is, the value as a
handle, the value numified, and the result of a call that succeeded. Today
`checkArg` and `checkOperand` each spell this out by hand, and they spell it
differently.

- **internal/infer/infer.go:567 `skipUnknownOperand`**: delete it and its two
  call sites (calls.go:227, infer.go:599). `TypeSatisfies(Unknown, x)` is
  already true when not Strict and false under Strict (types.go:350), so
  `satisfies` covers it in both modes. Saves 7 lines. Confidence 95.
- **infer.go:604 `if expected == types.Any`**: delete it. `typeSatisfies`
  returns true for `required == Any` first thing (types.go:346). Saves 3
  lines. Confidence 95.
- **ops.go:218-220 and calls.go:226-228**: `checkOperand` re-inlines
  `argSatisfies` without the as-is and handle clauses, then gets the as-is
  check back later inside `operandMismatch` (infer.go:608). Replace
  `argSatisfies` with one
  `accepts(n, actual, expected)` that loops over `{actual, succeeded(n, actual)}`
  and tests `satisfies || handleSatisfies || satisfies(numified)`. `checkArg`
  and `checkOperand` both call it. `operandMismatch` then shrinks to severity
  plus message: it loses its `opts` parameter, its `bool` result, the narrowing,
  the skip, the Any check and the satisfies check. Confidence 85.
- **finite is only ever called as `numified(finite(x))`** (4 sites): call
  `finite` inside `numified`, after its `TypeSatisfies` early return. The
  early return must come first, or `Numified(Scalar, Str)` breaks
  `TestBooleanInStrSlotIsSilent`; I checked this. Saves about 4 lines.
  Confidence 85.
- **The genuine-aggregate count is written four times**: infer.go:593
  (operandMismatch), infer.go:883 (`builtinArgActual`), ops.go:170 (`narrow`)
  and ops.go:417 (`valueOf`). Each is `if t == Array || t == Hash || t == List
  { NarrowByContext(t, ScalarCtx) }`, and for exactly those three inputs
  `NarrowByContext` always returns Int. Replace all four with a single
  `counted(t)` that returns Int for those types. `builtinArgActual` then
  becomes `if name == "scalar" { actual = counted(actual) }` in `checkArg`, and
  `checkOperand` applies `counted` when `expected != List`. Saves about 20
  lines. Confidence 90.
- **infer.go:648-652**: the severity computation in `operandMismatch` is a
  copy of `argumentSeverity` (infer.go:704). Call the function instead. Saves
  5 lines. Confidence 95.
- **infer.go:665-671**: the `int` special case in `argDirectedReturn` should
  go. Its comment is now false: CORE.tpm:451 declares `sub int (Num $number =
  $_) Int`, so `callReturns` already gives Int. Saves 8 lines. Confidence 90.

The verified diff for this group is 38 lines added and 102 deleted, a net
saving of 64. Keeping the long justification comment from `operandMismatch`
(arity ordering, why only a genuine aggregate) by moving it onto `counted`
brings the net saving to about 48. It is a true comment and must not be
dropped. Two helpers stay because they hold real rules no other check
covers: `scalarValue`, a call or assignment in a `$` slot, and `flattened`,
a slurpy slot. `lossyCoercion` stays as the last chance before a report.

#### 2. Six SymbolTable.Update* methods share one loop (verified)

**internal/infer/symbols.go:166-260**: `UpdateType`, `UpdateElemType`,
`UpdateKeyTypes`, `UpdateReturnType`, `UpdateParamTypes` and
`UpdateClassType` each repeat the same 8-line scope-chain walk. Add a private
`update(name, func(*Symbol)) bool` that does the walk, and make each method a
one-line closure over it. Separately, `CurrentPackage` (symbols.go:264-271)
loops only to find root, so it can be `return st.root.Name`. Saves 22 lines.
Confidence 95.

#### 3. negateGuard's De Morgan is redundant, and it is also a bug (verified)

**flow.go:635-651**: `negateGuard` applies De Morgan itself, but only one
level deep. It sets `Negated` on a child that may be a compound, and
`flattenGuards` (infer.go:1200) ignores `Negated` on a compound node. Perl
shows the effect:

```perl
my $x = undef; my $y = 1; my $z;
if (!((defined $x && defined $y) || defined $z)) { print $x + 1 }
```

perl 5.42 -w warns "Use of uninitialized value $x in addition". PSC narrows
$x to defined inside the body and reports nothing. The same body under
`if (!(defined $z))` reports correctly.

The fix: `negateGuard` becomes "copy and flip `Negated`" for every guard,
and `flattenGuards` gets one line, `negate = negate != guard.Negated`,
before its De Morgan switch. Saves about 10 lines and fixes the missed
diagnostic. The probe reports the diagnostic after the change, and all
suites stay green. Confidence 90.

#### 4. checkFile reads each file twice (psc)

**internal/psc/check_command.go:118-126**: `os.ReadFile(path)` runs, then
`sess.ParseFileFrom(path, roots...)` reads the file again
(parse/module.go:209). Use `root := sess.ParseSourceFrom(path, source,
roots...)`, as lsp.go:58 already does. That drops the second read and its
error branch. It also guarantees the line:col that `FormatDiagnostic`
computes from `source` refers to the same bytes that were parsed. Saves 4
lines. Confidence 95.

#### 5. test/e2e/helpers/psc.go is dead (114 lines)

The whole file can go:
- `isPSCAvailable` always returns true.
- `SkipIfNoPSC` is called only from this file.
- `SkipIfNoPSCIntegration`, `AssertPSCSucceedsOrSkipTODO`, `TestPSCCommand`
  and `handlePSCDependencyError` have no callers.
- The skip messages still say "run 'make tree-sitter'".

The PR edited a comment in this file but left the tree-sitter-era machinery
behind. The `containsAny` defined in internal/updater is that package's own
copy. Saves 114 lines. Confidence 95.

#### 6. walk should be nil-safe, so operand() and about ten nil guards go

**ops.go:426 `operand`** exists only to nil-guard `walk`. About 10 sites
still write `if X != nil { a.walk(X) }` by hand: flow.go:39, 61, 110, 121,
152, 156, 188; vars.go:53, 63, 97, 121; calls.go:27, 30. Put `if n == nil {
return types.Unknown }` at the top of `walk`, then delete `operand` and the
guards. This puts one guard in the shared function. Saves about 25 lines.
Confidence 90.

#### 7. Smaller duplications

- **flow.go:506-518**: the signature branch of `subParam` duplicates the one
  in `subParams` (calls.go:525-532). Replace it with `if len(s.Sig) > 0 {
  if p := subParams(s); len(p) == 1 { return p[0] }; return "" }`. Saves 6
  lines. Confidence 90.
- **infer.go:1049-1079**: `collectPreIfTypes` is a wrapper over
  `collectLeafVarTypes` used once. Merge them. Saves about 8 lines.
  Confidence 90.
- **infer.go:1127-1149**: in `applyBranchJoin`, the `if xTypes != nil`
  checks are redundant because reading a nil map is safe in Go. Saves about 8
  lines. Confidence 95.
- **infer.go:1242-1244**: the `return nil` after `flattenGuards`' exhaustive
  switch is unreachable. Saves 2 lines. Confidence 90.
- **flow.go:260-267 and flow.go:309-312**: `blockWithGuard` and
  `logicalRight` both do EnterScope("guard") and then define each shadow with
  its ClassType. Extract `enterGuards(shadows, classNames)`. Saves about 4
  lines. Confidence 85.
- **special.go:16-19 `mapGrepSort`**: it only walks the children and returns
  Unknown. Fold it into walk's untyped case at infer.go:177 and keep its
  comment. Saves 5 lines. Confidence 90.
- **infer.go:852-855 and calls.go:156-159**: `builtinArgType` and
  `slotContext` repeat the same "pick param i of this candidate, skip if
  absent" lines. Extract `paramAt(s, i)`. Saves about 4 lines. Confidence 80.
- **infer.go:897-900 and 942-945**: `shapedOperatorCandidates` and
  `shapedCallReturns` both expand UnknownCtx into {Scalar, List, Void}.
  Extract `contexts(ctx)`. Saves about 4 lines. Confidence 85.
- **calls.go:279, calls.go:329, flow.go:676**: "take the last element of a
  List" is written three times. Extract `lastElem(n)`. Saves about 4 lines.
  Confidence 85.
- **ops.go:80-87 `spaceshipResult`**: its test `t&Num == 0 && t&NaN != 0`
  is `finite(t)&NaN != 0`. Saves 1 line. Confidence 80.
- **Hand-rolled recursive walkers**: overrides, returnsBuiltinResult,
  unpacksArgs, paramTypes, widenWritten, assignedNames, subReturnExpr and
  readsArg0 each write `var f func(lower.Node); f = func(...) { ... for c :=
  range lower.Children(n) { f(c) } }`. A `lower.Inspect(n, func(Node) bool)`,
  shaped like go/ast.Inspect, would remove about 3 lines from each. lower is
  half A's scope. Saves about 25 lines. Confidence 70.
- **check_command.go:40-61 and analyze_command.go:33-46**: both repeat the
  same WalkDir plus `isPerlFile` plus file-or-directory dispatch. Extract
  `eachPerlFile(target, fn)`. Saves about 12 lines. Confidence 85.

#### 8. Exported or declared, but nothing in production uses it

- **diagnostics.go:54 `CodeUnknownBuiltin`**: nothing emits it. Only
  diagnostics_test.go reads it (lines 105 and 149). Delete the constant and
  those two assertions. Saves 3 lines. Confidence 90.
- **lsp.go:99-108 `LSPServer.Annotations`**: no caller outside tests; the
  three uses in lsp_inference_test.go can use `TypeAtByte` or
  `doc.analysis`. In the same file:
  - lsp.go:20, `document.uri`, is never read.
  - lsp.go:78-80, `doc.analysis = nil` just before `delete`, does nothing.
  - lsp.go:200, `result := sym; return &result` can be `return &sym, true`.

  Together these save about 17 lines. Annotations predates the PR.
  Confidence 65.

#### Left alone on purpose

- **ParamTypes and ParamNames**: subParams, unpacksArgs, paramTypes and
  recordParamTypes (calls.go:520-684, about 120 lines) have no production
  reader. They exist only for TestLoweredParamTypes, which must stay.
- **ProjectIndex, NewProjectIndex, Prefetch and every `a.idx != nil` branch**:
  all of it is dead in production, because every caller passes `idx = nil`
  (lsp.go:66, check_command.go:128, project.go:83). It was the same on
  origin/pu. docs/plans/2026-03-19-cross-file-analysis.md names it, so I have
  not flagged it for deletion. It is an unwired plan, though, not a feature
  in use.
- **SymMethod**: lsp_handler.go uses it, and the LSP design spec names it.
- **An unflagged bug next to finding 1**: `succeeded` (calls.go:251) treats
  a builtin that `use subs` overrides as a builtin, so it strips Undef from
  a user sub's result. This is outside the ponytail lens; flagging it for the
  bug reviewers at confidence 60.

## Steps

- [ ] One commit per cut; behaviour unchanged (the infer goldens, precision and the m1 gates hold); `make test`; commit

## Acceptance criteria

- [ ] the cuts are applied with behaviour unchanged (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
