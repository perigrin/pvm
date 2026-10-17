# ABOUTME: Fixture whose one mismatch only Carp's declaration reveals, for TestPSCCheckUsesDeclarations.
# ABOUTME: Carp.tpm types carp's message as List[Str]; perl -w warns on the undef carp joins.
use strict;
use warnings;
use Carp;

carp("no name: ", undef);
