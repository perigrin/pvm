---
title: "Audit fixes, POSIX: failure returns admit Undef, math returns admit Inf and NaN, classifiers take them, optional parameters match perl"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.506751465Z
updated: 2026-10-20T20:21:06.052053702Z
sessions:
- start_sha: 9cbd111a2a7e8a7f3e4a7d6790f10c685f1121fe
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T20:21:06.052053702Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T20:21:06.052053702Z
---

crochet:review of PR #488, round 1 (2026-10-21). Signature audit (01a15aaa), group A.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Findings (measured; from the audit, perl 5.42.0)

#### Audit group A: POSIX.tpm

File: internal/parse/declarations/POSIX.tpm. Judge: perl 5.42.0 (/home/perigrin/.local/bin/perl), POSIX 2.23 and its POSIX.pod.
Probes: scratchpad/audit-A/{probe.pl,probe2.pl,cprobe.pl,usage.txt,none-results.txt}.
- Arity: every non-exempt function was called with 12 args; the croak names the XS or %reimpl signature, with defaults (usage.txt).
- Returns: probe.pl and probe2.pl record list and scalar context, IV/NV/PV flags, is_bool and undef.
- "0 but true" counts as Num here, matching internal/infer/testdata/observe_types.pl (looks_like_number).

#### Coverage
- Functions: all 190 non-None function names were checked, plus the 3 None-typed process enders (_exit, abort, exit).
- Constants: all 464 `()` constants were measured, not just a sample (cprobe.pl). Each one matches its declaration: 426 Int, 19 Num, 2 Boolean, 3 Inf, 2 NaN and 12 None. No findings.
- None subs: 106 subs are declared None. 103 of them were called with (), (1,2,3) and ('a'), and every call died (none-results.txt, 309/309).
  - Of those 103: 77 die with "Unimplemented", 12 with "vendor has not defined", and 14 with "is not a valid POSIX macro".
  - _exit, abort and exit were not called, because they end the process; None is correct for all three by POSIX.pm and the XS.

#### Re-judged under the brief's CORRECTION (RFC 0001, "What a parameter type means")
None of the parameter findings below is a "perl coerces it, so widen it" claim. Each one keeps its place under rule (a), or is an arity or optional-parameter finding that the correction leaves standing.
- **F3:** isnan and its siblings coerce their argument to a C double, which preserves NaN and Inf. Num therefore isn't what these operations coerce to or preserve, and isnan(NaN) would be reported as a lossy coercion although nothing is lost. The same reasoning holds in principle for every NV-taking function. It is reported only for the classifiers and payload functions, where NaN and Inf are the point of the call. Confidence lowered from 90 to 80.
- **F4 (mblen, mbtowc):** undef is a documented distinct meaning (NULL, which resets the shift state), not a coerced string, so the operation preserves Undef.
- **F5 (isatty, chdir):** both operations coerce to or preserve a filehandle. -t and CORE::chdir use the handle itself, not its stringification.
- **F5 (setpayload):** $nv is an output lvalue that is written, not coerced. The Void return stands on its own.
- **F5 (utime):** undef means "now"; it is preserved, not coerced to 0.

#### Findings

#### F1. Documented failure returns missing Undef: SysRet/SysRetLong and other undef-on-failure subs
The T_SYSRET typemap returns undef when the call yields -1 and "0 but true" when it yields 0. The POD for each says "Returns undef on failure".

