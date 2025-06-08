my Int|Undef $maybe_count;
if (defined $maybe_count) {
    my $safe_count = $maybe_count;  # Should narrow to Int
}
