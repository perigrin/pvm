---
title: "`&m('x')` opens a match: a quote-operator word after & is not read as a name"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:46:32.849199889Z
updated: 2026-09-29T16:01:57.309433562Z
sessions:
- start_sha: 8cf105136f0824455290a02f0d8f8fb3b89326e5
  end_sha: 5436629d157e3c7ed7335113d825ff3f9b4ecd44
  commits: 1
  started_at: 2026-09-29T15:46:34.605962646Z
  ended_at: 2026-09-29T16:01:57.309433562Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T15:46:34.605962646Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T16:01:57.309433562Z
observed_paths:
- internal/lexer/ampquotename_test.go
- internal/lexer/quote.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

After the `&` sigil a quote-operator word opens a quote instead of naming a sub: `&m('amper')` lexes as a match whose delimiter is `(`, and likewise for q, qq, qr, qw, qx, s, tr and y.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'sub m {1} sub s {2} sub y {3} my @r = (&m("a"), &s("b"), &y("c"), &qw("d"));'
    my(@r) = (&m('a'), &s('b'), &y('c'), &qw('d'));

perl.git t/comp/opsubs.t:122-179.

## Acceptance criteria

- [ ] A quote-operator word after `&` is a name (`go test ./internal/lexer/ -run '^TestAmpQuoteOpName$' -count=1 -v | grep -q '^--- PASS: TestAmpQuoteOpName'`)
