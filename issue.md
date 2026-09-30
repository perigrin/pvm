---
title: "Core XS shapes are invisible to T1: read .xs from the perl.git checkout"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T21:55:19.017136126Z
updated: 2026-09-30T22:06:31.644346589Z
sessions:
- start_sha: 0bc0f3fb5263c6cd3ab48bfb5280a344d6171f9d
  end_sha: 555b4ff88e27fe5ed3f19052f9d50091492432ac
  commits: 1
  started_at: 2026-09-30T21:55:21.511706334Z
  ended_at: 2026-09-30T22:06:31.644346589Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T21:55:21.511706334Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T22:06:31.644346589Z
observed_paths:
- internal/parse/exports.go
- internal/parse/indirectvar_test.go
- internal/parse/module.go
- internal/parse/perlgitt_test.go
- internal/parse/ratchet_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
- internal/parse/xs_test.go
---

An installed perl ships no .xs files, so with perl 5.42's @INC as T1's loader, XS subs have no visible shape. Storable's `dclone` is `$` only in dist/Storable/Storable.xs. perigrin decided (2026-09-30) to read core modules' .xs from the perl.git checkout: T1's roots gain its dist/, cpan/ and ext/ after @INC.

Two gaps came out of it:
- A package an .xs declares beyond the module's own wasn't known. cpan/Compress-Raw-Bzip2/Bzip2.xs declares Compress::Raw::Bunzip2, and perl 5.42 reads `new Compress::Raw::Bunzip2(1, 1)` as `'Compress::Raw::Bunzip2'->new(1, 1)`.
- A computed glob assigned a variable reference counted as possibly defining subs. Test/More.pm:210's `*{"$to\::TODO"} = \our $TODO;` left every Test::More file's sub table incomplete.

PerlOnJava unit/storable.t:255 and unit/compress_raw_bzip2.t:29.

## Acceptance criteria

- [ ] An XS-declared package is known (`go test ./internal/parse/ -run '^TestXSPackageIsKnown$' -count=1 -v | grep -q '^--- PASS: TestXSPackageIsKnown'`)
- [ ] A variable glob keeps the sub table complete (`go test ./internal/parse/ -run '^TestVariableGlobKeepsTableComplete$' -count=1 -v | grep -q '^--- PASS: TestVariableGlobKeepsTableComplete'`)
