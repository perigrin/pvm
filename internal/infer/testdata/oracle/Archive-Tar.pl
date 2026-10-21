# ABOUTME: Call cases for Archive::Tar's subs, which the return-type oracle calls to hold Archive::Tar.tpm to perl.
# ABOUTME: Every path is relative, so every file Archive::Tar reads or writes stays inside the oracle's scratch directory.
#
# a.txt and b.txt are real files to archive; src.tar is a real tar
# built from them, written directly (not through the shared instance
# below) so it exists before INSTANCE's new(file) and the class
# methods (create_archive/iter/list_archive/extract_archive) each need
# a real archive to read. INSTANCE's shared instance starts already
# populated by that same read, through new(file) -- add_data/add_files
# add to it, chmod/chown/rename/replace_content/setcwd modify it,
# contains_file/extract/extract_file/get_content/get_files/list_files
# read it, read() reloads it fresh from src.tar partway through
# (alphabetical order), and clear is exempt (oracleExempt) since
# wiping it there would break every call after it.
open my $af, '>', 'a.txt' or die $!;
print $af "hello\n";
close $af;
open my $bf, '>', 'b.txt' or die $!;
print $bf "world\n";
close $bf;
my $setup = Archive::Tar->new;
$setup->add_files( 'a.txt', 'b.txt' );
$setup->write('src.tar');
(
    COMPRESS_GZIP => [ [] ],
    COMPRESS_BZIP => [ [] ],
    COMPRESS_XZ   => [ [] ],

    new                  => [ [] ],
    read                 => [ [ 'src.tar' ] ],
    contains_file        => [ [ 'a.txt' ] ],
    extract              => [ [] ],
    extract_file         => [ [ 'a.txt', 'extracted.txt' ] ],
    list_files           => [ [] ],
    get_files            => [ [] ],
    get_content          => [ [ 'a.txt' ] ],
    replace_content      => [ [ 'a.txt', 'new content' ] ],
    rename               => [ [ 'a.txt', 'a2.txt' ] ],
    chmod                => [ [ 'b.txt', '0644' ] ],
    chown                => [ [ 'b.txt', 'nobody' ] ],
    remove               => [ [ 'b.txt' ] ],
    # write's 0-arg candidate (returns the archive as a Str) and its
    # 1-3 arg candidate (writes to a file, returning an Int flag) are
    # each exercised.
    write                => [ [ 'out.tar' ], [] ],
    # add_files with zero files short-circuits to undef (`my @files =
    # @_ or return;`, measured), so this case names one.
    add_files            => [ [ 'a.txt' ] ],
    add_data             => [ [ 'c.txt', 'cee' ] ],
    error                => [ [] ],
    setcwd               => [ [ '.' ] ],
    create_archive       => [ [ 'out2.tar', 0, 'a.txt' ] ],
    iter                 => [ [ 'src.tar' ] ],
    list_archive         => [ [ 'src.tar' ] ],
    extract_archive      => [ [ 'src.tar' ] ],
    has_io_string               => [ [] ],
    has_perlio                  => [ [] ],
    has_zlib_support            => [ [] ],
    has_bzip2_support           => [ [] ],
    has_xz_support               => [ [] ],
    can_handle_compressed_files  => [ [] ],
    INSTANCE => { ctor => 'new', args => [ 'src.tar', 0 ], subs => [qw(
        add_data add_files chmod chown contains_file error extract extract_file
        get_content get_files list_files read remove rename replace_content
        setcwd write
    )] },
);
