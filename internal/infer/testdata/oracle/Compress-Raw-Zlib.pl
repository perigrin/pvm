# ABOUTME: Call cases for Compress::Raw::Zlib's subs, which the return-type oracle calls to hold
# ABOUTME: Compress::Raw::Zlib.tpm to perl. Each sub maps to a list of argument lists.
#
# The driver only requires the module, never imports it, and measured, a
# bare require leaves its 45 constants ungenerated -- this bare `use`
# (every one of them is :export) is what generates them, so the driver's
# own by-name symbol lookup finds each one.
use Compress::Raw::Zlib;
(
    DEF_WBITS      => [ [] ],
    MAX_MEM_LEVEL  => [ [] ],
    MAX_WBITS      => [ [] ],
    OS_CODE        => [ [] ],
    WANT_GZIP           => [ [] ],
    WANT_GZIP_OR_ZLIB   => [ [] ],
    ZLIBNG_VERNUM       => [ [] ],
    ZLIBNG_VERSION      => [ [] ],
    ZLIBNG_VER_MAJOR    => [ [] ],
    ZLIBNG_VER_MINOR    => [ [] ],
    ZLIBNG_VER_MODIFIED => [ [] ],
    ZLIBNG_VER_REVISION => [ [] ],
    ZLIBNG_VER_STATUS   => [ [] ],
    ZLIB_VERNUM         => [ [] ],
    ZLIB_VERSION        => [ [] ],
    Z_ASCII                => [ [] ],
    Z_BEST_COMPRESSION     => [ [] ],
    Z_BEST_SPEED           => [ [] ],
    Z_BINARY               => [ [] ],
    Z_BLOCK                => [ [] ],
    Z_BUF_ERROR            => [ [] ],
    Z_DATA_ERROR           => [ [] ],
    Z_DEFAULT_COMPRESSION  => [ [] ],
    Z_DEFAULT_STRATEGY     => [ [] ],
    Z_DEFLATED             => [ [] ],
    Z_ERRNO                => [ [] ],
    Z_FILTERED             => [ [] ],
    Z_FINISH               => [ [] ],
    Z_FIXED                => [ [] ],
    Z_FULL_FLUSH           => [ [] ],
    Z_HUFFMAN_ONLY         => [ [] ],
    Z_MEM_ERROR            => [ [] ],
    Z_NEED_DICT            => [ [] ],
    Z_NO_COMPRESSION       => [ [] ],
    Z_NO_FLUSH             => [ [] ],
    Z_NULL                 => [ [] ],
    Z_OK                   => [ [] ],
    Z_PARTIAL_FLUSH        => [ [] ],
    Z_RLE                  => [ [] ],
    Z_STREAM_END           => [ [] ],
    Z_STREAM_ERROR         => [ [] ],
    Z_SYNC_FLUSH           => [ [] ],
    Z_TREES                => [ [] ],
    Z_UNKNOWN              => [ [] ],
    Z_VERSION_ERROR        => [ [] ],

    # DEF_WBITS is declared None: measured on this build, every call
    # dies ("Your vendor has not defined Zlib macro DEF_WBITS"), which
    # this case proves rather than assumes.

    adler32 => [ ['hello'], [ 'hello', 1 ] ],
    crc32   => [ ['hello'], [''] ],
);
