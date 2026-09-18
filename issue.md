---
title: "A phaser brace is a block: BEGIN { } leaves the lexer in the wrong state"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-18T22:28:03.205765477Z
updated: 2026-09-18T22:28:03.205765477Z
---

The lexer has no phaser table, so BEGIN is an ordinary bareword, leaving XTerm. Its brace is classified as an anonymous hash rather than a block, so its closing brace leaves XOperator instead of XState, and a bare block after it is read as a hashref too. Placeholder body; full text follows in an edit.
