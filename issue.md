---
title: "`use feature 'a', 'b'` turns nothing on: a bare comma list of names is not read"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:12:08.069544922Z
updated: 2026-09-29T15:12:09.817332259Z
sessions:
- start_sha: 9fdecbffea973c5465219b1cf4275e67bc552d02
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T15:12:09.817332259Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T15:12:09.817332259Z
---

`use feature 'evalbytes', 'unicode_eval';` turns nothing on: literalNameList reads qw(), one quoted string and a parenthesised list, but not the bare comma list perl allows after `use`, which parses as a comma Binary. So a parenless `evalbytes "..."` after it refuses.

Measured with perl 5.42:

    $ perl -MO=Deparse -e "use feature 'evalbytes', 'unicode_eval'; my \$x = evalbytes \"s\", 1;"
    my $x = evalbytes 's', '???';

perl.git t/op/evalbytes.t:19, then 32-35.

## Acceptance criteria

- [ ] A bare comma list of quoted names is literal (`go test ./internal/parse/ -run '^TestCommaSeparatedNameList$' -count=1 -v | grep -q '^--- PASS: TestCommaSeparatedNameList'`)
