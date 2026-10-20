# ABOUTME: Call cases for HTTP::Tiny's methods, which the return-type oracle calls to hold
# ABOUTME: HTTP/Tiny.tpm to perl.
#
# new, www_form_urlencode and can_ssl never open a connection or read
# instance state; the rest do and are exempt (returntypes_test.go's
# oracleExempt).
(
    new                => [ [] ],
    www_form_urlencode => [ [ { a => 1, b => 'x y' } ], [ [ a => 1, b => 2 ] ] ],
    can_ssl            => [ [] ],
);
