---
title: "Audit fixes, pure-Perl modules: carp, cluck, warn and DynaLoader::croak pass a lone reference, longmess and shortmess, Getopt::Std, Config, File::stat"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.681982426Z
updated: 2026-10-20T23:16:55.37430505Z
sessions:
- start_sha: e597bd91da76f184c4a4181487ff77fd5443cd4e
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T23:16:55.37430505Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T23:16:55.37430505Z
---

crochet:review of PR #488, round 1 (2026-10-21). Signature audit (01a15aaa), group E.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

Measured 2026-10-21: carp({c=>42}) and warn({c=>42}) deliver the hash reference itself to $SIG{__WARN__}, as croak/die do to $@; with more than one argument all stringify. So carp, cluck, CORE warn (CORE.tpm:874) and DynaLoader::croak take croak's parameter type; longmess/shortmess pass a lone ref through (multis). The import-semantics findings (Getopt::Long, Mojolicious::Lite, Moose) are their own issue.

## Findings (measured; from the audit, perl 5.42.0)

#### .tpm audit, group E

Judge: /home/perigrin/.local/bin/perl (5.42.0). Repo: wt-cmt (read-only). Probes: scratchpad/audit-E/.
Parameter findings were re-judged under the brief's CORRECTION (a parameter type is what the operation coerces to or preserves).
psc itself could not be run: there is no Go toolchain on this host, so `go build ./cmd/psc` failed. The two import findings (Getopt::Long, Mojolicious::Lite) come from reading internal/parse/use.go and exports.go (importsFrom, literalNameList, the declared-import branch at use.go:377). I did not run them through psc.

Format: `Module::sub — file:line — .tpm says — perl/POD shows (command) — fix — confidence`.

#### Carp (3)

1. Carp::longmess: return type. Carp.tpm:19 says `Str`. longmess_heavy (Carp.pm:567) is `return wantarray ? @_ : $_[0]` when `ref $_[0]`, so a lone ref comes back as the ref itself. Probe audit-E/carp.pl: `longmess({x=>1},'b')` gives `HASH(0x..)` (a Ref) in scalar context and 2 elements in list context. Fix: a multi. The scalar candidate is `Str|Ref`. The list candidate is `List[Str|Ref|...]`. The parameter also preserves a lone ref, as croak's does, so make it `List[Str|Boolean|DualVar|Ref]`. Confidence 90.
2. Carp::shortmess: return type. Carp.tpm:20 says `Str`. shortmess_heavy (Carp.pm:682) is `return @_ if ref $_[0]`. Measured: `shortmess({x=>1},'b')` returns 2 (the count) in scalar context and the 2 arguments, ref included, in list context. The count is an Int, which is a Str, so scalar context holds. List context does not: it returns `@_` unchanged. Fix: a multi. The list candidate is `List[Str|Boolean|DualVar|Ref]`, the scalar one `Str`. Widen the parameter to croak's type (rule a: the ref is passed through). Confidence 85.
3. Note on carp/cluck, held to the CORRECTION, so not a finding. carp is `warn shortmess @_`, so a lone ref reaches `$SIG{__WARN__}` as the ref itself, not stringified. Measured: `carp({code=>42})` gives a handler that sees `REF HASH`; `cluck([1])` gives `REF ARRAY`. That is "preserves", not "stringifies", the same pass-through as croak to die. CORE.tpm:874 declares CORE::warn `List[Str]` too, so the two agree. perigrin should rule on warn and carp together. Confidence that the behaviour is as described: 95.

#### Exporter (1)

4. Exporter: two documented methods are omitted. Exporter.pm:68 defines `export_fail` and Exporter.pm:89 defines `require_version`. Both are documented in POD ("Module Version Checking", "Managing Unknown Symbols") and inherited by every `@ISA = ('Exporter')` module. Measured with audit-E/exp.pl:
   - `Foo->export_fail('x','y')` returns its arguments: a list in list context, the count 2 in scalar context.
   - `Foo->require_version(1)` returns `$VERSION` ('1.5').
   - `Foo->require_version(2)` dies "version 2 required".

   Fix: declare `multi sub export_fail (Str $class: List[Str] @syms) List[Str]` plus an Int scalar candidate, and `sub require_version (Str $class: Str $wanted) Str`. Confidence 70.

#### XSLoader (0)

