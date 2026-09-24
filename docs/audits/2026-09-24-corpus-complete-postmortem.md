# corpus-complete: postmortem

**Date:** 2026-09-24
**Milestone:** `corpus-complete`, 13/13
**Range:** `cbaacf10..fd1073d0`, 18 commits

## Telemetry

    MPG                 1.38 commits/issue
    speed               107.3 issues/week
    buffer              6.5 total, 4.4 burned
    fever               GREEN
    time-in-chain       24%
    shadow work         76%
    rework rate         23%
    review cycles       0.23/issue
    verification cov    100%
    WIP violations      1

Two numbers carry the story: **76% shadow work** and **23% rework**.

## 1. What worked well

**The gates found what I did not.** Four defects shipped and three were
caught by a PAAD review rather than by me. Every one was a plausible
argument that measurement refuted:

- a vacuous token fact, in the commit that fixed two others
- a header passage pasted into four files without checking whether it
  applied -- written as the fix for a header passage pasted into four
  files without checking whether it applied
- a four-name exception wrong on `__CLASS__` and missing `__SUB__`,
  which made a corpus file pass for the wrong reason
- a "paired negative" exemption built to keep ten facts, which one
  measurement (delete `->` from the lexer; only positives fail) showed
  was a rationalization

Review cycles are 0.23/issue and rework is 23%. Those look like bad
numbers. They are the system working: without the gates the defects
would have shipped silently and the corpus would assert less than it
appears to.

**Measurement kept overruling framing, and the framing kept being
wrong.** The issue that tracked "four parser gaps" was three-quarters
one bug. `#line` "already worked." `print FOO, "\n"` refuses the same
as `print __PACKAGE__, "\n"`, so the problem was never compile-time
tokens. Each correction came from running something, not from thinking
harder.

**The corpus checks itself.** 158 Go tests over `conformance/`, 84 of
them tier checks. A construct cannot be introduced without being
declared, an adjacency file cannot omit what its tier introduces, and
as of this milestone a token fact cannot assert nothing.

**TDD held where it was applied.** Every parser change had a failing
test first, and two of them (`startsTerm`, the comma rule) turned out
to fix far more than the case that motivated them -- `base/num.t` went
44 refusals to 6.

## 2. What didn't work

**I argue past tells.** The pairing exemption is the clearest case: I
noticed that one file did the same job with no arrow fact, and that
another had its identical fact deleted for want of a partner, and wrote
a paragraph explaining why the asymmetry was fine.
It was not fine. The rule was tracking which negatives another test
happened to require.

The pattern: when a measurement contradicts a structure I have just
built, I reach for a reason the structure survives rather than
re-running the measurement against it.

**Plausible prose is this codebase's failure mode and I kept producing
it.** The `->` passage read as careful reasoning and was exactly
backwards. `scanWord` emits source bytes; no keyword table can make
`push` arrive as `unshift`. Both were sentences that sounded like
measurement and were not.

**Name lists where a rule exists.** The `compileTimeToken` list was
wrong twice over because it encoded a judgement per name that perl
encodes structurally. The root-cause fix -- ask what FOLLOWS the word
-- is smaller, needs no maintenance, and gets feature-gated tokens
right in both states.

**76% shadow work.** Three-quarters of what got done was not in the
chain when the milestone started. Some of that is healthy discovery.
Some is that the milestone's issues were written from a five-axis
measurement that had never been checked against the corpus's own
internal consistency -- so the largest finding (49 vacuous facts) was
invisible to the plan.

**One WIP violation**, reported by sanbao. I could not identify which
pair caused it: the two issues I suspected have non-overlapping session
windows (`01a0cfb2` closed 21:33, `01a0cf3e` started 22:36), and
`issue list --format json` returns an empty array for done issues, so
the session data is not reachable after close. Recording the metric
without a cause rather than guessing one.

## 3. What surprised us

**The corpus was lying to itself at scale.** 49 of 115 negative facts
could never fail. The corpus's entire value is that its claims are
checkable, and 43% of one claim class was not.

