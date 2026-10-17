# ABOUTME: Fixture passing undef to cluck, a sub with one declared candidate, for TestPSCCheckSingleCandidateMismatchIsCoercion.
# ABOUTME: The arity fits, so the candidate is selected and its argument coerced, warning under perl -w.
use strict;
use warnings;
use Carp qw(cluck);

cluck(undef);