| sub | line | .tpm | perl (measured) | fix | conf |
|---|---|---|---|---|---|
| access | 547 | Num | access("/nonexist",0) -> undef | Num\|Undef | 95 |
| close | 575 | Num | close(99) -> undef | Num\|Undef | 95 |
| dup | 586 | Int | dup(99) -> undef | Num\|Undef | 95 |
| dup2 | 587 | Int | dup2(99,98) -> undef | Num\|Undef | 95 |
| open | 720 | Int | open("/nonexist") -> undef | Num\|Undef | 95 |
| creat | 580 | Int | creat("/nonexist/x",0) -> undef (via &open) | Num\|Undef | 95 |
| fcntl | 603 | Int | on a closed fh -> undef; success "0 but true" | Num\|Undef | 90 |
| lchown | 683 | Num | lchown(0,0,"/nonexist") -> undef | Num\|Undef | 95 |
| mkfifo | 710 | Int | mkfifo("/nonexist/x",0) -> undef | Num\|Undef | 95 |
| read | 733 | Int | read(99,$b,1) -> undef; EOF -> "0 but true" | Num\|Undef | 95 |
| write | 837 | Int | write(99,"x",1) -> undef | Num\|Undef | 95 |
| fpathconf | 621 | Int | fpathconf(99,0) -> undef | Num\|Undef | 95 |
| pathconf | 722 | Int | pathconf("/nonexist",0) -> undef | Num\|Undef | 95 |
| sysconf | 804 | Int | sysconf(99999) -> undef | Num\|Undef | 95 |
| setuid | 757 | Boolean | setuid($<) -> "0 but true"; setuid(0) as user -> undef; never is_bool | Num\|Undef | 95 |
| setgid | 750 | Boolean | setgid($(+0) -> "0 but true" | Num\|Undef | 95 |
| setpgid | 755 | Int | setpgid(999999999,0) -> undef | Num\|Undef | 95 |
| nice | 718 | Int | nice(0) -> "0 but true"; POD: undef on failure; the XS pushes undef | Num\|Undef | 85 |
| tcdrain/tcflow/tcflush/tcsendbreak/tcsetpgrp | 808,809,810,813,815 | Int | each on fd 99 -> undef | Num\|Undef | 95 |
| sigaction | 759 | Int | query -> "0 but true"; SysRet | Num\|Undef | 85 |
| sigpending | 763 | Int | -> "0 but true"; SysRet | Num\|Undef | 85 |
| sigsuspend | 766 | Int | SysRet alias of sigpending (POSIX.xs ALIAS sigsuspend = 1). sigsuspend(2) returns only -1, so the sub returns undef | Undef (or Num\|Undef) | 75 |
| pause | 723 | Int | SysRet pause(); pause(2) returns only -1, so the sub returns undef. POD: "Returns undef on failure" | Undef (or Num\|Undef) | 75 |
| mktime | 711 | Num | POD: "Returns C<undef> on failure". Measured mktime(0,0,0,1,0,70) -> "0 but true". Undef was not reproduced on glibc | Num\|Undef | 65 |
| getcwd | 641 | Str | in a removed cwd -> undef | Str\|Undef | 95 |
| getgrgid (scalar) | 646 | Str | getgrgid(99999) -> undef | Str\|Undef | 95 |
| getgrnam (scalar) | 648 | Int | getgrnam("nxgrp") -> undef | Int\|Undef | 95 |
| getpwnam (scalar) | 656 | Int | getpwnam("nxusr") -> undef | Int\|Undef | 95 |
| getpwuid (scalar) | 658 | Str | getpwuid(99999) -> undef | Str\|Undef | 95 |

sigprocmask (764), which is already Num|Undef, is correct.

#### F2. Math returns are declared Num but produce Inf or NaN for finite Num arguments
CORE.tpm already sets the precedent: `sub exp (...) Num|Inf;`, against POSIX.tpm's `exp ... Num`. Every value below was measured with finite arguments.

| sub | line | measured | fix |
|---|---|---|---|
| exp | 598 | exp(1000) = Inf | Num\|Inf |
| exp2 | 599 | exp2(2000) = Inf | Num\|Inf |
| expm1 | 600 | expm1(1000) = Inf | Num\|Inf |
| cosh | 579 | cosh(1000) = Inf | Num\|Inf |
| sinh | 768 | sinh(1000) = Inf | Num\|Inf |
| ldexp | 684 | ldexp(1,2000) = Inf | Num\|Inf |
| scalbn | 747 | scalbn(1,5000) = Inf | Num\|Inf |
| hypot | 664 | hypot(1.7e308,1.7e308) = Inf | Num\|Inf |
| fdim | 604 | fdim(1.7e308,-1.7e308) = Inf | Num\|Inf |
| fma | 615 | fma(1e308,10,0) = Inf | Num\|Inf |
| atanh | 557 | atanh(1) = Inf | Num\|Inf\|NaN |
| log10 | 692 | log10(0) = -Inf; log10(-1) = NaN | Num\|Inf\|NaN |
| log2 | 694 | log2(0) = -Inf | Num\|Inf\|NaN |
| log1p | 693 | log1p(-1) = -Inf | Num\|Inf\|NaN |
| logb | 695 | logb(0) = -Inf | Num\|Inf\|NaN |
| y0 | 838 | y0(0) = -Inf | Num\|Inf\|NaN |
| y1 | 839 | y1(0) = -Inf | Num\|Inf\|NaN |
| yn | 840 | yn(1,0) = -Inf | Num\|Inf\|NaN |
| lgamma | 686 | lgamma(0) = Inf | Num\|Inf |
| tgamma | 816 | tgamma(0) = Inf; tgamma(-1) = NaN | Num\|Inf\|NaN |
| pow | 726 | pow(10,400) = Inf; pow(-1,.5) = NaN | Num\|Inf\|NaN |
| acos | 548 | acos(2) = NaN | Num\|NaN |
| asin | 552 | asin(2) = NaN | Num\|NaN |
| acosh | 549 | acosh(0) = NaN | Num\|NaN |
| fmod | 618 | fmod(1,0) = NaN | Num\|NaN |
| remainder | 737 | remainder(1,0) = NaN | Num\|NaN |
| remquo (list) | 739 | remquo(1,0) = (NaN, 21856) | List[Num\|NaN] |
| strtod (both multis) | 794-795 | strtod("inf") = Inf; "nan" parses as NaN | Num\|Inf\|NaN; List[Num\|Inf\|NaN] |
| strtold (both multis) | 799-800 | same as strtod | Num\|Inf\|NaN; List[Num\|Inf\|NaN] |

Confidence: 90.

#### F3. Classification and NaN-payload functions reject NaN and Inf, the values they exist to classify
- Affected: fpclassify 622, isfinite 667, isinf 670, isnan 674, isnormal 675, issignaling 676, signbit 762, isgreater/isgreaterequal/isless/islessequal/islessgreater/isunordered 668-677, ilogb 665 and getpayload 652.
- Each takes `Num $x`, and types.go excludes NaN and Inf from Num.
- Measured: isnan(nan()) = 1, fpclassify(9**9**9) = 1 and getpayload(nan(7)) = 7.
- POD on getpayload: the payload is read from a NaN made by nan(). The oracle's own getpayload case passes POSIX::nan().
- Fix: `Num|NaN|Inf $x`. Confidence: 80.

#### F4. Missing optional parameters or defaults (perl accepts these arguments and the POD documents them)
| sub | line | .tpm | perl / POD | fix | conf |
|---|---|---|---|---|---|
| nan | 714 | `()` | Usage "POSIX::nan(payload = 0)"; POD: "With an argument, can be used to generate a NaN with payload". nan(5) works | (Num $payload = 0) NaN | 95 |
| open | 720 | 3 required | Usage "open(filename, flags = O_RDONLY, mode = 0666)"; POD example `POSIX::open("foo")` | `Int $flags = O_RDONLY, Int $mode = 0666` | 95 |
| strtol / strtoul | 797-798, 801-802 | `Int $base` required | Usage "strtol(str, base = 0)"; POD: "When the base is zero or omitted"; strtol("42x") works | `Int $base = 0` | 95 |
| mblen | 701 | (Str $s, Int $n) | Usage "mblen(s, n = ~0)"; POD: "Use undef as the first parameter"; mblen(undef) -> 0 | (Str\|Undef $s, Int $n = ~0) | 90 |
| mbtowc | 703 | (Any, Str $s, Int $n) | Usage "mbtowc(pwc, s, n = ~0)"; POD: "Use undef as the second parameter"; mbtowc(undef,undef) -> -1 | (Any $pwc, Str\|Undef $s, Int $n = ~0) | 90 |
| perror | 724 | (Str $message) required | Pure Perl: `print STDERR "@_: " if @_`. perror() and perror("a","b") both return 1 | (List[Str] @message) | 80 |
| asctime / mktime / strftime | 551, 711, 785 | `List[Int] @time` (strftime adds `Str $fmt`) | Usage "asctime(sec, min, hour, mday, mon, year, wday = 0, yday = 0, isdst = -1)". strftime is the same after fmt, with wday = -1, yday = -1 and isdst = 0. Fewer than 6 or more than 9 args croaks "Usage", and the slurpy admits both | 6 required + 3 optional Int, if the checker still accepts `strftime($f, localtime)` flattening; otherwise keep the slurpy and note why | 65 |

#### F5. Parameter types narrower than what perl and the POD accept
| sub | line | .tpm | perl / POD | fix | conf |
|---|---|---|---|---|---|
| isatty | 666 | (Int $filehandle) | POD: "whether the specified filehandle is connected to a tty". Under a pty (script -qc), isatty(\*STDIN), isatty(0) and isatty("STDIN") all return true | Int\|FileHandle (\|Str) | 90 |
| chdir | 570 | (Str $directory) | %reimpl calls CORE::chdir($_[0]). `POSIX::chdir($dirhandle)` changed to /tmp. CORE.tpm:157 declares `Str\|FileHandle $dir` | Str\|FileHandle | 75 |
| setpayload / setpayloadsig | 753, 754 | (Num $nv, Int $payload) Boolean\|Undef | POD: `setpayload($var, $payload)` "Sets the NaN payload of var". $nv is an output lvalue: `my $x; setpayload($x,5)` leaves $x = NaN, and `my $x = 1.5` works too. The XS is void and returns the empty list | (Any $nv, Int $payload) Void | 80 |
| utime | 829 | (Str, Int $atime, Int $mtime) | %reimpl calls CORE::utime($_[1],$_[2],$_[0]); POD: "identical to Perl's builtin utime()". `POSIX::utime("f", undef, undef)` succeeds under -w with no warning (it means "now") | Int\|Undef $atime/$mtime | 65 |

#### Counts
- POSIX.tpm: 5 finding classes covering 102 sub-signature defects. Lines with more than one defect, such as open, mktime and strtod, are counted once per class.
  - F1 (missing Undef): 33.
  - F2 (missing Inf/NaN in returns): 31.
  - F3 (NaN/Inf rejected by classifiers): 15.
  - F4 (missing optional params, slurpy arity): 12 (nan, open, strtol ×2, strtoul ×2, mblen, mbtowc, perror, asctime, mktime, strftime).
  - F5 (params too narrow): 5 (isatty, chdir, setpayload, setpayloadsig, utime).
- None subs: 0 findings. Constants: 0 findings (464/464 measured).

AUDIT: FINDINGS

## Steps
- [ ] Re-measure each finding before fixing; withdraw any that does not reproduce, saying why in the commit
- [ ] Add oracle cases for each changed signature
- [ ] Hold tests, lib baseline and precision unchanged or better

## Acceptance criteria

### Positive Scenarios
- [ ] TestPOSIXFailureReturnsAdmitUndef (`go test ./internal/infer/ -run '^TestPOSIXFailureReturnsAdmitUndef$' -count=1 -v | grep -q '^--- PASS: TestPOSIXFailureReturnsAdmitUndef'`)
- [ ] TestPOSIXMathReturnsAdmitInfNaN (`go test ./internal/infer/ -run '^TestPOSIXMathReturnsAdmitInfNaN$' -count=1 -v | grep -q '^--- PASS: TestPOSIXMathReturnsAdmitInfNaN'`)
- [ ] TestPOSIXClassifiersTakeInfNaN (`go test ./internal/infer/ -run '^TestPOSIXClassifiersTakeInfNaN$' -count=1 -v | grep -q '^--- PASS: TestPOSIXClassifiersTakeInfNaN'`)
- [ ] TestPOSIXPrototypes (`go test ./internal/parse/ -run '^TestPOSIXPrototypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXPrototypes'`)
- [ ] TestPOSIXExports (`go test ./internal/parse/ -run '^TestPOSIXExports$' -count=1 -v | grep -q '^--- PASS: TestPOSIXExports'`)
- [ ] TestPOSIXReturnTypes (`go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestPOSIXOptionalParametersMatchPerl (`go test ./internal/infer/ -run '^TestPOSIXOptionalParametersMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestPOSIXOptionalParametersMatchPerl'`)
