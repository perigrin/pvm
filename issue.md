---
title: "Audit fixes, XS utilities and libraries: Storable flags, List::Util returns, Cwd undef, too-wide parameters, scalar-context returns"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.636678343Z
updated: 2026-10-21T02:43:05.994920641Z
sessions:
- start_sha: dc52c671706a020f6d843153446dac0efa30792b
  end_sha: 767b931606fcd66f615e628013510ad201e28d83
  commits: 8
  started_at: 2026-10-20T22:31:19.013943179Z
  ended_at: 2026-10-21T02:43:05.994920641Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T22:31:19.013943179Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T02:43:05.994920641Z
observed_paths:
- internal/infer/declared_test.go
- internal/infer/oraclestderr_test.go
- internal/infer/parametertypes_test.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Archive-Tar.pl
- internal/infer/testdata/oracle/Compress-Raw-Zlib.pl
- internal/infer/testdata/oracle/Cwd.pl
- internal/infer/testdata/oracle/Data-Dumper.pl
- internal/infer/testdata/oracle/Encode.pl
- internal/infer/testdata/oracle/File-Spec-Functions.pl
- internal/infer/testdata/oracle/File-Spec-Unix.pl
- internal/infer/testdata/oracle/File-Spec.pl
- internal/infer/testdata/oracle/File-Temp.pl
- internal/infer/testdata/oracle/HTTP-Tiny.pl
- internal/infer/testdata/oracle/IO-Handle.pl
- internal/infer/testdata/oracle/IO-Select.pl
- internal/infer/testdata/oracle/IO-Socket-IP.pl
- internal/infer/testdata/oracle/IO-Socket.pl
- internal/infer/testdata/oracle/List-Util.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/Scalar-Util.pl
- internal/infer/testdata/oracle/Socket.pl
- internal/infer/testdata/oracle/Storable.pl
- internal/infer/testdata/oracle/Symbol.pl
- internal/infer/testdata/oracle/Term-ReadLine.pl
- internal/infer/testdata/oracle/Tie-Hash.pl
- internal/infer/testdata/oracle/Time-HiRes.pl
- internal/infer/testdata/parameter_oracle.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/declarations/Archive/Tar.tpm
- internal/parse/declarations/Compress/Raw/Zlib.tpm
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/Data/Dumper.tpm
- internal/parse/declarations/Encode.tpm
- internal/parse/declarations/ExtUtils/Manifest.tpm
- internal/parse/declarations/File/Spec.tpm
- internal/parse/declarations/File/Spec/Functions.tpm
- internal/parse/declarations/File/Spec/Unix.tpm
- internal/parse/declarations/File/Temp.tpm
- internal/parse/declarations/FileHandle.tpm
- internal/parse/declarations/HTTP/Tiny.tpm
- internal/parse/declarations/IO/File.tpm
- internal/parse/declarations/IO/Handle.tpm
- internal/parse/declarations/IO/Select.tpm
- internal/parse/declarations/IO/Socket.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/List/Util.tpm
- internal/parse/declarations/Net/Ping.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Storable.tpm
- internal/parse/declarations/Symbol.tpm
- internal/parse/declarations/Term/ReadLine.tpm
- internal/parse/declarations/Tie/Hash.tpm
- internal/parse/declarations/Time/HiRes.tpm
- internal/parse/heldtoperl_test.go
- internal/perltest/constantdifferences.go
- internal/perltest/constantdifferences_test.go
- internal/perltest/pathunder.go
- internal/perltest/pathunder_test.go
---

crochet:review of PR #488, round 1 (2026-10-21). Signature audit (01a15aaa), group D.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

D's report predates the parameter-rule correction: re-judge its parameter findings under the rule above before fixing (e.g. sum("abc") and encode with a ref are lossy coercions, withdrawn; max returning its argument unchanged is a return finding and stands). Quality S6 (Cwd chdir default argument) is here.

## Findings (measured; from the audit, perl 5.42.0)

#### .tpm audit, group D

Judge: /home/perigrin/.local/bin/perl (5.42.0). Probes: scratchpad/audit-D/{lu,su,en2,st,th,mb,cw,sy}.pl.
Paths are relative to internal/parse/declarations/. Format: sub, file:line, what the .tpm says, what perl/POD shows, fix, confidence.

#### List::Util (8)

