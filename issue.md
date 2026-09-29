---
title: "XS subs are unknown: a module's .xs names and prototypes are never read"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T23:08:35.758163908Z
updated: 2026-09-29T23:15:29.674660654Z
sessions:
- start_sha: fa2f9a5dc1174e5734013c5815469461ce48ea07
  end_sha: 85a1b24708885d3b391109f081d755998f949c7c
  commits: 1
  started_at: 2026-09-29T23:08:37.184290914Z
  ended_at: 2026-09-29T23:15:29.674660654Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T23:08:37.184290914Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T23:15:29.674660654Z
observed_paths:
- internal/parse/module.go
- internal/parse/perlgitt_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
- internal/parse/xs.go
- internal/parse/xs_test.go
---

Subs a module implements in XS are invisible to the parser. `Time::HiRes::usleep $big_delay;`, `Dump \%hash;` after `use Devel::Peek`, and `mro::get_isarev $args[0]` all refuse, because neither the name nor its prototype appears in any `.pm`. perl knows both once the module loads, and both come from the module's `.xs` by xsubpp's rules (ExtUtils::ParseXS; Node.pm's proto_string):
- prototypes are off unless `PROTOTYPES: ENABLE`;
- each parameter is `$`;
- a `;` goes before the first parameter with a default;
- `...` adds `;@`;
- a `PROTOTYPE:` line wins.

Two gaps stand in the way:
1. The resolver never reads XS.
2. perl.git's unbuilt checkout keeps these modules at the top of their distribution directories (ext/Devel-Peek/Peek.pm, dist/Time-HiRes/HiRes.pm, ext/mro/mro.pm). MakeMaker installs them from there, but DirLoader looks only under `lib/`.

perl.git t/io/eintr_print.t:61, t/op/hash-clear-placeholders.t:26, t/mro/isarev.t:19 and t/mro/isarev_utf8.t:22.

## Acceptance criteria

- [ ] XSSubs reads names and prototypes by xsubpp's rules (`go test ./internal/parse/ -run '^TestXSSubs$' -count=1 -v | grep -q '^--- PASS: TestXSSubs'`)
- [ ] A module in source-tree layout loads with its XS subs known (`go test ./internal/parse/ -run '^TestXSModuleLoads$' -count=1 -v | grep -q '^--- PASS: TestXSModuleLoads'`)