#### DynaLoader (1)

5. DynaLoader::croak: parameter. DynaLoader.tpm:13 says `List[Str] @message`. DynaLoader.pm:108 is `Carp::croak(@_)`, which passes a lone ref through as the exception. Measured: `eval { DynaLoader::croak({x=>1}) }; ref $@` gives HASH. Fix: `List[Str|Boolean|DualVar|Ref] @message`, matching Carp::croak. Confidence 80.

#### Test::More (2)

6. Test::More::new_ok: return type. More.tpm:60 says `Object`. The source (More.pm:732) returns `$obj`, which is whatever `$class->new` returned, and is undef when new died. Measured with audit-E/tm.pl: when `Bar::new` returns undef, `new_ok('Bar')` returns undef. Fix: `Object|Undef`. `Scalar` would be strictly accurate. Confidence 90.
7. Header version. More.tpm:2 says "Test::More 1.302209". The installed version is 1.302210 (`perl -MTest::More -e 'print $Test::More::VERSION'`). Fix: update the ABOUTME line, or re-measure. Confidence 60.

#### Config (1)

8. Config::config_re: return type. Config.tpm:38 says `Str`, and the comment says "one string whichever context". Config_heavy.pl:1500 is `return map { chomp; $_ } grep ...`. Measured: `my @l = config_re("ccflags.*")` gives 3 elements, and `scalar config_re("ccflags.*")` gives 3 (the count). Fix: a multi of `List[Str]` and `Int`, and correct the comment. Confidence 95.

#### Getopt::Long (2)

9. Getopt::Long's own `import` is not declared, so the common `:config` idiom installs nothing. Getopt/Long.pm's `import` does two things:
   - It moves everything after `:config` to Configure.
   - It calls `SUPER::import(@syms)`, adding `&GetOptions` to `@syms` whenever `@syms` is non-empty.

   Measured on perl:
   - `use Getopt::Long qw(:config bundling)` installs GetOptions. `@syms` is empty, so the full @EXPORT is installed.
   - `use Getopt::Long qw(GetOptionsFromArray)` also installs GetOptions, but not `$REQUIRE_ORDER`.

   psc's importsFrom (exports.go:533) maps `:config` to an unknown tag and drops `bundling`, so it installs nothing, and `GetOptions(...)` goes unresolved. Fix: declare an `import`. The existing declared-import rule (any list installs the :export set) is right for the `:config` form, but wrong for `qw(GetOptionsFromArray)`. The second case needs a mechanism extension (the named names plus GetOptions) or a ruling. Confidence 90 on the perl side; the psc side is from reading the code.
10. Getopt::Long::VersionMessage and HelpMessage: return type. Long.tpm:35-36 says `None`. The POD for `-exitval` says: 'or else the string "NOEXIT" to indicate that control should simply be returned without terminating the invoking process'. Measured: `VersionMessage(-exitval => "NOEXIT", -message => "m")` prints and returns 1. With `None`, psc marks the code after a documented call unreachable. Fix: a non-None return, such as `Int` (print's 1). Exempting the subs from the oracle still holds. Confidence 75.

#### Getopt::Std (1)

11. Getopt::Std::getopt: return type. Std.tpm:12 says `Boolean`, and the comment says "both leave a Boolean". The last statement is `unless (ref $hash) { ... }`, so with a hashref the sub's value is `ref $hash`. Measured: `getopt("a", \%h)` returns the Str "HASH" (is_bool 0), with or without options in @ARGV. Without a hashref it returns a Boolean. Fix: `Boolean|Str`, and correct the comment. Confidence 95.

#### Text::ParseWords (0)

shellwords never yields undef: its trailing undef is popped. quotewords and parse_line do, and are declared `Str|Undef`. Scalar context gives a count, or undef on an unmatched quote. All measured.

#### Text::Wrap (1)

12. `$Text::Wrap::break`: variable type. Wrap.tpm:11 says `our Str $break`. The POD says to set it 'to a string such as '[\s:]' ... or a pre-compiled regexp such as qr/[\s']/'. Measured: `$break = qr/[\s:]/; wrap(...)` breaks at the colons. psc would flag the documented assignment. Fix: `our Str|Regex $break`, and correct the comment ("never stores it as a Regexp ref" is about the default value, not the documented API). Confidence 80.

