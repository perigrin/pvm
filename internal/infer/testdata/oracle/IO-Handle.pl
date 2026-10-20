# ABOUTME: Call cases for IO::Handle's subs, which the return-type oracle calls to hold IO::Handle.tpm to perl.
# ABOUTME: Every case is a method call with the class name itself as its invocant, as the oracle calls it; format_write, print, printf and say die that way and are exempt.
#
# flush and the destructive methods below it (binmode, clearerr,
# error, fcntl, fdopen, ioctl, read, stat, sync, sysread, syswrite,
# truncate, ungetc, untaint, write) each dereference the
# invocant as a glob too, so rather than exempt them INSTANCE below
# names a real instance, opened once on a scratch file relative to the
# oracle's own scratch directory (returnTypeFailures confines perl's
# cwd there), for return_oracle.pl to call each of them on instead of
# the bare class name.
use Fcntl ();
# $scratch is a package global, not a lexical: a `my` here would go out
# of scope -- and so be garbage-collected, closing its fd -- the moment
# this do-file finishes, before return_oracle.pl ever dups $fd into an
# instance.
our $scratch;
open( $scratch, '+>', 'io-handle-instance-scratch' ) or die "cannot open scratch: $!";
my $fd = fileno($scratch);
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
    binmode                      => [ [] ],
    clearerr                     => [ [] ],
    error                        => [ [] ],
    fcntl                        => [ [ Fcntl::F_GETFL(), 0 ] ],
    fdopen                       => [ [ $fd, 'r+' ] ],
    ioctl                        => [ [ 0, 0 ] ],
    read                         => [ [ "", 2 ] ],
    sysread                      => [ [ "", 2 ] ],
    stat                         => [ [] ],
    sync                         => [ [] ],
    syswrite                     => [ [ "ab" ] ],
    truncate                     => [ [ 3 ] ],
    ungetc                       => [ [ 90 ] ],
    untaint                      => [ [] ],
    write                        => [ [ "abc", 2 ] ],
    INSTANCE                     => {
        ctor => 'new_from_fd',
        args => [ $fd, 'r+' ],
        subs => [
            qw(flush binmode clearerr error fcntl fdopen ioctl
              read sysread stat sync syswrite truncate ungetc untaint write)
        ],
    },
);