**A one-line parser bug was worth more than the whole milestone's
corpus work, by one measure.** `startsTerm` calling every Word a
term-starter cost 14 T1 files and 38 refusals in `base/num.t` alone.
It was found only because a narrower fix exposed it.

**`#line` passes while unimplemented.** The corpus file's own header
warns about a lexer that treats the directive as a comment "silently."
Ours does. The file passes because its output is pinned from perl and
its parse succeeds by skipping the directive as trivia.

**Closing an issue can break a citation.** Three corpus files cite
issues that closed because their MILESTONE finished, not because
anything was fixed.

## 4. What to do differently

**Write the falsification before the structure.** For the pairing
exemption the question was one command: "what broken lexer does this
catch that the positive alone does not?" Asking it takes a minute and
would have prevented the commit.

**Prefer the rule to the list.** When perl distinguishes two cases
structurally, encode the structure. Every name list in this milestone
was wrong or incomplete within one commit of being written.

**A header claim is a claim.** Corpus headers carry measurements and
this milestone shipped false ones repeatedly. They should be checked
the way facts are -- at minimum, every `MEASURED` block should be
re-runnable and re-run before commit.

**Check the corpus against itself, not only against Perl.** The
five-axis measurement asked "what of Perl is missing." Nothing asked
"do the claims we already make hold up." That second question produced
the milestone's biggest result and was nobody's issue.

**Make citation liveness mechanical.** Filed as `01a0d0c7`. A citation
nothing reads is a citation that rots.

## Carried forward

Six open issues, each filed with its measurement:

    01a0ce57   the lexer does not form `x=`
    01a0cf64   `-e` lexes as a minus and a word
    01a0d087   ungated `defer` is declined as a statement form
    01a0d0b0   `#line` is lexed as a comment; its case passes anyway
    01a0d0b1   `print {$fh} "x"` refuses
    01a0d0c7   five refusal citations point at closed issues

`01a0d1fa` was filed here and is CANCELLED: it recorded three false
measured claims in `.t` headers, and those files no longer exist. The
mdtest copies were repaired when the port found them.

The plan's two deliberate deferrals -- the `t/` sweep and adjacency
beyond the adjacent pair -- belong to the next milestone and are named
in the plan rather than dropped.

## Written before the format port

This postmortem was written against a corpus of 212 `.t` files. That
corpus was then ported to 65 markdown topic files in Chalk's mdtest
format and the `.t` files deleted (`a6b33a2a`), which is the largest
single thing that has happened to it.

Everything above still holds. The findings are about what the corpus
CLAIMS and the port moved claims without changing them -- 212 cases
against 212 files, sources byte-identical, 24 recorded refusals on both
sides, all measured rather than asserted.

WHAT THE PORT CONFIRMED, and it is this document's fourth
recommendation reached independently. "Check the corpus against itself,
not only against Perl" was written here as a lesson; the port then
demonstrated the cheap version of it. Every tier was ported by one
agent and verified by a second, and the verify step MEASURED each prose
claim it carried across rather than trusting it. That found six false
claims -- three introduced by the port, three that had been sitting in
`.t` headers all along and were only noticed because someone had to
re-measure them to move them.

Two of the three inherited ones survived because they cited a NUMBER,
which reads as evidence. That is the same observation this document
makes about plausible prose, sharpened: a number is prose that looks
like measurement.

WHAT IT ALSO CONFIRMED about "prefer the rule to the list". The port
deleted 24 tests and 10 hand-maintained Go maps -- 2,657 lines -- every
one of which derived a construct's identity from a FILENAME and then
grepped for the spelling that identity mapped to. They were not wrong;
they were compensating for a format where a file was a construct. The
claim they asserted is now visible in the source, and perigrin had
called the machinery over-engineered before the port without needing
the measurement.

One thing this document could not have known and should be read with:
"76% shadow work" was measured against a milestone whose issues never
mentioned a format change. The port is more shadow work by that
definition, and calling it that would be wrong -- it was the right work
found by doing the wrong work carefully. The metric measures deviation
from a plan, not whether the deviation was correct.
