---
title: A call before the file's own sub of the same name is typed from the import
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-17T00:21:53.977752562Z
updated: 2026-10-17T00:21:53.977752562Z
---

A call written before the file's own definition of a sub with the same name is still typed from the imported declaration: `use My::Multi; my $r = lone([]); sub lone { "s" }` is typed Int, though perl 5.42.0 calls the local sub for calls both before and after its definition (measured in 5700's issue). This can produce false-positive coercion mismatches (`carp(undef); sub carp {...}`). Per perigrin's ruling call resolution belongs to the parser: `internal/parse/arity.go:139-152` markResolution should clear Signatures on calls when the file defines a package sub of the same name anywhere, keeping the Scalar::Util builtin-alias exception. Found by 5700's review gate (confidence 65), 2026-10-17.
