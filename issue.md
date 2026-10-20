---
title: "Audit fixes, sockets: IO::Socket's constructor and methods, list-context unpackers, documented methods, dualvar errors"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.552145908Z
updated: 2026-10-20T19:18:42.552145908Z
---

crochet:review of PR #488, round 1 (2026-10-21). Signature audit (01a15aaa), group B.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

The report's Int-parameter policy question is settled by the lossy-coercion ruling (unknown-content Str in a numeric slot is silent by default).

## Findings (measured; from the audit, perl 5.42.0)

#### Audit group B: Socket, IO::Socket, IO::Socket::IP, Fcntl, Errno

Judge: /home/perigrin/.local/bin/perl (5.42.0). Socket 2.038, IO::Socket 1.55, IO::Socket::IP 0.43, Fcntl 1.20, Errno 1.38.
Paths below are relative to wt-cmt/internal/parse/declarations/. Probes are in scratchpad/audit-B/.

#### Constants (all of them, not a sample)

audit-B/consts.pl calls every `()` constant in Socket.tpm (285), Fcntl.tpm (151) and Errno.tpm (134) and compares the result with the declared Int, Str or None. There were **0 mismatches**: every None dies, every Int returns an integer, and every Str (INADDR_*, IN6ADDR_*, CR, LF, CRLF) returns a string.
Command: `perl audit-B/consts.pl Socket .../Socket.tpm` (and the same for Fcntl and Errno). Each prints only `checked N`.
Errno: @EXPORT_OK has 134 names, the POSIX tag has 74, and Errno.tpm declares 134 subs. They match.

#### Socket

1. **Socket::getaddrinfo** (Socket.tpm:325-326)
   - Declared: `Str $host`. Perl: the POD says "Given just a service name, this function attempts to resolve it". `getaddrinfo(undef,"80",{flags=>AI_PASSIVE})` returns err "" and 6 results, with no warning.
   - Fix: `Str|Undef $host`.
   - Confidence: 90.
2. **Socket::getaddrinfo** (Socket.tpm:325-326)
   - Declared: the list return is `List[Str|HashRef]` and the scalar return is `HashRef|Str`. Neither admits DualVar.
   - Perl: the POD says "The error value will be a dualvar". `isdual($err)` is 1 on success ("" / 0) and on failure ("Name or service not known" / -2). In scalar context a failure returns that dualvar.
   - The oracle cannot see this: observe_types.pl never emits DualVar.
   - Fix: `List[DualVar|HashRef]` (or `List[Str|DualVar|HashRef]`) and `HashRef|DualVar`.
   - Confidence: 80.
3. **Socket::getnameinfo** (Socket.tpm:327-328)
   - Declared: `List[Str|Int|Undef]` and `Str|Int|Undef`. Neither has DualVar.
   - Perl: the POD says "The error value will be a dualvar". `getnameinfo($addr, NI_NUMERICHOST|NI_NUMERICSERV)` gives `isdual($n[0])` = 1. `getnameinfo("garbage")` returns 1 element, a dualvar ("ai_family not supported"), so scalar context also yields a DualVar on failure.
   - Fix: add DualVar to both candidates.
   - Confidence: 80.
4. **Socket::sockaddr_un** (Socket.tpm:350)
   - Declared: one sub, `(Str $path) Str`, with no list-context candidate.
   - Perl: Socket.pm:876-883 calls `unpack_sockaddr_un` when `wantarray`. The POD has `($path) = sockaddr_un $sockaddr`. `my @l = sockaddr_un(pack_sockaddr_un("/tmp/x"))` gives `("/tmp/x")`. A path passed in list context dies ("Bad address family").
   - Fix: make it a multi like sockaddr_in: `multi (Str $path) Str;` and `multi (Str $sun) List[Str];`.
   - Confidence: 85.
5. **Socket::pack_sockaddr_in, pack_sockaddr_in6, sockaddr_in and sockaddr_in6 (pack form)** (Socket.tpm:336, 337, 346, 348)
   - Declared: `Int $port`.
   - Perl: the POD says "An undefined $port argument is taken as zero". `pack_sockaddr_in(undef, inet_aton("127.0.0.1"))` unpacks to port 0 with no warning, and pack_sockaddr_in6(undef, IN6ADDR_ANY) also works silently.
   - Fix: `Int|Undef $port`.
   - Confidence: 80.
