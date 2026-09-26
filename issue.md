---
title: git zhi issue add blocks on stdin forever when no body is given
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-26T09:57:11.104571753Z
updated: 2026-09-26T09:57:11.104571753Z
---

`git zhi issue add` with a positional title and no `--body` blocks on stdin
forever instead of erroring. It cost a 120s timeout in the m1 execution loop
and created no issue.

## Measured

    ZHI_ACTOR=agent:m1a-orch git zhi issue add "some title" --milestone m1-parse-the-core

No output, no issue created, process still live at 120s when the harness
moved it to the background. Adding `< /dev/null` and `--body-file <path>`
returns immediately:

    Created 01a0dd25: The identifier scanner eats the digits of x3, ...

## Why it is a defect and not user error

`issue add --help` documents `--body` as "issue body text (used with
positional title arg; overrides stdin)". So stdin IS the documented default
source, and reading it is correct. Blocking on it with no prompt, no
diagnostic and no TTY check is not: a non-interactive caller gets a hang
rather than "no body given; pass --body, --body-file, or pipe one".

This is the SECOND silent-input defect in this command pair. The first is
already recorded: `issue add --body -` silently stores a literal dash,
while only `issue edit --body -` reads stdin. Same surface, same class --
the body argument's input contract differs between two sibling subcommands
and neither reports when it gets nothing.

## Scope

1. When `add` has no `--body`/`--body-file` and stdin is a TTY (or is empty
   and not a pipe), fail with a message naming the three ways to supply a
   body. Do not block.
2. Decide whether `add --body -` should mean stdin, as `edit` does. Either
   answer is defensible; the current split is not, because it silently
   stores `-` as the body.

## Context

- commands: `git zhi issue add --help`, `git zhi issue edit --help`
- reproduce: `git zhi issue add "t" --milestone <m>` from a non-interactive
  shell with no stdin redirect

## Acceptance criteria

- [ ] `add` with no body and no stdin exits nonzero with a message rather than blocking (`timeout 10 git zhi issue add "probe" < /dev/null; test $? -ne 0 && test $? -ne 124`)
- [ ] The `--body -` contract is the same for `add` and `edit`, or the difference is documented in both help texts (`git zhi issue add --help 2>&1 | grep -q -- "--body -"`)
