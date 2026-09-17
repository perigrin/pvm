---
title: Remove phantom PSC commands and flags from help system
state: in-progress
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:32.068658-04:00
updated: 2026-09-17T15:35:57.552005572Z
sessions:
- start_sha: 67a6240c1b6d9189aa21e438a788a540fa16179e
  end_sha: ""
  commits: 0
  started_at: 2026-09-17T15:35:57.552005572Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-17T15:35:57.552005572Z
---

The PSC CLI help displays commands and flags that do not exist in the codebase:

Phantom commands: compile, def, format, generate-type, import-type
Phantom flags on check: --recursive, --format, --strict, --dump-ast, --show-inferred

The actual PSC subcommands (from `internal/psc/command.go`) are: parse, analyze, check, lsp.

## Acceptance

- [ ] `psc --help` shows only: parse, analyze, check, lsp, version
- [ ] `psc check --help` shows only flags registered by `newCheckCommand()`
- [ ] Help system reflects actual cobra command tree, not aspirational content

Key files: internal/cli/help.go, internal/psc/command.go
See also: GitHub #407


## Verified fixed, 2026-09-17

Every criterion above is satisfied at 3c997a6b. Measured by building
`./cmd/psc` and running it (the binary must be NAMED `psc` --
`cli.DetectComponent` dispatches on the invocation name, so a build to any
other filename prints pvm's help, which is not a defect).

`psc --help` lists exactly:

    analyze <file|directory>   Analyze Perl files and list their dependencies
    check <file|directory>     Type-check Perl files and report diagnostics
    lsp                        Start a Language Server Protocol server
    parse <file>               Parse a Perl file and display its syntax tree
    version                    Print version information

No `compile`, `def`, `generate-type` or `import-type`. Grepping
`internal/cli/help.go` for those four names returns nothing; the two hits for
"format" are prose ("JSON format", "formatted help"), not commands.

`psc check --help` lists only `--strict` beyond the global flags, and
`--strict` is real: `internal/psc/check_command.go:75`. Of the five flags this
issue called phantom, `--recursive`, `--dump-ast` and `--show-inferred` exist
nowhere in `internal/psc/`, and `--format` is a real flag on `parse`
(`parse_command.go:56`), not on `check`.

The help system reads the cobra tree: `internal/psc/command.go` registers the
four subcommands and the help renderer walks what is registered.

Closing as done rather than cancelled: the outcome this issue asked for is the
outcome in the tree. Which commit delivered it was not traced -- the issue
predates this milestone and the fix was not attributed here.
