<!-- ABOUTME: The formatting rules git-zhi verify enforces silently on an issue's acceptance criteria. -->
<!-- ABOUTME: Read before writing an issue's criteria; each rule was found by a criterion verify never ran. -->

---
stability: 1
covers: []
---

`git zhi verify` runs each acceptance criterion's command and fails the
milestone on any that fails. It finds those commands by reading the issue
body line by line, and four rules decide whether a criterion is run at all.
Breaking one does not fail verify: the criterion is skipped, sometimes with
no message. Each rule below cost a milestone criteria that were never run
(issues 01a0c546 and 01a0c666, m3-conformance-corpus).

## 1. The header must say `criteria`

verify looks for a heading containing the word `criteria`. A block headed
any other way is not read, so its criteria contribute nothing and verify
reports `total: 0` for the issue.

    ## Acceptance criteria   found
    ## Acceptance            not found

## 2. The command goes on the SAME LINE as its checkbox

Each criterion is one line: `- [ ]`, the claim, and the command in
parentheses and backticks, however long the line becomes. A command
wrapped onto a continuation line is invisible to the line-based extraction,
and verify reports the criterion `unverifiable`.

    - [ ] the lint sees inside a named sub (`go test ./internal/conformance/ -run TestOpsOfSeesInsideNamedSub -count=1`)

## 3. Nothing follows the criteria block

A `###` subsection after the criteria makes the whole block invisible, not
only the part after it, and verify says nothing. Put any note, correction
or scope remark before the `## Acceptance criteria` heading, and leave the
criteria last in the body.

    ## Acceptance criteria
    - [ ] trivially true (`true`)              discovered

    ## Acceptance criteria
    - [ ] trivially true (`true`)
    ### A following subsection                 contributes NOTHING

## 4. A criterion does not verify its own milestone

A criterion that runs `git zhi verify` on the milestone it belongs to is
verify running inside itself, and reports as `unverifiable`. Count or list
criteria with `git zhi verify <milestone> --dry-run`, which cannot recurse.
