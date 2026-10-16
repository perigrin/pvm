# A list assignment writes each of its variables: encoding.pm.
use strict;
use warnings;

sub locale_encoding {
    my $locale_encoding;
    $locale_encoding = $ENV{NO_SUCH_LOCALE_VAR};
    return $locale_encoding if defined $locale_encoding;
    my $locale = "en_US.UTF-8";
    if ($locale =~ /^([^.]+)\.([^.@]+)(?:@.*)?$/) {
        my $country_language;
        ($country_language, $locale_encoding) = ($1, $2);
        if (lc($locale_encoding) eq 'euc') {
            $locale_encoding = 'euc-jp';
        }
    }
    return $locale_encoding;
}

print locale_encoding() // "none", "\n";
