---
title: "RFC 0001: *STDOUT{IO} is a blessed reference that measures as an Object, not an IO"
state: in-progress
urgency: normal
milestone: type-libraries
created: 2026-10-05T21:44:42.006185086Z
updated: 2026-10-05T21:46:04.322217427Z
sessions:
- start_sha: cb757bbc63faca15b99fe95baf733420e1f8a374
  end_sha: ""
  commits: 0
  started_at: 2026-10-05T21:46:04.322217427Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-05T21:46:04.322217427Z
---

## Context

- paths: `docs/rfcs/0001-type-libraries.md` (section "Type names", around lines 164-169)
- commands: `make test`

RFC 0001's "Type names" section says "`*STDOUT{IO}` is an IO object". Under perigrin's ruling (a) of 2026-10-05, IO is the slot type (a top-level leaf like Code, never a value), and the value `*STDOUT{IO}` -- a reference of reftype IO blessed into IO::File -- measures as an Object, as qr// does (pvm 02df00ad; witnessed on the lattice oracle's Object <: Ref edge). Found by the PAAD alignment gate on issue 01a0fd9b-3fbf-7c6e-90a7-919d8670f758.

## Steps

- [ ] Reword the sentence to say `*STDOUT{IO}` is a blessed reference to the IO slot's contents, which measures as an Object, and that IO names the slot type; cite the oracle witness
- [ ] `make test`; commit

## Acceptance criteria

- [ ] the RFC no longer calls the value an IO object (`sh -c '! grep -q "is an IO object" docs/rfcs/0001-type-libraries.md'`)
