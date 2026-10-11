---
title: An ampersand before a sigil is not read as a call, so &$subref refuses
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T18:50:21.499634715Z
updated: 2026-10-11T10:39:29.849811379Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  commits: 0
  started_at: 2026-10-11T10:39:29.580400369Z
  ended_at: 2026-10-11T10:39:29.849811379Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T10:39:29.580400369Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T10:39:29.849811379Z
---

`&$subref` refuses. `&` followed by a SIGIL is not read as a call, while `&`
before a brace or a word is. 36 of perl.git t/'s files use it, and it is one of
the first-failure shapes in the current 302.

Found by bucketing the first Unknown of every dirty file after the require and
label fixes took the goal to 318 of 620.

## Measured at `20c20c3c`

    &$subref;          Unknown=1   canon "&$subref;"
    &$s;               Unknown=1   canon "&$s;"
    &$s();             Unknown=1   canon "&$s();"
    &$s(1,2);          Unknown=1   canon "&$s(1,2);"
    &{$s};             Unknown=0   canon "&{$s};"        <- a BRACE is fine
    $s->();            Unknown=0   canon "$s -> ();"     <- the arrow is fine
    &name;             Unknown=0   canon "&name;"        <- a WORD is fine

So it is `&` immediately followed by `$`. All four failing spellings refuse;
declaring the sub changes nothing, because the callee is a scalar rather than a
name.

Perl reads it as an ordinary call:

    perl -MO=Deparse -e 'my $subref = sub {1}; &$subref;'   ->  &$subref;

## The code says so

`internal/lexer/scan.go:735`'s `scanAmp` has:

    // `&&` is an operator even where a term is expected -- it is the
    // low-precedence form. Only a single `&` before a NAME is a sigil.

"before a name" is accurate about what it does and is the gap. A `&` before `$`
falls through to the operator reading, so `&$s` lexes as bitwise-and applied to
`$s`, and the statement has no term where one belongs.

## Reach

    36 of 620 files contain `&$name` (grep -rlE '&\$[A-Za-z_]')

It is the old calling convention -- `&$coderef` passes the caller's `@_`
implicitly, which is why perl's own suite uses it in tests about `@_` and
`goto &sub`. `t/comp/*.t` and `t/op/sub*.t` are the dense users.

## The `@_` semantics are NOT this issue

`&$s;` without parens passes the CURRENT `@_` to the callee; `&$s();` passes an
empty list. That is a real semantic difference and the corpus should eventually
assert it, but both spellings refuse here identically, so the parse fix is one
rule and the semantic assertion is a separate corpus case. Do not conflate them.

`conformance/07_subroutines/03_call_forms.t` already documents the `&f` form and
its own prose admits the claim is unobservable there, because all four spellings
in that case print the same value -- recorded in the m3 milestone notes. A case
for `&$coderef` should avoid that trap and make the `@_` difference visible.

## Scope

1. `scanAmp` should take `&` followed by a sigil as the call sigil, the way it
   takes `&` followed by a name. Measure the boundary: `&&` is already excluded,
   and `$a & $b` must stay bitwise-and -- `wantsTerm()` is the existing guard
   and may be sufficient.
2. Measure `&@a`, `&%h` and `&&$s`. If perl rejects them, our refusing them is
   CORRECT and the rule is `&` before `$` only.
3. Check `&$s[0]` and `&$s{k}` -- a subscript after the scalar. Perl may read
   those as an element of `@s`/`%s` rather than a call, which would make the
   rule narrower than it looks.
4. Canon must re-emit the `&`. It does today only because the statement is
   Unknown and echoes its bytes.

## Context

- paths: `internal/lexer/scan.go:735` (`scanAmp`), `internal/lexer/expect.go`
  (`wantsTerm`)
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`
- related: `01a0df00` (caret variables -- the other sigil-shaped lexer gap found
  in the same bucketing pass), `01a0deeb` (no ratchet gates this goal)

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `&$subref;`, `&$s;`, `&$s();` and `&$s(1,2);` all score Unknown=0 and canon with the `&$` kept, as perl's Deparse does. Fixed by c241ef8a.

The criteria that grepped `^ok` after `-run '^TestAmpersandScalarCall$'` were vacuous: `go test` prints `ok` when `-run` matches nothing. They now grep the test's own PASS line. The lexer guard names TestLeadingPackageSeparator, which asserts `$a & $b` lexes as `Operator(&)`. TestAmpersandScalarCall also asserts that `&$s[0]`, `&$s{k}` and `&@a` keep refusing, as perl 5.42.0 does (Scope items 2 and 3).

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] `&$subref;` parses as a call (`go test ./internal/parse/ -run '^TestAmpersandScalarCall$' -count=1 -v | grep -q '^--- PASS: TestAmpersandScalarCall'`)
- [ ] All four spellings parse: bare, `()`, `(args)`, and with the sub declared (`go test ./internal/parse/ -run '^TestAmpersandScalarCall$' -count=1 -v | grep -q '^--- PASS: TestAmpersandScalarCall'`)
- [ ] `$a & $b` is still bitwise-and (`go test ./internal/lexer/ -run '^TestLeadingPackageSeparator$' -count=1 -v | grep -q '^--- PASS: TestLeadingPackageSeparator'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 -timeout=30m`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1`)
