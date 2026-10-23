---
title: Re-pin the parse oracle to perl 5.44.0 and record the starting counts
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a16602-5e01-7e7a-92c5-27d7da7d6f36
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
created: 2026-10-23T01:08:31.856618478Z
updated: 2026-10-23T01:09:05.015015581Z
---

## Context

- paths: `internal/parseoracle/testdata/corpus.pin`, `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/m3start.txt` (new), `.github/scripts/parseoracle-pin.sh`, `.github/workflows/parse-fidelity.yml`, `.github/workflows/ci.yml`, `internal/parseoracle/ci_test.go`, `internal/parseoracle/README.md`
- commands: `PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parseoracle/testdata/corpus.pin`, `internal/parseoracle/corpus.go` (ReadPin, shim build), `internal/parseoracle/ci_test.go` (TestPinScriptAgreesWithReadPin, TestCIPinsTheWorldItMeasuresIn)

**perigrin's ruling (2026-10-23): judge the oracle with perl 5.44.0.** The corpus is perl5 blead at 94e50866, which is perl 5.45.0. Each perl was measured with its own shim (`perl -c`, plus `-T` for the taint tests). 5.42.0 compiles 589 of the 594 non-excluded files, and 5.44.0 compiles all 594. Under 5.42, class/inherit.t and perf/opcount.t segfault; class/method.t and op/signatures.t fail on the `experimental::signature_named_parameters` warnings category; and op/for-many.t has a syntax error at line 474. The 26 excluded files (porting/, win32/, bigmem/, op/goto_xs.t) stay excluded. **Scope:** only the parse oracle's pin moves (corpus.pin, the shim, and CI's ratchet and fidelity jobs). The `.tpm` libraries, the conformance corpus and the CORE.tpm check stay on 5.42.0. ci.yml reads the same pin script today (ci.yml:42, "Read the perl pin"), so those jobs need a pin of their own that stays at 5.42.0.

This issue is the chain's root. Every starting number (the rv2hv wider count, the site floor, the compile and graded counts) is re-measured here or downstream at the new pin. It records the starting counts in `testdata/ratchet/m3start.txt` as `key = value` lines. Later criteria compare against that file, never against a number typed into an issue (389, 593). Local runs put 5.44.0 at `$HOME/.local/perl-5.44.0/bin/perl`, and every oracle command in this milestone prefixes that directory onto PATH. The milestone's resolution command needs the same prefix.

## Steps

- [ ] Install perl 5.44.0 at `$HOME/.local/perl-5.44.0`, from `$HOME/.cache/pvm/binaries/5.44.0/linux-amd64/perl-5.44.0-linux-amd64.tar.gz` or a pvm install. Check `$]` and `useithreads`, which must be `define` (see corpus.pin's comment on threads).
- [ ] Rebuild the shim at `$HOME/.cache/pvm/oracleshim` from 5.44's @INC.
- [ ] Measure with that perl: run `perl -c` (`-T` where the file needs it) over the 594 non-excluded files in the shim. Record compiled and graded counts in m3start.txt (`interpreter`, `compiled`, `graded`). Quote any file that fails, with its stderr, in the commit message.
- [ ] RED: write TestCIWorkflowsPinTheirPerls in ci_test.go. parse-fidelity.yml installs corpus.pin's interpreter. ci.yml's conformance and CORE.tpm jobs install 5.42.0 from their own pin. Run it after changing only corpus.pin and watch it fail.
- [ ] GREEN: set corpus.pin to `interpreter = 5.044000`, with threads and revision as observed. Give ci.yml's 5.42 jobs their own pin source, and update the corpus.pin comment that explains for-many.t:474 under 5.42. Run TestCIWorkflowsPinTheirPerls, TestPinScriptAgreesWithReadPin and TestCIPinsTheWorldItMeasuresIn, and watch them pass.
- [ ] RED/GREEN: write TestM3StartRecordsThePin. m3start.txt has `interpreter`, `compiled` and `graded` keys, and its `interpreter` equals corpus.pin's.
- [ ] Re-baseline baseline.txt with `-parseoracle.update` under 5.44.0. Then re-check every other ratchet that reads the oracle (the exact-match floor TestGoParserExactFloor, and any fidelity ratchet parse-fidelity.yml runs), and re-baseline any that moves. The commit message explains every moved file (for example, op/signatures.t no-answer -> graded).
- [ ] Record the sweep's wall time. re/pat_psycho.t times out at PARSEORACLE_TIMEOUT=8m on every sweep (errors are not cached), and that counts against `git zhi verify --timeout 900`. If the sweep alone comes within 2 minutes of 900s, stop and raise it with perigrin before going on.
- [ ] Update internal/parseoracle/README.md's "Ground-truth ceiling" sentence (line ~287) to the compiled count measured with 5.44.0. Leave the historical sentence at ~309 alone.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Perl 5.44.0 with threads is installed where the milestone's commands expect it (`sh -c '$HOME/.local/perl-5.44.0/bin/perl -MConfig -e "exit !(\$] == 5.044000 && \$Config{useithreads} eq q(define))"'`)
- [ ] The oracle pin names 5.44.0 (`grep -q '^interpreter = 5.044000' internal/parseoracle/testdata/corpus.pin`)
- [ ] The starting counts are recorded against the pin (`go test ./internal/parseoracle/ -run '^TestM3StartRecordsThePin$' -count=1 -v | grep -q '^--- PASS: TestM3StartRecordsThePin'`)
- [ ] CI's oracle jobs install the pinned 5.44.0, and the conformance and CORE.tpm jobs stay on 5.42.0 (`go test ./internal/parseoracle/ -run '^TestCIWorkflowsPinTheirPerls$' -count=1 -v | grep -q '^--- PASS: TestCIWorkflowsPinTheirPerls'`)
- [ ] The pin script still agrees with the pin file (`go test ./internal/parseoracle/ -run '^TestPinScriptAgreesWithReadPin$' -count=1 -v | grep -q '^--- PASS: TestPinScriptAgreesWithReadPin'`)
- [ ] The per-file baseline was regenerated under 5.44.0 (`grep -q '^interpreter = 5.044000' internal/parseoracle/testdata/ratchet/baseline.txt`)
