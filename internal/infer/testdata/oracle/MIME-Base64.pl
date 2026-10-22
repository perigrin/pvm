# ABOUTME: Call cases for MIME::Base64's subs, which the return-type oracle calls to hold
# ABOUTME: MIME/Base64.tpm to perl.
(
    encode_base64          => [ ['abc'], [ 'abc', '' ] ],
    decode_base64          => [ ['YWJj'], ['not valid base64 !!!'] ],
    encode_base64url       => [ ['abc'] ],
    decode_base64url       => [ ['YWJj'] ],
    encoded_base64_length  => [ ['abc'], [ 'abc', '' ] ],
    decoded_base64_length  => [ ['YWJj'] ],
);
