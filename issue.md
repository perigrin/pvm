---
title: "Audit fixes, IO and OO: invocants, IO::Select handles, inherited Tie::Hash methods, scalar-context returns, documented OO interfaces"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.594449107Z
updated: 2026-10-21T00:05:01.458193232Z
sessions:
- start_sha: 21162e2a33e98fd419c99aaa4fd58ed72f3db1b8
  end_sha: a2621ad1916c9a56ca43a75931e9c15d8f8929fa
  commits: 5
  started_at: 2026-10-20T21:35:00.99907194Z
  ended_at: 2026-10-21T00:05:01.458193232Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T21:35:00.99907194Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T00:05:01.458193232Z
observed_paths:
- internal/infer/parametertypes_test.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Archive-Tar.pl
- internal/infer/testdata/oracle/File-Spec-Functions.pl
- internal/infer/testdata/oracle/File-Spec-Unix.pl
- internal/infer/testdata/oracle/File-Spec.pl
- internal/infer/testdata/oracle/File-Temp.pl
- internal/infer/testdata/oracle/HTTP-Tiny.pl
- internal/infer/testdata/oracle/IO-Handle.pl
- internal/infer/testdata/oracle/IO-Select.pl
- internal/infer/testdata/oracle/IO-Socket-IP.pl
- internal/infer/testdata/oracle/IO-Socket.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/Socket.pl
- internal/infer/testdata/oracle/Term-ReadLine.pl
- internal/infer/testdata/oracle/Tie-Hash.pl
- internal/infer/testdata/oracle/Time-HiRes.pl
- internal/infer/testdata/parameter_oracle.pl
- internal/parse/declarations/Archive/Tar.tpm
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
- internal/parse/declarations/Net/Ping.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Term/ReadLine.tpm
- internal/parse/declarations/Tie/Hash.tpm
- internal/parse/heldtoperl_test.go
- internal/perltest/missingsubs.go
- internal/perltest/missingsubs_test.go
- internal/perltest/noskips_test.go
- internal/perltest/pathunder.go
- internal/perltest/pathunder_test.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/perltest/referencebuild_test.go
- internal/perltest/staleexempt.go
- internal/perltest/staleexempt_test.go
---

crochet:review of PR #488, round 1 (2026-10-21). Signature audit (01a15aaa), group C.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

C's autoflush claim was measured: psc reports nothing on $fh->autoflush(1); the fix there is Int $on (perl's $| integerises), not a coercion-rule change.

## Findings (measured; from the audit, perl 5.42.0)

#### .tpm audit, group C (report only)

Perl: /home/perigrin/.local/bin/perl (5.42.0). Probes: scratchpad/audit-C/{ioh.pl,ft.pl,em/}.
Parameter findings were re-judged under the brief's CORRECTION (RFC 0001: a parameter type is the type the operation coerces to). #10 and #23 are withdrawn. #4, #9, #12 and #17 stand under case (a): the operation preserves a handle or object that the declared type excludes.
Paths below are relative to internal/parse/declarations/.
`TypeSatisfies` (internal/types/types.go:328) has no Int->Boolean coercion and no GlobRef->Object
coercion, so "Int passed to a Boolean param" and "GlobRef passed to an Object invocant" are both rejections.

#### IO::Handle

