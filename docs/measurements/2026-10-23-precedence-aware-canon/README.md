<!-- ABOUTME: Scratch measurement behind perigrin's round-7 ruling on m3-oracle-at-scale: precedence-aware canon. -->
<!-- ABOUTME: Kept so NU and NL (issues 01a1690b-1d10, 01a1690b-4745) can reproduce the method and its numbers. -->

This directory records the measurement that decided the round-7 chain-review ruling for milestone `m3-oracle-at-scale`. `measure-canon.diff` is the scratch implementation (named unary ops written by precedence, `ArgParen`/`NamedUnary` flags, the call-paren rule narrowed for named unary ops). The `zz*_test.go.txt` files are the measurement tests; copy one to `internal/parse/` as `*_test.go` to run it. `base.tsv`, `narrow.tsv` and `design.tsv` are the per-file results for perl.git t/, and `fixtures.out` the wrong-tree fixture results. None of this is production code.

# Named-unary canon design: scratch measurement

Worktree: detached at 054e75e6, now removed. Nothing was committed or pushed, and the main checkout was not touched. perl 5.44.0 was the judge. Go is go1.24.4, and the corpus is PERL5_CORPUS=~/dev/perl5.

Artifacts in this directory:
- `measure-canon-full.diff` is the whole scratch change. `measure-canon.diff` is canon.go plus fidelity.go only.
- `zzmeasure_test.go` walks the corpus file by file. It uses perlGitTFiles, perlGitTRoots and Session.ParseFileFrom, skips perlRejects, and runs parse.Faithful.
- `zzstmt_test.go` runs Faithful on every Statement node under three configurations.
- `zzfixtures_test.go` holds the fixtures and the hand-built wrong trees. `zzshapes_test.go` prints the unary set and the binding powers.
- `base.tsv`, `narrow.tsv` and `design.tsv` give the per-file verdict and the first difference. `stmts.tsv` lists the statements whose verdict changes. `fixtures.out` is the full fixture output.

## Sources used for the named-unary set and its precedence

- **Set, tree side:** `p.namedUnaryHere(spelled, text)` in call.go. This is the parser's own classification: keywordShape against CORE.tpm's derived shapes, which respects features and `CORE::`. File tests come from `lexer.IsFileTest`, where term.go already parses them as UNIOP.
- **Set, Faithful side:** Faithful sees tokens, not nodes, so it uses an approximation. A token counts when it is a file test, or when `coreShapes()[keywordName(tok)] == ShapeUnary`, and it is not preceded by `->`. That gives 84 builtins plus the file tests: abs alarm caller chdir chomp chop chr chroot close closedir cos dbmclose defined delete do each eof eval evalbytes exists exit exp fc fileno getc getgrgid getgrnam gethostbyname getnetbyname getpeername getpgrp getprotobyname getpwnam getpwuid getsockname gmtime goto hex int keys lc lcfirst length localtime lock log lstat oct ord pop pos prototype quotemeta rand readdir readline readlink readpipe ref reset rewinddir rmdir scalar sethostent setnetent setprotoent setservent shift sin sleep sqrt srand stat study tell telldir tied uc ucfirst umask undef untie values write.
  - The approximation counts a feature-gated `fc` or `evalbytes` as the builtin even where the feature is off.
  - It also counts a user sub that overrides a builtin as the builtin.
- **Precedence:** `bpNamedUnary` = 160, derived from CORE.tpm by derivePowers. For comparison: `<` is 140, `eq` 130, `<<` 170, `+` and `.` 180, `*` 190, `||` 90, `or` 10, prefix `not` 30, bpListOp 40.

## Scratch change

The parser change is 11 added lines and 2 removed:
- **parse.go:** adds `NamedUnary, ArgParen bool` to Node. Before this, the tree did not record whether the source wrote `NAME(`. Node.Paren is the node's own grouping, not the argument list.
- **call.go:** the paren path sets ArgParen, and sets NamedUnary from namedUnaryHere. The operand-less path and the namedUnary case set NamedUnary.
- **term.go:** a file test sets NamedUnary, and sets ArgParen on its paren path.

