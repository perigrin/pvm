---
title: Inc/dec, logical not, unary plus and the file tests, none as a subject
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.684116637Z
updated: 2026-09-23T19:06:34.274227011Z
sessions:
- start_sha: c2adc1f75f71cd96e8cf5067a659da2199c28462
  end_sha: d95cf4eec7360982afdd81e0ca82243e4e78eac9
  commits: 5
  started_at: 2026-09-23T17:44:20.265778898Z
  ended_at: 2026-09-23T19:06:34.274227011Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T17:44:20.265778898Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T19:06:34.274227011Z
observed_paths:
- conformance/04_operators/00_adjacency.t
- conformance/04_operators/27_incdec.t
- conformance/04_operators/28_string_increment.t
- conformance/04_operators/29_logical_not.t
- conformance/04_operators/30_unary_plus.t
- conformance/04_operators/31_file_test.t
- conformance/04_operators/32_chained_comparison.t
- conformance/04_operators/33_nonassoc_refusal.t
- conformance/04_operators/README.md
- conformance/07_subroutines/03_call_forms.t
- conformance/07_subroutines/README.md
- conformance/09_regex/README.md
- conformance/11_oo/16_classname_ungated.t
- conformance/11_oo/17_defer_ungated.t
- conformance/12_packages/10_compile_tokens.t
- conformance/12_packages/11_line_directive.t
- conformance/README.md
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier04_test.go
---

Four gaps at perlop's unary and named-unary levels, grouped because each
is small and they share a failure mode: an operator the corpus never
writes as a SUBJECT.

**`++` and `--` (level 3).** The corpus emits `postinc` only, as a side
effect inside `07_args_alias.t` and an adjacency file. No `preinc`, no
`predec`, no `postdec` anywhere. No file's subject is inc/dec, nothing
claims the pre-versus-post return value, and the magic string increment
(`my $s = "Az"; $s++` gives `"Ba"`) is unclaimed -- a behaviour no other
operator in the language has.

**`!` (level 5).** The most common operator in Perl that this corpus
never writes. Grep finds only `!=` and the `!` delimiter in `q!...!` and
`m!...!`. That is itself the hazard: `!` is a delimiter character in this
corpus and an operator nowhere.

**Unary `+` (level 5).** Absent, and it is the ONLY disambiguator for
`print (1+2)*3` -- where `print (1+2)` is a complete call and `*3` is
discarded. perlop documents `+` before a paren as the fix. A corpus that
never writes it cannot claim the ambiguity it resolves.

**File-test operators (level 10).** `-e -d -f -s -z -r -w -x -M -A -C`
and the rest: an entire named-unary family, zero coverage. `-e $file`
against `-e($file)` against a bareword `-e` is a real lexer fork, and
tier 01 already found that `-1` is two tokens -- this is the same
question in operator position.

## What catches these

Token facts, mostly. `$n++` and `++$n` differ in token ORDER, which the
stream shows and the output can also show when both appear. The string
increment is pinned by output. The `print (1+2)*3` case is pinned by
output, and sharply: it prints `3`, not `9`.

## Files

Four, in `04_operators`: inc/dec with both fixings and the string form;
`!` with its negation of a runtime value; unary `+` with the
`print (...)` disambiguation pinned by output; and one file-test operator
with its argument in both spellings.

## Acceptance criteria

- [ ] Tier 04 holds a file whose subject is inc/dec, including the string increment (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t | grep -q '++'`)
- [ ] A file pins that print (1+2)*3 prints 3 rather than 9, which only output can show (`go test ./internal/conformance/ -run '^TestCorpus/04_operators' -count=1 -v | grep -q '^    --- PASS: TestCorpus/04_operators'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
