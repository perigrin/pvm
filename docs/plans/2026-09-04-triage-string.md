# Triage: `expected Str` / `expected Regex` coercion-mismatch diagnostics

123 diagnostics; 94 carry a file:line (the remaining 29 lines in the input are
shell noise from a broken login profile, not PSC output).

Every Perl claim below was executed with `/home/perigrin/.local/bin/perl`
(v5.42.0) under `-w`. Witness scripts are in
`/tmp/claude-1001/-home-perigrin-dev-pvm/a75eda1c-4653-49a5-bee9-f4a5f57eebbb/scratchpad/w/`.

**Headline:** the overwhelming majority are false positives, and they trace to
three mechanisms, not thirty. One signature-table entry
(`internal/types/signatures.go:178`) accounts for the single largest group;
one lexer bug (`/x` regex modifier read as the `x` operator) and one parser
bug (`format` blocks) account for most of the rest. The one group where PSC
genuinely agrees with perl is undef-in-string-context.

---

## Finding 1: `=~` requires a compiled Regex, but Perl compiles strings into patterns
- Instances: 33 (`expected Regex, got Str`: 25; `got Int`: 6; `got Bool`: 2 — all the same cause)
- Example: `/home/perigrin/dev/perl5/t/uni/greek.t:53` — `ok("\xC1"    =~ '\xC1',       '\xC1 to \'\xC1\'');`
- Also: `t/op/coreamp.t:89` — `elsif ($p =~ '^;([$*]+)\z') {`; `t/uni/latin2.t:123` — `ok("\xC1" =~ $re, ...)` where `$re` is a plain string; `t/comp/fold.t:166` — `$@ =~ "Modification of a read-only value attempted at"`
- Verdict: FALSE POSITIVE
- Perl says: all forms match, silently. `perl -w w/regex_str.pl` printed `A: match` through `E: match`. That covers a single-quoted literal pattern (greek.t), a pattern in a scalar variable (latin2.t), an interpolated double-quoted pattern (pat_advanced.t), and a bare string on the right of `=~` (fold.t). The only warning emitted was `Useless (?g)`, which is the diagnostic `t/re/pat_advanced.t:132` deliberately provokes and asserts on — so PSC is not even flagging the thing perl objects to there.
  The `got Int` / `got Bool` variants are the same rule reached by a different path: in `t/comp/fold.t:107` (`ok scalar $jing =~ (1 ? /foo/ : /bar/)`) and `t/re/pat.t:1195` (`ok('a1b' =~ ('xyz' =~ /y/))`) the right operand is the *result* of an inner match, which Perl stringifies and compiles like any other string. `perl -w w/fold.pl` → `A: result=1`, `B: result=1`, `C: result=1`, `D: result=1`, no warnings. (D is true because the empty pattern reuses the last successful one — exactly the semantics `pat.t` is asserting.)
- Root cause: `internal/types/signatures.go:178` declares `"=~": {Left: Str, Right: Regex, Result: Bool}` (and `!~` on line 179). `Str` is not a subtype of `Regex`, and `internal/types/coercion.go` has no `Regex` entry in `coercionTargets` — so the mismatch is unavoidable for every string pattern. The signature encodes a language that requires `qr//`, which is not Perl.
- Suggested fix: widen the right operand of `=~`/`!~` to `Regex|Str` (Perl accepts any stringifiable scalar as a pattern), rather than adding Regex as a coercion target.

## Finding 2: `/x` regex modifier lexed as the binary `x` repetition operator
- Instances: 6 (5 `left operand of "x": expected Str, got Bool`, 1 `got Regex`)
- Example: `/home/perigrin/dev/perl5/t/re/reg_eval_scope.t:273` — `$l = __LINE__; "1" =~ /^1$c/x and warn "foo";`
- Also: `t/loc_tools.pl:67` — `elsif (   $number !~ / ^ -? \d+ $ /x`
- Verdict: FALSE POSITIVE
- Perl says: `perl -w w/xmod.pl` → `A: yes`, `B: match`, no warnings. There is no `x` operator on these lines at all; `/x` is the extended-whitespace modifier closing the match. The "left operand" PSC reports is the match's Bool result, which tells you it has consumed the `/` as a pattern terminator and then read the following `x` as an infix operator.
- Root cause: the lexer does not treat a character in the modifier position after a regex terminator as part of the regex literal. Note this group is only visible where the modifier is `x`; the same bug is silent for `/i`, `/g`, `/m`, `/s` because those letters are not operators — so the true blast radius is wider than 6, this is just where it surfaces as a type error.
- Suggested fix: consume the full trailing modifier run (`[msixpodualngcer]*`) as part of the regex token before resuming operator lexing.
- Caveat: `t/op/inc.t:254` — `[(qr/.../) x 2]` — is a *real* `x` operator applied to a Regex, so it is a distinct (also false) positive: `perl -w` built the 2-element list silently (`C: 2 [Regexp]`). Perl's `x` in list context repeats elements without stringifying them, so requiring `Str` on the left is wrong there too.

