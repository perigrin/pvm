---
title: Tier READMEs, because the dependency lint needs an INTRODUCES set that cannot drift
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-6e43-7c0d-8828-ef7659402e39
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:50:20.295669503Z
updated: 2026-10-23T23:40:17.280470419Z
sessions:
- start_sha: 0ce515cbfe0ba1f48eb9b741cb1fc5d8c389454d
  end_sha: 0ce515cbfe0ba1f48eb9b741cb1fc5d8c389454d
  commits: 0
  started_at: 2026-09-21T18:40:43.104838839Z
  ended_at: 2026-09-21T18:40:53.409352971Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T18:40:43.104838839Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T18:40:53.409352971Z
---

Listed in the spec's tooling list and left as one of two open questions:
"Does each tier get a README stating what it introduces and why it sits
there? Recommended yes. The spike wrote no README and did not miss one,
which is weak evidence at three files."

## The argument that settles it, which is not the documentation argument

A tier README is not primarily for readers. **The dependency lint needs a
machine-readable INTRODUCES set per tier**, and that set has to live
somewhere. In a Go table, the tier's contents and its declared contents are
two lists that drift. In the tier directory, drift is visible in the same
diff.

So the README is where the lint's input lives, and the prose comes free.
That also answers "why does this tier sit here" -- check 2 says a tier sits
where it does because the next cannot proceed without it, and that claim
should be written where it can be argued against.

## The second open question, folded in

"Does `04_operators` need internal grouping by precedence level?" It is the
largest tier by a distance. **Do not decide now.** Decide it when the tier
is written and the directory is actually unreadable; a grouping invented
before the files exist will be the wrong grouping. The deferral is recorded
against the tier 04 issue rather than left in the spec's open questions.

## This issue also enforces adjacency across every tier

The spec's rule is per tier -- "every construct the tier introduces, in one
body, each adjacent to another" -- and its argument is that a
one-construct-per-file corpus CANNOT reach the ADJUST-class bug at all:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

Each tier issue carries its own adjacency criterion, but a per-tier
criterion can be forgotten one tier at a time. This issue owns the generic
check, because it already owns the INTRODUCES set the check needs. Where a
tier declares a prerequisite the adjacency file pairs with THAT, not with
tier N-1 -- pairing 12_packages with 11_oo would assert nothing, since the
spec says packages needs nothing from the tier before it.

### The format is already settled, and tier 01's README exists

The dependency lint landed first and defines what this issue writes
thirteen more of. `conformance/01_literals/README.md` is the worked
example; `readTierOps` in internal/conformance/lint.go is the reader.

The machine-read part is one block:

    ## INTRODUCES

        const enter leave multiconcat nextstate padsv padsv_store print pushmark

Indented so it renders as a code block. Everything else in the file is
prose a human may reword freely.

Two rules the lint enforces, both learned the hard way:

- **The block takes op names alone.** Prose inside it became ops --
  "These are the ops: const and nextstate." yielded seven entries -- which
  silently WIDENED the tier. Anything not matching `^[a-z][a-z0-9_]*$` is
  now rejected by name.
- **Claim the ops perl EMITS, not the ones the source implies.**
  `print "$x\n"` compiles to `multiconcat`, not `concat`. Measure with
  `perl -MO=Concise,-exec` rather than reading the file.

`TestCorpusLints` also reports a README claiming an op no corpus file
emits, so a tier cannot be widened by an aspirational entry.

### Progress: the machinery landed, the fourteen tiers did not

Commit `a1d221c2` adds the second machine-read block and the check that
uses it.

**`## DEPENDS ON`** takes one tier name or the word `nothing`, read as the
first indented block after the heading. Declared rather than assumed to be
tier N-1, because the spec names tiers where N-1 asserts nothing --
`09_regex` needs nothing from `08_references`, `12_packages` nothing from
`11_oo` -- and an adjacency file pairing those with N-1 would prove
neither.

**`checkTierNecessity`** rejects a prerequisite at a later tier number, a
tier depending on itself, a name no tier provides, and a name that is not
a tier at all. AC6 is met by `TestTierNecessityCheck` (6 cases) with
`TestRealCorpusNecessityHolds` running it over the corpus.

**A defect found while adding it, which `reIntroduces` had all along.**
Both block regexes read to the next `## ` heading rather than to the end of
the indented block, so prose following the block became part of the value.
INTRODUCES survived only because tier 01's block happens to be followed
immediately by a heading; a sentence after it would have given the tier ops
named "Those" and "This". Both now read the first indented block and stop.
Pinned by `TestReadTierOpsIgnoresTrailingProse`.

**Adjacency file naming is settled: `00_adjacency.t`.** Construct files
start at `01`, so `00` sorts ahead of them and marks the file as the check
that the tier's constructs compose rather than as one of them. The constant
is `adjacencyFile` in lint.go.

### Remaining, and it is the bulk of the issue

Thirteen tier READMEs (02 through 14) and fourteen `00_adjacency.t` files.
Each README needs an INTRODUCES set MEASURED with `perl -MO=Concise,-exec`
rather than inferred from the source, and each adjacency file needs Perl
whose constructs are genuinely adjacent and whose ops cover the tier's
declared set.

