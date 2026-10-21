# ABOUTME: Call cases for IO::Select's methods, which the return-type oracle calls to hold
# ABOUTME: IO/Select.tpm to perl.
#
# new is called with the bare class name, as every class method is;
# every other method needs a real instance (measured: IO::Select->add
# on the bare class name dies "Can't use string ... as an ARRAY ref").
# INSTANCE names the constructor return_oracle.pl calls once --
# IO::Select->new(\*STDIN), measured side-effect free: it only stores
# the handle, and select() with a zero timeout never blocks -- and
# every other case below runs against that one shared instance, in the
# order perl calls them (alphabetical by name): add registers STDIN,
# so it is present for can_read, can_write, count, exists, handles and
# has_exception, and remove takes it back out last.
(
    new           => [ [] ],
    add           => [ [ \*STDIN ] ],
    remove        => [ [ \*STDIN ] ],
    exists        => [ [ \*STDIN ] ],
    handles       => [ [] ],
    can_read      => [ [0] ],
    can_write     => [ [0] ],
    has_exception => [ [0] ],
    has_error     => [ [0] ],
    count         => [ [] ],
    bits          => [ [] ],
    as_string     => [ [] ],
    INSTANCE      => {
        ctor => 'new',
        args => [ \*STDIN ],
        subs => [qw(add remove exists handles can_read can_write has_exception has_error count bits as_string)],
    },
);
