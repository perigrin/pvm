# ABOUTME: Classifies bison's precedence resolutions of perly.y by what a chart recognizer needs to apply them.
# ABOUTME: Reads a bison report made with default reductions off; prints the counts, then each state-dependent pair.
use v5.36;

# Usage: perl count.pl perly.output
# The report must come from bison -Dlr.default-reduction=accepting
# --report=itemset,lookaheads,solved, so every state lists the lookaheads of
# its reductions; count.sh makes it from the pinned perly.y.
#
# Each resolution is a (rule, token) pair. "shift" blocks completing the rule
# before the token (the reduce side); "reduce" blocks the rule's last operand
# having the token's operator on top (the shift side); "error" blocks both. A
# pair is global when applying it in every state, regardless of the parser
# state, blocks nothing bison allows elsewhere.

my $report = shift // die "usage: count.pl perly.output\n";
my (%st, $s);
open my $fh, '<', $report or die "$report: $!\n";
while (<$fh>) {
    if (/^State (\d+)$/) { $s = $1; $st{$s} = { final => {}, shift => {}, conf => {} }; next }
    next unless defined $s;
    if (/^\s+(\d+)\s+(?:\S+:|\|).*\.\s+\[(.*)\]\s*$/) { $st{$s}{final}{$1} = { map { $_ => 1 } split /,\s*/, $2 }; next }
    if (/^\s+(\d+)\s+(?:\S+:|\|).*\.\s*$/)         { $st{$s}{final}{$1} //= {} if $1 != 0; next }   # rule 0 accepts
    if (/^\s+(\S+)\s+shift, and go to state/)      { $st{$s}{shift}{$1} = 1; next }
    if (/Conflict between rule (\d+) and token (\S+) resolved as (?:an )?(\w+)/) { $st{$s}{conf}{"$1 $2"} = $3 }
}
die "$report: no lookaheads; run bison with -Dlr.default-reduction=accepting\n"
    if grep { my $f = $_->{final}; grep { !%$_ } values %$f } values %st;

my (%act, %where);
for my $s (keys %st) {
    for my $k (keys $st{$s}{conf}->%*) { $act{$k}{ $st{$s}{conf}{$k} }++; $where{$k}{$s} = 1 }
}
my $inconsistent = grep { keys $act{$_}->%* > 1 } keys %act;
my (%n, @dep);
for my $k (sort keys %act) {
    my ($r, $t) = split ' ', $k;
    my ($act) = keys $act{$k}->%*;
    my @bad;
    for my $s (sort { $a <=> $b } keys %st) {
        next if $where{$k}{$s};
        my $f = $st{$s}{final}{$r} or next;
        push @bad, "reduce\@$s" if ($act eq 'shift' || $act eq 'error') && $f->{$t};
        push @bad, "shift\@$s"  if ($act eq 'reduce' || $act eq 'error') && $st{$s}{shift}{$t} && !$f->{$t};
    }
    $n{$act}{ @bad ? 'state-dependent' : 'global' }++;
    push @dep, join "\t", $r, $t, $act, join ',', @bad if @bad;
}
say "pairs\t", scalar keys %act;
say "states\t", scalar grep { keys $st{$_}{conf}->%* } keys %st;
say "inconsistent\t$inconsistent";
for my $act (qw(shift reduce error)) {
    say "$act\t", join "\t", map { "$_ " . ($n{$act}{$_} // 0) } qw(global state-dependent);
}
say "state-dependent\t$_" for @dep;
