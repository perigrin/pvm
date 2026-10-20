# ABOUTME: Call cases for Storable's subs, which the return-type oracle calls to hold Storable.tpm
# ABOUTME: to perl. Every fixture file and filehandle this stages is written under the oracle's
# ABOUTME: scratch directory (its cwd), never a path elsewhere, which TestXSUtilitiesOracleStaysInScratch holds.
use Storable;
use IO::File;

Storable::store( { a => 1 }, 'fixture.store' );
my $frozen = Storable::freeze( [ 1, 2, 3 ] );

open my $fd_retrieve_fh,   '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $retrieve_fd_fh,   '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $fd_retrieve_fh2,  '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $retrieve_fd_fh2,  '<', 'fixture.store' or die "cannot open fixture.store: $!";
open my $store_fd_fh,      '>', 'store_fd.out'  or die "cannot open store_fd.out: $!";
open my $nstore_fd_fh,     '>', 'nstore_fd.out' or die "cannot open nstore_fd.out: $!";

# store_fd/nstore_fd/fd_retrieve/retrieve_fd also take a blessed
# filehandle object, dereferenced the same way a bareword or lexical
# one is (measured on 5.42.0).
my $store_fd_obj = IO::File->new('store_fd_obj.out', 'w') or die "cannot open store_fd_obj.out: $!";
my $fd_retrieve_fh_obj = IO::File->new('fixture.store', 'r') or die "cannot open fixture.store: $!";

(
    store         => [ [ { a => 1 }, 'store.out' ] ],
    nstore        => [ [ { a => 1 }, 'nstore.out' ] ],
    lock_store    => [ [ { a => 1 }, 'lock_store.out' ] ],
    lock_nstore   => [ [ { a => 1 }, 'lock_nstore.out' ] ],
    store_fd      => [ [ { a => 1 }, $store_fd_fh ], [ { a => 1 }, $store_fd_obj ] ],
    nstore_fd     => [ [ { a => 1 }, $nstore_fd_fh ] ],
    # $flags, documented on thaw/retrieve and the rest of this group
    # ("set the 2nd argument for thaw/retrieve to 0"), is exercised
    # here with 0 and 1 (lock_retrieve takes no lock on a non-flock
    # platform, so either value still succeeds).
    retrieve      => [ ['fixture.store'], ['fixture.store', 1] ],
    lock_retrieve => [ ['fixture.store'], ['fixture.store', 0] ],
    retrieve_fd   => [ [$retrieve_fd_fh], [$retrieve_fd_fh2, 0] ],
    fd_retrieve   => [ [$fd_retrieve_fh], [$fd_retrieve_fh2, 0], [$fd_retrieve_fh_obj] ],
    freeze        => [ [ [ 1, 2, 3 ] ] ],
    nfreeze       => [ [ [ 1, 2, 3 ] ] ],
    thaw          => [ [$frozen], [$frozen, 0] ],
    dclone        => [ [ { a => [ 1, 2 ] } ] ],
    file_magic    => [ ['fixture.store'] ],
    read_magic    => [ [$frozen], [$frozen, 0] ],
    BLESS_OK         => [ [] ],
    TIE_OK           => [ [] ],
    FLAGS_COMPAT     => [ [] ],
    stack_depth      => [ [] ],
    stack_depth_hash => [ [] ],
);