The two tests for those ACs are written and were RED for the right reasons
(14 subtests each, naming every missing directory). They are held in the
session scratchpad rather than committed, because committing them would
leave the package red against the pristine-output rule. They restore in
under a minute:

    scratchpad/red_readme_test.txt      TestEveryTierHasAReadme
    scratchpad/red_adjacency_test.txt   TestEveryTierHasAnAdjacencyFile

`specTiers` in lint_test.go is already committed and is the list both use.
It is deliberately written out rather than read from the corpus directory:
a list derived from the thing it checks cannot report an absence.

AC1 and AC3 named tests that never existed. Both commands have been
corrected in the Acceptance section above; see the progress note below.

## Progress, as of 0ce515cb

All fourteen tiers are written: a README per tier whose INTRODUCES set was
MEASURED with `perl -MO=Concise,-exec` rather than reasoned, the construct
files that emit it, and one adjacency file per tier. 122 files, 295
subtests, suite green.

The two AC commands above have been corrected. They named
`TestIntroducesReadFromReadmes` and `TestAspirationalIntroducesReported`,
neither of which ever existed; the real checks are
`TestOpTableDerivedFromReadmes` and, for the aspirational-op report, a
check living inside `TestCorpusLints`. The tests were right and the
acceptance criteria were wrong, so the criteria moved.

Four things the gate changed that are worth carrying forward:

**The lint cannot see inside a sub.** `opsOf` runs `perl -MO=Concise,-exec
<path>` with no sub named, so it measures the main optree only; a sub body
is a separate CV. Tier 07 dropped from seven claimed ops to two and tier 11
dropped `methstart`, both independently. argcheck, argdefelem, argelem,
leavesub and return are really emitted and are invisible to the gate.
Teaching `opsOf` to enumerate and dump each sub would make them claimable.
DEFERRED TO: a new issue, not yet filed -- it changes the instrument for
all fourteen tiers, so it wants measuring before being assumed an
improvement.

**"A tier claims an op once" is unenforced.** Twelve ops came back claimed
by two tiers each and were resolved by hand toward the tier that emits them
first. The lint checks that a claimed op IS emitted, not that it is claimed
ONCE, so a tier could silently re-claim one and pass. DEFERRED TO: a new
issue, not yet filed.

**A stray `--- expect output` inside a heredoc or data section is read
silently as a real section.** Measured against `ParseFile`: a stray
`--- not a marker` fails loudly with `unknown section`, but a line spelling
a REAL section name truncates the source and is not reported at all.
DEFERRED TO: a new issue, not yet filed.

**Two pre-commit hooks corrupt corpus files that are correct as written.**
`end-of-file-fixer` strips the blank line that `--- expect output` takes
its trailing newline from, which is why `expect output` may never be a
file's last section; `TestExpectOutputIsNeverTheLastSection` now enforces
that. `trailing-whitespace` strips the space from a line ending in a false
comparison's empty string, so tier 04 brackets every result. Both are
recorded in `conformance/README.md`.

Fifteen files carry `STATUS refuses`, each citing itself per the form
`file.go` documents for a gap found by writing the file. Tier 11's
adjacency file reproduces the ADJUST bug the adjacency design was written
for; tier 04's refuses where every one of its constructs parses alone;
tier 13's heredocs refuse outright.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files;  Left unchanged here: `TestEveryTierHasAnAdjacencyFile`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestEveryTierHasAnAdjacencyFile` -> `TestEveryTierHasAnAdjacencyTopic` and `TestAdjacencyCoversItsTier`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] The lint reads its per-tier INTRODUCES set from the READMEs rather than a hand-maintained Go table (`go test ./internal/conformance/ -run '^TestOpTableDerivedFromReadmes$' -count=1 -v | grep -q '^--- PASS: TestOpTableDerivedFromReadmes'`)
- [ ] Every tier directory has a README stating what it introduces and what it depends on (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] A construct in a README that no file in that tier introduces is reported (`go test ./internal/conformance/ -run '^TestCorpusLints$' -count=1 -v | grep -q '^--- PASS: TestCorpusLints'`)
- [ ] The spec's open-questions section is gone, both questions having been answered rather than merely asked (`! grep -q '^## Open questions' docs/plans/2026-09-21-graded-conformance-corpus.md`)
- [ ] Every tier has an adjacency topic covering its full INTRODUCES set, checked generically so no tier can omit one (`go test ./internal/conformance/ -run '^(TestEveryTierHasAnAdjacencyTopic|TestAdjacencyCoversItsTier)$' -count=1 -v | grep -cE '^--- PASS: (TestEveryTierHasAnAdjacencyTopic|TestAdjacencyCoversItsTier) ' | grep -qx 2`)
- [ ] A tier README declaring a prerequisite at a later tier number fails the lint, which is check 2 given teeth (`go test ./internal/conformance/ -run '^TestTierNecessityCheck$' -count=1 -v | grep -q '^--- PASS: TestTierNecessityCheck'`)
