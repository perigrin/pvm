---
title: "Correct the documents the delivery left behind: possible-undef in the RFC, the warn flags in help, stale RFC lines, guide errors"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.465984087Z
updated: 2026-10-20T19:19:21.15613291Z
sessions:
- start_sha: cf08c0acdb986b7a59c463c2ed994e0721307294
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T19:19:21.15613291Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T19:19:21.15613291Z
---

crochet:review of PR #488, round 1 (2026-10-21). Coverage 2-9 and quality S8, S9, S12: the RFC never states the possible-undef default (perigrin 2026-10-20: silent unless proven, reported under --warn-possible-undef); internal/data/help/types.md documents neither --warn-possible-undef nor --warn-lossy-coercion; RFC :72 lists only Moose, Moose::Role, Mojolicious::Lite as shipped; RFC :945-947 still says Call sites waits on m2; RFC :1153-1157 names only CORE.tpm as generated from perl (skeleton mode too); RFC :1078 claims the oracle runs with no network (nothing isolates it -- state what does); RFC :1068-1071 overclaims the typed-our refusal; guide :132 shows Str $import (shipped: ArrayRef); guide :52-54 says workers strip & (the driver does); IO/Socket/IP.tpm header names no module version; narrow_lib_test.go:365 comment points at a scratchpad path.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Fix each claim against the code; no new claims without a test behind them

## Acceptance criteria

### Positive Scenarios
- [ ] The help documents --warn-possible-undef (`grep -q 'warn-possible-undef' internal/data/help/types.md`)
- [ ] The help documents --warn-lossy-coercion (`grep -q 'warn-lossy-coercion' internal/data/help/types.md`)
- [ ] The RFC states the possible-undef default (`grep -q 'warn-possible-undef' docs/rfcs/0001-type-libraries.md`)
- [ ] IO::Socket::IP's header names its version (`head -3 internal/parse/declarations/IO/Socket/IP.tpm | grep -qE '[0-9]+[.][0-9]+'`)

### Negative Scenarios
- [ ] The RFC no longer says Call sites waits on m2 (`! grep -q 'It waits on .infer. running on' docs/rfcs/0001-type-libraries.md`)
- [ ] The guide no longer shows Str $import (`! grep -q 'Str \$import = undef' docs/contributing/core-module-declarations.md`)
- [ ] No test comment points into a scratchpad (`! grep -rq 'scratchpad' internal/infer/narrow_lib_test.go`)