6. **Socket::pack_ip_mreq / pack_ip_mreq_source** (Socket.tpm:333-334)
   - Declared: `Str $interface_addr = "\0\0\0\0"`.
   - Perl: an explicit undef interface is accepted silently. `pack_ip_mreq($m, undef)` returns 8 bytes and `pack_ip_mreq_source($m, INADDR_LOOPBACK, undef)` returns 12 bytes, with no warning.
   - Fix: `Str|Undef $interface_addr = ...`.
   - Confidence: 65.
7. **Socket::sockaddr_in, legacy 6-arg scalar form** (Socket.tpm:346-347)
   - Declared: no candidate for it.
   - Perl: Socket.pm:851 has `if (@_ == 6 && !wantarray)`. `scalar sockaddr_in(AF_INET,80,127,0,0,1)` returns a 16-byte sockaddr and warns "6-ARG sockaddr_in call is deprecated". The form is undocumented and deprecated.
   - Fix: either a candidate `(Int $af, Int $port, Int $a, Int $b, Int $c, Int $d) Str`, or a comment that it is deliberately left out.
   - Confidence: 60.

The other Socket functions were measured and are fine: inet_aton/ntoa/pton/ntop, sockaddr_family, pack_ipv6_mreq, all unpack_* in both contexts, the getnameinfo flags and xflags defaults, and the getaddrinfo hints (undef accepted, a non-hash dies).

#### IO::Socket

1. **IO::Socket::new** (IO/Socket.tpm:202)
   - Declared: `(Str $class:) Object`, which takes no arguments.
   - Perl: IO/Socket.pm:42 has `my($class,%arg) = @_`. The POD says "new optionally takes arguments" and "If the constructor fails it will return undef".
   - Measured: `IO::Socket->new(Domain=>AF_INET, Type=>SOCK_STREAM, Listen=>1, ...)` returns an IO::Socket::INET object. `IO::Socket->new(Domain=>AF_INET, ..., PeerPort=>1)` returns undef with errstr "connect: Connection refused".
   - Fix: `(Str $class: List @args) Object|Undef`.
   - Confidence: 95.
2. **IO::Socket::socket / connect / bind / listen** (IO/Socket.tpm:203-206)
   - Declared: each returns `Boolean`.
   - Perl: each returns `$sock` or undef (IO/Socket.pm:91, 165, 223, 233). Measured: socket, connect, bind and listen each returned an `IO::Socket` object.
   - Fix: `Object|Undef` for all four.
   - Confidence: 95.
3. **IO::Socket::listen** (IO/Socket.tpm:206)
   - Declared: `Optional[Int] $queue`.
   - Perl: `$queue = 5 unless $queue && $queue > 0`. `listen(undef)` returned the object with no warning.
   - Fix: `Optional[Int|Undef]`.
   - Confidence: 70.
4. **IO::Socket::accept** (IO/Socket.tpm:207)
   - Declared: one candidate, `Object|Undef`.
   - Perl: the POD says "In a list context a two-element array is returned containing the new socket and the peer address; the list will be empty upon failure". Measured: `my @a = $l->accept` gives n=2 (an object and a 16-byte peer address). On timeout the list form gives n=0 and the scalar form gives undef.
   - `$pkg = shift || $sock` means an object is also accepted where a class name is expected.
   - Fix: `multi (...) List[Object|Str]` for list context and `multi (...) Object|Undef` for scalar. The param becomes `Optional[Str|Object] $class`.
   - Confidence: 90.
5. **IO::Socket::send** (IO/Socket.tpm:208)
   - Declared: `Int $flags`, required.
   - Perl: `@_ >= 2 && @_ <= 4` and `$flags = $_[2] || 0`. The POD shows `send(BUF, [FLAGS, [TO]])`. Measured: `$c->send("hi")` returns 2.
   - Fix: `Optional[Int] $flags`.
   - Confidence: 95.
6. **IO::Socket::sockatmark** (IO/Socket.tpm:190)
   - Declared: returns `Boolean|Undef`.
   - Perl: on a connected socket it returns the string "0 but true" (`is_bool` is 0). On a closed handle it returns undef. `$sock->atmark` returns the same value.
   - Fix: `Str|Undef` (or `Int|Str|Undef`).
   - Confidence: 85.