1. **Instance methods (autoflush, format_page_number, format_lines_per_page, format_lines_left, format_name, format_top_name, input_line_number, format_write, print, printf, say, getline, getlines, printflush, flush, opened, fileno, eof, getc, close)**, IO/Handle.tpm:7,11-16,26-32,41,65-69. The .tpm says `Object $self`. Perl calls these on unblessed handles: `open my $fh,'>',...; $fh->autoflush(1)` works and ref($fh) is `GLOB` (ioh.pl #3), and so does `STDOUT->autoflush(0)` (#4). A GlobRef/Glob is not an Object (types.go:117: "a blessed IO::File is an Object"; FileHandle = Glob|GlobRef|IO). Fix: `Object|FileHandle $self`. Confidence 75.
2. **output_field_separator, output_record_separator, input_record_separator, format_line_break_characters, format_formfeed**, IO/Handle.tpm:8-10,17-18. The .tpm says `Object $self`. These are class methods. The POD says "not supported on a per-filehandle basis" and shows `IO::Handle->output_field_separator( [STR] )`. The source runs `carp "... not supported on a per-handle basis" if ref($_[0])`, and ioh.pl #1 measured that warning on an object. Fix: `Str $class`. Confidence 85.
3. **autoflush**, IO/Handle.tpm:7. The .tpm says `Boolean $on = 1`. The operation is `$| = $_[1]`, and `$|` coerces its value to an integer, not to a truth value. Measured: `$| = "yes"` leaves 0 and `$| = 0.5` leaves 0, though both values are true, while `"2abc"` leaves 1. So Boolean is not what the operation coerces to (RFC 0001, "What a parameter type means"). Fix: `Int $on = 1`. Confidence 70.
4. **new_from_fd**, IO/Handle.tpm:64. The .tpm says `Int $fd`. The POD (via fdopen) says the FD is "a file handle name, an IO::Handle object, or a file descriptor number". Measured: `new_from_fd(\*STDOUT,'w')`, `new_from_fd($io_file_obj,'w')` and `new_from_fd('STDOUT','w')` each return an IO::Handle (ioh.pl #6-8). Fix: `Str|FileHandle|Object $fd`. Confidence 85.
5. **Omitted documented methods**: fdopen, ungetc, write, read, sysread, syswrite, truncate, stat, fcntl, ioctl, error, clearerr, sync, blocking, binmode, untaint, setbuf, setvbuf. The POD's METHODS section lists all of them. The .tpm comment leaves them out because the oracle's bare-class call dies, but flush already uses an INSTANCE oracle (oracle/IO-Handle.pl), so these can be declared and checked the same way. For example, `IO::File->new('/dev/null')->binmode(':raw')` returns 1 and `$f->write("abc",2)` returns 1 (ioh.pl #25-26). Confidence 70.

#### FileHandle

6. The same invocant defect as #1, at FileHandle.tpm:87,90-97,105-106,139-143,146-150 (`Object $self` excludes GLOB/bareword handles). FileHandle.pm aliases the same code refs. Confidence 75.
7. The same class-method defect as #2, at FileHandle.tpm:88,89,98,99,100. Fix: `Str $class`. Confidence 85.
8. The same as #3 (`Boolean $on`; the operation coerces to Int), at FileHandle.tpm:87. Confidence 70.
9. The same as #4 (`new_from_fd Int $fd`), at FileHandle.tpm:138. Confidence 85.
10. WITHDRAWN under the CORRECTION: seek numifies its position, so Int is the coercion.

#### IO::File

11. **open**, IO/File.tpm:54. The .tpm says `Object $self`. Perl accepts a GLOB/bareword invocant, as in #1. Fix: `Object|FileHandle $self`. Confidence 70.

#### IO::Select

12. **Handle element type in new/add/remove/exists/handles/can_read/can_write/has_exception**, IO/Select.tpm:10,14,15,18,21,27-32. The .tpm says `FileHandle` (Glob|GlobRef|IO). The POD says "Each handle can be an IO::Handle object, an integer or an array reference where the first element is an IO::Handle or an integer." Measured: `new($io_handle_obj)` gives count 1, `add(0)` returns 1, `add([\*STDERR,"data"])` returns 1, and `handles`/`can_write` return `0,IO::Handle,ARRAY`. This rejects every IO::Handle/IO::Socket object, which is the most common use. Fix: `FileHandle|Object|Int|ArrayRef` in params and in return element types; `exists` should return the same type `|Undef`. Confidence 90.
13. **Omitted documented methods**:
    - `select` is a static/class method: `IO::Select->select($r,$w,$e,$t)`. In list context it returns 3 ArrayRefs or (); in scalar context it returns 3 or 0 (measured).
    - `bits` returns Str, or Undef when empty (measured).
    - `has_error` is deprecated, and `as_string` exists.

    Fix: `select (Str $class: Object|Undef $r, Object|Undef $w, Object|Undef $e, Optional[Num] $t)` as a list multi `List[ArrayRef]` plus a scalar multi `Int`, and `bits (Object $self:) Str|Undef`. Confidence 80.

#### File::Temp

14. **The OO interface is omitted** (POD "OBJECT-ORIENTED INTERFACE"):
    - `File::Temp->new(%opts)` returns File::Temp, and `File::Temp->newdir` returns File::Temp::Dir (both class invocant).
    - `$tmp->filename` returns Str and `$dir->dirname` returns Str (both instance).
    - `$tmp->unlink_on_destroy([BOOL])` returns 1 (instance).
    - `File::Temp->safe_level([LEVEL])` returns Int and `File::Temp->top_system_uid([UID])` returns Int (both class).

    All of these were measured in ft.pl. Confidence 90.
15. **Omitted documented functions that are not exported**: `tempnam($dir,$prefix)` returns Str, `cmpstat($fh,$path)` returns Int (1 measured), and `unlink1($fh,$path)` returns Int (1 measured). The POD sections are "ADDITIONAL FUNCTIONS" and "UTILITY FUNCTIONS". Confidence 80.
16. **tmpfile**, File/Temp.tpm:28. The .tpm says `GlobRef`. The source is `unlink0($fh, $file) or return undef;`. Fix: `GlobRef|Undef`. Confidence 85.
17. **unlink0**, File/Temp.tpm:41. The .tpm says `GlobRef $fh`. Perl only stat()s the handle. `unlink0($file_temp_obj, $obj->filename)` returns 1 (ft.pl), and a bareword Glob works too. Fix: `FileHandle|Object $fh`. Confidence 65.

#### Archive::Tar

18. **The whole method interface is omitted**. Archive/Tar.tpm declares only the 3 constants. The POD documents these methods:
    - Class: `Archive::Tar->new([$file,$compressed])`, `create_archive`, `iter`, `list_archive`, `extract_archive`, `has_io_string`, `has_perlio`, `has_zlib_support`, `has_bzip2_support`, `has_xz_support`, `can_handle_compressed_files`.
    - Instance: `read`, `contains_file`, `extract`, `extract_file`, `list_files`, `get_files`, `get_content`, `replace_content`, `rename`, `chmod`, `chown`, `remove`, `clear`, `write`, `add_files`, `add_data`, `error`, `setcwd`.

    These are the POD head2s at Archive/Tar.pm lines 96-1968. Confidence 90.

    The constants are correct: `COMPRESS_GZIP`=9, `COMPRESS_BZIP`='bzip2', `COMPRESS_XZ`='xz' (measured).

#### ExtUtils::Manifest

19. **manicheck, filecheck, skipcheck**, ExtUtils/Manifest.tpm:10,11,15. The .tpm says `List[Str]` only. In scalar context `return @missfile` gives the count. Measured in scratch em/: `my $m = manicheck()` gives 1 and `my $f = filecheck()` gives 2. Fix: add a scalar multi returning `Int`, as File::Spec.tpm does for no_upwards. Confidence 65.

#### Net::Ping

20. **pingecho**, Net/Ping.tpm:9. The .tpm says `Boolean|Undef`. It returns `$p->ping($host)` in the caller's context. In list context it returns `($ret, $duration, $ip)` (measured: `1 0.000257... 127.0.0.1`), or `()` when the host is unresolved (measured: 0 elements). In scalar context it returns `1`, and builtin::is_bool is false of it (measured), so it is an Int, not a Boolean. Fix: a list multi `List[Int|Num|Str]` and a scalar multi `Int|Boolean|Undef`. Confidence 85.
21. **wakeonlan**, Net/Ping.tpm:13. The .tpm says `Boolean|Undef`. The source is `return 1;` and `|| return undef`, and a literal 1 is not is_bool. Fix: `Int|Undef`. Confidence 80.
22. **The OO interface is omitted**: `Net::Ping->new([$proto[,$timeout[,$bytes[,$device[,$tos[,$ttl[,$family]]]]]]])`, `$p->ping($host[,$timeout[,$family]])` (list: ret, duration, ip), `close`, `open`, `ack`, `nack`, `bind`, `hires`, `time`, `source_verify`, `service_check`, `tcp_service_check`, `retrans`, `port_number`, `mselect`, `ntop`, `socket_blocking_mode`, `message_type`. All are documented. They can be exempt from the oracle the way pingecho is, but they are still a typed surface. Confidence 80.

#### File::Spec / File::Spec::Unix / File::Spec::Functions

23. WITHDRAWN under the CORRECTION: File::Spec::Unix::splitpath tests `if ( $nofile )` (Unix.pm:298), a truth test, so Boolean is the coercion.
24. **no_upwards, splitpath, splitdir, path in Functions**, File/Spec/Functions.tpm:11,13,17,18. Each is declared list-only. The functions are `sub {&$sub('File::Spec', @_)}`, so the context propagates. Measured scalar results: `no_upwards(qw(. .. a b))` gives 2, `splitdir("a/b/c")` gives 3, `splitpath("/a/b/c")` gives 'c', and `path()` gives 11. File/Spec.tpm already has these scalar multis. Fix: mirror them. Confidence 80.
25. **catfile/join with zero parts**, at File/Spec.tpm:7-8, File/Spec/Unix.tpm:21-23 and File/Spec/Functions.tpm:7. The .tpm says `List[Str] @parts` with return `Str`, which admits zero args. Measured: `catfile()` and `File::Spec->catfile()` both return undef. The comment acknowledges this and chooses not to cover it. Fix: return `Str|Undef`, or require one part. Confidence 60.

The invocants are correct here: these are class methods, and `Str $class` matches the POD.

#### HTTP::Tiny

26. **www_form_urlencode**, HTTP/Tiny.tpm:15. The .tpm says `Str $class`. The POD shows only `$params = $http->www_form_urlencode( $data );`, an instance method. It works on an object (measured `a=1`). Fix: `Object|Str $self` (the documented form is an Object). Confidence 75.
27. **can_ssl**, HTTP/Tiny.tpm:21-22. The .tpm says `Str $class`. The POD documents both `HTTP::Tiny->can_ssl` and `($ok,$why) = $http->can_ssl;`, and on an object it runs the extra CA check. Fix: `Str|Object $self`. Confidence 80.
28. **connected**, HTTP/Tiny.tpm:39. The .tpm says `Str|Undef`. The POD: "In list context, returns the peer host and port or an empty list". The source is `wantarray ? (peerhost, peerport) : join(':',...)`. Fix: add a list multi `List[Str|Int]`. Confidence 85.
29. **Omitted accessors**: cookie_jar, default_headers, http_proxy, https_proxy, keep_alive, local_address, max_redirect, max_size, proxy, no_proxy, SSL_options, verify_SSL. The POD says "An accessor/mutator method exists for each attribute.", and the source generates them in a BEGIN block. Measured: `$h->max_size(10)` returns 10. Confidence 85.

#### Term::ReadLine

30. **ReadLine, findConsole, addhistory, MinLine, Attribs, Features**, Term/ReadLine.tpm:14,18-19,21,24,26,27. The .tpm says `Str $class`. The POD says: "All the supported functions should be called as methods ... `$term->addhistory('row')`; where $term is a return value of Term::ReadLine->new()". Measured: `$t->ReadLine` and `$t->addhistory("x")` both work on the object. Fix: `Object $self` (ReadLine and findConsole also work on the class, so use `Object|Str`). Confidence 85.
31. **new**, Term/ReadLine.tpm:11. The .tpm says `(Str $class: Str $appname)`. The POD says it "Optionally can be followed by two arguments for IN and OUT filehandles". The source dies unless `@_==2 or @_==4`. Measured: `Term::ReadLine->new("x", \*STDIN, \*STDOUT)` returns an object. Fix: add `Optional[FileHandle] $in, Optional[FileHandle] $out` (both or neither). Confidence 85.
32. **Omitted documented methods**: `IN` and `OUT` return FileHandle|Undef (measured GLOB). The stub's documented extras `ornaments` (measured 'us,ue,md,me'), `newTTY`, `tkRunning` and `event_loop` are also omitted. Confidence 75.

#### Tie::Hash

33. **EXISTS, CLEAR**, Tie/Hash.tpm:25,28. The .tpm says `Str $class`. The POD says "EXISTS this, key" and "CLEAR this", and the source does `ref $_[0]` / `$self->FIRSTKEY`. These are instance methods. Fix: `Object $self`. Confidence 90.
34. **new, TIEHASH, CLEAR returning None**, Tie/Hash.tpm:19,22,28. These methods exist to be inherited (POD: "provides a new method, as well as methods TIEHASH, EXISTS and CLEAR"), and Tie::Hash's own body returns normally when reached from a subclass. Measured:
    - A subclass that defines only TIEHASH: `My->new` returns a `My` object, because Tie::Hash::new runs.
    - A subclass that defines only new: `N->TIEHASH` returns an `N` object.
    - `(tied %h)->CLEAR` on a subclass with FIRSTKEY/NEXTKEY/DELETE clears the hash and returns ''.

    The .tpm comment ("A subclass that defines its own TIEHASH or new is dispatched to that override") is false for new and CLEAR. Fix: `new` and `TIEHASH` return `Object`, and `CLEAR` returns `Any`. Only EXISTS is truly None. Confidence 80.
35. **Tie::StdHash and Tie::ExtraHash are omitted**. They are documented in Tie::Hash's POD and defined in Tie/Hash.pm (TIEHASH, STORE, FETCH, FIRSTKEY, NEXTKEY, EXISTS, DELETE, CLEAR, SCALAR). Confidence 65.

#### Counts

| Module | Findings |
|---|---|
| IO::Handle | 5 |
| FileHandle | 4 |
| IO::File | 1 |
| IO::Select | 2 |
| File::Temp | 4 |
| Archive::Tar | 1 |
| ExtUtils::Manifest | 1 |
| Net::Ping | 3 |
| File::Spec / ::Unix / ::Functions | 2 (shared across the three files) |
| HTTP::Tiny | 4 |
| Term::ReadLine | 3 |
| Tie::Hash | 3 |

AUDIT: FINDINGS

## Steps
- [ ] Re-measure each finding before fixing; withdraw any that does not reproduce, saying why in the commit
- [ ] Add oracle cases for each changed signature
- [ ] Hold tests, lib baseline and precision unchanged or better

## Acceptance criteria

### Positive Scenarios
- [ ] TestIOHandleSeparatorsAreClassMethods (`go test ./internal/infer/ -run '^TestIOHandleSeparatorsAreClassMethods$' -count=1 -v | grep -q '^--- PASS: TestIOHandleSeparatorsAreClassMethods'`)
- [ ] TestIOHandleMethodsTakeUnblessedHandles (`go test ./internal/infer/ -run '^TestIOHandleMethodsTakeUnblessedHandles$' -count=1 -v | grep -q '^--- PASS: TestIOHandleMethodsTakeUnblessedHandles'`)
- [ ] TestIOSelectTakesDocumentedHandles (`go test ./internal/infer/ -run '^TestIOSelectTakesDocumentedHandles$' -count=1 -v | grep -q '^--- PASS: TestIOSelectTakesDocumentedHandles'`)
- [ ] TestDocumentedOOInterfacesDeclared (`go test ./internal/parse/ -run '^TestDocumentedOOInterfacesDeclared$' -count=1 -v | grep -q '^--- PASS: TestDocumentedOOInterfacesDeclared'`)
- [ ] TestOOModulesReturnTypes (`go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestOOModulesReturnTypes'`)
- [ ] TestIOExportersReturnTypes (`go test ./internal/infer/ -run '^TestIOExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestIOExportersReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestTieHashInheritedMethodsReturn (`go test ./internal/infer/ -run '^TestTieHashInheritedMethodsReturn$' -count=1 -v | grep -q '^--- PASS: TestTieHashInheritedMethodsReturn'`)
