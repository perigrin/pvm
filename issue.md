---
title: A scalar typed our outside a .tpm is not refused, though perl dies No such class
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T22:01:53.202126568Z
updated: 2026-10-16T22:01:53.202126568Z
---

In a program, `our Str $CRLF :export_ok;` gets no refusal, while perl 5.42 dies "No such class Str". The parser accepts `our Foo $x` for any package name and cannot know statically whether package `Str` exists, so refusing it would be a heuristic. Found by dadb's review gate (F6, confidence 65), 2026-10-16. Decide whether the typed-`our` rule should name this limit in RFC 0001, or whether a typed-Perl type name (Str, Int, Hash...) in that position should be refused outside a .tpm when no such package is declared or loaded.
