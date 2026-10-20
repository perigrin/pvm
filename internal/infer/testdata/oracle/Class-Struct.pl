# ABOUTME: Call cases for Class::Struct's struct, which the return-type oracle calls to hold
# ABOUTME: Class::Struct.tpm to perl. Each case names a fresh class: struct refuses to redefine one.
(
    struct => [ [ 'Oracle::Struct::Pt1', { x => '$', y => '@' } ], [ 'Oracle::Struct::Pt2', [ x => '$' ] ] ],
);
