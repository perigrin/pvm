#!/usr/bin/env perl
# Test file for pure type inference examples

sub process_user_data($input) {
    return unless $input && ref($input) eq 'HASH';

    my $user_id = $input->{user_id};
    return unless defined($user_id) && $user_id =~ /^\d+$/;

    my $numeric_user_id = int($user_id);
    my $db_result = get_user_from_db($numeric_user_id);
    return unless $db_result;

    my $processed = {
        id => $db_result->{id},
        name => $db_result->{name},
        email => $db_result->{email},
        created => $db_result->{created_at} || time(),
    };

    return $processed;
}

sub analyze_data_types($data) {
    my $type = ref($data);
    my $is_defined = defined($data);
    my $keys = keys(%$data) if ref($data) eq 'HASH';
    my $length = length($data) if !ref($data);

    return {
        type => $type,
        defined => $is_defined,
        keys => $keys,
        length => $length,
    };
}

sub math_operations($x, $y) {
    my $sum = $x + $y;
    my $product = $x * $y;
    my $quotient = $x / $y;
    my $int_val = int($x);
    my $abs_val = abs($y);

    return ($sum, $product, $quotient, $int_val, $abs_val);
}
