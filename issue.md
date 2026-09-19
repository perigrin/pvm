---
title: A quote delimiter is a character, not a byte
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-19T00:47:32.887135684Z
updated: 2026-09-19T00:47:32.887135684Z
---

A quote-like operator's delimiter is matched one BYTE at a time, so a multi-byte delimiter closes on its first byte and the rest falls out as an Error token. Placeholder; full text follows in an edit.
