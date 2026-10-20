# ABOUTME: Call cases for ExtUtils::MakeMaker's subs, which the return-type oracle calls to
# ABOUTME: hold ExtUtils::MakeMaker.tpm to perl. WriteMakefile and prompt are exempt; see there.
# Every path named is relative, so each stays inside the oracle's scratch
# cwd (TestPurePerlExportersOracleStaysInScratch's rule).
(
    WriteEmptyMakefile => [ [ 'NAME', 'Oracle::Fixture', 'VERSION', '0.01' ] ],
    open_for_writing   => [ ['oracle-open-for-writing.tmp'] ],
    write_file_via_tmp => [ [ 'oracle-write-file-via-tmp.tmp', [ 'line one', 'line two' ] ] ],
    neatvalue          => [ [undef], ['a string'], [ [ 1, 2 ] ], [ { a => 1 } ] ],
    _sprintf562        => [ [ '%1$s-%2$s', 'a', 'b' ] ],
    os_unsupported      => [ [] ],
    mkbootstrap         => [ [] ],
    mksymlists          => [ [] ],
);
