---
title: Re-pin the parse oracle to perl 5.44.0 and record the starting counts
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16666-9a1b-76b3-9389-e4dbfb01ab9d
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a16602-5e01-7e7a-92c5-27d7da7d6f36
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
- 01a16602-6146-76ab-bb68-f3e8b645ae79
- 01a1664d-6fdb-7a09-8211-882e814db87c
created: 2026-10-23T01:08:31.856618478Z
updated: 2026-10-23T01:56:39.493426906Z
---

## Context

- paths: `internal/parseoracle/testdata/corpus.pin`, `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/m3start.txt` (new), `internal/parseoracle/testdata/ratchet/compiled.txt` (new), `internal/parse/compiledset_test.go` (new), `.github/scripts/parseoracle-pin.sh`, `.github/workflows/parse-fidelity.yml`, `.github/workflows/ci.yml`, `internal/parseoracle/ci_test.go`, `internal/parseoracle/README.md`
- commands: `PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parseoracle/testdata/corpus.pin`, `internal/parseoracle/corpus.go` (ReadPin, shim build), `internal/parseoracle/ci_test.go` (TestPinScriptAgreesWithReadPin, TestCIPinsTheWorldItMeasuresIn)

**perigrin's ruling (2026-10-23): judge the oracle with perl 5.44.0.** The corpus is perl5 blead at 94e50866, which is perl 5.45.0. Each perl was measured with its own shim (`perl -c`, plus `-T` for the taint tests). 5.42.0 compiles 589 of the 594 non-excluded files, and 5.44.0 compiles all 594. Under 5.42, class/inherit.t and perf/opcount.t segfault; class/method.t and op/signatures.t fail on the `experimental::signature_named_parameters` warnings category; and op/for-many.t has a syntax error at line 474. The 26 excluded files (porting/, win32/, bigmem/, op/goto_xs.t) stay excluded. **Scope:** only the parse oracle's pin moves (corpus.pin, the shim, and CI's ratchet and fidelity jobs). The `.tpm` libraries, the conformance corpus and the CORE.tpm check stay on 5.42.0. ci.yml reads the same pin script today (ci.yml:42, "Read the perl pin"), so those jobs need a pin of their own that stays at 5.42.0.

This issue is the chain's root. Every starting number (the rv2hv wider count, the site floor, the compile and graded counts) is re-measured here or downstream at the new pin. It records the starting counts in `testdata/ratchet/m3start.txt` as `key = value` lines. Later criteria compare against that file, never against a number typed into an issue (389, 593). Locally 5.44.0 is at `$HOME/.local/share/pvm/versions/5.44.0/bin/perl`. Tests and the oracle find it through `perltest.PinnedPerl` (the pinned-perl helper issue), so no command needs a PATH prefix.

**The compiled set (goal (4): "measured against the current compile count").** This issue records the set BY PATH in `testdata/ratchet/compiled.txt`, one path per line. A file is in it when the pinned perl accepts it under `perl -c`, with `-T` where the file's shebang line has it. Files in perl.git's `perlRejects` list (internal/parse/perlgitt_test.go) are left out even when `perl -c` exits 0. comp/final_line_num.t is the case: it ends in `print 1+`, and its BEGIN `__DIE__` handler prints `ok 1` and exits 0 (measured at 5.44: `1..1 / ok 1`, rc 0). re/pat_psycho.t IS in the set. `perl -c` accepts it (rc 0, ~2:50 wall at 5.44), and parsing it needs no oracle, even though the oracle sweep errors on it at its 8m budget. m3start.txt's `compiled` is the line count of compiled.txt. `graded` is recorded too, but the two are different sets and nothing compares them as if they were one.

## Steps

- [ ] Use the pvm-installed perl 5.44.0 at `$HOME/.local/share/pvm/versions/5.44.0` (present on this machine, measured threads=define on 2026-10-23); install it with pvm only if absent. Check `$]` and `useithreads`, which must be `define` (see corpus.pin's comment on threads).
- [ ] Rebuild the shim at `$HOME/.cache/pvm/oracleshim` from 5.44's @INC.
- [ ] Measure with that perl: run `perl -c` (`-T` where the shebang has it) over the 594 non-excluded files in the shim. Write the accepted paths, minus `perlRejects`, to compiled.txt. Record `interpreter`, `compiled` (compiled.txt's line count) and `graded` in m3start.txt. Quote any file that fails, with its stderr, in the commit message.
- [ ] Create `internal/parse/compiledset_test.go` with two `// ABOUTME:` lines. It holds `readCompiledSet`, the one reader of compiled.txt that the parse-without-error issue reuses. RED/GREEN TestCompiledSetExcludesPerlRejects: no `perlRejects` path is in compiled.txt, compiled.txt has no duplicate or blank lines, and its line count equals m3start.txt's `compiled`.
- [ ] RED: write TestCIWorkflowsPinTheirPerls in ci_test.go. parse-fidelity.yml installs corpus.pin's interpreter. ci.yml's 5.42 jobs install 5.42.0 from their own pin. Those are the Test matrix, the conformance and CORE.tpm steps, and "Ratchets (corpus-backed)", which runs internal/parse with `PVM_REQUIRE_PERL542=1`. The test names every job and the perl it installs, so a job added later must be classified. Run it after changing only corpus.pin and watch it fail.
- [ ] GREEN: set corpus.pin to `interpreter = 5.044000`, with threads and revision as observed. Give ci.yml's 5.42 jobs their own pin source, and update the corpus.pin comment that explains for-many.t:474 under 5.42. Run TestCIWorkflowsPinTheirPerls, TestPinScriptAgreesWithReadPin and TestCIPinsTheWorldItMeasuresIn, and watch them pass.
- [ ] RED/GREEN: write TestM3StartRecordsThePin. m3start.txt has `interpreter`, `compiled` and `graded` keys, and its `interpreter` equals corpus.pin's. Add TestM3StartRejectsInterpreterSkew: a fixture m3start whose interpreter differs from the pin fails.
- [ ] RED/GREEN: write TestCIWorkflowsPinTheirPerlsRejectsSharedPin. A fixture workflow where a 5.42 job reads corpus.pin, or the 5.44 job reads the 5.42 pin, fails.
- [ ] The pinned-perl helper issue is done first, so the sweep, `ObservePin` and parse_facts.pl's inner compiles all run 5.44.0 with no PATH prefix. Its change to parse_facts.pl changes the script's hash, which keys the oracle cache, so this sweep re-measures every file once; that is expected.
- [ ] Re-baseline baseline.txt with `-parseoracle.update` under 5.44.0. Re-baseline the CI gate's fixture baseline, `testdata/ratchet/fixture.txt` (today `interpreter = 5.042000`), at 5.44.0 too, since the gate's sweep CheckPins against it. Then re-check every other ratchet that reads the oracle (the exact-match floor TestGoParserExactFloor, and any fidelity ratchet parse-fidelity.yml runs), and re-baseline any that moves. The commit message explains every moved file (for example, op/signatures.t no-answer -> graded).
- [ ] Record the sweep's wall time. re/pat_psycho.t times out at PARSEORACLE_TIMEOUT=8m on every sweep (errors are not cached), and that counts against `git zhi verify --timeout 900`. If the sweep alone comes within 2 minutes of 900s, stop and raise it with perigrin before going on.
- [ ] Update internal/parseoracle/README.md's "Ground-truth ceiling" sentence (line ~287) to the compiled count measured with 5.44.0. Leave the historical sentence at ~309 alone.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Precondition (passes before this issue's work): perl 5.44.0 with threads is installed in pvm's versions directory (`sh -c '$HOME/.local/share/pvm/versions/5.44.0/bin/perl -MConfig -e "exit !(\$] == 5.044000 && \$Config{useithreads} eq q(define))"'`)
- [ ] The oracle pin names 5.44.0 (`grep -q '^interpreter = 5.044000' internal/parseoracle/testdata/corpus.pin`)
- [ ] The starting counts are recorded against the pin (`go test ./internal/parseoracle/ -run '^TestM3StartRecordsThePin$' -count=1 -v | grep -q '^--- PASS: TestM3StartRecordsThePin'`)
- [ ] CI's oracle jobs install the pinned 5.44.0, and the conformance and CORE.tpm jobs stay on 5.42.0 (`go test ./internal/parseoracle/ -run '^TestCIWorkflowsPinTheirPerls$' -count=1 -v | grep -q '^--- PASS: TestCIWorkflowsPinTheirPerls'`)
- [ ] Regression guard: the pin script still agrees with the pin file (`go test ./internal/parseoracle/ -run '^TestPinScriptAgreesWithReadPin$' -count=1 -v | grep -q '^--- PASS: TestPinScriptAgreesWithReadPin'`)
- [ ] The per-file baseline was regenerated under 5.44.0 (`grep -q '^interpreter = 5.044000' internal/parseoracle/testdata/ratchet/baseline.txt`)
- [ ] The compiled set is recorded by path, leaves out perl.git's perlRejects files, and matches m3start's count (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestCompiledSetExcludesPerlRejects$' -count=1 -v | grep -q '^--- PASS: TestCompiledSetExcludesPerlRejects'`)
- [ ] The CI gate's fixture baseline is regenerated at the 5.44.0 pin (`grep -q '^interpreter = 5.044000' internal/parseoracle/testdata/ratchet/fixture.txt`)

### Negative Scenarios

- [ ] An m3start.txt whose interpreter differs from corpus.pin fails (`go test ./internal/parseoracle/ -run '^TestM3StartRejectsInterpreterSkew$' -count=1 -v | grep -q '^--- PASS: TestM3StartRejectsInterpreterSkew'`)
- [ ] A 5.42 job reading the oracle pin, or the oracle job reading the 5.42 pin, fails the workflow check (`go test ./internal/parseoracle/ -run '^TestCIWorkflowsPinTheirPerlsRejectsSharedPin$' -count=1 -v | grep -q '^--- PASS: TestCIWorkflowsPinTheirPerlsRejectsSharedPin'`)
- [ ] comp/final_line_num.t is not in the compiled set even though perl -c exits 0 on it (`sh -c 'f=internal/parseoracle/testdata/ratchet/compiled.txt; test -s $f && ! grep -qx comp/final_line_num.t $f && grep -qx re/pat_psycho.t $f'`)
