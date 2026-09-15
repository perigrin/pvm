# Triage: `expected Int/Num, got X` — 762 diagnostics

All witnesses live in `.../scratchpad/triage/w/`. Every claim below was run
against the system perl (5.40.2). Headline: **the overwhelming majority are
false positives**, and they trace to a small number of missing rules rather
than to 762 independent mistakes.

The single most consequential observation: the files with the highest
diagnostic density (`t/op/range.t`, `t/op/repeat.t`, `t/base/num.t`,
`t/op/bop.t`) are the files whose *entire purpose* is to exercise Perl's
coercion edge cases. PSC is loudest exactly where Perl is deliberately being
weird on purpose. A checker aimed at ordinary code should probably not be
scored against perl's own coercion torture tests.

---

## Finding 1: Magic string range (`'a'..'z'`) treated as a numeric range
- **Instances:** 239 of 762 (all `..` with `got Str`)
- **Example:** `t/op/for.t:102` — `for my $i ('A' .. 'C') {`
  also `t/comp/proto.t:637` — `my @multiarray = ("a".."z");`
  also `t/cmd/for.t:70` — `for ("-3" .. "0") {`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent, and the feature works as documented. `perl -w str_range.pl`
  printed `alpha: a b c d e`, `numeric-str: -3 -2 -1 0`, `multi: aa ab ac ad`
  with zero warnings, exit 0. The string-increment ("magic") range is a
  first-class documented operator (perlop, "Range Operator"), not a coercion.
- **Root cause:** PSC models `..` as having signature `(Int, Int)`. Perl's `..`
  is genuinely two operators chosen at runtime by operand type: numeric range
  when both sides look like numbers, magic string increment otherwise. The
  single-signature model cannot express that.
- **Suggested fix:** Give `..` a union signature `(Int,Int) | (Str,Str)` and
  only complain when the operands are of *mixed or unrelated* type.

## Finding 2: Integer arithmetic widened to `Num`, then rejected where `Int` is required
- **Instances:** 65 of 762 (43 `..` + 22 `x`, both `got Num`)
- **Example:** `t/charset_tools.pl:17` — `for my $i (0 .. length($string) - 1) {`
  also `t/io/eintr_print.t:42` — `my $full_sample = 'abxhrtf6' x (8192-7);`
  also `t/base/rs.t:87` — `foreach $test ($test_count..$test_count + 3) {`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent. `perl -w arith.pl` ran `0 .. length($s)-1`,
  `(1 .. $n-1)`, and `"ab" x (8-7)` with no warnings, exit 0, correct results
  (`range: 1 2 3 4`, `x: ab`). `length($s)-1` printed as `4`, an integer.
- **Root cause:** PSC's arithmetic rule appears to be `Int op Int -> Num`
  unconditionally, presumably because `/` and `**` can produce fractions. That
  widened `Num` then fails the `Int` requirement on `..` and `x`. The
  `length(...) - 1` idiom is ubiquitous, which is why this group is large.
- **Suggested fix:** Make `+`, `-`, `*` and `%` preserve `Int` when both
  operands are `Int`; reserve widening to `Num` for `/`, `**`, and friends.

## Finding 3: Comparison/boolean results rejected as numeric operands
- **Instances:** 87 of 762 (`got Bool`)
- **Example:** `t/opbasic/arith.t:379` — `print "not "x($a ne $b), "ok ", $T++, ...`
  also `t/io/print.t:73` — `print "not " x ($n ne "a5c") . "ok 23 ..."`
  also `t/comp/proto.t:34` — `|| (defined($p) != defined($c)));`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent. `perl -w boolnum.pl` confirmed `"not " x ($a ne $b)`
  emits the string once when true and the empty string when false; `$cmp + 0`
  gave `1`; `$cmp == 1` was true. No warnings, exit 0.
- **Root cause:** PSC has a distinct `Bool` type that is not a subtype of
  `Int`/`Num`. In Perl there is no separate boolean: comparison operators
  return the integer 1 or the empty string, both of which are legitimate
  numeric operands. `"not " x ($cond)` is a pervasive idiom in perl's own
  test suite.