canon.go gains 42 lines, about 15 of them comments:
- The Call case gains a branch for NamedUnary without ArgParen.
  - It writes `NAME OPERAND`, with the operand emitted at bpNamedUnary+1. A looser operand gets wrapped: `defined ($x || $y)`.
  - It falls back to call parens only when the operator right after the call is tighter than 160. That is `f > bpNamedUnary`, where f is the "follow" power.
  - With no operand it writes the bare name.
- The follow power is threaded through emit:
  - Binary: the left operand gets the operator's power, and the right operand inherits.
  - Ternary: the first two parts get `?`, and the last inherits.
  - CmpChain: each link gets the operator, and the last inherits.
  - Unary operands, declaration initialisers, LoopControl expressions and modifier conditions inherit.
  - Postfix gets its own operator. Index children get atomBP.
  - Every other context gets 0. A node that is wrapped in parens resets the power to 0.
- Shortcut: the follow power is a package variable. Canon is therefore not concurrency-safe. A real change would thread it through emit's signature.
- I used the follow power instead of `outer` because a prefix op's operand extends to the right. Its outer power at a right-operand position, such as `!defined $x` or `$a . lc $b`, says nothing about what follows it, and using it would over-parenthesise correct trees.

fidelity.go gains 29 lines and loses 4:
- `namedUnaryAt` turns off the call-paren forgiveness in both places it applies: the main loop and matchForgivingSemicolons.
- Measurement toggles were added, along with a `faithfulToks` extraction so that one parse can be scored under several configurations.

## 1. Baseline at 054e75e6

perl.git t/ is **342/619** faithful. The M3 dirs (op mro uni lib test_pl) are **178/347**. Both match the expected numbers.

## 2. Narrowed Faithful only (canon unchanged)

perl.git t/ is **35/619**. The M3 dirs are **27/347**. The earlier estimate was 41 and 28; my set excludes method names after `->`.

- 307 files become unfaithful.
  - Every first difference is `canon has "("`.
  - None of them is a grouping error; each is a parenless named unary that canon parenthesised.
- 0 files become faithful.

## 3. Precedence-aware canon plus narrowed Faithful

perl.git t/ is **342/619**. The M3 dirs are **178/347**.

**0 files change in either direction**, so there is no sample of newly unfaithful files to give. Three files keep their verdict but have a different first difference:
- `lib/universal.t`: the trailing-token count moved from 369 to 365.
- `re/reg_fold.t`: the first difference moved from token 688 to 689.
- `run/locale.t`: the first difference moved from token 160 to 167.

Faithful reports only the first difference in a file, so I also scored every Statement node without a heredoc, 73,670 statements in all:

| config | faithful statements |
|---|---|
| baseline | 67,002 |
| narrowed only | 56,436 |
| design | 67,011 |

- 0 statements went from faithful to unfaithful under the design. Its precedence emission therefore reproduced every source spelling, and the old forgiveness hid **no** named-unary misgrouping anywhere in perl.git t/, at the granularity of a statement.
- 9 statements became faithful. All are in files that stay unfaithful for other reasons:
  - signature defaults in op/signatures.t: `($a = undef)`, `($a = caller)`, `goto &t130`
  - `delete $::{...}` and `= undef` in uni/gv.t and mro/package_aliases_utf8.t
  - I did not trace exactly why each one moved.

## 4. Fixtures

Canon output and Faithful verdict are under the design. Baseline canon is shown in brackets. Wrong trees were built by regrouping the parsed tree by hand. Where it matters, perl 5.44 Deparse,-p confirms that each wrong tree's canon is a different program.

