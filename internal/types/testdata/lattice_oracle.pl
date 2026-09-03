# ABOUTME: Measures every PSC lattice edge against real Perl and emits verdicts as JSON.
# ABOUTME: The verdict is COMPUTED from both membership factors, never written down.
use v5.36;
use Scalar::Util qw(reftype blessed);

# WHAT THIS FILE IS FOR.
#
# paper_alignment_test.go asserts IsSubtype(Int, Num) and friends against a
# table transcribed by hand from the paper. That is a change detector: it
# catches an edit to the lattice nobody reflected in the tests. What it cannot
# catch is a table that was WRONG WHEN TRANSCRIBED, because the table is its
# own oracle. Chalk's equivalent stayed green for the whole time its
# `Boolean => 'Str'` edge was wrong, and this paper's own appendix verifier
# certified `Boolean | Int = Str` over all 729 pairs after the paper had
# abandoned that claim.
#
# So here PERL is the oracle. Each edge's verdict is COMPUTED from two
# measurements and printed; the Go side requires its lattice to agree.
#
#     A <: B  iff  (A subset-of B) and (every op of B holds for values of A)
#
# THE VERDICT IS NOT WRITTEN DOWN ANYWHERE. The only judgement a human makes
# is which WITNESSES to test, and that is checkable by reading them.
#
# Boolean under Str is the worked example of why both factors are needed:
# factor 2 passes (every string op works on a boolean uncoerced) while factor 1
# fails, on `false` alone. Testing one factor, or testing only `true`, reaches
# the wrong answer -- and both mistakes were actually made in this paper's
# history.

# --- FACTOR 1: MEMBERSHIP ---------------------------------------------------
#
# Two oracles, because the lattice holds two kinds of type and one question
# cannot reach both.
#
# INTERPRETIVE (the scalar spine): v is in B iff interpreting it as B directly
# agrees with interpreting it as B via some admissible detour S. This is what
# catches `false`, which stringifies directly to "" but reaches "0" through
# every numeric detour.
#
# STRUCTURAL (references, globs): a CodeRef does not "interpret" as anything --
# asking what it stringifies to measures its address, not its type. Membership
# there is a question about what a value IS, which perl answers via
# ref/reftype/blessed.
my %INTERPRET_AS = (
    Str => sub { no warnings; "" . $_[0] },
    Num => sub { no warnings; 0 + $_[0] },
    Int => sub { no warnings; int( 0 + $_[0] ) },
);

# ONLY S = B is excluded. Interpreting as B and then as B again is the identity
# on every value of B, so it would certify membership for anything -- the
# vacuous test the paper's well-formedness note forbids. S = A (the child) is
# NOT excluded: the circularity rule bans a detour that makes B's definition
# depend on B, not one passing through a proper subtype of B.
sub admissible_detours_for ($B) { grep { $_ ne $B } sort keys %INTERPRET_AS }

sub interpretive_member ( $v, $B ) {
    my $direct = $INTERPRET_AS{$B}->($v);
    for my $S ( admissible_detours_for($B) ) {
        my $detour = $INTERPRET_AS{$B}->( $INTERPRET_AS{$S}->($v) );
        return 1 if $detour eq $direct;
    }
    return 0;
}

