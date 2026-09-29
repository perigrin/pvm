---
title: "`new version \"1.2\"` refuses: packages perl defines at startup are not known"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T18:54:48.492764776Z
updated: 2026-09-29T18:54:50.163512884Z
sessions:
- start_sha: 340a21572ad8fd4aa9adc5e298703387c8aa01da
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T18:54:50.163512884Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T18:54:50.163512884Z
---

The indirect object reading requires its class to be a package the parse knows -- declared by `package` or `class`, or loaded by `use` or `require`. The packages perl defines at interpreter start are known in every program but not to the parser, so `new version "1.2"` and `find PerlIO::Layer 'perlio'` refuse.

perl.c's perl_parse runs boot_core_PerlIO (`PerlIO::Layer::find`), boot_core_UNIVERSAL (universal.c's these_details, and vxs.inc's `version` methods under PERL_CORE) and boot_core_builtin. Measured with perl 5.42, nothing loaded:

    $ perl -MO=Deparse -e 'my $v = new version "1.2"; my $l = PerlIO::Layer->find("perlio");'
    my $v = 'version'->new('1.2');

perl.git t/op/bop.t:574 and t/io/binmode.t:17, 27.

## Acceptance criteria

- [ ] Interpreter-defined packages are known to the indirect reading (`go test ./internal/parse/ -run '^TestInterpreterPackagesAreKnown$' -count=1 -v | grep -q '^--- PASS: TestInterpreterPackagesAreKnown'`)
