---
title: "A block deref loses its braces: ${$pkg . '::'} re-emits as $$pkg . '::'"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T00:31:06.168570065Z
updated: 2026-10-11T11:25:22.632139659Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: ""
  commits: 0
  started_at: 2026-10-11T11:25:22.632139659Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T11:25:22.632139659Z
---

## Measured finding

`${EXPR}` -- the block form of a scalar deref -- loses its braces, so the
emission re-groups around the sigil. Measured at `34737c46` in a clean
`git archive` of HEAD, independent of any in-flight change:

    source                                  first canon                      second canon
    my $x = ${$pkg . '::'};                 $$pkg . '::'                     ($$pkg) . '::'
    my $x = !!${$pkg . '::'}{$1 . '::'};    !!($$pkg . '::'){$1 . '::'}      !!(($$pkg) . '::'){$1 . '::'}
    f(${$pkg . '::'});                      f($$pkg . '::')                  f(($$pkg) . '::')

The braces of `${EXPR}` GROUP -- `term.go:206` says so in its own comment --
and the whole expression inside them is what the sigil applies to. Dropping
them makes `$` apply to `$pkg` alone and leaves `. '::'` as a concatenation
outside the deref, which is a different program. perl, measured on 5.42.0 at
`/home/perigrin/.local/bin/perl`:

    $ perl -MO=Deparse -e 'no strict "refs"; my $pkg = "main"; my $x = ${$pkg . "::"};'
    my $x = ${$pkg . '::';};

perl's own deparser writes the braces back, because there is no brace-free
spelling of this.

A SECOND defect in the same file: `$::{$pkg}` -- the main-stash subscript --
canons as `$:;:{$pkg}`, a `$:` variable followed by a stray `;:`. That one is a
LEXER-level split of the `$::` sigil and is noted here rather than filed apart
because the same corpus file carries both.

## How it was found

Found by fixing 01a0e013, whose canon defect had been DELETING every `return`
operand -- `return EXPR;` canon'd as `return ;`. With the operand emitted, this
became visible. `package_decl_stash.t` in the T1 corpus is the file:

    return !!${$pkg . '::'}{$1 . '::'};
    return !!$::{$pkg . '::'};

It sat in `t1canon.ratchet` at 0 only because BOTH emissions deleted the whole
operand, which made a wrong tree look like a fixpoint. It is `0 -> 1` after
01a0e013 and the ratchet records why.

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `my $x = ${$pkg . '::'};` canons as itself and is a fixpoint, as are `!!${$pkg . '::'}{$1 . '::'}` and `f(${$pkg . '::'})`. `$::{$pkg}` and `$::{$pkg . '::'}` canon as themselves, not `$:;:{$pkg}`. `package_decl_stash.t` is 0 in `t1canon.ratchet`. Fixed by 06f95ce2, the braces, and 9b7b1311, the main stash's empty package name.

TestBlockDerefKeepsItsBraces never existed. The test is TestDerefCanonKeepsBraces, from 06f95ce2; the issue's own shapes were added to it in test/pgt-reconcile, and they fail when canon's brace rule or the lexer's `$::` rule is removed. The guard's `%{$h}` and `${$h}{k}` canon as `%$h` and `$$h{k}`, the same programs, so the claim is reworded from "unchanged" to "keep their meaning". The commands were bare backticks without parens, which verify does not run.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] `my $x = ${$pkg . '::'};` canons with its braces and is a fixpoint (`go test ./internal/parse/ -run '^TestDerefCanonKeepsBraces$' -count=1 -v | grep -q '^--- PASS: TestDerefCanonKeepsBraces'`)
- [ ] `$::{$pkg}` canons as itself rather than `$:;:{$pkg}` (`go test ./internal/parse/ -run '^TestDerefCanonKeepsBraces$' -count=1 -v | grep -q '^--- PASS: TestDerefCanonKeepsBraces' && go test ./internal/lexer/ -run '^TestMainStashSubscript$' -count=1 -v | grep -q '^--- PASS: TestMainStashSubscript'`)
- [ ] `package_decl_stash.t` is 0 in `t1canon.ratchet` and held there by TestCanonTokenIdentity (`grep -q '^0 package_decl_stash.t$' internal/parse/testdata/t1canon.ratchet && go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 -v | grep -q '^--- PASS: TestCanonTokenIdentity' && go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] REGRESSION GUARD: `$$r`, `@{$r}`, `%{$h}` and `${$h}{k}` keep their meaning (`go test ./internal/parse/ -run '^TestDerefCanonKeepsBraces$' -count=1 -v | grep -q '^--- PASS: TestDerefCanonKeepsBraces'`)
- [ ] the whole tree is green (`go test ./... -count=1 -timeout=30m`)