## Finding 3: `format` blocks misparsed; diagnostics attributed to unrelated following lines
- Instances: 6 (5 `right operand of "."`, 1 `right operand of "." got Bool`)
- Example: `/home/perigrin/dev/perl5/t/op/rt119311.t:140` — `undef $_;`
- Also: `t/comp/decl.t:12` — `print "1..9\n";`
- Verdict: FALSE POSITIVE (and a misreported location, which is the more serious defect)
- Perl says: neither line contains a `.` operator, so the diagnostic cannot be about the code it points at. Both sit immediately after a `format NAME = ... .` block, whose picture lines are terminated by a lone `.` on its own line. `perl -w w/fmt.pl` ran a format block and the following `undef $_;` cleanly: output `hello` / `no concat happened on the undef line`, no warnings.
- Root cause: PSC does not recognise `format` block syntax and parses the terminating `.` as a concatenation operator, which then swallows the next statement as its right operand. This is why the reported line is always the statement *after* the format block.
- Suggested fix: lex `format NAME =` through the terminating lone `.` as an opaque token, the way heredoc bodies are handled.

## Finding 4: reference stringification treated as a mismatch
- Instances: 20 (`eq`/`GlobRef` 12, `.`/`ScalarRef` 3+3, `=~`/ArrayRef 1, `=~`/HashRef 1, `.`/HashRef 1, `.`/CodeRef via postfixderef 1, `.`/Regex 2 — grouped by shared cause)
- Example: `/home/perigrin/dev/perl5/t/op/proto.t:530` — `print "not " unless $_[0] eq \*FOO;`
- Also: `t/op/lvref.t:135` — `is \$a[0].\$a[1], \$_.\$_, '\@array[indices]';`; `t/re/pat.t:1188` — `::ok([] =~ /^ARRAY/, "Array ref stringification")`; `t/op/qr.t:108` — `my $str = "".qr//;`
- Verdict: FALSE POSITIVE
- Perl says: `perl -w w/refstr.pl` → `A: eq`, `B: concatenated ok`, `C: match`, `D: match`, `E: match`, `F: (?^:)`. Every reference stringifies silently under `-w`. Several of these tests (`pat.t:1188` "Array ref stringification", `qr.t:108`) exist *specifically* to assert that stringification behaviour.
- Root cause: `internal/types/coercion.go` is correct here — its comment explicitly says "Every reference stringifies to `HASH(0x55f1...)`, so HashRef is coercible to Str". The problem is the *policy* in `CoercionMismatch`, which fires on `IsCoercible && !TypeSatisfies`. For a lossy-but-intended conversion like ref stringification that predicate is true by construction, so the diagnostic cannot distinguish "stringified a ref by accident" from "stringified a ref on purpose".
- Suggested fix: unclear — needs a policy decision. Ref-to-Str is coercible-but-not-subtype *by design*, so either `.`/`eq` should accept the full stringifiable domain without complaint, or this diagnostic needs a suppression for contexts where stringification is the evident intent.

## Finding 5: overloaded objects in string context
- Instances: 15 (`eq`/Object 10, `=~`/Object 5)
- Example: `/home/perigrin/dev/perl5/lib/overload_fallback.t:15` — `ok ($x eq 'stringvalue', 'fallback worked');`
- Also: `t/mro/c3_with_overload.t:40` — `ok(($y eq 'OverloadingTest stringified'), '... eq was handled correctly');`; `t/op/overload.t:21` — `my ($one) = $o =~ /(.*)/g;`
- Verdict: FALSE POSITIVE
- Perl says: `perl -w w/ovl.pl` → `A: eq`, `B: 'stringvalue'`, `C: ne`, no warnings. An object with `'""'` overloading *is* a string in every context that matters. Even a non-overloaded object compares silently (`C: ne`) — perl does not warn about objects in string context at all.
- Root cause: PSC has no model of `use overload`. Every one of these files is a dedicated overloading test, so the `Object` here is precisely the case where stringification is defined and correct. Same `CoercionMismatch` policy as Finding 4.
- Suggested fix: when a package declares `use overload '""'`, treat its instances as satisfying `Str` rather than merely coercible to it.