1. **sum (multi)**, List/Util.tpm:42-43. Candidates are `(Num $x, List @rest) Num` and `() Undef`. Together they leave out every non-empty call whose first argument is not Num. Measured: `sum("abc")`=0, `sum(undef)`=0, `sum(true)`=1, `sum([1])`=the address. All return without dying. The return type `Num` also misses `sum("inf")`=Inf, `sum("nan")`=NaN and `sum(1e308,1e308)`=Inf. An object with overloaded `+` comes back as that object: `ref(sum($o,1))` is "O". Fix: `(Scalar $x, List @rest) Num|Inf|NaN|Object`, and keep `() Undef`. Confidence 85.
2. **sum0**, :44. Says `Int`. Measured `sum0(1.5)`=1.5 and `sum0("inf")`=Inf, and an overloaded object is returned as itself. Fix: `Num|Inf|NaN|Object`. Confidence 90.
3. **product**, :49. Says `Num`. Measured `product("inf")`=Inf, and an overloaded object is returned as itself. Fix: `Num|Inf|NaN|Object`. Confidence 75.
4. **max/min**, :45-46. Say `Num|Undef`. Both return the winning argument itself, without numifying it. Measured `max("abc")`='abc', `max("inf")`='inf', `ref max([1])`=ARRAY, and an overloaded object comes back as itself. Fix: `Scalar` (or `Any`). Confidence 85.
5. **maxstr/minstr**, :47-48. Say `Str|Undef`. Measured `ref maxstr([1])`=ARRAY, and `maxstr(true)` is a builtin bool. Each returns the element unchanged. Fix: `Scalar`. Confidence 80.
6. **head/tail/sample**, :62-64. `Int $n` rejects arguments that perl numifies without dying. Measured `head("2",1,2,3)` gives 2 elements, `head(1.7,...)` gives 1, `head(undef,...)` gives 0, `head("abc",...)` gives 0, and `sample("2",...)` and `sample(1.5,...)` behave the same way. Fix: `Str|Undef $n`, or keep Int if Str→Int coercion is meant to be diagnosed (types.go TypeSatisfies treats that gap as a diagnostic). Confidence 60.
7. **Scalar-context returns missing**, :24,28,29,30,54-57,92,93. POD documents scalar-context results for several of these subs, and a single `List` return does not state them:
   - `uniq`, `uniqnum`, `uniqstr` and `uniqint` return a count ("In scalar context, returns the number of elements").
   - `pairmap` returns a count. `pairgrep` returns the number of pairs.
   - `pairfirst` returns "a simple boolean". Measured: BOOL(1), or undef when nothing matches.
   - Measured, not in POD: `pairkeys` and `pairvalues` give the last element, `reductions` gives the last accumulation, and `shuffle` gives an element.

   Fix: add scalar-context multis the way IO::Select and Data::Dumper do: `Int` for the uniq\*, pairmap and pairgrep subs, and `Boolean|Undef` for pairfirst. Confidence 70.
8. **first**, :13, and **reduce**, :23. Both return `Any`. No defect found. Listed only to record that they were checked.

#### Scalar::Util (4)

