# ABOUTME: Call cases for Storable's subs, which the return-type oracle calls to hold Storable.tpm
# ABOUTME: to perl. Every fixture file and filehandle this stages is written under the oracle's
# ABOUTME: scratch directory (its cwd), never a path elsewhere, which TestXSUtilitiesOracleStaysInScratch holds.
use Storable;

Storable::store( { a => 1 }, 'fixture.store' );
my $frozen = Storable::freeze( [ 1, 2, 3 ] );

open my $fd_retrieve_fh,   '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $retrieve_fd_fh,   '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $store_fd_fh,      '>', 'store_fd.out'  or die "cannot open store_fd.out: $!";
open my $nstore_fd_fh,     '>', 'nstore_fd.out' or die "cannot open nstore_fd.out: $!";

(
    store         => [ [ { a => 1 }, 'store.out' ] ],
    nstore        => [ [ { a => 1 }, 'nstore.out' ] ],
    lock_store    => [ [ { a => 1 }, 'lock_store.out' ] ],
    lock_nstore   => [ [ { a => 1 }, 'lock_nstore.out' ] ],
    store_fd      => [ [ { a => 1 }, $store_fd_fh ] ],
    nstore_fd     => [ [ { a => 1 }, $nstore_fd_fh ] ],
    retrieve      => [ ['fixture.store'] ],
    lock_retrieve => [ ['fixture.store'] ],
    retrieve_fd   => [ [$retrieve_fd_fh] ],
    fd_retrieve   => [ [$fd_retrieve_fh] ],
    freeze        => [ [ [ 1, 2, 3 ] ] ],
    nfreeze       => [ [ [ 1, 2, 3 ] ] ],
    thaw          => [ [$frozen] ],
    dclone        => [ [ { a => [ 1, 2 ] } ] ],
    file_magic    => [ ['fixture.store'] ],
    read_magic    => [ [$frozen] ],
    BLESS_OK         => [ [] ],
    TIE_OK           => [ [] ],
    FLAGS_COMPAT     => [ [] ],
    stack_depth      => [ [] ],
    stack_depth_hash => [ [] ],
);