| source | correct tree: canon / faithful | wrong tree: canon / faithful | baseline wrong-tree verdict |
|---|---|---|---|
| `defined $x \|\| $y;` | `defined $x \|\| $y;` / yes [`defined($x) \|\| $y`] | defined over `\|\|`: `defined ($x \|\| $y);` / **no** (Deparse: `defined(($x or $y))` vs `(defined($x) or $y)`) | forgiven (yes) |
| `ref $x eq 'A';` (a 1-link CmpChain) | `ref $x eq 'A';` / yes | ref over eq: `ref ($x eq 'A');` / **no** | forgiven |
| `chdir 't' if -d 't';` | `chdir 't' if -d 't';` / yes | no regrouping of the same leaves exists (nothing follows either operand) | n/a |
| `lc $s;` | `lc $s;` / yes | none constructible | n/a |
| `scalar @a;` | `scalar @a;` / yes | none constructible | n/a |
| `exists $h{a};` | `exists $h{a};` / yes | subscript outside: `exists($h){a};` / **no** (perl rejects that text) | forgiven |
| `shift;` | `shift;` / yes [`shift()`] | none (no operand) | n/a |
| `-e $f;` | `-e $f;` / yes | none constructible | n/a |
| `defined($x) \|\| $y;` | `defined($x) \|\| $y;` / yes (ArgParen kept) | defined over `\|\|` with ArgParen: `defined($x \|\| $y);` / no | no (this one was already caught) |

`defined($x) || $y;` has one more case: a tree with the right grouping but the ArgParen flag dropped emits `defined $x || $y;`, which is unfaithful. The baseline accepted that tree.

`not $x or 1;`:
- I classify it as **not a named unary**. The parser reads it as a prefix Unary (NOTOP) at power 30, not as a Call, and it is not in CORE.tpm's unary shapes.
- The correct tree gives `not $x or 1;` / yes.
- The wrong tree, not over `or`, gives `not ($x or 1);` and is **still forgiven**. `isName("not")` is true and namedUnaryAt excludes it.
- So `not` is a remaining hole, in the baseline as well.

List ops were left unchanged, as the task asked:
- `push @a, 1 or die;`
  - The correct tree gives `push(@a , 1) or die();` / yes.
  - The wrong tree, push over `or`, gives `push(@a , 1 or die());` / **still forgiven**. perl reads that text as `push(((@a,1) \|\| die()))`.
- `sub foo; sub bar; foo 1, bar 2, 3;` (bare `foo 1, bar 2, 3;` with no declarations parses as Unknown)
  - The correct tree gives `foo(1 , bar(2 , 3));` / yes.
  - The wrong tree, `foo(1, bar(2), 3)`, gives `foo(1 , bar(2) , 3);` / **still forgiven**. perl judges it different: `foo(1, bar(2), 3)` against `foo(1, bar(2, 3))`.

## 5. Effort

- **Lines changed:** 11 added and 2 removed in the parser (parse.go, call.go, term.go). 42 added in canon.go. 29 added and 4 removed in fidelity.go; without the scratch toggles and the extraction, the design is about 18 lines.
- **Precedence gaps:** there were none for this set, because bpNamedUnary already existed and is derived. The real work was the right-edge problem: a prefix op's parens depend on what FOLLOWS it, which emit's `outer` does not carry. That needed the follow power described above.
- **Named ops this change does not cover:**
  - `not` is NOTOP, so a wrong tree that misgroups it is still forgiven.
  - `require` has its own path.
  - `goto` goes through control.go.
  - User subs with a `($)` prototype take the knowsShape route. They are not flagged, so they are still always parenthesised and forgiven.
  - The `eval`/`do` BLOCK forms are skipped by design.
- **Test churn:** 11 existing parse tests fail because they pin the old spelling (`length("x")`, `-e($f)`, `keys(%h)`, `defined($a) , $b`, `-e()` for an operand-less file test, and so on):
  - TestBarewordFilehandleCase
  - TestCoreUnaryAndListOperatorParseDiffer
  - TestCorePrefixedKeyword
  - TestCoreFeatureGatedBuiltinTypedParseUnchanged
  - TestEmptyCommaSlot
  - TestFeatureGatedFcArity
  - TestMainStashEmptyName
  - TestOperandlessFiletest
  - TestParenlessBuiltinArity
  - TestPrintScalarBeforeModifier
  - TestSortComparator
- TestOperandlessFiletest's "swallowed the terminator" check reads `-e or` as a defect. Perl reads the bare spelling the same way, so that check would need to change, not the code.
- Every other internal/ package passes. internal/parseoracle fails only when PATH's `perl` is the system 5.38; with 5.42 first on PATH, it passes.

## 6. Ratchets with the change

