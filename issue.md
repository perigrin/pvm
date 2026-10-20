---
title: The return-type oracle cannot hold a sub that dispatches on arity and context together
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T02:39:32.766151981Z
updated: 2026-10-20T02:39:32.766151981Z
---

The return-type oracle (internal/infer/returntypes_test.go calls(), testdata/return_oracle.pl) calls every case of a sub in every context any candidate answers; it does not pair a case's argument count with the context its candidate answers. A sub that dispatches on arity and wantarray together (Socket::sockaddr_in, sockaddr_in6: the pack form dies in list context, the unpack form in scalar) therefore cannot be held and is exempt. Extend the case format so a case names (or is matched by arity to) the context it is called in. Found by dc48's review gate (70), 2026-10-20.
