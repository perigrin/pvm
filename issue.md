---
title: "Core XS shapes are invisible to T1: read .xs from the perl.git checkout"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T21:55:19.017136126Z
updated: 2026-09-30T21:55:19.017136126Z
---

An installed perl ships no .xs files, so with perl 5.42's @INC as T1's loader, XS subs have no visible shape. Storable's `dclone` is `$` only in dist/Storable/Storable.xs. perigrin decided (2026-09-30) to read core modules' .xs from the perl.git checkout: T1's roots gain its dist/, cpan/ and ext/ after @INC.

Two gaps came out of it:
- A package an .xs declares beyond the module's own wasn't known. cpan/Compress-Raw-Bzip2/Bzip2.xs declares Compress::Raw::Bunzip2, and perl 5.42 reads `new Compress::Raw::Bunzip2(1, 1)` as `'Compress::Raw::Bunzip2'->new(1, 1)`.
- A computed glob assigned a variable reference counted as possibly defining subs. Test/More.pm:210's `*{"$to\::TODO"} = \our $TODO;` left every Test::More file's sub table incomplete.

PerlOnJava unit/storable.t:255 and unit/compress_raw_bzip2.t:29.

## Acceptance criteria

- [ ] An XS-declared package is known (`go test ./internal/parse/ -run '^TestXSPackageIsKnown$' -count=1 -v | grep -q '^--- PASS: TestXSPackageIsKnown'`)
- [ ] A variable glob keeps the sub table complete (`go test ./internal/parse/ -run '^TestVariableGlobKeepsTableComplete$' -count=1 -v | grep -q '^--- PASS: TestVariableGlobKeepsTableComplete'`)