| test | result | files moved |
|---|---|---|
| TestCanonRatchet | PASS | 0 |
| TestCanonFidelityRatchet (T1) | PASS | 0 |
| TestParseRatchet | PASS | 0 |
| TestPerlGitTRatchet | PASS | 0 |
| TestCanonQuotientIsNarrow | PASS | 0 |

These ratchets fail on movement in either direction, so a pass means nothing moved.

# `not`, list ops, calls and niladics: scratch measurement (issue 01a1690b-4745)

Worktree: detached at f09a07a0, after the named-unary issue merged; removed after the run. perl 5.44.0 judged the facts quoted here.

Artifacts, all prefixed `nl-`:
- `nl-design.diff` is the whole scratch change, with three toggles: `measureRuleOff` turns off the call-paren forgiveness, `measureDesign` writes the design, and `measureTC` adds the trailing-comma variant below.
- `nl-zzmeasure_test.go.txt`, `nl-zzstmt_test.go.txt` and `nl-zzexport_test.go.txt` are the harness, extended from the files above to three configurations.
- `nl-files.tsv` gives the per-file verdicts with the trailing-comma variant on. Its columns are base, rule-off and design, then each one's first difference.
- `nl-stmts.tsv` lists every statement whose verdict differs between baseline and design, with the variant off.

The design:
- A Call without `ArgParen` and not a named unary is written bare: `NAME ARGS`. Its arguments are read at bpListOp+1, and the last one carries the follow power.
  - The call parens come back where the follow power is tighter than a list op, or where the arguments' emission opens with a `(` that closes early.
  - A call with no arguments is written as its bare name. That covers niladic builtins (`time`, `wantarray`) and compile-time tokens.
- The parser sets `ArgParen` on a method call's `->name(` and on an indirect call's `(` (expr.go, indirect.go). An indirect call without one is written `METHOD CLASS ARGS`.
- `not` is written by the follow power, as landed in this issue.
- A comma-separated element other than the last gets the comma's power as its follow.
- Variant (`measureTC`): a bare call whose last argument carries the source's trailing comma stays bare whatever follows.

| configuration | perl.git t/ | M3 dirs | statements (73,670) |
|---|---|---|---|
| baseline | 342/619 | 178/347 | 67,012 |
| rule off alone | 75/619 | 49/347 | 50,935 |
| design, rule off | 348/619 (1 lost, 7 gained) | 183/347 | 67,183 (3 lost, 174 gained) |
| design + trailing-comma variant | 349/619 (0 lost, 7 gained) | 184/347 | 67,186 (0 lost, 174 gained) |

The stop condition does not fire.
- All 23 files fresh5 lost are faithful under the design except op/utf8cache.t, and the variant recovers it.
- No other file or statement goes from faithful to unfaithful.
- Gained files: comp/use.t, mro/isa_c3.t, mro/package_aliases_utf8.t, op/isa.t, op/not.t, uni/attrs.t, uni/method.t.

Triage:
- **op/utf8cache.t:70**, `substr $x, 0, 1, = "\x{100}";`, is a list op ended by its trailing comma. perl reads `substr($x, 0, 1) = "\x{100}"`. The design writes call parens because `=` (power 60) is tighter than a list op (40). So the file is lost unless the tree's trailing comma keeps the call bare.
  - Measured at 5.44, a trailing comma ends the list before `=`, `.=`, `||=`, `&&=`, `==` and `||`. `print 1, 2, || die` is `(print(1, 2) or die)`.
  - It does not end the list before `-=`, `*=`, `%=`, `/=`, `&=`, `**=`, `x=`, `//=` or `<<=`. Each of those is a syntax error or re-lexes as a term.
  - The variant keeps only the source's own spelling. It is right for correct trees, which is all the corpus holds.
  - io/open.t:281, `ok open(...), '...', || _diag $!;`, is the same case. That file is unfaithful for other reasons.
- **uni/parser.t** (a heredoc) stays faithful under the design. Its heredoc statements are outside the statement count.
- The 17 niladic files and the 2 indirect-call files are recovered by writing zero-argument calls bare and keeping the indirect spelling.

The landed `not` differs from the scratch in one place. A `not` whose own Paren flag is set keeps its pair: `my @l = (not $x, 1)` is a probe row.