1. **dualvar**, Scalar/Util.tpm:16. `Num $num` (and `Str $str`). Measured `dualvar("abc","x")` works, `dualvar(undef,undef)` returns a defined value, and `dualvar([1],[2])` works. Any scalar is accepted. Fix: `Scalar $num, Scalar $str`. Confidence 75.
2. **set_prototype**, :61. `Str $proto` rejects undef. POD: "If $prototype is undef, then the function's prototype is removed". Measured: `set_prototype(sub{1}, undef)` leaves no prototype. Fix: `Str|Undef $proto`. Confidence 90.
3. **openhandle**, :32. Says `GlobRef|Undef`. Measured:
   - `openhandle(*STDIN)` returns the bare glob '\*main::STDIN' (`ref \` is GLOB, a Glob and not a GlobRef).
   - `ref openhandle(IO::File->new(...))` is IO::File.
   - `ref openhandle(*STDIN{IO})` is IO::File.

   Fix: `Glob|GlobRef|Object|Undef`. Confidence 90.
4. **weaken/unweaken**, :69-70. `Any $ref` is too wide. Measured `weaken($x)` with $x=1 dies "Can't weaken a nonreference", `unweaken` dies the same way, and `weaken(5)` dies. `weaken(undef)` is fine. Fix: `Ref|Undef $ref`. Confidence 80.

Checked and clean: the refaddr/reftype multis (Ref ∪ Any covers everything; measured `reftype(*STDIN)` is undef, and the Any candidate handles that), blessed, isdual, isvstring, isweak, looks_like_number (all BOOL), readonly and tainted (Int, readonly(5)=134283264).

#### Encode (6)

1. **encode (multi)**, Encode.tpm:58-59. The `$string` candidates are `Str` and `Str|Undef`. Neither covers Boolean, DualVar or Ref, and perl stringifies all three: measured `encode("utf8",[1])`='ARRAY(0x...)' and `encode("utf8",true)`='1'. Fix: candidate 1 takes `Str|Boolean|DualVar|Ref` and returns Str, candidate 2 takes `Scalar` and returns Str|Undef. Confidence 75. The same narrowing applies to decode:45, bytes2str:37, str2bytes:38, decode_utf8:46 and encode_utf8:64: measured `decode("utf8",[1])`, `decode_utf8([1])` and `encode_utf8([1])` all return a string.
2. **$check parameter**, :37,38,45,46,58,59,78. `Int|CodeRef|Undef` rejects values that perl accepts. Measured `encode("ascii",$x,"1")`, `encode("ascii",$x,true)`, `encode("ascii",$x,1.5)`, `encode("ascii",$x,"abc")` and `encode("ascii",$x,[1])` all return 'x' without dying. Fix: `Str|Boolean|CodeRef|Undef` at minimum. Confidence 65.
3. **find_encoding**, :73 (and find_mime_encoding, :74). `Boolean $skip_external = 0` rejects Int. Measured `find_encoding("utf8",1)` and `find_encoding("utf8","x")` both return Encode::utf8. TypeSatisfies is subset-based, so an Int argument fails a Boolean parameter, and the declared default `0` is itself an Int. Fix: `Scalar $skip_external = 0`. Confidence 70.
4. **perlio_ok**, :89. Says `Int`, and the comment claims "a plain 0 or 1, not a boolean". Measured `perlio_ok("ascii")` is BOOL(1) (is_bool true). `perlio_ok("utf8")` is '1'. Fix: `Int|Boolean`. Confidence 85.
5. **Missing names**: is_8bit, is_16bit, utf8_upgrade and utf8_downgrade. All four are in @EXPORT_OK and `%EXPORT_TAGS{all}`, and `use Encode qw(is_8bit)` installs a stub (`exists &Encode::is_8bit` is 1), but every call dies with "Undefined subroutine &Encode::is_8bit". The .tpm declares none of them, so a `use Encode ':all'` caller gets an unresolved name. Fix: `sub is_8bit :export_ok(all) (List @args) None;` and the same for the other three, following the DEF_WBITS precedent. Confidence 70.
6. **from_to**, :78. `Str $to_encoding`/`$from_encoding` reject objects. Measured `from_to($s, "latin1", find_encoding("utf8"))`=1 and `from_to($s, find_encoding("latin1"), "utf8")`=1. Every other ENCODING parameter in this file is `Str|Object`. Fix: `Str|Object`. Confidence 70.

Checked and clean: the encodings multis (scalar context is undef, as stated), clone_encoding, resolve_alias, is_utf8 (BOOL, never undef), _utf8_on/_utf8_off, and the constants (sampled).

#### Storable (4)

1. **thaw, retrieve, lock_retrieve, fd_retrieve, retrieve_fd**, Storable.tpm:23-26,35. These omit the documented second `$flags` argument. POD: "set the 2nd argument for thaw/retrieve to 0". In the source, `_retrieve($file,$use_locking,$flags)`, `fd_retrieve { my ($file,$flags) }` and `thaw { my ($frozen,$flags) }`. Measured `thaw(freeze({}),0)` and `retrieve("f",1)` both return a HASH. Fix: add `Optional[Int] $flags`. Confidence 90.
2. **read_magic**, :45. Omits the documented second argument. POD: "=item $info = Storable::read_magic( $buffer, $must_be_file )". Fix: add `Optional[Boolean|Int] $must_be_file`. Confidence 90.
3. **store, nstore, lock_store, lock_nstore, store_fd, nstore_fd**, :11-19. Say `Boolean`. POD: "The routine returns C<undef> for I/O problems or other internal error" and "Normal errors are reported by having store() or retrieve() return C<undef>". `_store` returns undef when close() fails (`close($FILE) or undef $ret`) and when flock emulation is broken. Fix: `Boolean|Undef`. Confidence 75.
4. **store_fd, nstore_fd, fd_retrieve, retrieve_fd**, :18-19,25-26. `FileHandle` (Glob|GlobRef|IO) rejects IO::Handle objects. Measured `store_fd({a=>1}, IO::File->new(...,"w"))` is BOOL(1), and `fd_retrieve(IO::File->new(...))` is a HASH. Fix: `FileHandle|Object`. Confidence 80.

Checked and clean: dclone (dies on a non-ref, so Ref is right), freeze/nfreeze (die on a non-ref), file_magic, and the five `()` constants (2, 4, 6, 512, 256).

#### Time::HiRes (5)

1. **nanosleep**, Time/HiRes.tpm:103. Says return `Int` and param `Int`. Measured `nanosleep(1.5)` returns '1.5'. Fix: `(Num $nanoseconds) Num`. Confidence 85.
2. **clock_nanosleep**, :74. Says return `Int`. Measured `clock_nanosleep(CLOCK_REALTIME, 1500.5)` returns '1500.5'. Fix: `Num $nanoseconds`, return `Num`. Confidence 85.
3. **stat/lstat (multis)**, :97-100. `Optional[Str|Glob] $file` rejects filehandle references. Measured: scalar `stat(\*STDIN)`, `stat($lexical_fh)` and `lstat($fh)` are all BOOL(1). POD says "=item stat FH". Fix: `Optional[Str|FileHandle|Object]`. Confidence 85.
4. **utime**, :124. `List[Str] @files` rejects filehandles. POD says "As L<perlfunc/utime>", and CORE utime takes handles. Measured `Time::HiRes::utime(1.5,1.5,$fh)` is 1. Fix: `List[Str|FileHandle] @files`. Confidence 80.
5. **getitimer/setitimer scalar candidates**, :86,88. Say `Num`. Measured `scalar getitimer(999)` and `scalar setitimer(999,0)` are both undef. Fix: `Num|Undef`. Confidence 70 (measured; the POD is silent).

Checked and clean: the d_* and clock constants (sampled; the None ones die with "Your vendor has not defined"), alarm, ualarm, usleep (integer result), time, clock, gettimeofday, tv_interval, sleep, and clock_gettime/clock_getres (-1 on a bad clock is still Num).

#### Compress::Raw::Zlib (1)

1. **crc32/adler32**, Compress/Raw/Zlib.tpm:69-70. `(List @args)` is too wide. Measured `crc32()` dies "Usage: Compress::Raw::Zlib::crc32(buf, crc=CRZ_crcInitial, offset=0)", `crc32([1])` dies "buffer parameter is not a SCALAR reference", and a 3rd argument out of range dies. POD: `crc32($buffer [,$crc])`, "The buffer parameters can either be a scalar or a scalar reference. If the $crc parameters is undef, the crc value will be reset." Fix: `:listop (Str|ScalarRef|Undef $buffer, Optional[Int|Undef] $crc, Optional[Int] $offset) Int`. Confidence 65.

The constants were sampled and all 47 exports are present.

#### MIME::Base64 (1)

1. **encode_base64, encoded_base64_length, encode_base64url, decode_base64, decode_base64url**, MIME/Base64.tpm:12-26. `Str $data` is the Carp defect class: it rejects Boolean and Ref, which perl stringifies. Measured `encode_base64(true)`='MQ==' and `encode_base64([1])` encodes 'ARRAY(0x...)'. Fix: `Str|Boolean|DualVar|Ref $data`. Confidence 60.

Note, not a .tpm defect: `encoded_base64_length(undef)` segfaults perl 5.42 (rc 139). The .tpm's `Str` correctly excludes Undef there; do not widen that one to Undef.

#### Data::Dumper (1)

1. **Documented OO interface not declared**, Data/Dumper.tpm (whole file). `Data::Dumper->new(...)`, `->Dump`, `Data::Dumper->Dump([..],[..])` (measured: it prints `$x = 1;`), and the config methods and variables (`$Data::Dumper::Indent`, `Sortkeys`, `Terse`, ...) are all documented and all absent. Only the exports are declared. Fix: declare `new (Str $class: ArrayRef $vals, Optional[ArrayRef] $names) Object` and `Dump`, or record the omission. Confidence 60.

The Dumper/DumperX context multis measure correctly.

#### Cwd (2)

1. **getcwd/fastcwd**, Cwd.tpm:12-13. Say `Str`. POD for getcwd: "On error returns C<undef>, with C<$!> set". POD for fastcwd: "If fastcwd encounters a problem it will return undef". Measured in a removed directory: `getcwd()` and `fastcwd()` are both undef. Fix: `Str|Undef`. Confidence 90. (cwd and fastgetcwd return '' there, so Str is right for them.)
2. **chdir**, :21. `Str $dir` is required. The source reads `my $newdir = @_ ? shift : '';  # allow for no arg (chdir to HOME dir)`, and measured `Cwd::chdir()` returns 0 without dying. Fix: `Optional[Str] $dir` (also consider FileHandle: measured `Cwd::chdir($dirhandle)` is 1). Confidence 65.

Checked and clean: abs_path, realpath, fast_abs_path and fast_realpath (undef for a missing path, and a no-argument call uses the cwd).

#### SDBM_File (0)

DIRFEXT='.dir', PAGFEXT='.pag', PAIRMAX=1008, each with a "" prototype. Clean.

#### Symbol (4)

1. **qualify (multi), Ref candidate**, Symbol.tpm:18. Returns `GlobRef`, but perl returns the reference unchanged. Measured `ref qualify({})` is HASH and `ref qualify([])` is ARRAY. Fix: return `Ref`, or narrow the parameter to `GlobRef|Object` to match the POD's "assumption that they are glob references". Confidence 80.
2. **qualify (multi) coverage**, :18-19. `Ref` ∪ `Str` leaves out Boolean, DualVar, Glob and Undef. Measured `qualify(true)`='main::1' and `qualify(*STDOUT)`='\*main::STDOUT'. `qualify(undef)`='main::' with uninitialized warnings, so leaving it out is deliberate. Fix: Str candidate takes `Str|Boolean|DualVar|Glob $name`. Confidence 70.
3. **qualify_to_ref**, :25. `Str|Ref $name` is too wide. Measured `qualify_to_ref({})`, `qualify_to_ref(\"x")` and `qualify_to_ref(sub{})` all die "Not a GLOB reference". The return `GlobRef` is too narrow: measured `ref qualify_to_ref(IO::Handle->new)` is IO::Handle. Fix: `(Str|Boolean|Glob|GlobRef|Object $name, ...) GlobRef|Object`. Confidence 75.
4. **delete_package**, :34. Says `Str|Undef`. The value returned is a bare glob: measured `ref \(delete_package("Foo"))` is GLOB. The comment admits the oracle's ref()-based check reports Str. Fix: `Glob|Undef`. This is an oracle limitation, so the decision is perigrin's. Confidence 65.

Checked and clean: gensym (GLOB), geniosym (IO::File object), ungensym.

#### Multi coverage summary (as asked)

- **reftype/refaddr**: covered. `Ref` ∪ `Any` is the whole domain.
- **sum**: not covered. Non-Num first arguments (Str, Undef, Boolean, Ref) match neither candidate, and the return type misses Inf/NaN/Object.
- **encode**: not covered. A Boolean, DualVar or Ref `$string` matches neither candidate.

#### Count

List::Util 7 (item 8 is not a finding), Scalar::Util 4, Encode 6, Storable 4, Time::HiRes 5, Compress::Raw::Zlib 1, MIME::Base64 1, Data::Dumper 1, Cwd 2, SDBM_File 0, Symbol 4. Total 35.

AUDIT: FINDINGS

## Steps
- [ ] Re-measure each finding before fixing; withdraw any that does not reproduce, saying why in the commit
- [ ] Add oracle cases for each changed signature
- [ ] Hold tests, lib baseline and precision unchanged or better

## Acceptance criteria

### Positive Scenarios
- [ ] TestStorableFlagsArgument (`go test ./internal/infer/ -run '^TestStorableFlagsArgument$' -count=1 -v | grep -q '^--- PASS: TestStorableFlagsArgument'`)
- [ ] TestListUtilReturnsArgumentUnchanged (`go test ./internal/infer/ -run '^TestListUtilReturnsArgumentUnchanged$' -count=1 -v | grep -q '^--- PASS: TestListUtilReturnsArgumentUnchanged'`)
- [ ] TestCwdFailureReturnsUndef (`go test ./internal/infer/ -run '^TestCwdFailureReturnsUndef$' -count=1 -v | grep -q '^--- PASS: TestCwdFailureReturnsUndef'`)
- [ ] TestListUtilScalarContextReturns (`go test ./internal/infer/ -run '^TestListUtilScalarContextReturns$' -count=1 -v | grep -q '^--- PASS: TestListUtilScalarContextReturns'`)
- [ ] TestXSUtilitiesReturnTypes (`go test ./internal/infer/ -run '^TestXSUtilitiesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesReturnTypes'`)
- [ ] TestXSLibrariesReturnTypes (`go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestTooWideParametersNarrowed (`go test ./internal/infer/ -run '^TestTooWideParametersNarrowed$' -count=1 -v | grep -q '^--- PASS: TestTooWideParametersNarrowed'`)