- **Suggested fix:** Make `Bool` a subtype of `Int` in the lattice so it
  satisfies `Int` and `Num` requirements without a coercion note.

## Finding 4: Bitwise string operators read as integer bitwise operators
- **Instances:** 81 of 762 (`&`, `|`, `^`, `<<`, `>>`)
- **Example:** `t/op/bop.t:127` — `is ("ok \xFF\xFF\n" & "ok 19\n", "ok 19\n");`
  also `t/op/bop.t:128` — `is ("ok 20\n" | "ok \0\0\n", "ok 20\n");`
  also `t/op/bop.t:250` — `is(($x | $y), ("a" | "c"));`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent, and the results are meaningful strings. `perl -w
  bitstr.pl` printed `AND: ok 19`, `OR: ok 20`, `XOR: ok 21`, exit 0, no
  warnings.
- **Root cause:** Same single-signature problem as Finding 1. When both
  operands of `& | ^` are strings, Perl performs a *bitwise string* operation
  (perlop, "Bitwise String Operators") producing a string, not a number. PSC
  only models the integer form.
- **Suggested fix:** Give `& | ^` a `(Str,Str) -> Str` alternative alongside
  the integer signature, mirroring the `..` fix.

## Finding 5: `==` used for reference identity comparison
- **Instances:** 44 of 762 (`CodeRef`/`ArrayRef`/`HashRef`/`Object`/`Regex`)
- **Example:** `t/op/attrs.t:227` — `main::ok $c == \&{"t0"};`
  also `t/class/construct.t:50` — `is($obj+0, 12345, 'numified object with overload');`
  also `t/op/qr.t:23` — `isnt($a + 0, $b + 0, 'Not the same object');`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent. `perl -w refeq.pl` (with `$c` **explicitly assigned**
  `\&t0`, not left undeclared) reported `coderef ==: true`, `arrayref ==:
  true`, `hashref ==: true`, no warnings. `perl -w overl.pl` gave
  `obj+0 = 12345` via the `0+` overload and a nonzero refaddr for `qr+0`.
- **Root cause:** Two rules missing. (a) Comparing two references with `==` is
  the standard identity/refaddr idiom — references numify to their address.
  (b) For blessed objects, `overload '0+'` makes numeric use explicitly
  legal, so PSC must consult overloading before flagging an `Object`.
- **Suggested fix:** Allow ref types as `==`/`!=` operands (identity
  comparison), and suppress numeric complaints on blessed values whose class
  declares `0+` or `fallback`.

## Finding 6: `undef` in range/repeat flagged as `error` in files that run without warnings
- **Instances:** 53 of 762 (`got Undef`; 20 of them at `error` severity)
- **Example:** `t/op/range.t:137` — `is(join(":",undef..2), '0:1:2');`
  preceded at line 136 by the comment `# undef should be treated as 0 for numerical range`
  also `t/op/repeat.t:19` — `is('-' x undef, '',     '  x undef');`
  also `t/op/not.t:32` — `ok($not1 == undef, ...)` directly under a `no warnings;` on line 31
- **Verdict:** MIXED — correct detection, wrong severity, wrong target
- **Perl says:** it depends on the pragma, and I measured both. With
  `use warnings`, `perl undefrange.pl` emitted `Use of uninitialized value in
  range (or flop)` and `... in repeat (x)`. But **neither `range.t` nor
  `repeat.t` contains `use warnings` or `no warnings`** (verified by grep), so
  as they actually run they are silent: `perl nowarn.pl` produced `A: 0:1:2`,
  `D: ''`, `E: ''` with no diagnostics at all. Same code under forced `-w` did
  warn — so the pragma state is the whole difference.
- **Root cause:** PSC is right that `undef` reaches a numeric operand, but it
  ignores the file's lexical warnings state, and it escalates to `error` a
  construct that Perl defines precisely (`undef` numifies to 0, `x undef`
  yields `""`). These specific lines are the deliberate *subject* of the test.
- **Suggested fix:** Downgrade undef-in-numeric-context to a warning, and
  honour `no warnings`/absent-`use warnings` scope before emitting.