#### ExtUtils::MakeMaker (1)

13. ExtUtils::MakeMaker::neatvalue: return type. MakeMaker.tpm:41 says `Str`. The source has `return $v unless $t eq 'HASH'`, after the ARRAY branch. Measured: `neatvalue(sub{1})` returns a CODE ref, `neatvalue(bless {}, "Foo")` a Foo object, `neatvalue(\"x")` a SCALAR ref. Fix: `Str|ScalarRef|CodeRef|GlobRef|Object`, or `Str|Ref`. Confidence 95.

#### File::stat (2)

14. File::stat::stat and lstat: parameter. stat.tpm:27-28 says `Str $file`. Both call CORE::stat/lstat on the argument, and stat also falls back to a qualified glob. The handle is preserved, not stringified; CORE.tpm:767 declares CORE::stat's parameter `Str|Glob`. Measured: `stat(\*STDIN)`, `stat(*STDIN)` and `stat($lexical_fh)` each return a File::stat object. Fix: `Str|FileHandle $file`. For lstat, `Str|Glob` matches CORE; perl warns "lstat() on filehandle" but returns an object. Confidence 85 (stat), 70 (lstat).
15. File::stat::stat_cando: first parameter. stat.tpm:35 says `Object $stat_object`. The POD says stat_cando "takes an arrayref containing the return values of stat or lstat as its first argument". Measured: `stat_cando([CORE::stat "/etc/passwd"], S_IRUSR, 1)` returns 1. A plain ArrayRef is not an Object, and cando indexes the argument as an array, which is the type it uses. Fix: `ArrayRef|Object $stat_object`. Confidence 90. (`Boolean $effective` stands under the CORRECTION: the operation tests truthiness.)

#### File::Path (0)

All four were measured, in list and scalar context, with old and new APIs and an options hashref.

#### File::Basename (1)

16. File::Basename::fileparse: parameter. Basename.tpm:15-16 says `List[Str] @args`. The POD (Basename.pm:88-92) says suffixes are 'patterns (... or qr//)' and gives `fileparse("/foo/bar/baz.txt", qr/\.[^.]*/)`. Measured: that call returns `(baz, /foo/bar/, .txt)`. The suffixes are used as patterns, so a Regex is preserved as a pattern, not lossily stringified. Fix: `List[Str|Regex] @args`. Confidence 70 (rule (a) as I read it; perigrin may class qr→pattern as lossless Str coercion).

#### File::Copy (0)

All four return a plain Int 1 or 0, never a Boolean, and copy accepts filehandles. All measured.

#### File::Find (1)

17. File::Find::find and finddepth: return type. Find.tpm:9-10 says `Boolean`, and the comment says "a true Boolean". `_find_opt` begins `return unless @_`. Measured:
   - `find(sub{})` with no directories returns undef in scalar context.
   - A normal walk returns a false Boolean (""), not a true one.

   Fix: `Boolean|Undef`, and correct the comment. Confidence 80.

#### Time::Local (0)

#### Class::Struct (0)

Note at confidence 60, not counted: the documented `use Class::Struct CLASS => [...]` form makes `import` call `struct` at compile time and install nothing (measured). psc treats that list as computed and opens the symbols, which is safe but loses the generated class.

#### Locale::Maketext::Simple (1)