7. **IO::Socket: documented methods left out** (IO/Socket.tpm:202-210)
   - The .tpm leaves out these methods: configure, socketpair, blocking, sockname, peername, connected, shutdown, setsockopt, getsockopt, sockopt, atmark, timeout, sockdomain, socktype, protocol, register_domain. Each is defined at IO/Socket.pm:56-395, and all except configure and register_domain have a `=head2` in the POD.
   - Measured returns:

     | Method | Returns |
     |---|---|
     | socketpair | list of 2 objects, or () |
     | blocking | 1 (not is_bool) |
     | sockname / peername / connected | 16-byte Str, or undef |
     | timeout | old value, undef initially |
     | sockdomain / socktype / protocol | Int (or undef) |
     | getsockopt / sockopt | Int, or Str |
     | setsockopt | Boolean |
     | shutdown | 1 |

   - Fix: declare them.
   - Confidence: 85.
8. **IO::Socket::sockaddr_in (pack form)** (IO/Socket.tpm:179)
   - Declared: `Str|Undef $addr`. This differs from Socket.tpm:346 (`Str $addr`), even though the header says every value is the one Socket.tpm measures.
   - Perl: `scalar Socket::sockaddr_in(80, undef)` dies "Undefined address for Socket::pack_sockaddr_in". The type is too wide: it admits a value perl dies on.
   - Fix: `Str $addr`, and the port should be `Int|Undef` as in Socket finding 5.
   - Confidence: 90.
9. **IO::Socket re-exports of Socket's sockaddr_un, getaddrinfo, ...** (IO/Socket.tpm:183)
   - sockaddr_un has the same missing list-context candidate as Socket finding 4. pack_sockaddr_in, pack_sockaddr_in6 and sockaddr_in6 have the same `Int $port` problem as Socket finding 5.
   - Separately, IO::Socket::import forwards any import list to `Exporter::export 'Socket'` (IO/Socket.pm:37). Measured: `use IO::Socket qw(getaddrinfo :addrinfo inet_pton)` installs getaddrinfo, NI_NUMERICHOST and inet_pton. The .tpm declares only the bare `:export` set and marks nothing `:export_ok`, so every Socket export_ok name and tag is missing from IO::Socket.tpm.
   - Fix: mirror Socket's `:export_ok`/tags, or declare `import`.
   - Confidence: 75.

#### IO::Socket::IP

1. **IO::Socket::IP::connect** (IO/Socket/IP.tpm:209)
   - Declared: returns `Object|Undef`.
   - Perl: IP.pm:702-781 never returns `$self`. It returns core `connect` (Boolean), 1, 0, undef, or `$self->setup` (1/0/undef). The POD says "While connect returns false ... Once the socket has been connected to the peer, connect will return true".
   - Measured: a non-blocking `connect()` with no args returns 1 (not is_bool). `connect($addr)` returns Boolean true. A refused connect returns undef.
   - Fix: `Boolean|Int|Undef`.
   - Confidence: 95.
2. **IO::Socket::IP::new** (IO/Socket/IP.tpm:208)
   - Declared: returns `Object`.
   - Perl: `IO::Socket::IP->new(PeerHost=>"127.0.0.1", PeerPort=>1, Type=>SOCK_STREAM)` returns undef (errstr "Connection refused"). IP.pm:629 has `return undef`.
   - Fix: `Object|Undef`.
   - Confidence: 95.
3. **IO::Socket::IP: documented methods left out** (IO/Socket/IP.tpm:208-213)
   - The .tpm leaves out documented methods: sockhost_service, peerhost_service, sockhostname, sockservice, peerhostname, peerservice, sockaddr, peeraddr, as_inet, split_addr, join_addr (each has a `=head2` at IP.pm:830-1186). It also leaves out the overrides accept (wantarray, IP.pm:965-975), connected (Boolean) and socket.
   - Measured returns:

     | Method | Returns |
     |---|---|
     | peerhost_service | list of 2 |
     | sockaddr / peeraddr | 4-byte Str |
     | `split_addr("[::1]:80")` | ("::1", "80") |
     | `split_addr("x")` | ("x", undef) |
     | `join_addr("::1", 80)` | "[::1]:80" |
     | `join_addr("h", undef)` | "h" |

   - Also left out: the documented `-register` import option (IP.pm:128), which registers IP for AF_INET/AF_INET6 and installs nothing.
   - Confidence: 85.