# Each predicate is written from perl's own classification, never from the
# lattice -- otherwise the table would be grading its own homework again.
my %STRUCTURALLY_IS = (
    Ref       => sub { ref( $_[0] ) ne '' },
    ScalarRef => sub { ( reftype( $_[0] ) // '' ) =~ /^(?:SCALAR|REF)$/ },
    ArrayRef  => sub { ( reftype( $_[0] ) // '' ) eq 'ARRAY' },
    HashRef   => sub { ( reftype( $_[0] ) // '' ) eq 'HASH' },
    CodeRef   => sub { ( reftype( $_[0] ) // '' ) eq 'CODE' },
    GlobRef   => sub { ( reftype( $_[0] ) // '' ) eq 'GLOB' },
    Object    => sub { defined blessed( $_[0] ) },
    Regex     => sub { ( reftype( $_[0] ) // '' ) eq 'REGEXP' },
    Undef     => sub { !defined $_[0] },
    Scalar    => sub { 1 },    # anything holding a single value

    # THE COLLECTION BRANCH IS AN ARITY QUESTION -- a third oracle, not a
    # variant of the other two. List is not something a value interprets as,
    # nor a structural kind ref() reports: it is "how many values does this
    # yield in list context". Every value yields some number, including zero,
    # so membership in List is satisfied by any arity. That makes
    # Scalar <: List the claim that one value is a list of length one.
    List => sub { 1 },
);

sub structural_member ( $v, $B ) {
    my $p = $STRUCTURALLY_IS{$B} or return undef;    # no oracle for this type
    return $p->($v) ? 1 : 0;
}

# Structural first (the more direct question), falling back to interpretive.
# undef means NO ORACLE, reported as untested rather than silently passing.
sub member_of ( $v, $B ) {
    my $s = structural_member( $v, $B );
    return $s if defined $s;
    return interpretive_member( $v, $B ) if exists $INTERPRET_AS{$B};
    return undef;
}

# --- FACTOR 2: SUBSTITUTABILITY ---------------------------------------------
# Every operation of B must work on a value of A with no coercion. Run for
# effect: a die means the operation does not hold.
my %OPERATIONS_OF = (
    Str => [
        [ concat => sub { my $r = 'x' . $_[0] . 'y'; 1 } ],
        [ length => sub { my $n = length $_[0];      1 } ],
        [ uc     => sub { my $s = uc $_[0];          1 } ],
        [ eq     => sub { my $b = ( $_[0] eq $_[0] ); 1 } ],
    ],
    Num => [
        [ add    => sub { no warnings; my $n = $_[0] + 1;      1 } ],
        [ mul    => sub { no warnings; my $n = $_[0] * 2;      1 } ],
        [ numcmp => sub { no warnings; my $b = ( $_[0] <=> 0 ); 1 } ],
    ],
    Int => [
        [ modulo => sub { no warnings; my $n = $_[0] % 2; 1 } ],
        [ bitand => sub { no warnings; my $n = $_[0] & 1; 1 } ],
    ],
    Ref => [
        [ ref       => sub { my $k = ref $_[0];               1 } ],
        [ boolean   => sub { my $b = $_[0] ? 1 : 0;           1 } ],
        [ stringify => sub { no warnings; my $s = "$_[0]";    1 } ],
    ],
    ScalarRef => [ [ deref => sub { my $v = ${ $_[0] }; 1 } ] ],
    Object    => [
        [ blessed => sub { my $c = blessed $_[0]; 1 } ],
        [ can     => sub { my $m = $_[0]->can('x'); 1 } ],
    ],
    Scalar => [
        [ assign  => sub { my $copy = $_[0];        1 } ],
        [ boolean => sub { my $b = $_[0] ? 1 : 0;   1 } ],
        [ defined => sub { my $d = defined $_[0];   1 } ],
    ],
    List => [
        [ flatten => sub { my @l = ( $_[0] );              1 } ],
        [ count   => sub { my $n = scalar( () = ( $_[0] ) ); 1 } ],
    ],
);

# --- THE EDGES: witnesses only, the verdict is computed below ---------------
#
# A type is only as tested as its witnesses are representative, so a type whose
# members disagree (Bool: `true` passes Str membership, `false` does not) must
# list both. That is the one judgement a human still makes here.
my $scalar_x   = 5;
my $scalar_ref = \$scalar_x;
my $ref_to_ref = \$scalar_ref;

my @EDGES = (
    # --- the scalar spine, interpretive oracle ---
    { child => 'Int', parent => 'Num',
      witnesses => [ [ '42', 42 ], [ '0', 0 ], [ '-7', -7 ] ] },
    { child => 'Num', parent => 'Str',
      witnesses => [ [ '3.5', 3.5 ], [ '0.0', 0.0 ], [ '-2.25', -2.25 ] ] },
    # Transitive, asserted directly rather than inferred: if it held only by
    # transitivity through Num, a broken Int <: Num would hide here.
    { child => 'Int', parent => 'Str',
      witnesses => [ [ '42', 42 ], [ '0', 0 ] ] },

    # THE EDGE THE PAPER GOT WRONG TWICE. Both polarities, because `true`
    # passes and `false` does not.
    { child => 'Bool', parent => 'Str',
      witnesses => [ [ 'true', !!1 ], [ 'false', !!0 ] ] },

    # NaN and Inf: the placement this alignment corrected. They pass Str
    # membership (the round trip holds) and are excluded from Num by the
    # semantic component alone, which factor 2 measures.
    { child => 'NaN', parent => 'Str',
      witnesses => [ [ 'nan', 9**9**9 - 9**9**9 ] ] },
    { child => 'Inf', parent => 'Str',
      witnesses => [ [ 'inf', 9**9**9 ] ] },

    # --- direct children of Scalar ---
    { child => 'Str',   parent => 'Scalar',
      witnesses => [ [ 'hi', 'hi' ], [ 'empty', '' ] ] },
    { child => 'Bool',  parent => 'Scalar',
      witnesses => [ [ 'true', !!1 ], [ 'false', !!0 ] ] },
    { child => 'Undef', parent => 'Scalar',
      witnesses => [ [ 'undef', undef ] ] },
    { child => 'Ref',   parent => 'Scalar',
      witnesses => [ [ 'ref-to-ref', $ref_to_ref ] ] },

    # --- the reference branch, structural oracle ---
    { child => 'ScalarRef', parent => 'Ref',
      witnesses => [ [ '\\$x', $scalar_ref ] ] },
    { child => 'ArrayRef', parent => 'Ref',
      witnesses => [ [ '[1,2]', [ 1, 2 ] ], [ '[]', [] ] ] },
    { child => 'HashRef', parent => 'Ref',
      witnesses => [ [ '{a=>1}', { a => 1 } ], [ '{}', {} ] ] },
    { child => 'CodeRef', parent => 'Ref',
      witnesses => [ [ 'sub{1}', sub { 1 } ] ] },
    { child => 'GlobRef', parent => 'Ref',
      witnesses => [ [ '\\*STDOUT', \*STDOUT ] ] },
    { child => 'Object', parent => 'Ref',
      witnesses => [ [ 'bless {}', bless( {}, 'Lattice::Witness' ) ] ] },
    { child => 'Regex', parent => 'Object',
      witnesses => [ [ 'qr/x/', qr/x/ ] ] },

    # --- the collection branch: the arity claim ---
    { child => 'Scalar', parent => 'List',
      witnesses => [ [ '42', 42 ], [ 'hi', 'hi' ], [ 'undef', undef ] ] },
);

# --- COMPUTE the verdict for every edge --------------------------------------
sub measure_edge ($e) {
    my ( $A, $B ) = ( $e->{child}, $e->{parent} );
    my ( @f1_fail, @f1_untested, @f2_fail );

    for my $w ( $e->{witnesses}->@* ) {
        my ( $label, $v ) = $w->@*;

        my $m = member_of( $v, $B );
        if    ( !defined $m ) { push @f1_untested, $label }
        elsif ( !$m )         { push @f1_fail,     $label }

        for my $op ( ( $OPERATIONS_OF{$B} // [] )->@* ) {
            my ( $name, $code ) = $op->@*;
            eval { $code->($v); 1 } or push @f2_fail, "$label/$name";
        }
    }

    # Subtyping needs BOTH factors. An edge with no membership oracle is
    # reported untested rather than counted as holding.
    my $factor1 = ( !@f1_fail && !@f1_untested ) ? 1 : 0;
    my $factor2 = ( !@f2_fail ) ? 1 : 0;

    return {
        child       => $A,
        parent      => $B,
        holds       => ( $factor1 && $factor2 ) ? 1 : 0,
        factor1     => $factor1,
        factor2     => $factor2,
        f1_fail     => \@f1_fail,
        f1_untested => \@f1_untested,
        f2_fail     => \@f2_fail,
    };
}

# --- emit JSON (hand-rolled: no non-core dependency) ------------------------
sub json_str ($s) { $s =~ s/(["\\])/\\$1/g; qq{"$s"} }
sub json_arr ($a) { '[' . join( ',', map { json_str($_) } $a->@* ) . ']' }

my @out;
for my $e (@EDGES) {
    my $r = measure_edge($e);
    push @out, join( ',',
        q{"child":} . json_str( $r->{child} ),
        q{"parent":} . json_str( $r->{parent} ),
        qq{"holds":$r->{holds}},
        qq{"factor1":$r->{factor1}},
        qq{"factor2":$r->{factor2}},
        q{"f1_fail":} . json_arr( $r->{f1_fail} ),
        q{"f1_untested":} . json_arr( $r->{f1_untested} ),
        q{"f2_fail":} . json_arr( $r->{f2_fail} ),
    );
}
print '[', join( ',', map { "{$_}" } @out ), "]\n";
