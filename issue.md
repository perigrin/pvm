---
title: TestGoParserExactFloor silently scores lower when PATH's perl is not 5.42
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T23:22:07.383876759Z
updated: 2026-10-16T23:22:07.383876759Z
---

The parse oracle runs bare `perl` from PATH (internal/parseoracle/oracle.go:129), and TestGoParserExactFloor does not check the interpreter version. Under /usr/bin/perl 5.38.2 the nine class/ tests (accessor.t, class.t, construct.t, destruct.t, field.t, gh22169.t, phasers.t, threads.t, utf8.t) fail to compile, go to no-answer, and stay in the denominator, so exact falls from 47 to 38 of 56 with no error. Measured 2026-10-16 at a9ce48f0 and cc141681: identical per-file results under perl 5.42.0, 38 under 5.38.2. A gate that silently scores lower under the wrong perl can hide or fake a regression. Fix: the test resolves the oracle perl and fails (or skips naming the version) when it is not the corpus's perl, as the parse-fidelity workflow's skew check does.