4. **IO::Socket::IP inherited re-exports** (IO/Socket/IP.tpm:186-201)
   - These carry the same problems as IO::Socket findings 8 and 9. sockaddr_un has no list-context candidate, the pack forms' `Int $port` should admit Undef, and only the bare `:export` set is declared, even though `use IO::Socket::IP qw(getaddrinfo)` goes through IO::Socket::import to Socket.
   - IP.tpm:192 has `Str $addr`, which is correct, unlike IO/Socket.tpm:179.
   - Confidence: 75.

#### Fcntl

No findings at 60 or above. Measured:
- S_ISREG/ISDIR/... return is_bool values, which matches `Boolean`.
- S_IMODE and S_IFMT($m) return Int, and `S_IFMT()` returns 61440, which matches both S_IFMT candidates.
- S_ISWHT dies "Usage: Fcntl::AUTOLOAD()", which matches `None`.
- S_ISTXT dies, which matches `None`.
- A zero-arg S_IS*/S_IMODE call warns "uninitialized" and does not read `$_`, so requiring `$mode` is right.
- Extra args are silently ignored (`S_ISREG($m,1)` is true). Declaring arity 1 is stricter than perl here, but I am not reporting it (confidence < 60).

#### Errno

No findings at 60 or above. All 134 constants return Int, the export and POSIX-tag counts match, and Errno.pm defines no other public subs. TIEHASH, FETCH, STORE, EXISTS, FIRSTKEY and NEXTKEY are the tie interface behind `%!` and are not exported, so leaving them out is reasonable.

#### Cross-cutting (policy question, confidence 60)

Every numeric parameter in these files is typed `Int` (ports, flags, af, scope_id, ifindex, $mode). Perl numifies a string without error: `pack_sockaddr_in("80", ...)`, `pack_sockaddr_in6("80", IN6ADDR_ANY, "1", "2")`, `S_ISREG("$m")` and `getnameinfo($addr, "1")` all succeed silently. A port read from config or argv is a Str, so psc would reject a call that perl accepts. Whether `Int` params should admit numeric `Str` is a project-wide policy call, not specific to these modules.

#### Counts

| Module | Findings |
|---|---|
| Socket | 7 |
| IO::Socket | 9 |
| IO::Socket::IP | 4 |
| Fcntl | 0 |
| Errno | 0 |
| Cross-cutting note | 1 |

AUDIT: FINDINGS

## Steps
- [ ] Re-measure each finding before fixing; withdraw any that does not reproduce, saying why in the commit
- [ ] Add oracle cases for each changed signature
- [ ] Hold tests, lib baseline and precision unchanged or better

## Acceptance criteria

### Positive Scenarios
- [ ] TestIOSocketConstructorTakesArgs (`go test ./internal/infer/ -run '^TestIOSocketConstructorTakesArgs$' -count=1 -v | grep -q '^--- PASS: TestIOSocketConstructorTakesArgs'`)
- [ ] TestIOSocketMethodsReturnSocketOrUndef (`go test ./internal/infer/ -run '^TestIOSocketMethodsReturnSocketOrUndef$' -count=1 -v | grep -q '^--- PASS: TestIOSocketMethodsReturnSocketOrUndef'`)
- [ ] TestSocketListContextUnpackers (`go test ./internal/infer/ -run '^TestSocketListContextUnpackers$' -count=1 -v | grep -q '^--- PASS: TestSocketListContextUnpackers'`)
- [ ] TestIOSocketDocumentedMethodsDeclared (`go test ./internal/parse/ -run '^TestIOSocketDocumentedMethodsDeclared$' -count=1 -v | grep -q '^--- PASS: TestIOSocketDocumentedMethodsDeclared'`)
- [ ] TestConstantExportersPrototypes (`go test ./internal/parse/ -run '^TestConstantExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersPrototypes'`)
- [ ] TestConstantExportersReturnTypes (`go test ./internal/infer/ -run '^TestConstantExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersReturnTypes'`)
- [ ] TestIOExportersReturnTypes (`go test ./internal/infer/ -run '^TestIOExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestIOExportersReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestIOSocketSendFlagsOptional (`go test ./internal/infer/ -run '^TestIOSocketSendFlagsOptional$' -count=1 -v | grep -q '^--- PASS: TestIOSocketSendFlagsOptional'`)
