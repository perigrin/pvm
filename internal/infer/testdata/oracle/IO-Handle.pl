# ABOUTME: Call cases for IO::Handle's subs, which the return-type oracle calls to hold IO::Handle.tpm to perl.
# ABOUTME: Every case is a method call with the class name itself as its invocant, as the oracle calls it; format_write, print, printf and say die that way and are exempt.
#
# flush dereferences the invocant as a glob too, but rather than exempt
# it INSTANCE below names a real instance, built once from
# new_from_fd(1, 'w'), for return_oracle.pl to call it on instead of
# the bare class name (measured side-effect free: it only dups
# whatever fd 1 is already pointed at when the oracle runs).
(
    autoflush                    => [ [1] ],
    output_field_separator       => [ [] ],
    output_record_separator      => [ [] ],
    input_record_separator       => [ [] ],
    input_line_number            => [ [] ],
    format_page_number           => [ [] ],
    format_lines_per_page        => [ [] ],
    format_lines_left            => [ [] ],
    format_name                  => [ [] ],
    format_top_name               => [ [] ],
    format_line_break_characters => [ [] ],
    format_formfeed               => [ [] ],
    getline                      => [ [] ],
    getlines                      => [ [] ],
    printflush                   => [ [ 'x' ] ],
    flush                        => [ [] ],
    SEEK_SET                     => [ [] ],
    SEEK_CUR                     => [ [] ],
    SEEK_END                     => [ [] ],
    _IOFBF                       => [ [] ],
    _IOLBF                       => [ [] ],
    _IONBF                       => [ [] ],
    new                          => [ [] ],
    new_from_fd                  => [ [ -1, 'r' ] ],
    opened                       => [ [] ],
    fileno                       => [ [] ],
    eof                          => [ [] ],
    getc                         => [ [] ],
    close                        => [ [] ],
    INSTANCE                     => { ctor => 'new_from_fd', args => [ 1, 'w' ], subs => [qw(flush)] },
);
