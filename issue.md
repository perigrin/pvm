---
title: An export list built from %EXPORT_TAGS imports nothing
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T20:35:20.810131487Z
updated: 2026-09-30T20:35:21.342324736Z
sessions:
- start_sha: 19a92bef3e16a52ec1d6d7740b40c69a8e58401d
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T20:35:21.342324736Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T20:35:21.342324736Z
---

`use Hash::Util::FieldHash qw(fieldhash id); fieldhash my %fields;` refuses. perl 5.42's Hash/Util/FieldHash.pm builds its export list from its tags: `our %EXPORT_TAGS = ( 'all' => [ qw(fieldhash ...) ] ); our @EXPORT_OK = ( @{ $EXPORT_TAGS{'all'} } );`. The reader took that list as computed, marked the module opaque, and imported nothing. `fieldhash` is declared in Perl with prototype `(\%)`, and perl measures `prototype "main::fieldhash"` as `\%` after the import.

`:tag` in an import list was also ignored, silently leaving the sub table "complete" without the tag's names. A tag the reader can't resolve now marks the table incomplete. A quiet pragma's arguments are exempt, since `use open qw(:utf8 :std)` names layers, not Exporter tags.

PerlOnJava unit/hash_util_fieldhash_identity.t:7 and unit/interpreter_chained_method_invocant_lifetime.t:9.

## Acceptance criteria

- [ ] %EXPORT_TAGS-built export lists and `:tag` imports resolve (`go test ./internal/parse/ -run '^TestExportTags$' -count=1 -v | grep -q '^--- PASS: TestExportTags'`)
- [ ] A pragma's `:tag`-shaped argument leaves the table complete (`go test ./internal/parse/ -run '^TestPragmaTagIsNotAnExport$' -count=1 -v | grep -q '^--- PASS: TestPragmaTagIsNotAnExport'`)
