---
title: "Import semantics the declarations lose: Getopt::Long :config and GetOptionsFromArray, POSIX, Mojolicious::Lite -signatures, Moose -traits"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.4290902Z
updated: 2026-10-21T02:51:05.56377892Z
sessions:
- start_sha: a2621ad1916c9a56ca43a75931e9c15d8f8929fa
  end_sha: 66180e7e19fc8ad1b81e1c0fb54bcc62942ffd1c
  commits: 5
  started_at: 2026-10-21T00:16:22.762511393Z
  ended_at: 2026-10-21T02:51:05.56377892Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T00:16:22.762511393Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T02:51:05.56377892Z
observed_paths:
- internal/infer/declared_test.go
- internal/infer/oraclestderr_test.go
- internal/infer/parametertypes_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/oracle/Compress-Raw-Zlib.pl
- internal/infer/testdata/oracle/Cwd.pl
- internal/infer/testdata/oracle/Data-Dumper.pl
- internal/infer/testdata/oracle/Encode.pl
- internal/infer/testdata/oracle/IO-Handle.pl
- internal/infer/testdata/oracle/List-Util.pl
- internal/infer/testdata/oracle/Scalar-Util.pl
- internal/infer/testdata/oracle/Storable.pl
- internal/infer/testdata/oracle/Symbol.pl
- internal/infer/testdata/oracle/Time-HiRes.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/declarations/Compress/Raw/Zlib.tpm
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/Data/Dumper.tpm
- internal/parse/declarations/Encode.tpm
- internal/parse/declarations/Getopt/Long.tpm
- internal/parse/declarations/List/Util.tpm
- internal/parse/declarations/Mojolicious/Lite.tpm
- internal/parse/declarations/Moose.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/declarations/Storable.tpm
- internal/parse/declarations/Symbol.tpm
- internal/parse/declarations/Time/HiRes.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/use.go
- internal/perltest/constantdifferences.go
- internal/perltest/constantdifferences_test.go
---

crochet:review of PR #488, round 1 (2026-10-21). Audit E and quality S2. Measured: `use Getopt::Long qw(:config bundling)` installs GetOptions (psc installs nothing); `qw(GetOptionsFromArray)` also installs GetOptions; `use Mojolicious::Lite -signatures` installs all 16 DSL subs (psc treats the list as computed and loses them); `use Moose -traits => []` loses the 13 sugar subs; POSIX's own import semantics are lost (S2, internal/parse/declarations.go:88). Use the declared-import mechanism (perigrin: the parameters are the signature to import(); the signature names the keys that alter installs) where it expresses the case; where it cannot (GetOptionsFromArray installing GetOptions), stop and bring the mechanism question to perigrin rather than special-casing a module.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Measure each use-line's installs with perl first
- [ ] Declare unmarked imports where the mechanism expresses it

## Acceptance criteria

### Positive Scenarios
- [ ] TestGetoptLongConfigInstallsGetOptions (`go test ./internal/parse/ -run '^TestGetoptLongConfigInstallsGetOptions$' -count=1 -v | grep -q '^--- PASS: TestGetoptLongConfigInstallsGetOptions'`)
- [ ] TestMojoliciousLiteFlagInstallsDSL (`go test ./internal/parse/ -run '^TestMojoliciousLiteFlagInstallsDSL$' -count=1 -v | grep -q '^--- PASS: TestMojoliciousLiteFlagInstallsDSL'`)
- [ ] TestMooseTraitsInstallsSugar (`go test ./internal/parse/ -run '^TestMooseTraitsInstallsSugar$' -count=1 -v | grep -q '^--- PASS: TestMooseTraitsInstallsSugar'`)
- [ ] TestPOSIXImportMatchesPerl (`go test ./internal/parse/ -run '^TestPOSIXImportMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestPOSIXImportMatchesPerl'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestGetoptLongFromArrayInstallsGetOptions (`go test ./internal/parse/ -run '^TestGetoptLongFromArrayInstallsGetOptions$' -count=1 -v | grep -q '^--- PASS: TestGetoptLongFromArrayInstallsGetOptions'`)