18. Locale::Maketext::Simple::loc_lang: return type. Simple.tpm:32 says `Int`. The source returns `$loc_lang || sub { 1 }`. When Locale::Maketext::Lexicon is installed, loc_lang is `sub { $lh = $pkg->get_handle(@_) }` and returns a handle object. Measured: Int, with Lexicon absent here. The Object branch is from source only, because Lexicon (not core) is not installed. Fix: `Int|Object`, or leave it and record that the declaration assumes core-only. Confidence 60. (`loc`'s `List[Str]` stands under the CORRECTION: loc interpolates its arguments.)

#### Moose (1)

All 13 names and prototypes match: only `blessed` has one, `($)`. Measured by enumerating the caller stash after `use Moose`.

19. Moose import options. Measured: `use Moose -traits => []` still installs all 13. psc sees a computed list and opens the symbols, losing the 13 declarations. Fix: declare an unmarked `sub import (Str $class: List @args) Void;`, so any list installs the :export set. Confidence 65.

#### Moose::Role (0)

All 15 names and prototypes match (measured).

#### Mojolicious::Lite (1)

20. Mojolicious::Lite: the SYNOPSIS form drops the whole DSL. Measured, 9.49: the names and the `group (&)` prototype match all 16 the .tpm lists. But the module's own SYNOPSIS (Lite.pm:67, 86) uses `use Mojolicious::Lite -signatures;`, and its import ignores its arguments for DSL purposes: it monkey-patches all 16 and then hands `@_` to Mojo::Base. Measured: `use Mojolicious::Lite -signatures` installs `get`. psc's literalNameList does not read `-signatures` as a literal, so it opens the symbols and imports none of the 16, including `group`'s `&` prototype. Fix: declare an unmarked `sub import (Str $class: List @flags) Void;`. use.go:377 then installs the :export set for any list, which matches perl exactly. Confidence 85 (psc side read from code, not run).

#### Future::AsyncAwait (1)

`async` (sublike PREFIX), `await` (TERMEXPR_SCALARCTX) and `CANCEL` (ANONSUB, STMT) match AsyncAwait.xs 0.71 (lines 2369-2519, from ~/.cpanm/work/.../Future-AsyncAwait-0.71). CANCEL without `:experimental(cancel)` only warns; it parses (measured).

21. async's sublike flags are partly omitted. The XS sets `XS_PARSE_SUBLIKE_FLAG_PREFIX|BODY_OPTIONAL|ALLOW_PKGNAME` (AsyncAwait.xs:2370). The .tpm states only PREFIX, and syntax.go:51 accepts only `PREFIX`, so `async sub f;` forward declarations and `async sub Pkg::name {...}` are not stated. Fix: extend `:sublike(...)` to take these flags, or record why it is not needed. Confidence 60 (I could not run psc to see whether its plain-sub parse already covers both).

#### Unmeasured

Every module, the four non-core ones included, is installed under 5.42.0 and was measured. Nothing was audited from CPAN docs alone. The exceptions:
- psc behaviour for findings 9, 19 and 20 (no Go toolchain).
- The Lexicon branch of finding 18.

#### Counts

Carp 2 (plus 1 note), Exporter 1, XSLoader 0, DynaLoader 1, Test::More 2, Config 1, Getopt::Long 2, Getopt::Std 1, Text::ParseWords 0, Text::Wrap 1, ExtUtils::MakeMaker 1, File::stat 2, File::Path 0, File::Basename 1, File::Copy 0, File::Find 1, Time::Local 0, Class::Struct 0 (1 note), Locale::Maketext::Simple 1, Moose 1, Moose::Role 0, Mojolicious::Lite 1, Future::AsyncAwait 1. Total: 20 findings.

AUDIT: FINDINGS

## Steps
- [ ] Re-measure each finding before fixing; withdraw any that does not reproduce, saying why in the commit
- [ ] Add oracle cases for each changed signature
- [ ] Hold tests, lib baseline and precision unchanged or better

## Acceptance criteria

### Positive Scenarios
- [ ] TestCarpFamilyPassesALoneReference (`go test ./internal/infer/ -run '^TestCarpFamilyPassesALoneReference$' -count=1 -v | grep -q '^--- PASS: TestCarpFamilyPassesALoneReference'`)
- [ ] TestGetoptStdReturnsHashName (`go test ./internal/infer/ -run '^TestGetoptStdReturnsHashName$' -count=1 -v | grep -q '^--- PASS: TestGetoptStdReturnsHashName'`)
- [ ] TestConfigReReturnsByContext (`go test ./internal/infer/ -run '^TestConfigReReturnsByContext$' -count=1 -v | grep -q '^--- PASS: TestConfigReReturnsByContext'`)
- [ ] TestFileStatTakesHandles (`go test ./internal/infer/ -run '^TestFileStatTakesHandles$' -count=1 -v | grep -q '^--- PASS: TestFileStatTakesHandles'`)
- [ ] TestCarpReturnTypes (`go test ./internal/infer/ -run '^TestCarpReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestCarpReturnTypes'`)
- [ ] TestPurePerlExportersReturnTypes (`go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestShortmessListContextPassesArgs (`go test ./internal/infer/ -run '^TestShortmessListContextPassesArgs$' -count=1 -v | grep -q '^--- PASS: TestShortmessListContextPassesArgs'`)
