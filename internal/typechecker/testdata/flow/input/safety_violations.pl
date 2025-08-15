#!/usr/bin/env perl
# Test file for intentionally unsafe code patterns

sub unsafe_hash_access($api_response) {
    my $status = $api_response->{status};
    if ($status eq 'success') {
        debug($api_response->{body});  # ERROR: body field not checked
    }
}

sub unsafe_array_access($data) {
    my $first = $data->[0];        # ERROR: not proven to be array
    return $first;
}

sub uninitialized_usage($maybe_value) {
    my $result = $maybe_value * 2;     # ERROR: might be undef
    my $message = "Value: $maybe_value"; # ERROR: string interpolation
    return $result;
}

sub unsafe_method_call($maybe_object) {
    my $result = $maybe_object->method(); # ERROR: might be undef
    return $result;
}

sub double_hash_access($config) {
    my $db_host = $config->{database}->{host}; # ERROR: database might not exist
    return $db_host;
}

sub unsafe_numeric_context($hash) {
    my $price = $hash->{price};  # No exists check
    my $total = $price * 1.1;   # ERROR: $price might be undef
    return $total;
}