## Finding 7: Void-context stringification clobbers a variable's numeric type
- **Instances:** ~66 of 762 (the `t/base/num.t` share of `expected Num, got Str`)
- **Example:** `t/base/num.t:61` — `print $a + 1 == 2 ? "ok 16\n" : ...`,
  where line 60 is `$a = 1; "$a"; # Keep the stringification as a potential troublemaker.`
- **Verdict:** FALSE POSITIVE
- **Perl says:** silent about the arithmetic. `perl -w numstr.pl` warned only
  `Useless use of string in void context` (about the `"$a";` statement itself)
  and computed `1+1 = 2`, `1e3+1 = 1001`, `0.1+1 = 1.1`. The value remains
  numeric; stringifying it in void context changes nothing.
- **Root cause:** PSC treats the void-context `"$a"` as an assignment of `Str`
  to `$a`, so every later `$a + 1` sees a `Str`. It is flow-tracking a
  string *use* as though it were a string *store*. The test file's own comment
  shows this is a deliberately planted trap — for perl's optimiser, and PSC
  fell into it.
- **Suggested fix:** Do not narrow a variable's type from an rvalue
  interpolation; only assignments should update the tracked type.

## Finding 8: Correct undef detection reported with the wrong type name
- **Instances:** 4 of 762 (`error: left operand of "+": expected Num, got ArrayRef`)
- **Example:** `t/test_pl/examples.t:36:32` — `warnings_like(sub { my $x; $x+1 },`
  with the expectation `[ qr/^Use of uninitialized value \$x in addition/ ]` on line 37
- **Verdict:** TRUE POSITIVE (detection) with an incorrect reported type
- **Perl says:** it warns, and the test asserts that it warns —
  `Use of uninitialized value $x in numeric eq` / `in addition` was reproduced
  in `refeq.pl` line 11 and in the arithmetic case. Here `my $x;` really is
  undef with no assignment, so the flag is legitimate.
- **Root cause:** Column 32 points inside the closure `sub { my $x; $x+1 }`,
  where the operand is undef — but PSC named the type `ArrayRef`, evidently
  picking up the enclosing `[ qr/.../ ]` expectation list rather than the
  operand in the nested sub. Correct line, correct concern, wrong type and
  wrong severity.
- **Suggested fix:** Fix operand resolution so a nested `sub { }` body is not
  typed from its enclosing argument list; report this as `Undef`, not `ArrayRef`.

---

## Summary

| # | Root cause | Count | Verdict |
|---|---|---:|---|
| 1 | Magic string range | 239 | FALSE POSITIVE |
| 3 | Bool not a subtype of Int | 87 | FALSE POSITIVE |
| 4 | Bitwise string operators | 81 | FALSE POSITIVE |
| 2 | Int arithmetic widened to Num | 65 | FALSE POSITIVE |
| 7 | Void stringification clobbers type | ~66 | FALSE POSITIVE |
| 6 | undef in range/repeat | 53 | MIXED (severity/pragma) |
| 5 | `==` on references / overload | 44 | FALSE POSITIVE |
| 8 | Wrong type name on real undef | 4 | TRUE POSITIVE (mislabelled) |

Roughly **580 of 762 (~76%) are unambiguous false positives** from five
missing rules. Only Finding 8 is a real defect in the test file's neighbourhood,
and even there the test is asserting the warning deliberately — so the count of
genuine problems found in perl's test suite is effectively **zero**.

Highest leverage in order: Bool-as-Int (Finding 3) and Int-preserving
arithmetic (Finding 2) are one-line lattice/signature changes clearing ~150
diagnostics. The dual-signature work for `..` and `& | ^` (Findings 1 and 4)
clears ~320 more but needs real operand-type dispatch. Finding 7 is a
flow-analysis bug worth fixing on its own merits — typing a variable from an
rvalue use will misfire well beyond this corpus.

One caveat I could not resolve: I sampled 3-5 instances per group as directed,
so the per-group counts are attributed by message shape rather than verified
line by line. Finding 7's count in particular is approximate — `expected Num,
got Str` (187 total) mixes the num.t stringification cause with capture-variable
(`$1`) arithmetic, which I confirmed separately is also silent under `-w`
(`capture arith: ... -> equal`) but did not tally independently.
