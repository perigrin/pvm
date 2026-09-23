# Chain review: corpus-complete

**Date:** 2026-09-23
**Milestone:** `corpus-complete`, 11 issues, 59 acceptance criteria
**Reviewed at:** d20ac8df

The chain was written and reviewed in the same session, which is the
weakest position to review from. Two real defects were found, both of the
same class the milestone itself exists to fix: **a claim that lives in
prose and is not enforced.**

## Defect 1 — eight criteria whose command was not the extracted span

`git-zhi verify` extracts the FIRST backtick span on a `- [ ]` line and
runs it. Eight criteria were written with prose emphasis before the
command:

    - [ ] A corpus file pins what `state` does with its feature
          disabled ... (`go test ...`)

The extracted span is `state`. verify would run that as a shell command,
which is not a regression and not a pass -- it is the fourth silent rule
(no backtick inside the command) wearing a new disguise.

**My first check missed this.** It asked whether a backtick span existed
on the line, not whether the FIRST span was the command. A checker that
answers a weaker question than the tool it models is worth very little,
and this is the second time this session that a check passed on
incidental data.

Fixed: prose unquoted, so the command is the only span on every criterion
line. Re-verified at 0 of 59.

## Defect 2 — a dependency stated in prose and not declared

Issue `01a0cc04-333a` ("One pragma, two answers") carried the sentence:

> Depends on the bitwise/shift issue: a pragma pair cannot claim
> operators the corpus has not introduced.

No dependency was declared in the tracker. `crochet:execute` reads the
tracker, not the prose, so the chain would have offered the pragma pair
before the operators it gates existed.

A second, unstated dependency was found by reading the bodies against
each other: compound assignment (`01a0cc04-32c9`) includes `|=`, `&=`,
`^=`, `<<=`, which cannot be claimed before `|`, `&`, `^` and `<<` are
introduced -- the same constraint, one issue over.

Both are now declared. The remaining nine issues are genuinely
independent: they touch different tiers and claim different ops.

## What the review did NOT find

- No criterion names a test that does not exist. All eight named tests
  were confirmed present, which matters because `go test -run` of a
  nonexistent test exits 0.
- No issue is missing its `## Acceptance criteria` header, and nothing
  follows any criteria block -- the two silent rules that make a whole
  issue invisible.
- No two issues claim the same construct or the same tier's INTRODUCES
  entry.

## Sizing

Eleven issues, ~33 files, against a 25-35 target. Two issues add no files
(`09_call_forms`, `11_pattern_scope`) and exist to repair or record
claims rather than add coverage. That is the right shape: the measurement
found two files that document more than they assert, and a corpus whose
prose outruns its claims is the failure this milestone is correcting.

## The one thing I would still change

The criteria lean on `TestCorpusLints` and the full-suite green check,
which are regression guards rather than completion gates -- they pass
today, before any work. Each issue does carry at least one criterion that
FAILS today (a source grep for a construct not yet present), so the gates
exist; but the ratio is about one gating criterion to five guards.

That is defensible -- the guards are what stop a new file breaking an old
tier -- but it means `git-zhi verify` reporting "59/59" will be mostly
measuring that the suite still passes. Worth knowing when reading the
number.
