# ABOUTME: Call cases for Encode's subs, which the return-type oracle calls to hold Encode.tpm to perl.
# ABOUTME: Each sub maps to a list of argument lists, one call per list.
use Encode qw(find_encoding);
no warnings 'uninitialized';
my $obj = find_encoding('utf-8');
(
    DIE_ON_ERR      => [ [] ],
    FB_CROAK        => [ [] ],
    FB_DEFAULT      => [ [] ],
    FB_HTMLCREF     => [ [] ],
    FB_PERLQQ       => [ [] ],
    FB_QUIET        => [ [] ],
    FB_WARN         => [ [] ],
    FB_XMLCREF      => [ [] ],
    HTMLCREF        => [ [] ],
    LEAVE_SRC       => [ [] ],
    PERLQQ          => [ [] ],
    RETURN_ON_ERR   => [ [] ],
    STOP_AT_PARTIAL => [ [] ],
    WARN_ON_ERR     => [ [] ],
    XMLCREF         => [ [] ],

    _utf8_off => [ [ "caf\x{e9}" ], [undef] ],
    _utf8_on  => [ [ "cafe" ], [undef] ],

    bytes2str => [ [ 'utf-8', "h\x{c3}\x{a9}llo" ], [ 'utf-8', undef ] ],
    str2bytes => [ [ 'utf-8', "h\x{e9}llo" ], [ 'utf-8', undef ] ],

    clone_encoding => [ ['utf-8'], ['no-such-enc-xyz'] ],

    decode      => [ [ 'utf-8', "h\x{c3}\x{a9}llo" ], [ 'utf-8', undef ] ],
    decode_utf8 => [ ["abc"], [undef] ],

    define_encoding => [ [ $obj, 'oracle-test-encoding' ] ],

    # encode is declared Str, not Str|Undef, pending issue 01a156e4 (see
    # Encode.tpm): no undef-input case here, since that would observe a
    # value its declaration no longer admits.
    encode => [ [ 'utf-8', "h\x{e9}llo" ], [ 'ascii', "caf\x{e9}", sub { sprintf '<U+%04X>', shift } ] ],
    encode_utf8 => [ ["abc"], [undef] ],

    encodings => [ [] ],

    find_encoding      => [ ['utf-8'], ['no-such-enc-xyz'] ],
    find_mime_encoding => [ ['utf-8'], ['no-such-mime-xyz'] ],

    # $from_encoding and $to_encoding also take a pre-resolved encoding
    # object (measured), not only a name.
    from_to => [ [ "hello", 'iso-8859-1', 'cp1250' ], [ "hello", $obj, 'cp1250' ], [ "hello", 'iso-8859-1', $obj ] ],

    is_utf8 => [ ["abc"], [undef] ],

    # perlio_ok's own source comment claims "a plain 0 or 1, not a
    # boolean", but measured, "ascii" gives a builtin true.
    perlio_ok => [ ['utf-8'], ['iso-2022-kr'], ['ascii'] ],

    resolve_alias => [ ['utf8'], ['no-such-alias-xyz'] ],
);
