# Milestone m3-oracle-at-scale Postmortem

The last stretch ran in a cloud session after the session on shin stopped,
from feat/m3-oracle-at-scale at 7a4b0cc1. It closed 6030, 608d, 60e9, 75e8
and 61fe, and with them two v0.1 issues the same causes covered (01a16a7a,
01a16b51).

## What Worked Well

- **Per-fix review against perl, with new probes, caught every regression
  before merge.** 6030's first review found two regressions ae4be0f2 had
  introduced (a mid-line `__END__` dropping a `;`; `goto x` and `last
  default` read as calls, op/goto.t:755) and one it had made worse.
  The fixes went through three more rounds. Each round found defects in the
  previous round's own rules (`__END__::foo` read as data, `__END__ :` with
  a blank, `->` after a label, a modifier over an operator chain), and the
  fourth approved. The reviews added 129 probes to stmt-boundary.txt. 608d's
  review ran about 230 probes and found no grouping defect.
- **Fixing canon at the cause emptied the queue.** The eight
  canon-adds-paren rows had five causes, and none needed a forgiveness
  rule: a blank instead of a pair, a prefix operator on `**`'s right, the
  ternary middle at assignment power, a `->` call as a subscript's base,
  and one misgrouping in the parser (a declaration under a tighter
  operator). The misgrouping fix also closed 01a16a7a, and the
  subscript-base rule closed 01a16b51.
- **perl's own bits settled a wording question.** The first draft said `-
  -$x` and `-(-$x)` were "one tree". Concise showed `negate sK` and `negate
  sKP`: the pair is perl's OPf_PARENS. So the blank is not just faithful
  to the source, it is the source's tree.

## What Didn't Work

- **A targeted test run stood in for the full suite.** 608d was committed
  after running only the tests it named. The full `internal/parse` run
  then failed seven tests that pinned canon's old pair spelling or built
  fixtures from queue rows that no longer existed. The per-fix reviewer
  found the same seven independently. A test file that called PinnedPerl
  without a CI step requiring it was caught only by `make test`
  (TestEveryPinnedPerlTestIsRequired).
- **The fix rounds did not converge in one.** As in core-module-types,
  each round's new rules (the XState data section, the label operand)
  carried edge cases the next round's probes found.
- **The environment was not shin's.** The work started without it:
  - perl 5.44.0 was not installed. pvm has no release binary for it on
    this platform and no embedded checksum, so it was built from cpan.org
    source after checking the published sha256.
  - pvm's 5.42.0 release binary has a RUNPATH fixed to `/home/runner`,
    so it cannot run anywhere else.
  - pvm's HEAD-based binary check fails behind the proxy.
  - `T1_CORPUS` pointed at the PerlOnJava root, not `src/test/resources/unit`.
- **Two tests pin one machine's build.** TestConfigValuesFitDeclaredType
  (172 undef `%Config` values) and TestCoreModuleRanking (128 modules, 1427
  imports) fail on every 5.42.0 built here: from source, and by pvm,
  static libperl, 163 and 131/1437. They fail identically at 7a4b0cc1.
- **The tracker could not be pushed from the session.** Branch pushes go
  through the session's git proxy; `refs/zhi/*` pushes are refused with a
  403. Every issue state change was made locally and had to be pushed from
  elsewhere.

## What Puzzles Us

- Why the reviews' probe tables, run over more than 1,000 shapes, kept
  finding defects in rules a few lines long. Each defect was in a shape
  the rule's author had not listed: the rule was right for its examples
  and wrong at its edges.
- `make test` rewrites test/e2e/.perl-version on every run, the same local
  edit shin carried, so a worktree that runs it is left dirty.

## Where A Human Had To Act

- perigrin restored GitHub access mid-session. The `refs/zhi/*` push
  refusal needs a session or proxy setting, or a push from shin.
- The two machine-pinned tests need perigrin's ruling: pin them to a
  reproducible build, or loosen them to the property they measure.
- The pre-existing defects the reviews found are listed for v0.1 issues,
  not fixed here: `last when`, `last L, 1`, `next, print`, `return last or
  1`, `print "a\n"; use strict` with no last `;`, `\my $x :shared`, `\my Foo
  $x`, and `print STDOUT ("x\n")`, which Faithful passes but perl runs as a
  call to STDOUT.

## Recommendations

- Run the full package suite, not the named tests, before every commit;
  `make test` before every merge, as the chain's own rule says.
- Give each new rule its edge-case table at the time it is written: every
  token that can follow the word it tests, measured on perl.
- Make the environment check part of a session start: the pinned perl in
  pvm's versions directory, `T1_CORPUS` at `unit/`, and a test push of a
  `refs/zhi/` ref.