## Finding 6: undef in string context
- Instances: 11 (`.`/Undef 6 — but 4 of those are Finding 3's format misparse; `x`/Undef 3; `eq`/Undef 1; `=~`/Undef 1)
- Example: `/home/perigrin/dev/perl5/t/op/range.t:155` — `@foo=(); push @foo, $_ for -2..undef;`
- Also: `t/op/concat.t:83` — `ok($a eq "\x{1ff}", "bug id 20000901.092 (#4184), undef right");`; `t/re/pat_rt_report.t:37` — `ok(undef =~ /^([^\/]*)(.*)$/, ...)`
- Verdict: MIXED — leaning TRUE POSITIVE on the genuine undef cases
- Perl says: this is the one group where perl warns too. With an **explicitly assigned** `my $u = undef` (not merely an undeclared scalar), `perl -w w/undef.pl` emitted:
  - `Use of uninitialized value $u in concatenation (.) or string at line 5.`
  - `Use of uninitialized value in range (or flop) at line 8.`
  - `Use of uninitialized value in pattern match (m//) at line 11.`
  - `Use of uninitialized value $a in string eq at line 14.`
  PSC is reporting the same defect perl's own `uninitialized` warning reports, so by the brief's criterion these are true positives.
- Root cause: n/a — PSC is right. Two caveats on the *set*, though. First, the 4 `rt119311.t` instances counted under `.`/Undef are not really this finding; they are the format misparse of Finding 3 and should be subtracted. Second, the 3 `(undef) x N` instances (`t/op/repeat.t:71`, `t/op/stack.t:99`, `t/op/sub_lval.t:116`) are **false** positives: `perl -w w/xmod.pl` → `D: 2 defined0=0`, silent, because list-context `x` replicates undef elements without stringifying them. The tests themselves are deliberately exercising undef, so all of these are intentional in context — but that is a test-suite property, not a PSC error.
- Suggested fix: keep the diagnostic; exclude list-context `(undef) x N` from the `Str` requirement on `x`'s left operand.

## Finding 7: over-wide inferred types on plain string variables
- Instances: 4
- Example: `/home/perigrin/dev/perl5/t/porting/libperl.t:172` — `if ($nm_style eq 'gnu' && !defined $fake_style) {`
- Also: `libperl.t:188` — `if ($fake_input eq '-')`, whose reported type is the entire lattice (`Bool|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|NaN|Inf`)
- Verdict: FALSE POSITIVE
- Perl says: no runtime question to answer — the inferred type is refuted by the source. `grep -n 'nm_style\s*=' t/porting/libperl.t` shows every assignment is a string literal: `$nm_style = 'gnu'` (lines 114, 127, 129) and `$nm_style = 'darwin'` (116, 131). The variable is only ever a Str, so `eq` is unimpeachable.
- Root cause: the union widens to the top of the lattice rather than to the join of the observed assignments. The `libperl.t:188` type — literally every scalar type PSC knows, plus the `0x80000000` artifact in the `nm_style` union — reads as a fallback for "inference gave up" being reported as if it were a computed type. Distinct from Findings 1-6: this is an inference defect, not a signature or policy one.
- Suggested fix: unclear — needs investigation of why assignment-site information is discarded; the `0x80000000` in the union suggests a sentinel or bitmask leaking into the type display.

---

## Summary

| Finding | Instances | Verdict |
|---|---|---|
| 1. String-as-pattern rejected by `=~` | 33 | FALSE POSITIVE |
| 4. Reference stringification | 20 | FALSE POSITIVE |
| 5. Overloaded objects in string context | 15 | FALSE POSITIVE |
| 6. undef in string context | 11 | MIXED (mostly TRUE) |
| 2. `/x` modifier lexed as `x` operator | 6 | FALSE POSITIVE |
| 3. `format` block misparse | 6 | FALSE POSITIVE |
| 7. Over-wide inferred types | 4 | FALSE POSITIVE |

Roughly 85 of 94 located diagnostics are false positives.

Highest-value fixes, in order:
1. **`signatures.go:178-179`** — widen `=~`/`!~` right operand to `Regex|Str`. One line, clears 33.
2. **`CoercionMismatch` policy** — ref and overloaded-object stringification are coercible-by-design; firing on `IsCoercible && !TypeSatisfies` cannot distinguish intent. Clears ~35 across Findings 4 and 5.
3. **Regex modifier lexing** — correctness bug beyond these diagnostics; `/x` is merely where it becomes visible.
4. **`format` block lexing** — misattributes diagnostics to unrelated lines, which undermines trust in every location PSC reports.
